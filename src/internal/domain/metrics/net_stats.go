// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// NetStats represents network statistics for an interface.
//
// Captures detailed network traffic metrics including bytes, packets, errors,
// and drops for both transmit and receive directions.
type NetStats struct {
	// Interface is the interface name.
	Interface string
	// BytesSent is the total number of bytes transmitted.
	BytesSent uint64
	// BytesRecv is the total number of bytes received.
	BytesRecv uint64
	// PacketsSent is the total number of packets transmitted.
	PacketsSent uint64
	// PacketsRecv is the total number of packets received.
	PacketsRecv uint64
	// ErrorsIn is the number of receive errors.
	ErrorsIn uint64
	// ErrorsOut is the number of transmit errors.
	ErrorsOut uint64
	// DropsIn is the number of incoming packets dropped.
	DropsIn uint64
	// DropsOut is the number of outgoing packets dropped.
	DropsOut uint64
	// FIFOIn is the number of FIFO buffer errors on receive.
	FIFOIn uint64
	// FIFOOut is the number of FIFO buffer errors on transmit.
	FIFOOut uint64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// NewNetStats creates a new NetStats with essential fields.
//
// Params:
//   - iface: interface name
//   - timestamp: when this sample was taken
//
// Returns:
//   - NetStats: new network statistics instance
func NewNetStats(iface string, timestamp time.Time) NetStats {
	return NetStats{
		Interface: iface,
		Timestamp: timestamp,
	}
}

// TotalBytes returns the total bytes transferred (sent + received).
//
// Returns:
//   - uint64: sum of bytes sent and received
func (n *NetStats) TotalBytes() uint64 {
	return n.BytesSent + n.BytesRecv
}

// TotalPackets returns the total packets transferred (sent + received).
//
// Returns:
//   - uint64: sum of packets sent and received
func (n *NetStats) TotalPackets() uint64 {
	return n.PacketsSent + n.PacketsRecv
}

// TotalErrors returns the total number of errors.
//
// Returns:
//   - uint64: sum of input and output errors
func (n *NetStats) TotalErrors() uint64 {
	return n.ErrorsIn + n.ErrorsOut
}

// TotalDrops returns the total number of dropped packets.
//
// Returns:
//   - uint64: sum of input and output drops
func (n *NetStats) TotalDrops() uint64 {
	return n.DropsIn + n.DropsOut
}
