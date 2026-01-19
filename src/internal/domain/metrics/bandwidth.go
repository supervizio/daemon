// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

const (
	// bitsPerByte is the number of bits in one byte.
	bitsPerByte float64 = 8
)

// Bandwidth represents network bandwidth measurements for an interface.
//
// This is calculated from two NetStats samples taken at different times,
// providing throughput rates in bytes and packets per second.
type Bandwidth struct {
	// Interface is the interface name.
	Interface string
	// TxBytesPerSec is the transmit rate in bytes per second.
	TxBytesPerSec float64
	// RxBytesPerSec is the receive rate in bytes per second.
	RxBytesPerSec float64
	// TxPacketsPerSec is the transmit rate in packets per second.
	TxPacketsPerSec float64
	// RxPacketsPerSec is the receive rate in packets per second.
	RxPacketsPerSec float64
	// Duration is the time between the two samples.
	Duration time.Duration
	// Timestamp is when this measurement was calculated.
	Timestamp time.Time
}

// NewBandwidth creates a new Bandwidth with essential fields.
//
// Params:
//   - iface: interface name
//   - timestamp: when this measurement was calculated
//
// Returns:
//   - Bandwidth: new bandwidth instance
func NewBandwidth(iface string, timestamp time.Time) Bandwidth {
	// Return bandwidth with essential fields set.
	return Bandwidth{
		Interface: iface,
		Timestamp: timestamp,
	}
}

// TotalBytesPerSec returns the combined bandwidth in bytes per second.
//
// Returns:
//   - float64: sum of transmit and receive bytes per second
func (b Bandwidth) TotalBytesPerSec() float64 {
	// Sum transmit and receive rates to get total bandwidth.
	return b.TxBytesPerSec + b.RxBytesPerSec
}

// TotalPacketsPerSec returns the combined packet rate.
//
// Returns:
//   - float64: sum of transmit and receive packets per second
func (b Bandwidth) TotalPacketsPerSec() float64 {
	// Sum transmit and receive packet rates to get total packet rate.
	return b.TxPacketsPerSec + b.RxPacketsPerSec
}

// TxBitsPerSec returns the transmit rate in bits per second.
//
// Returns:
//   - float64: transmit rate in bits per second
func (b Bandwidth) TxBitsPerSec() float64 {
	// Convert bytes to bits by multiplying by bitsPerByte.
	return b.TxBytesPerSec * bitsPerByte
}

// RxBitsPerSec returns the receive rate in bits per second.
//
// Returns:
//   - float64: receive rate in bits per second
func (b Bandwidth) RxBitsPerSec() float64 {
	// Convert bytes to bits by multiplying by bitsPerByte.
	return b.RxBytesPerSec * bitsPerByte
}

// CalculateBandwidth calculates bandwidth between two NetStats samples.
//
// Params:
//   - prev: previous network statistics sample
//   - curr: current network statistics sample
//
// Returns:
//   - Bandwidth: calculated bandwidth metrics
func CalculateBandwidth(prev, curr *NetStats) Bandwidth {
	duration := curr.Timestamp.Sub(prev.Timestamp)
	// Check if duration is invalid (zero or negative).
	if duration <= 0 {
		// Return empty bandwidth with just interface and timestamp.
		return Bandwidth{
			Interface: curr.Interface,
			Timestamp: curr.Timestamp,
		}
	}

	seconds := duration.Seconds()
	// Calculate rates by dividing deltas by elapsed time in seconds.
	return Bandwidth{
		Interface:       curr.Interface,
		TxBytesPerSec:   float64(curr.BytesSent-prev.BytesSent) / seconds,
		RxBytesPerSec:   float64(curr.BytesRecv-prev.BytesRecv) / seconds,
		TxPacketsPerSec: float64(curr.PacketsSent-prev.PacketsSent) / seconds,
		RxPacketsPerSec: float64(curr.PacketsRecv-prev.PacketsRecv) / seconds,
		Duration:        duration,
		Timestamp:       curr.Timestamp,
	}
}
