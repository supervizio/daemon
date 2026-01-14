//go:build freebsd || openbsd || netbsd || dragonfly

// Package bsd provides system metrics collection for BSD variants
// using sysctl and other BSD-specific interfaces.
//
// Supported systems: FreeBSD, OpenBSD, NetBSD, DragonFly BSD
package bsd

import (
	"context"
	"errors"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// ErrNotImplemented is returned for methods not yet implemented.
var ErrNotImplemented = errors.New("BSD metrics collector not yet implemented")

// BSDProbe implements probe.SystemCollector for BSD systems.
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
func (p *BSDProbe) CPU() probe.CPUCollector {
	return p.cpu
}

// Memory returns the memory collector.
func (p *BSDProbe) Memory() probe.MemoryCollector {
	return p.memory
}

// Disk returns the disk collector.
func (p *BSDProbe) Disk() probe.DiskCollector {
	return p.disk
}

// Network returns the network collector.
func (p *BSDProbe) Network() probe.NetworkCollector {
	return p.network
}

// IO returns the I/O collector.
func (p *BSDProbe) IO() probe.IOCollector {
	return p.io
}

// CPUCollector collects CPU metrics on BSD systems.
// TODO: Implement using sysctl hw.ncpu, kern.cp_time
type CPUCollector struct{}

// CollectSystem collects system-wide CPU metrics.
func (c *CPUCollector) CollectSystem(_ context.Context) (probe.SystemCPU, error) {
	return probe.SystemCPU{}, ErrNotImplemented
}

// CollectProcess collects CPU metrics for a specific process.
func (c *CPUCollector) CollectProcess(_ context.Context, _ int) (probe.ProcessCPU, error) {
	return probe.ProcessCPU{}, ErrNotImplemented
}

// CollectAllProcesses collects CPU metrics for all visible processes.
func (c *CPUCollector) CollectAllProcesses(_ context.Context) ([]probe.ProcessCPU, error) {
	return nil, ErrNotImplemented
}

// CollectLoadAverage collects system load average.
func (c *CPUCollector) CollectLoadAverage(_ context.Context) (probe.LoadAverage, error) {
	return probe.LoadAverage{}, ErrNotImplemented
}

// CollectPressure collects CPU pressure metrics.
// Note: PSI is Linux-specific, BSD doesn't have an equivalent.
func (c *CPUCollector) CollectPressure(_ context.Context) (probe.CPUPressure, error) {
	return probe.CPUPressure{}, ErrNotImplemented
}

// MemoryCollector collects memory metrics on BSD systems.
// TODO: Implement using sysctl hw.physmem, vm.stats
type MemoryCollector struct{}

// CollectSystem collects system-wide memory metrics.
func (m *MemoryCollector) CollectSystem(_ context.Context) (probe.SystemMemory, error) {
	return probe.SystemMemory{}, ErrNotImplemented
}

// CollectProcess collects memory metrics for a specific process.
func (m *MemoryCollector) CollectProcess(_ context.Context, _ int) (probe.ProcessMemory, error) {
	return probe.ProcessMemory{}, ErrNotImplemented
}

// CollectAllProcesses collects memory metrics for all visible processes.
func (m *MemoryCollector) CollectAllProcesses(_ context.Context) ([]probe.ProcessMemory, error) {
	return nil, ErrNotImplemented
}

// CollectPressure collects memory pressure metrics.
// Note: PSI is Linux-specific, BSD doesn't have an equivalent.
func (m *MemoryCollector) CollectPressure(_ context.Context) (probe.MemoryPressure, error) {
	return probe.MemoryPressure{}, ErrNotImplemented
}

// DiskCollector collects disk metrics on BSD systems.
// TODO: Implement using geom, sysctl
type DiskCollector struct{}

// ListPartitions returns all mounted partitions.
func (d *DiskCollector) ListPartitions(_ context.Context) ([]probe.Partition, error) {
	return nil, ErrNotImplemented
}

// CollectUsage collects disk usage for a specific path.
func (d *DiskCollector) CollectUsage(_ context.Context, _ string) (probe.DiskUsage, error) {
	return probe.DiskUsage{}, ErrNotImplemented
}

// CollectAllUsage collects disk usage for all mounted partitions.
func (d *DiskCollector) CollectAllUsage(_ context.Context) ([]probe.DiskUsage, error) {
	return nil, ErrNotImplemented
}

// CollectIO collects I/O statistics for all block devices.
func (d *DiskCollector) CollectIO(_ context.Context) ([]probe.DiskIOStats, error) {
	return nil, ErrNotImplemented
}

// CollectDeviceIO collects I/O statistics for a specific device.
func (d *DiskCollector) CollectDeviceIO(_ context.Context, _ string) (probe.DiskIOStats, error) {
	return probe.DiskIOStats{}, ErrNotImplemented
}

// NetworkCollector collects network metrics on BSD systems.
// TODO: Implement using getifaddrs, netstat
type NetworkCollector struct{}

// ListInterfaces returns all network interfaces.
func (n *NetworkCollector) ListInterfaces(_ context.Context) ([]probe.NetInterface, error) {
	return nil, ErrNotImplemented
}

// CollectStats collects statistics for a specific interface.
func (n *NetworkCollector) CollectStats(_ context.Context, _ string) (probe.NetStats, error) {
	return probe.NetStats{}, ErrNotImplemented
}

// CollectAllStats collects statistics for all interfaces.
func (n *NetworkCollector) CollectAllStats(_ context.Context) ([]probe.NetStats, error) {
	return nil, ErrNotImplemented
}

// IOCollector collects I/O metrics on BSD systems.
type IOCollector struct{}

// CollectStats collects system-wide I/O statistics.
func (i *IOCollector) CollectStats(_ context.Context) (probe.IOStats, error) {
	return probe.IOStats{}, ErrNotImplemented
}

// CollectPressure collects I/O pressure metrics.
// Note: PSI is Linux-specific, BSD doesn't have an equivalent.
func (i *IOCollector) CollectPressure(_ context.Context) (probe.IOPressure, error) {
	return probe.IOPressure{}, ErrNotImplemented
}
