// Package scratch provides a minimal metrics collector for environments
// without /proc filesystem access (e.g., scratch containers, Windows).
// It returns zero values or errors for most operations.
package scratch

import (
	"context"
	"errors"

	"github.com/kodflow/daemon/internal/domain/probe"
)

// ErrNotSupported is returned when a metric is not available in scratch mode.
var ErrNotSupported = errors.New("metric collection not supported in scratch mode")

// ScratchProbe implements probe.SystemCollector with minimal functionality.
// It's designed for environments where system metrics are unavailable.
type ScratchProbe struct {
	cpu     *CPUCollector
	memory  *MemoryCollector
	disk    *DiskCollector
	network *NetworkCollector
	io      *IOCollector
}

// NewScratchProbe creates a new minimal probe for scratch environments.
func NewScratchProbe() *ScratchProbe {
	return &ScratchProbe{
		cpu:     &CPUCollector{},
		memory:  &MemoryCollector{},
		disk:    &DiskCollector{},
		network: &NetworkCollector{},
		io:      &IOCollector{},
	}
}

// CPU returns the CPU collector.
func (p *ScratchProbe) CPU() probe.CPUCollector {
	return p.cpu
}

// Memory returns the memory collector.
func (p *ScratchProbe) Memory() probe.MemoryCollector {
	return p.memory
}

// Disk returns the disk collector.
func (p *ScratchProbe) Disk() probe.DiskCollector {
	return p.disk
}

// Network returns the network collector.
func (p *ScratchProbe) Network() probe.NetworkCollector {
	return p.network
}

// IO returns the I/O collector.
func (p *ScratchProbe) IO() probe.IOCollector {
	return p.io
}

// CPUCollector provides minimal CPU metrics collection.
type CPUCollector struct{}

// CollectSystem returns an error as system CPU metrics are not available.
func (c *CPUCollector) CollectSystem(_ context.Context) (probe.SystemCPU, error) {
	return probe.SystemCPU{}, ErrNotSupported
}

// CollectProcess returns an error as process CPU metrics are not available.
func (c *CPUCollector) CollectProcess(_ context.Context, _ int) (probe.ProcessCPU, error) {
	return probe.ProcessCPU{}, ErrNotSupported
}

// CollectAllProcesses returns an error as process enumeration is not available.
func (c *CPUCollector) CollectAllProcesses(_ context.Context) ([]probe.ProcessCPU, error) {
	return nil, ErrNotSupported
}

// CollectLoadAverage returns an error as load average is not available.
func (c *CPUCollector) CollectLoadAverage(_ context.Context) (probe.LoadAverage, error) {
	return probe.LoadAverage{}, ErrNotSupported
}

// CollectPressure returns an error as CPU pressure metrics are not available.
func (c *CPUCollector) CollectPressure(_ context.Context) (probe.CPUPressure, error) {
	return probe.CPUPressure{}, ErrNotSupported
}

// MemoryCollector provides minimal memory metrics collection.
type MemoryCollector struct{}

// CollectSystem returns an error as system memory metrics are not available.
func (m *MemoryCollector) CollectSystem(_ context.Context) (probe.SystemMemory, error) {
	return probe.SystemMemory{}, ErrNotSupported
}

// CollectProcess returns an error as process memory metrics are not available.
func (m *MemoryCollector) CollectProcess(_ context.Context, _ int) (probe.ProcessMemory, error) {
	return probe.ProcessMemory{}, ErrNotSupported
}

// CollectAllProcesses returns an error as process enumeration is not available.
func (m *MemoryCollector) CollectAllProcesses(_ context.Context) ([]probe.ProcessMemory, error) {
	return nil, ErrNotSupported
}

// CollectPressure returns an error as memory pressure metrics are not available.
func (m *MemoryCollector) CollectPressure(_ context.Context) (probe.MemoryPressure, error) {
	return probe.MemoryPressure{}, ErrNotSupported
}

// DiskCollector provides minimal disk metrics collection.
type DiskCollector struct{}

// ListPartitions returns an empty list as partition enumeration is not available.
func (d *DiskCollector) ListPartitions(_ context.Context) ([]probe.Partition, error) {
	return nil, ErrNotSupported
}

// CollectUsage returns an error as disk usage metrics are not available.
func (d *DiskCollector) CollectUsage(_ context.Context, _ string) (probe.DiskUsage, error) {
	return probe.DiskUsage{}, ErrNotSupported
}

// CollectAllUsage returns an error as disk enumeration is not available.
func (d *DiskCollector) CollectAllUsage(_ context.Context) ([]probe.DiskUsage, error) {
	return nil, ErrNotSupported
}

// CollectIO returns an error as disk I/O metrics are not available.
func (d *DiskCollector) CollectIO(_ context.Context) ([]probe.DiskIOStats, error) {
	return nil, ErrNotSupported
}

// CollectDeviceIO returns an error as disk I/O metrics are not available.
func (d *DiskCollector) CollectDeviceIO(_ context.Context, _ string) (probe.DiskIOStats, error) {
	return probe.DiskIOStats{}, ErrNotSupported
}

// NetworkCollector provides minimal network metrics collection.
type NetworkCollector struct{}

// ListInterfaces returns an error as interface enumeration is not available.
func (n *NetworkCollector) ListInterfaces(_ context.Context) ([]probe.NetInterface, error) {
	return nil, ErrNotSupported
}

// CollectStats returns an error as interface statistics are not available.
func (n *NetworkCollector) CollectStats(_ context.Context, _ string) (probe.NetStats, error) {
	return probe.NetStats{}, ErrNotSupported
}

// CollectAllStats returns an error as interface enumeration is not available.
func (n *NetworkCollector) CollectAllStats(_ context.Context) ([]probe.NetStats, error) {
	return nil, ErrNotSupported
}

// IOCollector provides minimal I/O metrics collection.
type IOCollector struct{}

// CollectStats returns an error as I/O statistics are not available.
func (i *IOCollector) CollectStats(_ context.Context) (probe.IOStats, error) {
	return probe.IOStats{}, ErrNotSupported
}

// CollectPressure returns an error as I/O pressure metrics are not available.
func (i *IOCollector) CollectPressure(_ context.Context) (probe.IOPressure, error) {
	return probe.IOPressure{}, ErrNotSupported
}
