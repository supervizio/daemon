//go:build darwin

// Package darwin provides system metrics collection for macOS
// using sysctl and Mach APIs.
package darwin

import (
	"context"
	"errors"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// Sentinel errors for Darwin metrics collection.
var (
	// ErrNotImplemented is returned for methods not yet implemented.
	ErrNotImplemented = errors.New("Darwin metrics collector not yet implemented")
	// ErrInvalidPID indicates an invalid process ID was provided.
	ErrInvalidPID = errors.New("invalid pid")
	// ErrEmptyPath indicates an empty path was provided.
	ErrEmptyPath = errors.New("empty path")
	// ErrEmptyDevice indicates an empty device name was provided.
	ErrEmptyDevice = errors.New("empty device name")
	// ErrEmptyInterface indicates an empty interface name was provided.
	ErrEmptyInterface = errors.New("empty interface name")
)

// Probe implements metrics.SystemCollector for macOS.
type Probe struct {
	cpu     *CPUCollector
	memory  *MemoryCollector
	disk    *DiskCollector
	network *NetworkCollector
	io      *IOCollector
}

// NewProbe creates a new macOS metrics collector.
func NewProbe() *Probe {
	return &Probe{
		cpu:     &CPUCollector{},
		memory:  &MemoryCollector{},
		disk:    &DiskCollector{},
		network: &NetworkCollector{},
		io:      &IOCollector{},
	}
}

// CPU returns the CPU collector.
func (p *Probe) CPU() metrics.CPUCollector {
	return p.cpu
}

// Memory returns the memory collector.
func (p *Probe) Memory() metrics.MemoryCollector {
	return p.memory
}

// Disk returns the disk collector.
func (p *Probe) Disk() metrics.DiskCollector {
	return p.disk
}

// Network returns the network collector.
func (p *Probe) Network() metrics.NetworkCollector {
	return p.network
}

// IO returns the I/O collector.
func (p *Probe) IO() metrics.IOCollector {
	return p.io
}

// CPUCollector collects CPU metrics on macOS.
// TODO: Implement using sysctl hw.ncpu, host_processor_info
type CPUCollector struct{}

// CollectSystem collects system-wide CPU metrics.
func (c *CPUCollector) CollectSystem(ctx context.Context) (metrics.SystemCPU, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.SystemCPU{}, ctx.Err()
	}
	// Not implemented on Darwin.
	return metrics.SystemCPU{}, ErrNotImplemented
}

// CollectProcess collects CPU metrics for a specific process.
func (c *CPUCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.ProcessCPU{}, ctx.Err()
	}
	// Validate PID before returning not implemented.
	if pid <= 0 {
		// Return error for invalid process ID.
		return metrics.ProcessCPU{}, ErrInvalidPID
	}
	// Not implemented on Darwin, return PID for context.
	return metrics.ProcessCPU{PID: pid}, ErrNotImplemented
}

// CollectAllProcesses collects CPU metrics for all visible processes.
func (c *CPUCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessCPU, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not implemented on Darwin.
	return nil, ErrNotImplemented
}

// CollectLoadAverage collects system load average.
func (c *CPUCollector) CollectLoadAverage(ctx context.Context) (metrics.LoadAverage, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.LoadAverage{}, ctx.Err()
	}
	// Not implemented on Darwin.
	return metrics.LoadAverage{}, ErrNotImplemented
}

// CollectPressure collects CPU pressure metrics.
// Note: PSI is Linux-specific, macOS doesn't have an equivalent.
func (c *CPUCollector) CollectPressure(ctx context.Context) (metrics.CPUPressure, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.CPUPressure{}, ctx.Err()
	}
	// Not implemented on Darwin.
	return metrics.CPUPressure{}, ErrNotImplemented
}

// MemoryCollector collects memory metrics on macOS.
// TODO: Implement using sysctl hw.memsize, vm_statistics
type MemoryCollector struct{}

// CollectSystem collects system-wide memory metrics.
func (m *MemoryCollector) CollectSystem(ctx context.Context) (metrics.SystemMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.SystemMemory{}, ctx.Err()
	}
	// Not implemented on Darwin.
	return metrics.SystemMemory{}, ErrNotImplemented
}

// CollectProcess collects memory metrics for a specific process.
func (m *MemoryCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.ProcessMemory{}, ctx.Err()
	}
	// Validate PID before returning not implemented.
	if pid <= 0 {
		// Return error for invalid process ID.
		return metrics.ProcessMemory{}, ErrInvalidPID
	}
	// Not implemented on Darwin, return PID for context.
	return metrics.ProcessMemory{PID: pid}, ErrNotImplemented
}

// CollectAllProcesses collects memory metrics for all visible processes.
func (m *MemoryCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not implemented on Darwin.
	return nil, ErrNotImplemented
}

// CollectPressure collects memory pressure metrics.
// Note: PSI is Linux-specific, macOS doesn't have an equivalent.
func (m *MemoryCollector) CollectPressure(ctx context.Context) (metrics.MemoryPressure, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.MemoryPressure{}, ctx.Err()
	}
	// Not implemented on Darwin.
	return metrics.MemoryPressure{}, ErrNotImplemented
}

// DiskCollector collects disk metrics on macOS.
// TODO: Implement using statfs, diskutil
type DiskCollector struct{}

// ListPartitions returns all mounted partitions.
func (d *DiskCollector) ListPartitions(ctx context.Context) ([]metrics.Partition, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not implemented on Darwin.
	return nil, ErrNotImplemented
}

// CollectUsage collects disk usage for a specific path.
func (d *DiskCollector) CollectUsage(ctx context.Context, path string) (metrics.DiskUsage, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.DiskUsage{}, ctx.Err()
	}
	// Validate path before returning not implemented.
	if path == "" {
		// Return error for empty path.
		return metrics.DiskUsage{}, ErrEmptyPath
	}
	// Not implemented on Darwin, return path for context.
	return metrics.DiskUsage{Path: path}, ErrNotImplemented
}

// CollectAllUsage collects disk usage for all mounted partitions.
func (d *DiskCollector) CollectAllUsage(ctx context.Context) ([]metrics.DiskUsage, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not implemented on Darwin.
	return nil, ErrNotImplemented
}

// CollectIO collects I/O statistics for all block devices.
func (d *DiskCollector) CollectIO(ctx context.Context) ([]metrics.DiskIOStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not implemented on Darwin.
	return nil, ErrNotImplemented
}

// CollectDeviceIO collects I/O statistics for a specific device.
func (d *DiskCollector) CollectDeviceIO(ctx context.Context, device string) (metrics.DiskIOStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.DiskIOStats{}, ctx.Err()
	}
	// Validate device name before returning not implemented.
	if device == "" {
		// Return error for empty device name.
		return metrics.DiskIOStats{}, ErrEmptyDevice
	}
	// Not implemented on Darwin, return device for context.
	return metrics.DiskIOStats{Device: device}, ErrNotImplemented
}

// NetworkCollector collects network metrics on macOS.
// TODO: Implement using getifaddrs, nettop
type NetworkCollector struct{}

// ListInterfaces returns all network interfaces.
func (n *NetworkCollector) ListInterfaces(ctx context.Context) ([]metrics.NetInterface, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not implemented on Darwin.
	return nil, ErrNotImplemented
}

// CollectStats collects statistics for a specific interface.
func (n *NetworkCollector) CollectStats(ctx context.Context, iface string) (metrics.NetStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.NetStats{}, ctx.Err()
	}
	// Validate interface name before returning not implemented.
	if iface == "" {
		// Return error for empty interface name.
		return metrics.NetStats{}, ErrEmptyInterface
	}
	// Not implemented on Darwin, return interface for context.
	return metrics.NetStats{Interface: iface}, ErrNotImplemented
}

// CollectAllStats collects statistics for all interfaces.
func (n *NetworkCollector) CollectAllStats(ctx context.Context) ([]metrics.NetStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not implemented on Darwin.
	return nil, ErrNotImplemented
}

// IOCollector collects I/O metrics on macOS.
type IOCollector struct{}

// CollectStats collects system-wide I/O statistics.
func (i *IOCollector) CollectStats(ctx context.Context) (metrics.IOStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.IOStats{}, ctx.Err()
	}
	// Not implemented on Darwin.
	return metrics.IOStats{}, ErrNotImplemented
}

// CollectPressure collects I/O pressure metrics.
// Note: PSI is Linux-specific, macOS doesn't have an equivalent.
func (i *IOCollector) CollectPressure(ctx context.Context) (metrics.IOPressure, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.IOPressure{}, ctx.Err()
	}
	// Not implemented on Darwin.
	return metrics.IOPressure{}, ErrNotImplemented
}
