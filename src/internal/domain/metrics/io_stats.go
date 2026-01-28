// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// IOStats represents system-wide I/O statistics.
// This type captures total I/O operations and bytes transferred across all devices.
type IOStats struct {
	// ReadBytesTotal is the total bytes read across all devices.
	ReadBytesTotal uint64
	// WriteBytesTotal is the total bytes written across all devices.
	WriteBytesTotal uint64
	// ReadOpsTotal is the total read operations across all devices.
	ReadOpsTotal uint64
	// WriteOpsTotal is the total write operations across all devices.
	WriteOpsTotal uint64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// NewIOStats creates a new IOStats instance.
//
// Params:
//   - readBytes: total bytes read across all devices
//   - writeBytes: total bytes written across all devices
//   - readOps: total read operations across all devices
//   - writeOps: total write operations across all devices
//   - timestamp: when this sample was taken
//
// Returns:
//   - IOStats: the created IOStats instance
func NewIOStats(readBytes, writeBytes, readOps, writeOps uint64, timestamp time.Time) IOStats {
	return IOStats{
		ReadBytesTotal:  readBytes,
		WriteBytesTotal: writeBytes,
		ReadOpsTotal:    readOps,
		WriteOpsTotal:   writeOps,
		Timestamp:       timestamp,
	}
}

// TotalBytes returns the total bytes transferred (read + write).
//
// Returns:
//   - uint64: the sum of read and write bytes
func (i IOStats) TotalBytes() uint64 {
	return i.ReadBytesTotal + i.WriteBytesTotal
}

// TotalOps returns the total operations (read + write).
//
// Returns:
//   - uint64: the sum of read and write operations
func (i IOStats) TotalOps() uint64 {
	return i.ReadOpsTotal + i.WriteOpsTotal
}
