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
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.Partition: nil slice
//   - error: ErrNotSupported or context error
func (d *DiskCollector) ListPartitions(ctx context.Context) ([]metrics.Partition, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not supported in scratch mode.
	return nil, ErrNotSupported
}

// CollectUsage returns an error as disk usage metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//   - path: mount path to collect usage for
//
// Returns:
//   - metrics.DiskUsage: zero value with Path set for context
//   - error: ErrNotSupported, ErrEmptyPath, or context error
func (d *DiskCollector) CollectUsage(ctx context.Context, path string) (metrics.DiskUsage, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.DiskUsage{}, ctx.Err()
	}
	// Validate path before returning not supported.
	if path == "" {
		// Return error for empty mount path.
		return metrics.DiskUsage{}, ErrEmptyPath
	}
	// Not supported in scratch mode, return path for context.
	return metrics.DiskUsage{Path: path}, ErrNotSupported
}

// CollectAllUsage returns an error as disk enumeration is not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.DiskUsage: nil slice
//   - error: ErrNotSupported or context error
func (d *DiskCollector) CollectAllUsage(ctx context.Context) ([]metrics.DiskUsage, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not supported in scratch mode.
	return nil, ErrNotSupported
}

// CollectIO returns an error as disk I/O metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.DiskIOStats: nil slice
//   - error: ErrNotSupported or context error
func (d *DiskCollector) CollectIO(ctx context.Context) ([]metrics.DiskIOStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not supported in scratch mode.
	return nil, ErrNotSupported
}

// CollectDeviceIO returns an error as disk I/O metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//   - device: device name to collect I/O for
//
// Returns:
//   - metrics.DiskIOStats: zero value with Device set for context
//   - error: ErrNotSupported, ErrEmptyDevice, or context error
func (d *DiskCollector) CollectDeviceIO(ctx context.Context, device string) (metrics.DiskIOStats, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.DiskIOStats{}, ctx.Err()
	}
	// Validate device name before returning not supported.
	if device == "" {
		// Return error for empty device name.
		return metrics.DiskIOStats{}, ErrEmptyDevice
	}
	// Not supported in scratch mode, return device for context.
	return metrics.DiskIOStats{Device: device}, ErrNotSupported
}
