//go:build freebsd || openbsd || netbsd || dragonfly

// Package bsd provides system metrics collection for BSD variants
// using sysctl and other BSD-specific interfaces.
//
// Supported systems: FreeBSD, OpenBSD, NetBSD, DragonFly BSD
package bsd

import (
	"context"
	"errors"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// ErrNotImplemented is returned for methods not yet implemented.
var ErrNotImplemented = errors.New("BSD metrics collector not yet implemented")

// BSDProbe implements metrics.SystemCollector for BSD systems.
type BSDProbe struct {
	cpu     *CPUCollector
	memory  *MemoryCollector
	disk    *DiskCollector
	network *NetworkCollector
	io      *IOCollector
}

// NewBSDProbe creates a new BSD metrics collector.
func NewBSDProbe() *BSDProbe {
	return &BSDProbe{
		cpu:     &CPUCollector{},
		memory:  &MemoryCollector{},
		disk:    &DiskCollector{},
		network: &NetworkCollector{},
		io:      &IOCollector{},
	}
}

// CPU returns the CPU collector.
func (p *BSDProbe) CPU() metrics.CPUCollector {
	return p.cpu
}

// Memory returns the memory collector.
func (p *BSDProbe) Memory() metrics.MemoryCollector {
	return p.memory
}

// Disk returns the disk collector.
func (p *BSDProbe) Disk() metrics.DiskCollector {
	return p.disk
}

// Network returns the network collector.
func (p *BSDProbe) Network() metrics.NetworkCollector {
	return p.network
}

// IO returns the I/O collector.
func (p *BSDProbe) IO() metrics.IOCollector {
	return p.io
}

// CPUCollector collects CPU metrics on BSD systems.
// TODO: Implement using sysctl hw.ncpu, kern.cp_time
type CPUCollector struct{}

// CollectSystem collects system-wide CPU metrics.
func (c *CPUCollector) CollectSystem(_ context.Context) (metrics.SystemCPU, error) {
	return metrics.SystemCPU{}, ErrNotImplemented
}

// CollectProcess collects CPU metrics for a specific process.
func (c *CPUCollector) CollectProcess(_ context.Context, _ int) (metrics.ProcessCPU, error) {
	return metrics.ProcessCPU{}, ErrNotImplemented
}

// CollectAllProcesses collects CPU metrics for all visible processes.
func (c *CPUCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessCPU, error) {
	return nil, ErrNotImplemented
}

// CollectLoadAverage collects system load average.
func (c *CPUCollector) CollectLoadAverage(_ context.Context) (metrics.LoadAverage, error) {
	return metrics.LoadAverage{}, ErrNotImplemented
}

// CollectPressure collects CPU pressure metrics.
// Note: PSI is Linux-specific, BSD doesn't have an equivalent.
func (c *CPUCollector) CollectPressure(_ context.Context) (metrics.CPUPressure, error) {
	return metrics.CPUPressure{}, ErrNotImplemented
}

// MemoryCollector collects memory metrics on BSD systems.
// TODO: Implement using sysctl hw.physmem, vm.stats
type MemoryCollector struct{}

// CollectSystem collects system-wide memory metrics.
func (m *MemoryCollector) CollectSystem(_ context.Context) (metrics.SystemMemory, error) {
	return metrics.SystemMemory{}, ErrNotImplemented
}

// CollectProcess collects memory metrics for a specific process.
func (m *MemoryCollector) CollectProcess(_ context.Context, _ int) (metrics.ProcessMemory, error) {
	return metrics.ProcessMemory{}, ErrNotImplemented
}

// CollectAllProcesses collects memory metrics for all visible processes.
func (m *MemoryCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessMemory, error) {
	return nil, ErrNotImplemented
}

// CollectPressure collects memory pressure metrics.
// Note: PSI is Linux-specific, BSD doesn't have an equivalent.
func (m *MemoryCollector) CollectPressure(_ context.Context) (metrics.MemoryPressure, error) {
	return metrics.MemoryPressure{}, ErrNotImplemented
}

// DiskCollector collects disk metrics on BSD systems.
// TODO: Implement using geom, sysctl
type DiskCollector struct{}

// ListPartitions returns all mounted partitions.
func (d *DiskCollector) ListPartitions(_ context.Context) ([]metrics.Partition, error) {
	return nil, ErrNotImplemented
}

// CollectUsage collects disk usage for a specific path.
func (d *DiskCollector) CollectUsage(_ context.Context, _ string) (metrics.DiskUsage, error) {
	return metrics.DiskUsage{}, ErrNotImplemented
}

// CollectAllUsage collects disk usage for all mounted partitions.
func (d *DiskCollector) CollectAllUsage(_ context.Context) ([]metrics.DiskUsage, error) {
	return nil, ErrNotImplemented
}

// CollectIO collects I/O statistics for all block devices.
func (d *DiskCollector) CollectIO(_ context.Context) ([]metrics.DiskIOStats, error) {
	return nil, ErrNotImplemented
}

// CollectDeviceIO collects I/O statistics for a specific device.
func (d *DiskCollector) CollectDeviceIO(_ context.Context, _ string) (metrics.DiskIOStats, error) {
	return metrics.DiskIOStats{}, ErrNotImplemented
}

// NetworkCollector collects network metrics on BSD systems.
// TODO: Implement using getifaddrs, netstat
type NetworkCollector struct{}

// ListInterfaces returns all network interfaces.
func (n *NetworkCollector) ListInterfaces(_ context.Context) ([]metrics.NetInterface, error) {
	return nil, ErrNotImplemented
}

// CollectStats collects statistics for a specific interface.
func (n *NetworkCollector) CollectStats(_ context.Context, _ string) (metrics.NetStats, error) {
	return metrics.NetStats{}, ErrNotImplemented
}

// CollectAllStats collects statistics for all interfaces.
func (n *NetworkCollector) CollectAllStats(_ context.Context) ([]metrics.NetStats, error) {
	return nil, ErrNotImplemented
}

// IOCollector collects I/O metrics on BSD systems.
type IOCollector struct{}

// CollectStats collects system-wide I/O statistics.
func (i *IOCollector) CollectStats(_ context.Context) (metrics.IOStats, error) {
	return metrics.IOStats{}, ErrNotImplemented
}

// CollectPressure collects I/O pressure metrics.
// Note: PSI is Linux-specific, BSD doesn't have an equivalent.
func (i *IOCollector) CollectPressure(_ context.Context) (metrics.IOPressure, error) {
	return metrics.IOPressure{}, ErrNotImplemented
}
