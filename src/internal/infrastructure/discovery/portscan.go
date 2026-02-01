//go:build linux

// Package discovery provides infrastructure adapters for target discovery.
//
//nolint:ktn-struct-onefile // listeningPort is internal helper type for PortScanDiscoverer
package discovery

import (
	"bufio"
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// portscan parsing constants.
const (
	// ipv4ByteLength is the expected byte length for an IPv4 address.
	ipv4ByteLength int = 4

	// ipv6ByteLength is the expected byte length for an IPv6 address.
	ipv6ByteLength int = 16

	// minTCPFields is the minimum number of fields expected in a /proc/net/tcp line.
	minTCPFields int = 4

	// addressPartCount is the expected number of parts in a hex address (IP:port).
	addressPartCount int = 2

	// byteReversalDivisor divides byte length to find midpoint for reversal.
	byteReversalDivisor int = 2

	// portScanLabelCount is the number of labels added to port scan targets.
	portScanLabelCount int = 3

	// tcpStateFieldIndex is the index of the state field in /proc/net/tcp lines.
	tcpStateFieldIndex int = 3

	// hexBase is the base for hexadecimal number parsing.
	hexBase int = 16

	// portBitSize is the bit size for port numbers (16-bit).
	portBitSize int = 16
)

// Sentinel errors for port scan parsing.
var (
	// errInvalidAddressFormat is returned when address format is invalid.
	errInvalidAddressFormat error = errors.New("invalid address format")

	// errInvalidIPv4Length is returned when IPv4 address has wrong length.
	errInvalidIPv4Length error = errors.New("invalid IPv4 length")

	// errInvalidIPv6Length is returned when IPv6 address has wrong length.
	errInvalidIPv6Length error = errors.New("invalid IPv6 length")
)

// listeningPort represents a port in LISTEN state from /proc/net/tcp.
type listeningPort struct {
	Protocol  string `dto:"out,priv,pub" json:"protocol"`  // "tcp" or "udp"
	LocalAddr string `dto:"out,priv,pub" json:"localAddr"` // IP address
	LocalPort int    `dto:"out,priv,pub" json:"localPort"` // Port number
	State     string `dto:"out,priv,pub" json:"state"`     // Socket state (hex)
}

// PortScanDiscoverer discovers listening ports by parsing /proc/net/tcp.
// It scans local network interfaces for services listening on TCP ports
// and creates monitoring targets for them.
//
//nolint:ktn-struct-onefile // listeningPort is internal helper
type PortScanDiscoverer struct {
	// interfaces maps interface names to true for fast lookup.
	interfaces map[string]bool

	// excludePorts maps port numbers to exclude.
	excludePorts map[int]bool

	// includePorts maps port numbers to include exclusively.
	includePorts map[int]bool
}

// Addrser provides network address listing capability.
type Addrser interface {
	Addrs() ([]net.Addr, error)
}

// NewPortScanDiscoverer creates a new port scan discoverer.
//
// Params:
//   - cfg: the port scan discovery configuration.
//
// Returns:
//   - *PortScanDiscoverer: a new port scan discoverer.
func NewPortScanDiscoverer(cfg *config.PortScanDiscoveryConfig) *PortScanDiscoverer {
	discoverer := &PortScanDiscoverer{
		interfaces:   make(map[string]bool, len(cfg.Interfaces)),
		excludePorts: make(map[int]bool, len(cfg.ExcludePorts)),
		includePorts: make(map[int]bool, len(cfg.IncludePorts)),
	}

	// Build interface lookup map for fast filtering.
	for _, iface := range cfg.Interfaces {
		discoverer.interfaces[iface] = true
	}

	// Build exclude ports lookup map.
	for _, port := range cfg.ExcludePorts {
		discoverer.excludePorts[port] = true
	}

	// Build include ports lookup map.
	for _, port := range cfg.IncludePorts {
		discoverer.includePorts[port] = true
	}

	// Return configured discoverer.
	return discoverer
}

// Type returns the target type for port scan discovery.
//
// Returns:
//   - target.Type: TypeCustom.
func (d *PortScanDiscoverer) Type() target.Type {
	// Return custom type constant for port scan targets.
	return target.TypeCustom
}

// Discover finds all listening TCP ports on the system.
// It reads /proc/net/tcp and /proc/net/tcp6 to find ports in LISTEN state.
//
// Params:
//   - ctx: context for cancellation.
//
// Returns:
//   - []target.ExternalTarget: the discovered listening ports.
//   - error: any error during discovery.
func (d *PortScanDiscoverer) Discover(ctx context.Context) ([]target.ExternalTarget, error) {
	// Check for context cancellation before starting.
	if err := ctx.Err(); err != nil {
		// Return early if context is cancelled.
		return nil, err
	}

	// Collect all listening ports from proc files.
	allPorts, err := d.collectAllListeningPorts()
	// Check for port collection error.
	if err != nil {
		// Return error from collection.
		return nil, err
	}

	// Filter and convert to external targets.
	return d.filterAndConvertPorts(allPorts), nil
}

// collectAllListeningPorts parses TCP and TCP6 proc files.
//
// Returns:
//   - []listeningPort: all discovered listening ports.
//   - error: any error during parsing.
func (d *PortScanDiscoverer) collectAllListeningPorts() ([]listeningPort, error) {
	var allPorts []listeningPort

	// Parse IPv4 TCP connections.
	tcp4Ports, err := d.parseNetTCP("/proc/net/tcp", "tcp4")
	// Check for TCP4 parsing error.
	if err != nil {
		// Return error with file context.
		return nil, fmt.Errorf("parsing /proc/net/tcp: %w", err)
	}
	allPorts = append(allPorts, tcp4Ports...)

	// Parse IPv6 TCP connections.
	tcp6Ports, err := d.parseNetTCP("/proc/net/tcp6", "tcp6")
	// Check for TCP6 parsing error.
	if err != nil {
		// Return error with file context.
		return nil, fmt.Errorf("parsing /proc/net/tcp6: %w", err)
	}
	allPorts = append(allPorts, tcp6Ports...)

	// Return all collected ports.
	return allPorts, nil
}

// filterAndConvertPorts filters ports and converts them to targets.
//
// Params:
//   - allPorts: the ports to filter and convert.
//
// Returns:
//   - []target.ExternalTarget: the filtered and converted targets.
func (d *PortScanDiscoverer) filterAndConvertPorts(allPorts []listeningPort) []target.ExternalTarget {
	var targets []target.ExternalTarget
	seenPorts := make(map[string]bool, len(allPorts))

	// Iterate through each listening port.
	for _, port := range allPorts {
		// Skip if already processed this port.
		key := fmt.Sprintf("%s:%d", port.LocalAddr, port.LocalPort)
		// Check if port was already seen.
		if seenPorts[key] {
			continue
		}

		// Apply port and interface filters.
		if !d.shouldIncludePort(port.LocalPort) {
			continue
		}
		// Check interface filter if interfaces are configured.
		if len(d.interfaces) > 0 && !d.matchesInterface(port.LocalAddr) {
			continue
		}

		// Mark port as seen and convert to target.
		seenPorts[key] = true
		targets = append(targets, d.portToTarget(port))
	}

	// Return filtered targets.
	return targets
}

// parseNetTCP parses /proc/net/tcp format and extracts listening ports.
//
// Params:
//   - path: path to the file (/proc/net/tcp or /proc/net/tcp6).
//   - protocol: protocol name ("tcp4" or "tcp6").
//
// Returns:
//   - []listeningPort: list of ports in LISTEN state.
//   - error: any error during parsing.
func (d *PortScanDiscoverer) parseNetTCP(path, protocol string) ([]listeningPort, error) {
	// Open /proc/net/tcp file.
	f, err := os.Open(path)
	// Check for file open error.
	if err != nil {
		// Return nil for non-existent files.
		if os.IsNotExist(err) {
			// Return empty result for missing file.
			return nil, nil
		}
		// Return error for other file errors.
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// Parse file content into listening ports.
	ports, err := d.scanNetTCPFile(f, protocol)
	// Check for scanning error.
	if err != nil {
		// Return error with scan context.
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}

	// Return parsed ports.
	return ports, nil
}

// scanNetTCPFile scans lines from a /proc/net/tcp file.
//
// Params:
//   - f: the file to scan.
//   - protocol: protocol name ("tcp4" or "tcp6").
//
// Returns:
//   - []listeningPort: list of ports in LISTEN state.
//   - error: any scanner error.
func (d *PortScanDiscoverer) scanNetTCPFile(f *os.File, protocol string) ([]listeningPort, error) {
	var ports []listeningPort
	scanner := bufio.NewScanner(f)

	// Skip header line.
	_ = scanner.Scan()

	// Parse each line.
	for scanner.Scan() {
		port, ok := d.parseNetTCPLine(scanner.Text(), protocol)
		// Append port if parsing succeeded.
		if ok {
			ports = append(ports, port)
		}
	}

	// Return parsed ports with any scanner error.
	return ports, scanner.Err()
}

// parseNetTCPLine parses a single line from /proc/net/tcp.
//
// Params:
//   - line: the line to parse.
//   - protocol: protocol name ("tcp4" or "tcp6").
//
// Returns:
//   - listeningPort: the parsed port (if valid).
//   - bool: true if the line was a valid listening port.
func (d *PortScanDiscoverer) parseNetTCPLine(line, protocol string) (listeningPort, bool) {
	line = strings.TrimSpace(line)
	// Skip empty lines.
	if line == "" {
		// Return false for empty line.
		return listeningPort{}, false
	}

	// Split line into fields.
	fields := strings.Fields(line)
	// Check for minimum required fields.
	if len(fields) < minTCPFields {
		// Return false for malformed line.
		return listeningPort{}, false
	}

	// Only interested in LISTEN state (0A in hex).
	state := fields[tcpStateFieldIndex]
	// Check if socket is in LISTEN state.
	if state != "0A" {
		// Return false for non-listening socket.
		return listeningPort{}, false
	}

	// Parse local address (field 1).
	addr, port, err := d.parseAddress(fields[1], protocol)
	// Check for address parsing error.
	if err != nil {
		// Return false for invalid address.
		return listeningPort{}, false
	}

	// Return parsed listening port.
	return listeningPort{
		Protocol:  protocol,
		LocalAddr: addr,
		LocalPort: port,
		State:     state,
	}, true
}

// parseAddress parses hex-encoded address from /proc/net/tcp.
// Format: "0100007F:0035" (IP:port in hex, little-endian for IPv4).
//
// Params:
//   - hexAddr: hex-encoded address string.
//   - protocol: protocol name for proper IP parsing.
//
// Returns:
//   - string: IP address.
//   - int: port number.
//   - error: any parsing error.
func (d *PortScanDiscoverer) parseAddress(hexAddr, protocol string) (string, int, error) {
	// Split address into IP and port.
	parts := strings.Split(hexAddr, ":")
	// Check for valid address format.
	if len(parts) != addressPartCount {
		// Return error for invalid format.
		return "", 0, fmt.Errorf("%w: %s", errInvalidAddressFormat, hexAddr)
	}

	// Parse port (always big-endian hex).
	portVal, err := parseHexPort(parts[1])
	// Check for port parsing error.
	if err != nil {
		// Return error from port parsing.
		return "", 0, err
	}

	// Parse IP address based on protocol.
	ipAddr, err := parseHexIP(parts[0], protocol)
	// Check for IP parsing error.
	if err != nil {
		// Return error from IP parsing.
		return "", 0, err
	}

	// Return parsed address components.
	return ipAddr, portVal, nil
}

// parseHexPort parses a hex-encoded port number.
//
// Params:
//   - portHex: hex-encoded port string.
//
// Returns:
//   - int: port number.
//   - error: any parsing error.
func parseHexPort(portHex string) (int, error) {
	portVal, err := strconv.ParseUint(portHex, hexBase, portBitSize)
	// Check for parsing error.
	if err != nil {
		// Return error with context.
		return 0, fmt.Errorf("parsing port %s: %w", portHex, err)
	}
	// Return parsed port value.
	return int(portVal), nil
}

// parseHexIP parses a hex-encoded IP address.
//
// Params:
//   - ipHex: hex-encoded IP string.
//   - protocol: protocol name ("tcp4" or "tcp6").
//
// Returns:
//   - string: IP address string.
//   - error: any parsing error.
func parseHexIP(ipHex, protocol string) (string, error) {
	ipBytes, err := hex.DecodeString(ipHex)
	// Check for hex decoding error.
	if err != nil {
		// Return error with context.
		return "", fmt.Errorf("decoding IP %s: %w", ipHex, err)
	}

	// Handle protocol-specific IP parsing.
	if protocol == "tcp4" {
		// Return IPv4 parsing result.
		return parseIPv4Bytes(ipBytes)
	}
	// Return IPv6 parsing result.
	return parseIPv6Bytes(ipBytes)
}

// parseIPv4Bytes converts IPv4 bytes to string (little-endian).
//
// Params:
//   - ipBytes: byte slice containing IPv4 address.
//
// Returns:
//   - string: IP address string representation.
//   - error: any parsing error.
func parseIPv4Bytes(ipBytes []byte) (string, error) {
	// Check for valid IPv4 length.
	if len(ipBytes) != ipv4ByteLength {
		// Return error for invalid length.
		return "", fmt.Errorf("%w: %d", errInvalidIPv4Length, len(ipBytes))
	}
	// Reverse bytes for little-endian.
	for i := range len(ipBytes) / byteReversalDivisor {
		j := len(ipBytes) - 1 - i
		ipBytes[i], ipBytes[j] = ipBytes[j], ipBytes[i]
	}
	// Return string representation of IP.
	return net.IP(ipBytes).String(), nil
}

// parseIPv6Bytes converts IPv6 bytes to string.
//
// Params:
//   - ipBytes: byte slice containing IPv6 address.
//
// Returns:
//   - string: IP address string representation.
//   - error: any parsing error.
func parseIPv6Bytes(ipBytes []byte) (string, error) {
	// Check for valid IPv6 length.
	if len(ipBytes) != ipv6ByteLength {
		// Return error for invalid length.
		return "", fmt.Errorf("%w: %d", errInvalidIPv6Length, len(ipBytes))
	}
	// Return string representation of IP.
	return net.IP(ipBytes).String(), nil
}

// shouldIncludePort checks if a port should be included based on filters.
//
// Params:
//   - port: the port number to check.
//
// Returns:
//   - bool: true if the port should be included.
func (d *PortScanDiscoverer) shouldIncludePort(port int) bool {
	// If IncludePorts is set, only include those ports.
	if len(d.includePorts) > 0 {
		// Return true if port is in include list.
		return d.includePorts[port]
	}

	// Otherwise, exclude ports in ExcludePorts.
	return !d.excludePorts[port]
}

// matchesInterface checks if an IP address belongs to a configured interface.
//
// Params:
//   - addr: the IP address to check.
//
// Returns:
//   - bool: true if the address matches any configured interface.
func (d *PortScanDiscoverer) matchesInterface(addr string) bool {
	// Get all network interfaces.
	interfaces, err := net.Interfaces()
	// Allow address if interface listing fails.
	if err != nil {
		// Return true on error to allow the address.
		return true
	}

	// Parse target IP.
	targetIP := net.ParseIP(addr)
	// Reject if address cannot be parsed.
	if targetIP == nil {
		// Return false for invalid IP.
		return false
	}

	// Check each configured interface.
	for _, iface := range interfaces {
		// Skip interface if not in configured set.
		if !d.interfaces[iface.Name] {
			continue
		}
		// Check if interface has the target IP.
		if d.interfaceContainsIP(&iface, targetIP) {
			// Return true when IP is found.
			return true
		}
	}

	// Return false when no interface matches.
	return false
}

// interfaceContainsIP checks if an interface has the given IP address.
//
// Params:
//   - iface: the network interface to check.
//   - targetIP: the IP address to find.
//
// Returns:
//   - bool: true if the interface has the target IP.
func (d *PortScanDiscoverer) interfaceContainsIP(iface Addrser, targetIP net.IP) bool {
	addrs, err := iface.Addrs()
	// Return false if addresses cannot be retrieved.
	if err != nil {
		// Return false on error.
		return false
	}

	// Check each address on the interface.
	for _, ifaceAddr := range addrs {
		ipNet, ok := ifaceAddr.(*net.IPNet)
		// Check if address matches target.
		if ok && ipNet.IP.Equal(targetIP) {
			// Return true when IP matches.
			return true
		}
	}

	// Return false when no address matches.
	return false
}

// portToTarget converts a listening port to an ExternalTarget.
// It configures a TCP probe for the discovered port.
//
// Params:
//   - port: the listening port information.
//
// Returns:
//   - target.ExternalTarget: the external target.
func (d *PortScanDiscoverer) portToTarget(port listeningPort) target.ExternalTarget {
	// Format target address.
	address := fmt.Sprintf("%s:%d", port.LocalAddr, port.LocalPort)

	// Create unique ID for this port.
	id := fmt.Sprintf("portscan:%s", address)

	// Create name with protocol and address.
	name := fmt.Sprintf("%s:%d", port.Protocol, port.LocalPort)

	// Initialize target with port scan specific configuration.
	t := target.ExternalTarget{
		ID:               id,
		Name:             name,
		Type:             target.TypeCustom,
		Source:           target.SourceDiscovered,
		Labels:           make(map[string]string, portScanLabelCount),
		ProbeType:        "tcp",
		ProbeTarget:      health.NewTCPTarget(address),
		Interval:         defaultProbeInterval,
		Timeout:          defaultProbeTimeout,
		SuccessThreshold: defaultProbeSuccessThreshold,
		FailureThreshold: defaultProbeFailureThreshold,
	}

	// Add labels for filtering and querying.
	t.Labels["portscan.protocol"] = port.Protocol
	t.Labels["portscan.port"] = strconv.Itoa(port.LocalPort)
	t.Labels["portscan.address"] = port.LocalAddr

	// Return fully configured target with TCP probe.
	return t
}
