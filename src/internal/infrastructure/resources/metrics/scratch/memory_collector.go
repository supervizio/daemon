// Package scratch provides a minimal metrics collector for environments
// without /proc filesystem access (e.g., scratch containers, Windows).
package scratch

import (
	"context"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// MemoryCollector provides minimal memory metrics collection.
// All methods return ErrNotSupported as memory metrics are unavailable
// in scratch environments without /proc filesystem access.
type MemoryCollector struct{}

// NewMemoryCollector creates a new memory collector for scratch environments.
//
// Returns:
//   - *MemoryCollector: initialized memory collector
func NewMemoryCollector() *MemoryCollector {
	// Return zero-value memory collector
	return &MemoryCollector{}
}

// CollectSystem returns an error as system memory metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.SystemMemory: zero value
//   - error: ErrNotSupported or context error
func (m *MemoryCollector) CollectSystem(ctx context.Context) (metrics.SystemMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.SystemMemory{}, ctx.Err()
	}
	// Not supported in scratch mode.
	return metrics.SystemMemory{}, ErrNotSupported
}

// CollectProcess returns an error as process memory metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//   - _: process ID (unused in scratch mode)
//
// Returns:
//   - metrics.ProcessMemory: zero value
//   - error: ErrNotSupported or context error
func (m *MemoryCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.ProcessMemory{}, ctx.Err()
	}
	// Validate PID before returning not supported.
	if pid <= 0 {
		// Return error for invalid process ID.
		return metrics.ProcessMemory{}, ErrInvalidPID
	}
	// Not supported in scratch mode.
	return metrics.ProcessMemory{}, ErrNotSupported
}

// CollectAllProcesses returns an error as process enumeration is not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.ProcessMemory: nil slice
//   - error: ErrNotSupported or context error
func (m *MemoryCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessMemory, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return nil, ctx.Err()
	}
	// Not supported in scratch mode.
	return nil, ErrNotSupported
}

// CollectPressure returns an error as memory pressure metrics are not available.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.MemoryPressure: zero value
//   - error: ErrNotSupported or context error
func (m *MemoryCollector) CollectPressure(ctx context.Context) (metrics.MemoryPressure, error) {
	// Respect context cancellation.
	if ctx.Err() != nil {
		// Return context error when cancelled.
		return metrics.MemoryPressure{}, ctx.Err()
	}
	// Not supported in scratch mode.
	return metrics.MemoryPressure{}, ErrNotSupported
}
