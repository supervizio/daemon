// Package scratch provides a minimal metrics collector for environments
// without /proc filesystem access (e.g., scratch containers, Windows).
package scratch

import (
	"context"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// DiskCollector provides minimal disk metrics collection.
// All methods return ErrNotSupported as disk metrics are unavailable
// in scratch environments without filesystem enumeration capabilities.
type DiskCollector struct{}

// NewDiskCollector creates a new disk collector for scratch environments.
//
// Returns:
//   - *DiskCollector: initialized disk collector
func NewDiskCollector() *DiskCollector {
	// Return zero-value disk collector
	return &DiskCollector{}
}

// ListPartitions returns an empty list as partition enumeration is not available.
//
// Params:
//   - _: context (unused in scratch mode)
//
// Returns:
//   - []metrics.Partition: nil slice
//   - error: ErrNotSupported
func (d *DiskCollector) ListPartitions(_ context.Context) ([]metrics.Partition, error) {
	// Not supported in scratch mode
	return nil, ErrNotSupported
}

// CollectUsage returns an error as disk usage metrics are not available.
//
// Params:
//   - _: context (unused in scratch mode)
//   - _: mount path (unused in scratch mode)
//
// Returns:
//   - metrics.DiskUsage: zero value
//   - error: ErrNotSupported
func (d *DiskCollector) CollectUsage(_ context.Context, _ string) (metrics.DiskUsage, error) {
	// Not supported in scratch mode
	return metrics.DiskUsage{}, ErrNotSupported
}

// CollectAllUsage returns an error as disk enumeration is not available.
//
// Params:
//   - _: context (unused in scratch mode)
//
// Returns:
//   - []metrics.DiskUsage: nil slice
//   - error: ErrNotSupported
func (d *DiskCollector) CollectAllUsage(_ context.Context) ([]metrics.DiskUsage, error) {
	// Not supported in scratch mode
	return nil, ErrNotSupported
}

// CollectIO returns an error as disk I/O metrics are not available.
//
// Params:
//   - _: context (unused in scratch mode)
//
// Returns:
//   - []metrics.DiskIOStats: nil slice
//   - error: ErrNotSupported
func (d *DiskCollector) CollectIO(_ context.Context) ([]metrics.DiskIOStats, error) {
	// Not supported in scratch mode
	return nil, ErrNotSupported
}

// CollectDeviceIO returns an error as disk I/O metrics are not available.
//
// Params:
//   - _: context (unused in scratch mode)
//   - _: device name (unused in scratch mode)
//
// Returns:
//   - metrics.DiskIOStats: zero value
//   - error: ErrNotSupported
func (d *DiskCollector) CollectDeviceIO(_ context.Context, _ string) (metrics.DiskIOStats, error) {
	// Not supported in scratch mode
	return metrics.DiskIOStats{}, ErrNotSupported
}
