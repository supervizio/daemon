//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// cross-platform system metrics collection.
//
// The probe package acts as an adapter implementing the domain.SystemCollector
// interface by delegating to the Rust libprobe library via FFI. This enables
// cross-platform metrics collection without CGO dependencies on platform-specific
// system libraries.
//
// Usage:
//
//	if err := probe.Init(); err != nil {
//	    log.Fatal("failed to initialize probe:", err)
//	}
//	defer probe.Shutdown()
//
//	collector := probe.NewCollector()
//	cpu, err := collector.Cpu().CollectSystem(ctx)
//
// Limitations:
//   - Some detailed metrics (jiffies) are not available cross-platform
//   - PSI metrics are Linux-only (returns ErrNotSupported on other platforms)
//   - Process enumeration is not supported (use platform-specific collectors)
package probe

import (
	"github.com/kodflow/daemon/internal/domain/metrics"
)

// Collector implements metrics.SystemCollector via the Rust probe library.
// It provides cross-platform metrics collection for CPU, memory, and load.
type Collector struct {
	cpu        *CPUCollector
	memory     *MemoryCollector
	disk       *DiskCollector
	network    *NetworkCollector
	io         *IOCollector
	connection *ConnectionCollector
}

// NewCollector creates a new system metrics collector backed by the Rust probe.
// The probe library must be initialized via Init() before using the collector.
//
// Returns:
//   - *Collector: new system metrics collector instance
func NewCollector() *Collector {
	// Initialize all sub-collectors for comprehensive metrics collection.
	return &Collector{
		cpu:        NewCPUCollector(),
		memory:     NewMemoryCollector(),
		disk:       NewDiskCollector(),
		network:    NewNetworkCollector(),
		io:         NewIOCollector(),
		connection: NewConnectionCollector(),
	}
}

// Cpu returns the CPU metrics collector.
//
// Returns:
//   - metrics.CPUCollector: collector for CPU metrics
func (c *Collector) Cpu() metrics.CPUCollector {
	// Return the pre-initialized CPU collector.
	return c.cpu
}

// Memory returns the memory metrics collector.
//
// Returns:
//   - metrics.MemoryCollector: collector for memory metrics
func (c *Collector) Memory() metrics.MemoryCollector {
	// Return the pre-initialized memory collector.
	return c.memory
}

// Disk returns the disk metrics collector.
//
// Returns:
//   - metrics.DiskCollector: collector for disk metrics
func (c *Collector) Disk() metrics.DiskCollector {
	// Return the pre-initialized disk collector.
	return c.disk
}

// Network returns the network metrics collector.
//
// Returns:
//   - metrics.NetworkCollector: collector for network metrics
func (c *Collector) Network() metrics.NetworkCollector {
	// Return the pre-initialized network collector.
	return c.network
}

// Io returns the I/O metrics collector.
//
// Returns:
//   - metrics.IOCollector: collector for I/O metrics
func (c *Collector) Io() metrics.IOCollector {
	// Return the pre-initialized I/O collector.
	return c.io
}

// Connection returns the network connection collector.
// This provides TCP, UDP, and Unix socket enumeration with process resolution.
//
// Returns:
//   - *ConnectionCollector: collector for network connections
func (c *Collector) Connection() *ConnectionCollector {
	// Return the pre-initialized connection collector.
	return c.connection
}

// Ensure Collector implements metrics.SystemCollector.
var _ metrics.SystemCollector = (*Collector)(nil)
