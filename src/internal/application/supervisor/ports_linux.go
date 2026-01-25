//go:build linux

package supervisor

import (
	"bufio"
	"context"
	"encoding/hex"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

// getListeningPorts returns TCP/UDP ports the process is listening on.
// Reads from /proc/net/tcp, /proc/net/tcp6, /proc/net/udp, /proc/net/udp6.
// Also detects Docker container port mappings.
func getListeningPorts(pid int) []int {
	if pid <= 0 {
		return nil
	}

	ports := make(map[int]struct{})

	// First, try to detect Docker container ports.
	// This works even when running inside a devcontainer.
	dockerPorts := getDockerPorts(pid)
	for _, p := range dockerPorts {
		ports[p] = struct{}{}
	}

	// If we found Docker ports, return them.
	// Docker containers run in different namespace, /proc/net won't work.
	if len(ports) > 0 {
		return mapToSortedSlice(ports)
	}

	// Get inodes for this PID's file descriptors.
	inodes := getSocketInodes(pid)
	if len(inodes) == 0 {
		return nil
	}

	// Find listening ports matching these inodes.
	// Check TCP and UDP (both IPv4 and IPv6).
	for _, netFile := range []string{"/proc/net/tcp", "/proc/net/tcp6", "/proc/net/udp", "/proc/net/udp6"} {
		findListeningPorts(netFile, inodes, ports)
	}

	return mapToSortedSlice(ports)
}

// mapToSortedSlice converts a port map to a sorted slice.
func mapToSortedSlice(ports map[int]struct{}) []int {
	result := make([]int, 0, len(ports))
	for port := range ports {
		result = append(result, port)
	}

	// Sort ports.
	for i := 0; i < len(result)-1; i++ {
		for j := i + 1; j < len(result); j++ {
			if result[i] > result[j] {
				result[i], result[j] = result[j], result[i]
			}
		}
	}

	return result
}

// getDockerPorts detects if the process is a Docker container and returns its port mappings.
// Works by checking if the process is "docker run" and finding the container's published ports.
func getDockerPorts(pid int) []int {
	// Read the process command line.
	cmdline, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return nil
	}

	// cmdline is null-separated.
	args := strings.Split(string(cmdline), "\x00")
	if len(args) < 2 {
		return nil
	}

	// Check if this is a docker/podman run command.
	cmd := filepath.Base(args[0])
	if cmd != "docker" && cmd != "podman" {
		return nil
	}

	// Find "run" subcommand.
	isRun := false
	for _, arg := range args[1:] {
		if arg == "run" {
			isRun = true
			break
		}
	}
	if !isRun {
		return nil
	}

	// Find container name from --name flag or use child container detection.
	containerName := ""
	for i, arg := range args {
		if arg == "--name" && i+1 < len(args) {
			containerName = args[i+1]
			break
		}
		if strings.HasPrefix(arg, "--name=") {
			containerName = strings.TrimPrefix(arg, "--name=")
			break
		}
	}

	// If no name, try to find container by parent PID.
	if containerName == "" {
		containerName = findDockerContainerByPID(pid)
	}

	if containerName == "" {
		return nil
	}

	// Get port mappings from docker inspect.
	return getDockerContainerPorts(containerName)
}

// findDockerContainerByPID tries to find a Docker container associated with a PID.
func findDockerContainerByPID(pid int) string {
	// Use docker ps to find containers and match by checking if our PID is ancestor.
	ctx := context.Background()
	out, err := exec.CommandContext(ctx, "docker", "ps", "-q").Output()
	if err != nil {
		return ""
	}

	containers := strings.Fields(string(out))
	for _, containerID := range containers {
		// Get the container's PID.
		inspectOut, err := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Pid}}", containerID).Output()
		if err != nil {
			continue
		}

		containerPID, err := strconv.Atoi(strings.TrimSpace(string(inspectOut)))
		if err != nil {
			continue
		}

		// Check if our PID is the parent of the container's init process.
		// The docker run command's PID is the parent of containerd-shim which is parent of container PID.
		if isAncestor(pid, containerPID) {
			return containerID
		}
	}

	return ""
}

// isAncestor checks if ancestorPID is an ancestor of childPID.
func isAncestor(ancestorPID, childPID int) bool {
	// Walk up the process tree from child.
	currentPID := childPID
	for i := 0; i < 10; i++ { // Max 10 levels up.
		if currentPID <= 1 {
			return false
		}

		// Read parent PID.
		statFile := filepath.Join("/proc", strconv.Itoa(currentPID), "stat")
		data, err := os.ReadFile(statFile)
		if err != nil {
			return false
		}

		// Format: pid (comm) state ppid ...
		// Find closing paren to skip comm (which may contain spaces).
		s := string(data)
		idx := strings.LastIndex(s, ")")
		if idx < 0 {
			return false
		}

		fields := strings.Fields(s[idx+1:])
		if len(fields) < 2 {
			return false
		}

		ppid, err := strconv.Atoi(fields[1])
		if err != nil {
			return false
		}

		if ppid == ancestorPID {
			return true
		}

		currentPID = ppid
	}

	return false
}

// getDockerContainerPorts returns the host ports mapped for a container.
func getDockerContainerPorts(containerNameOrID string) []int {
	// Use docker port command to get mappings.
	out, err := exec.CommandContext(context.Background(), "docker", "port", containerNameOrID).Output()
	if err != nil {
		return nil
	}

	// Output format: "80/tcp -> 0.0.0.0:8082"
	var ports []int
	lines := strings.Split(string(out), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Find the host port after the last ":".
		parts := strings.Split(line, ":")
		if len(parts) < 2 {
			continue
		}

		portStr := strings.TrimSpace(parts[len(parts)-1])
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		if port > 0 {
			ports = append(ports, port)
		}
	}

	return ports
}

// getSocketInodes returns socket inodes for a PID from /proc/{pid}/fd.
func getSocketInodes(pid int) map[uint64]struct{} {
	inodes := make(map[uint64]struct{})

	fdDir := filepath.Join("/proc", strconv.Itoa(pid), "fd")
	entries, err := os.ReadDir(fdDir)
	if err != nil {
		return inodes
	}

	for _, entry := range entries {
		link, err := os.Readlink(filepath.Join(fdDir, entry.Name()))
		if err != nil {
			continue
		}

		// Socket links look like: socket:[12345]
		if strings.HasPrefix(link, "socket:[") && strings.HasSuffix(link, "]") {
			inodeStr := link[8 : len(link)-1]
			if inode, err := strconv.ParseUint(inodeStr, 10, 64); err == nil {
				inodes[inode] = struct{}{}
			}
		}
	}

	return inodes
}

// findListeningPorts reads a /proc/net/* file and finds listening ports.
// State 0A = TCP_LISTEN, state 07 = UDP (unconditionally listening).
func findListeningPorts(netFile string, inodes map[uint64]struct{}, ports map[int]struct{}) {
	f, err := os.Open(netFile)
	if err != nil {
		return
	}
	defer func() { _ = f.Close() }()

	isUDP := strings.Contains(netFile, "udp")
	scanner := bufio.NewScanner(f)

	// Skip header line.
	if !scanner.Scan() {
		return
	}

	for scanner.Scan() {
		fields := strings.Fields(scanner.Text())
		if len(fields) < 10 {
			continue
		}

		// Field 3 is state (hex).
		state := fields[3]

		// For TCP, only LISTEN state (0A). For UDP, state 07 means bound.
		if !isUDP && state != "0A" {
			continue
		}
		if isUDP && state != "07" {
			continue
		}

		// Field 9 is inode.
		inode, err := strconv.ParseUint(fields[9], 10, 64)
		if err != nil {
			continue
		}

		// Check if this inode belongs to our process.
		if _, ok := inodes[inode]; !ok {
			continue
		}

		// Field 1 is local_address (IP:PORT in hex).
		localAddr := fields[1]
		parts := strings.Split(localAddr, ":")
		if len(parts) != 2 {
			continue
		}

		// Parse port (hex).
		portHex := parts[1]
		portBytes, err := hex.DecodeString(portHex)
		if err != nil || len(portBytes) != 2 {
			continue
		}

		port := int(portBytes[0])<<8 | int(portBytes[1])
		if port > 0 {
			ports[port] = struct{}{}
		}
	}
}
