//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ProcessIO contains I/O statistics for a process.
//
// This value object captures the I/O throughput of a specific process measured
// in bytes per second. Values are calculated by the Rust probe based on deltas.
type ProcessIO struct {
	// PID is the process identifier.
	PID int
	// ReadBytesPerSec is the read throughput in bytes per second.
	ReadBytesPerSec uint64
	// WriteBytesPerSec is the write throughput in bytes per second.
	WriteBytesPerSec uint64
}
