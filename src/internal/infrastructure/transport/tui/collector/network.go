// Package collector provides data collectors for TUI snapshot.
package collector

import (
	"net"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// NetworkCollector gathers network interface information.
type NetworkCollector struct {
	// Previous stats for calculating rates.
	prevStats map[string]netStats
}

type netStats struct {
	rxBytes uint64
	txBytes uint64
}

// NewNetworkCollector creates a network collector.
func NewNetworkCollector() *NetworkCollector {
	return &NetworkCollector{
		prevStats: make(map[string]netStats, 8), // Pre-allocate for typical interface count.
	}
}

// CollectInto populates network interface information.
func (c *NetworkCollector) CollectInto(snap *model.Snapshot) error {
	interfaces, err := net.Interfaces()
	if err != nil {
		return err
	}

	snap.Network = make([]model.NetworkInterface, 0, len(interfaces))

	for _, iface := range interfaces {
		ni := model.NetworkInterface{
			Name:       iface.Name,
			IsUp:       iface.Flags&net.FlagUp != 0,
			IsLoopback: iface.Flags&net.FlagLoopback != 0,
		}

		// Get IP addresses.
		if addrs, err := iface.Addrs(); err == nil {
			for _, addr := range addrs {
				var ip net.IP
				switch v := addr.(type) {
				case *net.IPNet:
					ip = v.IP
				case *net.IPAddr:
					ip = v.IP
				}

				// Prefer IPv4.
				if ip != nil && ip.To4() != nil {
					ni.IP = ip.String()
					break
				}
			}
		}

		// Get stats (platform-specific).
		rxBytes, txBytes, speed := getInterfaceStats(iface.Name)
		ni.Speed = speed

		// Calculate rates.
		if prev, ok := c.prevStats[iface.Name]; ok {
			if rxBytes >= prev.rxBytes {
				ni.RxBytesPerSec = rxBytes - prev.rxBytes
			}
			if txBytes >= prev.txBytes {
				ni.TxBytesPerSec = txBytes - prev.txBytes
			}

			// Update adaptive speed based on observed throughput (for virtual interfaces).
			// Use max of RX and TX, convert to bits/sec.
			maxBytes := ni.RxBytesPerSec
			if ni.TxBytesPerSec > maxBytes {
				maxBytes = ni.TxBytesPerSec
			}
			if maxBytes > 0 {
				UpdateAdaptiveSpeed(iface.Name, maxBytes*8)
			}
		}

		// Store for next iteration.
		c.prevStats[iface.Name] = netStats{
			rxBytes: rxBytes,
			txBytes: txBytes,
		}

		snap.Network = append(snap.Network, ni)
	}

	return nil
}
