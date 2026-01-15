// Package scratch provides a minimal metrics collector for environments
// without /proc filesystem access (e.g., scratch containers, Windows).
package scratch

import (
	"context"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// IOCollector provides minimal I/O metrics collection.
// All methods return ErrNotSupported as I/O metrics are unavailable
// in scratch environments without /proc/pressure access.
type IOCollector struct{}

// NewIOCollector creates a new I/O collector for scratch environments.
//
// Returns:
//   - *IOCollector: initialized I/O collector
func NewIOCollector() *IOCollector {
	// Return zero-value I/O collector
	return &IOCollector{}
}

// CollectStats returns an error as I/O statistics are not available.
//
// Params:
//   - _: context (unused in scratch mode)
//
// Returns:
//   - metrics.IOStats: zero value
//   - error: ErrNotSupported
func (i *IOCollector) CollectStats(_ context.Context) (metrics.IOStats, error) {
	// Not supported in scratch mode
	return metrics.IOStats{}, ErrNotSupported
}

// CollectPressure returns an error as I/O pressure metrics are not available.
//
// Params:
//   - _: context (unused in scratch mode)
//
// Returns:
//   - metrics.IOPressure: zero value
//   - error: ErrNotSupported
func (i *IOCollector) CollectPressure(_ context.Context) (metrics.IOPressure, error) {
	// Not supported in scratch mode
	return metrics.IOPressure{}, ErrNotSupported
}
