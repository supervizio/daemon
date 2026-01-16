// Package scratch provides a minimal metrics collector for environments
// without /proc filesystem access (e.g., scratch containers, Windows).
package scratch

import (
	"context"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// CPUCollector provides minimal CPU metrics collection.
// All methods return ErrNotSupported as CPU metrics are unavailable
// in scratch environments without /proc filesystem access.
type CPUCollector struct{}

// NewCPUCollector creates a new CPU collector for scratch environments.
//
// Returns:
//   - *CPUCollector: initialized CPU collector
func NewCPUCollector() *CPUCollector {
	// Return zero-value CPU collector
	return &CPUCollector{}
}

// CollectSystem returns an error as system CPU metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.SystemCPU: zero value
//   - error: ErrNotSupported or context error
func (c *CPUCollector) CollectSystem(ctx context.Context) (metrics.SystemCPU, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.SystemCPU{}, ctx.Err()
	}
	// Not supported in scratch mode.
	return metrics.SystemCPU{}, ErrNotSupported
}

// CollectProcess returns an error as process CPU metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//   - _: process ID (unused in scratch mode)
//
// Returns:
//   - metrics.ProcessCPU: zero value
//   - error: ErrNotSupported or context error
func (c *CPUCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.ProcessCPU{}, ctx.Err()
	}
	// Validate PID before returning not supported.
	if pid <= 0 {
		// Return error for invalid process ID.
		return metrics.ProcessCPU{}, ErrInvalidPID
	}
	// Not supported in scratch mode.
	return metrics.ProcessCPU{}, ErrNotSupported
}

// CollectAllProcesses returns an error as process enumeration is not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.ProcessCPU: nil slice
//   - error: ErrNotSupported or context error
func (c *CPUCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessCPU, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not supported in scratch mode.
	return nil, ErrNotSupported
}

// CollectLoadAverage returns an error as load average is not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.LoadAverage: zero value
//   - error: ErrNotSupported or context error
func (c *CPUCollector) CollectLoadAverage(ctx context.Context) (metrics.LoadAverage, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.LoadAverage{}, ctx.Err()
	}
	// Not supported in scratch mode.
	return metrics.LoadAverage{}, ErrNotSupported
}

// CollectPressure returns an error as CPU pressure metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.CPUPressure: zero value
//   - error: ErrNotSupported or context error
func (c *CPUCollector) CollectPressure(ctx context.Context) (metrics.CPUPressure, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.CPUPressure{}, ctx.Err()
	}
	// Not supported in scratch mode.
	return metrics.CPUPressure{}, ErrNotSupported
}
