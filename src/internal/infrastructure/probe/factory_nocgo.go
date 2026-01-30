//go:build !cgo

package probe

import (
	"context"
	"fmt"

	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	"github.com/kodflow/daemon/internal/domain/metrics"
)

var errNoCGO = fmt.Errorf("probe requires CGO: rebuild with CGO_ENABLED=1")

// NewSystemCollector returns an error collector when CGO is disabled.
// The Rust probe requires CGO for FFI bindings.
//
// Returns:
//   - metrics.SystemCollector: a stub that returns errors
func NewSystemCollector() metrics.SystemCollector {
	return &noCGOCollector{}
}

// NewAppProcessCollector returns an error collector when CGO is disabled.
// The Rust probe requires CGO for FFI bindings.
//
// Returns:
//   - appmetrics.Collector: a stub that returns errors
func NewAppProcessCollector() appmetrics.Collector {
	return &noCGOProcessCollector{}
}

// DetectedPlatform returns "nocgo" when CGO is disabled.
//
// Returns:
//   - string: "nocgo"
func DetectedPlatform() string {
	return "nocgo"
}

// noCGOCollector is a stub SystemCollector for when CGO is not available.
type noCGOCollector struct{}

// CPU returns an error CPU collector.
func (c *noCGOCollector) CPU() metrics.CPUCollector {
	return &noCGOCPUCollector{}
}

// Memory returns an error memory collector.
func (c *noCGOCollector) Memory() metrics.MemoryCollector {
	return &noCGOMemoryCollector{}
}

// Disk returns an error disk collector.
func (c *noCGOCollector) Disk() metrics.DiskCollector {
	return &noCGODiskCollector{}
}

// Network returns an error network collector.
func (c *noCGOCollector) Network() metrics.NetworkCollector {
	return &noCGONetworkCollector{}
}

// IO returns an error I/O collector.
func (c *noCGOCollector) IO() metrics.IOCollector {
	return &noCGOIOCollector{}
}

// noCGOCPUCollector is a stub CPU collector.
type noCGOCPUCollector struct{}

// CollectSystem returns an error indicating CGO is required.
func (c *noCGOCPUCollector) CollectSystem(_ context.Context) (metrics.SystemCPU, error) {
	return metrics.SystemCPU{}, errNoCGO
}

// CollectProcess returns an error indicating CGO is required.
func (c *noCGOCPUCollector) CollectProcess(_ context.Context, _ int) (metrics.ProcessCPU, error) {
	return metrics.ProcessCPU{}, errNoCGO
}

// CollectAllProcesses returns an error indicating CGO is required.
func (c *noCGOCPUCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessCPU, error) {
	return nil, errNoCGO
}

// CollectLoadAverage returns an error indicating CGO is required.
func (c *noCGOCPUCollector) CollectLoadAverage(_ context.Context) (metrics.LoadAverage, error) {
	return metrics.LoadAverage{}, errNoCGO
}

// CollectPressure returns an error indicating CGO is required.
func (c *noCGOCPUCollector) CollectPressure(_ context.Context) (metrics.CPUPressure, error) {
	return metrics.CPUPressure{}, errNoCGO
}

// noCGOMemoryCollector is a stub memory collector.
type noCGOMemoryCollector struct{}

// CollectSystem returns an error indicating CGO is required.
func (c *noCGOMemoryCollector) CollectSystem(_ context.Context) (metrics.SystemMemory, error) {
	return metrics.SystemMemory{}, errNoCGO
}

// CollectProcess returns an error indicating CGO is required.
func (c *noCGOMemoryCollector) CollectProcess(_ context.Context, _ int) (metrics.ProcessMemory, error) {
	return metrics.ProcessMemory{}, errNoCGO
}

// CollectAllProcesses returns an error indicating CGO is required.
func (c *noCGOMemoryCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessMemory, error) {
	return nil, errNoCGO
}

// CollectPressure returns an error indicating CGO is required.
func (c *noCGOMemoryCollector) CollectPressure(_ context.Context) (metrics.MemoryPressure, error) {
	return metrics.MemoryPressure{}, errNoCGO
}

// noCGODiskCollector is a stub disk collector.
type noCGODiskCollector struct{}

// ListPartitions returns an error indicating CGO is required.
func (c *noCGODiskCollector) ListPartitions(_ context.Context) ([]metrics.Partition, error) {
	return nil, errNoCGO
}

// CollectUsage returns an error indicating CGO is required.
func (c *noCGODiskCollector) CollectUsage(_ context.Context, _ string) (metrics.DiskUsage, error) {
	return metrics.DiskUsage{}, errNoCGO
}

// CollectAllUsage returns an error indicating CGO is required.
func (c *noCGODiskCollector) CollectAllUsage(_ context.Context) ([]metrics.DiskUsage, error) {
	return nil, errNoCGO
}

// CollectIO returns an error indicating CGO is required.
func (c *noCGODiskCollector) CollectIO(_ context.Context) ([]metrics.DiskIOStats, error) {
	return nil, errNoCGO
}

// CollectDeviceIO returns an error indicating CGO is required.
func (c *noCGODiskCollector) CollectDeviceIO(_ context.Context, _ string) (metrics.DiskIOStats, error) {
	return metrics.DiskIOStats{}, errNoCGO
}

// noCGONetworkCollector is a stub network collector.
type noCGONetworkCollector struct{}

// ListInterfaces returns an error indicating CGO is required.
func (c *noCGONetworkCollector) ListInterfaces(_ context.Context) ([]metrics.NetInterface, error) {
	return nil, errNoCGO
}

// CollectStats returns an error indicating CGO is required.
func (c *noCGONetworkCollector) CollectStats(_ context.Context, _ string) (metrics.NetStats, error) {
	return metrics.NetStats{}, errNoCGO
}

// CollectAllStats returns an error indicating CGO is required.
func (c *noCGONetworkCollector) CollectAllStats(_ context.Context) ([]metrics.NetStats, error) {
	return nil, errNoCGO
}

// noCGOIOCollector is a stub I/O collector.
type noCGOIOCollector struct{}

// CollectStats returns an error indicating CGO is required.
func (c *noCGOIOCollector) CollectStats(_ context.Context) (metrics.IOStats, error) {
	return metrics.IOStats{}, errNoCGO
}

// CollectPressure returns an error indicating CGO is required.
func (c *noCGOIOCollector) CollectPressure(_ context.Context) (metrics.IOPressure, error) {
	return metrics.IOPressure{}, errNoCGO
}

// noCGOProcessCollector is a stub process collector.
type noCGOProcessCollector struct{}

// CollectCPU returns an error indicating CGO is required.
func (c *noCGOProcessCollector) CollectCPU(_ context.Context, _ int) (metrics.ProcessCPU, error) {
	return metrics.ProcessCPU{}, errNoCGO
}

// CollectMemory returns an error indicating CGO is required.
func (c *noCGOProcessCollector) CollectMemory(_ context.Context, _ int) (metrics.ProcessMemory, error) {
	return metrics.ProcessMemory{}, errNoCGO
}
