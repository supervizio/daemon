// Package collector provides data collectors for TUI snapshot.
package collector

import (
	"net"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

const (
	// typicalInterfaceCount is the typical number of network interfaces.
	typicalInterfaceCount int = 8

	// bitsPerByte is the number of bits in a byte for bandwidth calculations.
	bitsPerByte uint64 = 8
)

// NetworkCollector gathers network interface information.
// It tracks interface states, IP addresses, and bandwidth statistics.
type NetworkCollector struct {
	// Previous stats for calculating rates.
	prevStats map[string]netStats
}

// NewNetworkCollector creates a network collector.
//
// Returns:
//   - *NetworkCollector: configured network collector
func NewNetworkCollector() *NetworkCollector {
	// Pre-allocate map for typical interface count.
	return &NetworkCollector{
		prevStats: make(map[string]netStats, typicalInterfaceCount),
	}
}

// Gather populates network interface information.
//
// Params:
//   - snap: target snapshot to populate
//
// Returns:
//   - error: error if interfaces cannot be retrieved
func (c *NetworkCollector) Gather(snap *model.Snapshot) error {
	// Get all network interfaces.
	interfaces, err := net.Interfaces()
	// Handle interface retrieval error.
	if err != nil {
		// Failed to get interfaces.
		return err
	}

	snap.Network = make([]model.NetworkInterface, 0, len(interfaces))

	// Process each interface.
	for _, iface := range interfaces {
		ni := c.collectInterface(iface)
		snap.Network = append(snap.Network, ni)
	}

	// Successfully collected network data.
	return nil
}

// collectInterface collects data for a single network interface.
//
// Params:
//   - iface: network interface to collect
//
// Returns:
//   - model.NetworkInterface: populated interface data
func (c *NetworkCollector) collectInterface(iface net.Interface) model.NetworkInterface {
	ni := model.NetworkInterface{
		Name:       iface.Name,
		IsUp:       iface.Flags&net.FlagUp != 0,
		IsLoopback: iface.Flags&net.FlagLoopback != 0,
	}

	// Get IP addresses.
	ni.IP = c.getInterfaceIPv4(&iface)

	// Get stats (platform-specific).
	rxBytes, txBytes, speed := getInterfaceStats(iface.Name)
	ni.Speed = speed

	// Calculate rates if we have previous data.
	c.calculateRates(&ni, iface.Name, rxBytes, txBytes)

	// Store for next iteration.
	c.prevStats[iface.Name] = netStats{
		rxBytes: rxBytes,
		txBytes: txBytes,
	}

	// Return populated interface.
	return ni
}

// getInterfaceIPv4 extracts the first IPv4 address from an interface.
//
// Params:
//   - iface: network interface to scan (uses Addrser interface)
//
// Returns:
//   - string: IPv4 address or empty string
func (c *NetworkCollector) getInterfaceIPv4(iface Addrser) string {
	addrs, err := iface.Addrs()
	// Handle address retrieval error.
	if err != nil {
		// Cannot get addresses.
		return ""
	}

	// Find first IPv4 address.
	for _, addr := range addrs {
		var ip net.IP
		// Extract IP from address.
		switch ipAddr := addr.(type) {
		// Handle *net.IPNet type.
		case *net.IPNet:
			ip = ipAddr.IP
		// Handle *net.IPAddr type.
		case *net.IPAddr:
			ip = ipAddr.IP
		}

		// Prefer IPv4.
		if ip != nil && ip.To4() != nil {
			// Found IPv4.
			return ip.String()
		}
	}

	// No IPv4 found.
	return ""
}

// calculateRates calculates RX/TX rates and updates adaptive speed.
//
// Params:
//   - ni: network interface to update
//   - name: interface name
//   - rxBytes: current RX bytes counter
//   - txBytes: current TX bytes counter
func (c *NetworkCollector) calculateRates(ni *model.NetworkInterface, name string, rxBytes, txBytes uint64) {
	prev, ok := c.prevStats[name]
	// Skip if no previous data.
	if !ok {
		// No previous sample.
		return
	}

	// Calculate RX rate (handle counter wrap).
	if rxBytes >= prev.rxBytes {
		ni.RxBytesPerSec = rxBytes - prev.rxBytes
	}

	// Calculate TX rate (handle counter wrap).
	if txBytes >= prev.txBytes {
		ni.TxBytesPerSec = txBytes - prev.txBytes
	}

	// Update adaptive speed based on observed throughput.
	maxBytes := max(ni.RxBytesPerSec, ni.TxBytesPerSec)

	// Update speed estimation if there is traffic.
	if maxBytes > 0 {
		UpdateAdaptiveSpeed(name, maxBytes*bitsPerByte)
	}
}
