//go:build linux

// Package supervisor provides service orchestration for the process supervisor.
// This file contains Linux-specific port detection for managed processes.
package supervisor

import (
	"bufio"
	"context"
	"encoding/hex"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"
)

const (
	// dockerCommandTimeout is the maximum time to wait for Docker CLI commands.
	dockerCommandTimeout time.Duration = 5 * time.Second

	// maxAncestorDepth limits process tree traversal to prevent infinite loops.
	maxAncestorDepth int = 10

	// minDockerArgs is the minimum number of arguments for a valid docker command.
	minDockerArgs int = 2

	// minProcStatFields is the minimum fields expected in /proc/pid/stat after comm.
	minProcStatFields int = 2

	// minPortParts is the minimum parts when splitting port mapping output.
	minPortParts int = 2

	// defaultPortMapCapacity is the initial capacity for port maps.
	defaultPortMapCapacity int = 8

	// defaultInodeMapCapacity is the initial capacity for inode maps.
	defaultInodeMapCapacity int = 16

	// minNetFields is minimum fields in /proc/net/* line.
	minNetFields int = 10

	// portBytesLen is the expected length of hex-decoded port bytes.
	portBytesLen int = 2

	// bitsPerByte is used for port byte shifting.
	bitsPerByte int = 8

	// tcpListenState is the hex state for TCP LISTEN.
	tcpListenState string = "0A"

	// udpBoundState is the hex state for UDP bound.
	udpBoundState string = "07"
)

// getListeningPorts returns TCP/UDP ports the process is listening on.
// Reads from /proc/net/tcp, /proc/net/tcp6, /proc/net/udp, /proc/net/udp6.
// Also detects Docker container port mappings.
//
// Params:
//   - pid: process ID to check for listening ports
//
// Returns:
//   - []int: sorted slice of listening port numbers, nil if none found
func getListeningPorts(pid int) []int {
	// Validate PID is positive.
	if pid <= 0 {
		return nil
	}

	ports := make(map[int]struct{}, defaultPortMapCapacity)

	// Try Docker container port detection first.
	dockerPorts := getDockerPorts(pid)
	for _, port := range dockerPorts {
		ports[port] = struct{}{}
	}

	// Docker ports found - return them directly.
	if len(ports) > 0 {
		return mapToSortedSlice(ports)
	}

	// Get socket inodes for this PID.
	inodes := getSocketInodes(pid)
	if len(inodes) == 0 {
		return nil
	}

	// Scan /proc/net/* files for listening ports.
	netFiles := []string{"/proc/net/tcp", "/proc/net/tcp6", "/proc/net/udp", "/proc/net/udp6"}
	for _, netFile := range netFiles {
		findListeningPorts(netFile, inodes, ports)
	}

	return mapToSortedSlice(ports)
}

// mapToSortedSlice converts a port map to a sorted slice.
//
// Params:
//   - ports: map of ports to convert
//
// Returns:
//   - []int: sorted slice of port numbers
func mapToSortedSlice(ports map[int]struct{}) []int {
	result := slices.Collect(maps.Keys(ports))
	slices.Sort(result)

	return result
}

// getDockerPorts detects if the process is a Docker container and returns its port mappings.
//
// Params:
//   - pid: process ID to check
//
// Returns:
//   - []int: slice of mapped host ports, nil if not a Docker container
func getDockerPorts(pid int) []int {
	// Read process command line from procfs.
	cmdline, err := os.ReadFile(filepath.Join("/proc", strconv.Itoa(pid), "cmdline"))
	if err != nil {
		return nil
	}

	// Parse null-separated arguments.
	args := strings.Split(string(cmdline), "\x00")
	if len(args) < minDockerArgs {
		return nil
	}

	// Verify docker or podman command.
	cmd := filepath.Base(args[0])
	if cmd != "docker" && cmd != "podman" {
		return nil
	}

	// Must have "run" subcommand.
	if !hasRunSubcommand(args) {
		return nil
	}

	// Extract container name from arguments or PID lookup.
	containerName := findContainerName(args)
	if containerName == "" {
		containerName = findDockerContainerByPID(pid)
	}

	if containerName == "" {
		return nil
	}

	return getDockerContainerPorts(containerName)
}

// hasRunSubcommand checks if args contain a "run" subcommand.
//
// Params:
//   - args: command line arguments to check
//
// Returns:
//   - bool: true if "run" subcommand found
func hasRunSubcommand(args []string) bool {
	for _, arg := range args[1:] {
		if arg == "run" {
			return true
		}
	}

	return false
}

// findContainerName extracts container name from --name flag.
//
// Params:
//   - args: command line arguments to search
//
// Returns:
//   - string: container name if found, empty string otherwise
func findContainerName(args []string) string {
	for idx, arg := range args {
		// Check --name flag with separate value.
		if arg == "--name" && idx+1 < len(args) {
			return args[idx+1]
		}

		// Check --name=value format.
		if after, found := strings.CutPrefix(arg, "--name="); found {
			return after
		}
	}

	return ""
}

// findDockerContainerByPID tries to find a Docker container associated with a PID.
//
// Params:
//   - pid: process ID whose container we're looking for
//
// Returns:
//   - string: container ID if found, empty string otherwise
func findDockerContainerByPID(pid int) string {
	ctx, cancel := context.WithTimeout(context.Background(), dockerCommandTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "docker", "ps", "-q").Output()
	if err != nil {
		return ""
	}

	// Iterate over running containers.
	containers := strings.Fields(string(out))
	for _, containerID := range containers {
		if matchesContainerPID(ctx, containerID, pid) {
			return containerID
		}
	}

	return ""
}

// matchesContainerPID checks if pid is an ancestor of the container's init process.
//
// Params:
//   - ctx: context for timeout
//   - containerID: Docker container ID
//   - pid: process ID to check as ancestor
//
// Returns:
//   - bool: true if pid is ancestor of container's init process
func matchesContainerPID(ctx context.Context, containerID string, pid int) bool {
	inspectOut, err := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Pid}}", containerID).Output()
	if err != nil {
		return false
	}

	containerPID, err := strconv.Atoi(strings.TrimSpace(string(inspectOut)))
	if err != nil {
		return false
	}

	return isAncestor(pid, containerPID)
}

// isAncestor checks if ancestorPID is an ancestor of childPID.
//
// Params:
//   - ancestorPID: potential ancestor process ID
//   - childPID: child process ID to check
//
// Returns:
//   - bool: true if ancestorPID is an ancestor of childPID
func isAncestor(ancestorPID, childPID int) bool {
	currentPID := childPID

	// Walk up the process tree from child.
	for range maxAncestorDepth {
		// Reached init or invalid PID.
		if currentPID <= 1 {
			return false
		}

		ppid, ok := getParentPID(currentPID)
		if !ok {
			return false
		}

		// Found the ancestor.
		if ppid == ancestorPID {
			return true
		}

		currentPID = ppid
	}

	return false
}

// getParentPID reads the parent PID from /proc/pid/stat.
//
// Params:
//   - pid: process ID to get parent of
//
// Returns:
//   - int: parent process ID
//   - bool: true if successfully read
func getParentPID(pid int) (int, bool) {
	statFile := filepath.Join("/proc", strconv.Itoa(pid), "stat")
	data, err := os.ReadFile(statFile)
	if err != nil {
		return 0, false
	}

	// Format: pid (comm) state ppid ...
	// Find closing paren to skip comm field.
	statStr := string(data)
	idx := strings.LastIndex(statStr, ")")
	if idx < 0 {
		return 0, false
	}

	// Parse fields after comm.
	fields := strings.Fields(statStr[idx+1:])
	if len(fields) < minProcStatFields {
		return 0, false
	}

	ppid, err := strconv.Atoi(fields[1])
	if err != nil {
		return 0, false
	}

	return ppid, true
}

// getDockerContainerPorts returns the host ports mapped for a container.
//
// Params:
//   - containerNameOrID: container name or ID to query
//
// Returns:
//   - []int: slice of host ports mapped to the container
func getDockerContainerPorts(containerNameOrID string) []int {
	ctx, cancel := context.WithTimeout(context.Background(), dockerCommandTimeout)
	defer cancel()

	out, err := exec.CommandContext(ctx, "docker", "port", containerNameOrID).Output()
	if err != nil {
		return nil
	}

	return parseDockerPortOutput(string(out))
}

// parseDockerPortOutput parses "docker port" command output.
// Output format: "80/tcp -> 0.0.0.0:8082"
//
// Params:
//   - output: raw output from docker port command
//
// Returns:
//   - []int: slice of parsed host port numbers
func parseDockerPortOutput(output string) []int {
	var ports []int
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip empty lines.
		if line == "" {
			continue
		}

		// Extract host port after last colon.
		parts := strings.Split(line, ":")
		if len(parts) < minPortParts {
			continue
		}

		portStr := strings.TrimSpace(parts[len(parts)-1])
		port, err := strconv.Atoi(portStr)
		if err != nil {
			continue
		}

		// Only add valid ports.
		if port > 0 {
			ports = append(ports, port)
		}
	}

	return ports
}

// getSocketInodes returns socket inodes for a PID from /proc/{pid}/fd.
//
// Params:
//   - pid: process ID to get socket inodes for
//
// Returns:
//   - map[uint64]struct{}: set of socket inodes owned by the process
func getSocketInodes(pid int) map[uint64]struct{} {
	inodes := make(map[uint64]struct{}, defaultInodeMapCapacity)

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

		// Parse socket:[inode] format.
		if after, found := strings.CutPrefix(link, "socket:["); found {
			if inodeStr, ok := strings.CutSuffix(after, "]"); ok {
				if inode, err := strconv.ParseUint(inodeStr, 10, 64); err == nil {
					inodes[inode] = struct{}{}
				}
			}
		}
	}

	return inodes
}

// findListeningPorts reads a /proc/net/* file and finds listening ports.
// State 0A = TCP_LISTEN, state 07 = UDP (unconditionally listening).
//
// Params:
//   - netFile: path to /proc/net/* file
//   - inodes: socket inodes to match
//   - ports: map to store found ports
func findListeningPorts(netFile string, inodes map[uint64]struct{}, ports map[int]struct{}) {
	file, err := os.Open(netFile)
	if err != nil {
		return
	}
	defer func() { _ = file.Close() }()

	isUDP := strings.Contains(netFile, "udp")
	scanner := bufio.NewScanner(file)

	// Skip header line.
	if !scanner.Scan() {
		return
	}

	// Parse each connection line.
	for scanner.Scan() {
		port, ok := parseNetLine(scanner.Text(), inodes, isUDP)
		if ok && port > 0 {
			ports[port] = struct{}{}
		}
	}
}

// parseNetLine parses a line from /proc/net/* and extracts port if matching.
//
// Params:
//   - line: line from /proc/net/* file
//   - inodes: socket inodes to match
//   - isUDP: true if parsing UDP file
//
// Returns:
//   - int: port number if found and matching
//   - bool: true if valid port found
func parseNetLine(line string, inodes map[uint64]struct{}, isUDP bool) (int, bool) {
	fields := strings.Fields(line)
	if len(fields) < minNetFields {
		return 0, false
	}

	// Check connection state (field 3).
	state := fields[3]
	if !isUDP && state != tcpListenState {
		return 0, false
	}
	if isUDP && state != udpBoundState {
		return 0, false
	}

	// Parse and match inode (field 9).
	inode, err := strconv.ParseUint(fields[9], 10, 64)
	if err != nil {
		return 0, false
	}

	// Verify inode belongs to our process.
	if _, ok := inodes[inode]; !ok {
		return 0, false
	}

	// Extract port from local address (field 1).
	return parseHexPort(fields[1])
}

// parseHexPort extracts port number from hex address format.
//
// Params:
//   - localAddr: address in format IP:PORT (hex)
//
// Returns:
//   - int: port number
//   - bool: true if successfully parsed
func parseHexPort(localAddr string) (int, bool) {
	parts := strings.Split(localAddr, ":")
	if len(parts) != minPortParts {
		return 0, false
	}

	portBytes, err := hex.DecodeString(parts[1])
	if err != nil || len(portBytes) != portBytesLen {
		return 0, false
	}

	// Convert big-endian bytes to port number.
	port := int(portBytes[0])<<bitsPerByte | int(portBytes[1])

	return port, true
}
