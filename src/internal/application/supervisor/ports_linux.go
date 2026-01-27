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

	// decimalBase is the base for decimal number parsing.
	decimalBase int = 10

	// netFieldState is the index of the connection state field in /proc/net/*.
	netFieldState int = 3

	// netFieldInode is the index of the inode field in /proc/net/*.
	netFieldInode int = 9

	// bitSize64 is the bit size for 64-bit integers.
	bitSize64 int = 64
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
		// Invalid PID.
		return nil
	}

	ports := make(map[int]struct{}, defaultPortMapCapacity)

	// Try Docker container port detection first.
	dockerPorts := getDockerPorts(pid)
	// Add all Docker ports to the map.
	for _, port := range dockerPorts {
		ports[port] = struct{}{}
	}

	// Docker ports found - return them directly.
	if len(ports) > 0 {
		// Return Docker ports.
		return mapToSortedSlice(ports)
	}

	// Get socket inodes for this PID.
	inodes := getSocketInodes(pid)
	// No sockets found for this process.
	if len(inodes) == 0 {
		// No socket inodes.
		return nil
	}

	// Scan /proc/net/* files for listening ports.
	netFiles := []string{"/proc/net/tcp", "/proc/net/tcp6", "/proc/net/udp", "/proc/net/udp6"}
	// Check each protocol file for listening ports.
	for _, netFile := range netFiles {
		findListeningPorts(netFile, inodes, ports)
	}

	// Convert map to sorted slice.
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
	// Convert map keys to slice.
	result := slices.Collect(maps.Keys(ports))
	slices.Sort(result)

	// Return sorted slice.
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
	// Failed to read cmdline.
	if err != nil {
		// Unable to read cmdline.
		return nil
	}

	// Parse null-separated arguments.
	args := strings.Split(string(cmdline), "\x00")
	// Not enough arguments for docker command.
	if len(args) < minDockerArgs {
		// Insufficient arguments.
		return nil
	}

	// Verify docker or podman command.
	cmd := filepath.Base(args[0])
	// Not a docker/podman command.
	if cmd != "docker" && cmd != "podman" {
		// Not a container command.
		return nil
	}

	// Must have "run" subcommand.
	// Not a docker/podman run command.
	if !hasRunSubcommand(args) {
		// No run subcommand.
		return nil
	}

	// Extract container name from arguments or PID lookup.
	containerName := findContainerName(args)
	// No container name in args, try PID lookup.
	if containerName == "" {
		containerName = findDockerContainerByPID(pid)
	}

	// Container name not found.
	if containerName == "" {
		// Unable to find container.
		return nil
	}

	// Query Docker for container ports.
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
	// Check if "run" is in the arguments.
	return slices.Contains(args[1:], "run")
}

// findContainerName extracts container name from --name flag.
//
// Params:
//   - args: command line arguments to search
//
// Returns:
//   - string: container name if found, empty string otherwise
func findContainerName(args []string) string {
	// Search for --name flag in arguments.
	for idx, arg := range args {
		// Check --name flag with separate value.
		if arg == "--name" && idx+1 < len(args) {
			// Found --name flag.
			return args[idx+1]
		}

		// Check --name=value format.
		if after, found := strings.CutPrefix(arg, "--name="); found {
			// Found --name=value.
			return after
		}
	}

	// No --name flag found.
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

	// Get list of running container IDs.
	out, err := exec.CommandContext(ctx, "docker", "ps", "-q").Output()
	// Failed to list containers.
	if err != nil {
		// Unable to list containers.
		return ""
	}

	// Iterate over running containers.
	// Check each container for PID match.
	for containerID := range strings.FieldsSeq(string(out)) {
		// Found matching container.
		if matchesContainerPID(ctx, containerID, pid) {
			// Container matches PID.
			return containerID
		}
	}

	// No matching container found.
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
	// Get container's init PID.
	inspectOut, err := exec.CommandContext(ctx, "docker", "inspect", "-f", "{{.State.Pid}}", containerID).Output()
	// Failed to inspect container.
	if err != nil {
		// Inspect failed.
		return false
	}

	// Parse PID from output.
	containerPID, err := strconv.Atoi(strings.TrimSpace(string(inspectOut)))
	// Failed to parse PID.
	if err != nil {
		// Parse failed.
		return false
	}

	// Check if pid is an ancestor.
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
			// At init or below.
			return false
		}

		// Get parent PID from procfs.
		ppid, ok := getParentPID(currentPID)
		// Failed to get parent PID.
		if !ok {
			// Unable to get parent.
			return false
		}

		// Found the ancestor.
		if ppid == ancestorPID {
			// Ancestor found.
			return true
		}

		// Move up to parent.
		currentPID = ppid
	}

	// Reached max depth without finding ancestor.
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
	// Read /proc/pid/stat file.
	statFile := filepath.Join("/proc", strconv.Itoa(pid), "stat")
	data, err := os.ReadFile(statFile)
	// Failed to read stat file.
	if err != nil {
		// Read failed.
		return 0, false
	}

	// Format: pid (comm) state ppid ...
	// Find closing paren to skip comm field.
	statStr := string(data)
	idx := strings.LastIndex(statStr, ")")
	// Invalid stat format.
	if idx < 0 {
		// Malformed stat.
		return 0, false
	}

	// Parse fields after comm.
	fields := strings.Fields(statStr[idx+1:])
	// Not enough fields.
	if len(fields) < minProcStatFields {
		// Insufficient fields.
		return 0, false
	}

	// Parse parent PID (field 1 after comm).
	ppid, err := strconv.Atoi(fields[1])
	// Failed to parse PPID.
	if err != nil {
		// Parse failed.
		return 0, false
	}

	// Return parent PID.
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

	// Query Docker for port mappings.
	out, err := exec.CommandContext(ctx, "docker", "port", containerNameOrID).Output()
	// Failed to get port mappings.
	if err != nil {
		// Port query failed.
		return nil
	}

	// Parse port output.
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

	// Parse each line for port mapping.
	for line := range strings.SplitSeq(output, "\n") {
		line = strings.TrimSpace(line)
		// Skip empty lines.
		if line == "" {
			continue
		}

		// Extract host port after last colon.
		parts := strings.Split(line, ":")
		// Not enough parts for valid mapping.
		if len(parts) < minPortParts {
			continue
		}

		// Parse port number.
		portStr := strings.TrimSpace(parts[len(parts)-1])
		port, err := strconv.Atoi(portStr)
		// Failed to parse port.
		if err != nil {
			continue
		}

		// Only add valid ports.
		if port > 0 {
			ports = append(ports, port)
		}
	}

	// Return collected ports.
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
	// Read file descriptor directory.
	entries, err := os.ReadDir(fdDir)
	// Failed to read fd directory.
	if err != nil {
		// Unable to read fd dir.
		return inodes
	}

	// Check each file descriptor.
	for _, entry := range entries {
		// Read symlink target.
		link, err := os.Readlink(filepath.Join(fdDir, entry.Name()))
		// Failed to read symlink.
		if err != nil {
			continue
		}

		// Parse socket:[inode] format.
		// Check if link is a socket.
		if after, found := strings.CutPrefix(link, "socket:["); found {
			// Extract inode number.
			if inodeStr, ok := strings.CutSuffix(after, "]"); ok {
				// Parse inode as uint64.
				if inode, err := strconv.ParseUint(inodeStr, decimalBase, bitSize64); err == nil {
					inodes[inode] = struct{}{}
				}
			}
		}
	}

	// Return collected inodes.
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
	// Open network file.
	file, err := os.Open(netFile)
	// Failed to open file.
	if err != nil {
		// Unable to open netfile.
		return
	}
	defer func() { _ = file.Close() }()

	isUDP := strings.Contains(netFile, "udp")
	scanner := bufio.NewScanner(file)

	// Skip header line.
	if !scanner.Scan() {
		// No header found.
		return
	}

	// Parse each connection line.
	for scanner.Scan() {
		// Parse line for port and inode match.
		port, ok := parseNetLine(scanner.Text(), inodes, isUDP)
		// Add matching port.
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
	// Not enough fields.
	if len(fields) < minNetFields {
		// Insufficient fields.
		return 0, false
	}

	// Check connection state.
	state := fields[netFieldState]
	// Not TCP LISTEN state.
	if !isUDP && state != tcpListenState {
		// Wrong TCP state.
		return 0, false
	}
	// Not UDP bound state.
	if isUDP && state != udpBoundState {
		// Wrong UDP state.
		return 0, false
	}

	// Parse and match inode.
	inode, err := strconv.ParseUint(fields[netFieldInode], decimalBase, bitSize64)
	// Failed to parse inode.
	if err != nil {
		// Return invalid port on parse error.
		return 0, false
	}

	// Verify inode belongs to our process.
	// Inode not owned by this process.
	if _, ok := inodes[inode]; !ok {
		// Return invalid port for unmatched inode.
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
	// Invalid address format.
	if len(parts) != minPortParts {
		// Return invalid port on malformed address.
		return 0, false
	}

	// Decode hex port bytes.
	portBytes, err := hex.DecodeString(parts[1])
	// Failed to decode or wrong length.
	if err != nil || len(portBytes) != portBytesLen {
		// Return invalid port on decode error.
		return 0, false
	}

	// Convert big-endian bytes to port number.
	port := int(portBytes[0])<<bitsPerByte | int(portBytes[1])

	// Return parsed port.
	return port, true
}
