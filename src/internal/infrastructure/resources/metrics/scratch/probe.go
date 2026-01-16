// Package scratch provides a minimal metrics collector for environments
// without /proc filesystem access (e.g., scratch containers, Windows).
// It returns zero values or errors for most operations.
package scratch

import (
	"errors"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// Sentinel errors for scratch metrics package.
var (
	// ErrNotSupported is returned when a metric is not available in scratch mode.
	ErrNotSupported error = errors.New("metric collection not supported in scratch mode")
	// ErrInvalidPID is returned when an invalid process ID is provided.
	ErrInvalidPID error = errors.New("invalid process ID: must be positive")
	// ErrEmptyPath is returned when an empty mount path is provided.
	ErrEmptyPath error = errors.New("empty mount path")
	// ErrEmptyDevice is returned when an empty device name is provided.
	ErrEmptyDevice error = errors.New("empty device name")
	// ErrEmptyInterface is returned when an empty interface name is provided.
	ErrEmptyInterface error = errors.New("empty interface name")
)

// Probe implements metrics.SystemCollector with minimal functionality.
// It's designed for environments where system metrics are unavailable.
// All collector methods return ErrNotSupported to indicate the platform
// does not support system metrics collection.
type Probe struct {
	collCPU  *CPUCollector
	collMem  *MemoryCollector
	collDisk *DiskCollector
	collNet  *NetworkCollector
	collIO   *IOCollector
}

// NewProbe creates a new minimal probe for scratch environments.
//
// Returns:
//   - *Probe: initialized probe with stub collectors
func NewProbe() *Probe {
	// Return probe with initialized stub collectors
	return &Probe{
		collCPU:  NewCPUCollector(),
		collMem:  NewMemoryCollector(),
		collDisk: NewDiskCollector(),
		collNet:  NewNetworkCollector(),
		collIO:   NewIOCollector(),
	}
}

// CPU returns the CPU collector.
//
// Returns:
//   - metrics.CPUCollector: CPU metrics collector
func (p *Probe) CPU() metrics.CPUCollector {
	// Assign to local variable for interface implementation
	collector := p.collCPU
	// Return the CPU collector instance
	return collector
}

// Memory returns the memory collector.
//
// Returns:
//   - metrics.MemoryCollector: memory metrics collector
func (p *Probe) Memory() metrics.MemoryCollector {
	// Assign to local variable for interface implementation
	collector := p.collMem
	// Return the memory collector instance
	return collector
}

// Disk returns the disk collector.
//
// Returns:
//   - metrics.DiskCollector: disk metrics collector
func (p *Probe) Disk() metrics.DiskCollector {
	// Assign to local variable for interface implementation
	collector := p.collDisk
	// Return the disk collector instance
	return collector
}

// Network returns the network collector.
//
// Returns:
//   - metrics.NetworkCollector: network metrics collector
func (p *Probe) Network() metrics.NetworkCollector {
	// Assign to local variable for interface implementation
	collector := p.collNet
	// Return the network collector instance
	return collector
}

// IO returns the I/O collector.
//
// Returns:
//   - metrics.IOCollector: I/O metrics collector
func (p *Probe) IO() metrics.IOCollector {
	// Assign to local variable for interface implementation
	collector := p.collIO
	// Return the I/O collector instance
	return collector
}
