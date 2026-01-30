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
//	cpu, err := collector.CPU().CollectSystem(ctx)
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
func NewCollector() *Collector {
	return &Collector{
		cpu:        NewCPUCollector(),
		memory:     NewMemoryCollector(),
		disk:       NewDiskCollector(),
		network:    NewNetworkCollector(),
		io:         NewIOCollector(),
		connection: NewConnectionCollector(),
	}
}

// CPU returns the CPU metrics collector.
func (c *Collector) CPU() metrics.CPUCollector {
	return c.cpu
}

// Memory returns the memory metrics collector.
func (c *Collector) Memory() metrics.MemoryCollector {
	return c.memory
}

// Disk returns the disk metrics collector.
func (c *Collector) Disk() metrics.DiskCollector {
	return c.disk
}

// Network returns the network metrics collector.
func (c *Collector) Network() metrics.NetworkCollector {
	return c.network
}

// IO returns the I/O metrics collector.
func (c *Collector) IO() metrics.IOCollector {
	return c.io
}

// Connection returns the network connection collector.
// This provides TCP, UDP, and Unix socket enumeration with process resolution.
func (c *Collector) Connection() *ConnectionCollector {
	return c.connection
}

// Ensure Collector implements metrics.SystemCollector.
var _ metrics.SystemCollector = (*Collector)(nil)
