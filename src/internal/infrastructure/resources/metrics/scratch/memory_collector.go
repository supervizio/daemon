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
//   - _: context (unused in scratch mode)
//
// Returns:
//   - metrics.SystemMemory: zero value
//   - error: ErrNotSupported
func (m *MemoryCollector) CollectSystem(_ context.Context) (metrics.SystemMemory, error) {
	// Not supported in scratch mode
	return metrics.SystemMemory{}, ErrNotSupported
}

// CollectProcess returns an error as process memory metrics are not available.
//
// Params:
//   - _: context (unused in scratch mode)
//   - _: process ID (unused in scratch mode)
//
// Returns:
//   - metrics.ProcessMemory: zero value
//   - error: ErrNotSupported
func (m *MemoryCollector) CollectProcess(_ context.Context, _ int) (metrics.ProcessMemory, error) {
	// Not supported in scratch mode
	return metrics.ProcessMemory{}, ErrNotSupported
}

// CollectAllProcesses returns an error as process enumeration is not available.
//
// Params:
//   - _: context (unused in scratch mode)
//
// Returns:
//   - []metrics.ProcessMemory: nil slice
//   - error: ErrNotSupported
func (m *MemoryCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessMemory, error) {
	// Not supported in scratch mode
	return nil, ErrNotSupported
}

// CollectPressure returns an error as memory pressure metrics are not available.
//
// Params:
//   - _: context (unused in scratch mode)
//
// Returns:
//   - metrics.MemoryPressure: zero value
//   - error: ErrNotSupported
func (m *MemoryCollector) CollectPressure(_ context.Context) (metrics.MemoryPressure, error) {
	// Not supported in scratch mode
	return metrics.MemoryPressure{}, ErrNotSupported
}
