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
//   - ctx: context (unused in scratch mode)
//
// Returns:
//   - metrics.SystemCPU: zero value
//   - error: ErrNotSupported
func (c *CPUCollector) CollectSystem(_ context.Context) (metrics.SystemCPU, error) {
	// Not supported in scratch mode
	return metrics.SystemCPU{}, ErrNotSupported
}

// CollectProcess returns an error as process CPU metrics are not available.
//
// Params:
//   - ctx: context (unused in scratch mode)
//   - pid: process ID (unused in scratch mode)
//
// Returns:
//   - metrics.ProcessCPU: zero value
//   - error: ErrNotSupported
func (c *CPUCollector) CollectProcess(_ context.Context, _ int) (metrics.ProcessCPU, error) {
	// Not supported in scratch mode
	return metrics.ProcessCPU{}, ErrNotSupported
}

// CollectAllProcesses returns an error as process enumeration is not available.
//
// Params:
//   - ctx: context (unused in scratch mode)
//
// Returns:
//   - []metrics.ProcessCPU: nil slice
//   - error: ErrNotSupported
func (c *CPUCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessCPU, error) {
	// Not supported in scratch mode
	return nil, ErrNotSupported
}

// CollectLoadAverage returns an error as load average is not available.
//
// Params:
//   - ctx: context (unused in scratch mode)
//
// Returns:
//   - metrics.LoadAverage: zero value
//   - error: ErrNotSupported
func (c *CPUCollector) CollectLoadAverage(_ context.Context) (metrics.LoadAverage, error) {
	// Not supported in scratch mode
	return metrics.LoadAverage{}, ErrNotSupported
}

// CollectPressure returns an error as CPU pressure metrics are not available.
//
// Params:
//   - ctx: context (unused in scratch mode)
//
// Returns:
//   - metrics.CPUPressure: zero value
//   - error: ErrNotSupported
func (c *CPUCollector) CollectPressure(_ context.Context) (metrics.CPUPressure, error) {
	// Not supported in scratch mode
	return metrics.CPUPressure{}, ErrNotSupported
}
