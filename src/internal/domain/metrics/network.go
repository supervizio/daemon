// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// NetInterface represents a network interface.
type NetInterface struct {
	// Name is the interface name (e.g., "eth0", "en0", "lo").
	Name string
	// Index is the interface index.
	Index int
	// HardwareAddr is the MAC address (e.g., "00:11:22:33:44:55").
	HardwareAddr string
	// MTU is the Maximum Transmission Unit in bytes.
	MTU int
	// Flags describes the interface state (e.g., "up", "broadcast", "multicast").
	Flags []string
	// Addresses are the IP addresses assigned to this interface.
	Addresses []string
}

// IsUp returns true if the interface is up.
func (n NetInterface) IsUp() bool {
	for _, flag := range n.Flags {
		if flag == "up" {
			return true
		}
	}
	return false
}

// IsLoopback returns true if this is a loopback interface.
func (n NetInterface) IsLoopback() bool {
	for _, flag := range n.Flags {
		if flag == "loopback" {
			return true
		}
	}
	return false
}

// NetStats represents network statistics for an interface.
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

// TotalBytes returns the total bytes transferred (sent + received).
func (n NetStats) TotalBytes() uint64 {
	return n.BytesSent + n.BytesRecv
}

// TotalPackets returns the total packets transferred (sent + received).
func (n NetStats) TotalPackets() uint64 {
	return n.PacketsSent + n.PacketsRecv
}

// TotalErrors returns the total number of errors.
func (n NetStats) TotalErrors() uint64 {
	return n.ErrorsIn + n.ErrorsOut
}

// TotalDrops returns the total number of dropped packets.
func (n NetStats) TotalDrops() uint64 {
	return n.DropsIn + n.DropsOut
}

// Bandwidth represents network bandwidth measurements for an interface.
// This is calculated from two NetStats samples taken at different times.
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

// TotalBytesPerSec returns the combined bandwidth in bytes per second.
func (b Bandwidth) TotalBytesPerSec() float64 {
	return b.TxBytesPerSec + b.RxBytesPerSec
}

// TotalPacketsPerSec returns the combined packet rate.
func (b Bandwidth) TotalPacketsPerSec() float64 {
	return b.TxPacketsPerSec + b.RxPacketsPerSec
}

// TxBitsPerSec returns the transmit rate in bits per second.
func (b Bandwidth) TxBitsPerSec() float64 {
	return b.TxBytesPerSec * 8
}

// RxBitsPerSec returns the receive rate in bits per second.
func (b Bandwidth) RxBitsPerSec() float64 {
	return b.RxBytesPerSec * 8
}

// CalculateBandwidth calculates bandwidth between two NetStats samples.
func CalculateBandwidth(prev, curr NetStats) Bandwidth {
	duration := curr.Timestamp.Sub(prev.Timestamp)
	if duration <= 0 {
		return Bandwidth{
			Interface: curr.Interface,
			Timestamp: curr.Timestamp,
		}
	}

	seconds := duration.Seconds()
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
