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

type netStats struct {
	rxBytes uint64
	txBytes uint64
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

// CollectInto populates network interface information.
//
// Params:
//   - snap: target snapshot to populate
//
// Returns:
//   - error: error if interfaces cannot be retrieved
func (c *NetworkCollector) CollectInto(snap *model.Snapshot) error {
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
		ni := model.NetworkInterface{
			Name:       iface.Name,
			IsUp:       iface.Flags&net.FlagUp != 0,
			IsLoopback: iface.Flags&net.FlagLoopback != 0,
		}

		// Get IP addresses.
		if addrs, err := iface.Addrs(); err == nil {
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
				// Check for valid IPv4 address.
				if ip != nil && ip.To4() != nil {
					ni.IP = ip.String()
					// Stop after first IPv4.
					break
				}
			}
		}

		// Get stats (platform-specific).
		rxBytes, txBytes, speed := getInterfaceStats(iface.Name)
		ni.Speed = speed

		// Calculate rates if we have previous data.
		if prev, ok := c.prevStats[iface.Name]; ok {
			// Calculate RX rate (handle counter wrap).
			// Check if RX counter advanced.
			if rxBytes >= prev.rxBytes {
				ni.RxBytesPerSec = rxBytes - prev.rxBytes
			}
			// Calculate TX rate (handle counter wrap).
			// Check if TX counter advanced.
			if txBytes >= prev.txBytes {
				ni.TxBytesPerSec = txBytes - prev.txBytes
			}

			// Update adaptive speed based on observed throughput (for virtual interfaces).
			// Use max of RX and TX, convert to bits/sec.
			maxBytes := ni.RxBytesPerSec
			// Check if TX is higher.
			if ni.TxBytesPerSec > maxBytes {
				maxBytes = ni.TxBytesPerSec
			}
			// Update speed estimation if there is traffic.
			// Check for non-zero traffic.
			if maxBytes > 0 {
				UpdateAdaptiveSpeed(iface.Name, maxBytes*bitsPerByte)
			}
		}

		// Store for next iteration.
		c.prevStats[iface.Name] = netStats{
			rxBytes: rxBytes,
			txBytes: txBytes,
		}

		snap.Network = append(snap.Network, ni)
	}

	// Successfully collected network data.
	return nil
}
