//go:build linux

// Package discovery provides infrastructure adapters for target discovery.
package discovery

import (
	"bufio"
	"context"
	"encoding/hex"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/target"
)

// listeningPort represents a port in LISTEN state from /proc/net/tcp.
type listeningPort struct {
	Protocol  string // "tcp" or "udp"
	LocalAddr string // IP address
	LocalPort int    // Port number
	State     string // Socket state (hex)
}

// PortScanDiscoverer discovers listening ports by parsing /proc/net/tcp.
// It scans local network interfaces for services listening on TCP ports
// and creates monitoring targets for them.
type PortScanDiscoverer struct {
	// interfaces maps interface names to true for fast lookup.
	interfaces map[string]bool

	// excludePorts maps port numbers to exclude.
	excludePorts map[int]bool

	// includePorts maps port numbers to include exclusively.
	includePorts map[int]bool
}

// NewPortScanDiscoverer creates a new port scan discoverer.
//
// Params:
//   - cfg: the port scan discovery configuration.
//
// Returns:
//   - *PortScanDiscoverer: a new port scan discoverer.
func NewPortScanDiscoverer(cfg *config.PortScanDiscoveryConfig) *PortScanDiscoverer {
	d := &PortScanDiscoverer{
		interfaces:   make(map[string]bool, len(cfg.Interfaces)),
		excludePorts: make(map[int]bool, len(cfg.ExcludePorts)),
		includePorts: make(map[int]bool, len(cfg.IncludePorts)),
	}

	// Build interface lookup map for fast filtering.
	for _, iface := range cfg.Interfaces {
		d.interfaces[iface] = true
	}

	// Build exclude ports lookup map.
	for _, port := range cfg.ExcludePorts {
		d.excludePorts[port] = true
	}

	// Build include ports lookup map.
	for _, port := range cfg.IncludePorts {
		d.includePorts[port] = true
	}

	// Return configured discoverer.
	return d
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
		return nil, err
	}

	var allPorts []listeningPort

	// Parse IPv4 TCP connections.
	tcp4Ports, err := d.parseNetTCP("/proc/net/tcp", "tcp4")
	if err != nil {
		// Return error when IPv4 TCP parsing fails.
		return nil, fmt.Errorf("parsing /proc/net/tcp: %w", err)
	}
	allPorts = append(allPorts, tcp4Ports...)

	// Parse IPv6 TCP connections.
	tcp6Ports, err := d.parseNetTCP("/proc/net/tcp6", "tcp6")
	if err != nil {
		// Return error when IPv6 TCP parsing fails.
		return nil, fmt.Errorf("parsing /proc/net/tcp6: %w", err)
	}
	allPorts = append(allPorts, tcp6Ports...)

	// Filter and convert to external targets.
	targets := make([]target.ExternalTarget, 0)
	seenPorts := make(map[string]bool, len(allPorts))

	// Iterate over all discovered ports.
	for _, port := range allPorts {
		// Skip if already processed this port.
		key := fmt.Sprintf("%s:%d", port.LocalAddr, port.LocalPort)
		if seenPorts[key] {
			continue
		}

		// Apply port filters.
		if !d.shouldIncludePort(port.LocalPort) {
			continue
		}

		// Apply interface filters if configured.
		if len(d.interfaces) > 0 && !d.matchesInterface(port.LocalAddr) {
			continue
		}

		// Mark port as seen.
		seenPorts[key] = true

		// Create external target for this listening port.
		t := d.portToTarget(port)
		targets = append(targets, t)
	}

	// Return discovered targets.
	return targets, nil
}

// parseNetTCP parses /proc/net/tcp format and extracts listening ports.
//
// Format of /proc/net/tcp:
//
//	sl  local_address rem_address   st tx_queue rx_queue tr tm->when retrnsmt   uid  timeout inode
//	0: 0100007F:0035 00000000:0000 0A 00000000:00000000 00:00000000 00000000     0        0 17892
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
	if err != nil {
		// Ignore missing file (might not have IPv6 support).
		if os.IsNotExist(err) {
			return nil, nil
		}
		// Return error when file cannot be opened.
		return nil, err
	}
	defer f.Close()

	var ports []listeningPort
	scanner := bufio.NewScanner(f)

	// Skip header line.
	if scanner.Scan() {
		// Header line skipped.
	}

	// Parse each line.
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		// Skip empty lines.
		if line == "" {
			continue
		}

		// Split line into fields.
		fields := strings.Fields(line)
		// Need at least 4 fields: sl, local_address, rem_address, st.
		if len(fields) < 4 {
			continue
		}

		// Extract state (field 3).
		state := fields[3]
		// Only interested in LISTEN state (0A in hex).
		if state != "0A" {
			continue
		}

		// Parse local address (field 1).
		localAddr := fields[1]
		addr, port, err := d.parseAddress(localAddr, protocol)
		if err != nil {
			// Skip malformed addresses.
			continue
		}

		// Add to ports list.
		ports = append(ports, listeningPort{
			Protocol:  protocol,
			LocalAddr: addr,
			LocalPort: port,
			State:     state,
		})
	}

	// Check for scanner errors.
	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("scanning %s: %w", path, err)
	}

	// Return parsed listening ports.
	return ports, nil
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
	if len(parts) != 2 {
		return "", 0, fmt.Errorf("invalid address format: %s", hexAddr)
	}

	// Parse port (always big-endian hex).
	portHex := parts[1]
	portVal, err := strconv.ParseUint(portHex, 16, 16)
	if err != nil {
		return "", 0, fmt.Errorf("parsing port %s: %w", portHex, err)
	}

	// Parse IP address (little-endian for IPv4, big-endian for IPv6).
	ipHex := parts[0]
	ipBytes, err := hex.DecodeString(ipHex)
	if err != nil {
		return "", 0, fmt.Errorf("decoding IP %s: %w", ipHex, err)
	}

	var ipAddr string
	if protocol == "tcp4" {
		// IPv4: reverse byte order (little-endian).
		if len(ipBytes) != 4 {
			return "", 0, fmt.Errorf("invalid IPv4 length: %d", len(ipBytes))
		}
		// Reverse bytes for little-endian.
		for i := 0; i < len(ipBytes)/2; i++ {
			j := len(ipBytes) - 1 - i
			ipBytes[i], ipBytes[j] = ipBytes[j], ipBytes[i]
		}
		ipAddr = net.IP(ipBytes).String()
	} else {
		// IPv6: already in correct byte order.
		if len(ipBytes) != 16 {
			return "", 0, fmt.Errorf("invalid IPv6 length: %d", len(ipBytes))
		}
		ipAddr = net.IP(ipBytes).String()
	}

	// Return parsed IP and port.
	return ipAddr, int(portVal), nil
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
	if err != nil {
		// Cannot verify interfaces - include all.
		return true
	}

	// Parse target IP.
	targetIP := net.ParseIP(addr)
	if targetIP == nil {
		// Invalid IP - exclude.
		return false
	}

	// Check each interface.
	for _, iface := range interfaces {
		// Skip if not in configured interfaces.
		if !d.interfaces[iface.Name] {
			continue
		}

		// Get addresses for this interface.
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}

		// Check if target IP matches any address on this interface.
		for _, ifaceAddr := range addrs {
			// Parse interface address.
			ipNet, ok := ifaceAddr.(*net.IPNet)
			if !ok {
				continue
			}

			// Check if target IP matches.
			if ipNet.IP.Equal(targetIP) {
				return true
			}
		}
	}

	// No interface matched.
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
		Labels:           make(map[string]string, 3),
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
