//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

import (
	"context"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// CachedMemoryCollector provides memory metrics collection with caching support.
// It implements the metrics.MemoryCollector interface.
type CachedMemoryCollector struct{}

// NewCachedMemoryCollector creates a new cached memory collector.
//
// Returns:
//   - *CachedMemoryCollector: new cached memory collector instance
func NewCachedMemoryCollector() *CachedMemoryCollector {
	// Return a new empty collector instance.
	return &CachedMemoryCollector{}
}

// CollectSystem collects system-wide memory metrics with caching.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.SystemMemory: cached system-wide memory statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CachedMemoryCollector) CollectSystem(_ context.Context) (metrics.SystemMemory, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.SystemMemory{}, err
	}

	var cMem C.SystemMemory
	result := C.probe_collect_memory_cached(&cMem)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.SystemMemory{}, err
	}

	// Return cached memory metrics with current timestamp.
	return metrics.SystemMemory{
		Total:     uint64(cMem.total_bytes),
		Available: uint64(cMem.available_bytes),
		Used:      uint64(cMem.used_bytes),
		Cached:    uint64(cMem.cached_bytes),
		Buffers:   uint64(cMem.buffers_bytes),
		SwapTotal: uint64(cMem.swap_total_bytes),
		SwapUsed:  uint64(cMem.swap_used_bytes),
		SwapFree:  uint64(cMem.swap_total_bytes) - uint64(cMem.swap_used_bytes),
		Timestamp: time.Now(),
	}, nil
}

// CollectProcess collects memory metrics for a specific process.
// Process metrics are not cached.
//
// Params:
//   - ctx: context for cancellation
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessMemory: memory metrics for the process
//   - error: nil on success, error if collection fails
func (c *CachedMemoryCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	// Delegate to regular collector as process metrics are not cached.
	return (&MemoryCollector{}).CollectProcess(ctx, pid)
}

// CollectAllProcesses is not implemented.
//
// Params:
//   - ctx: context for cancellation (unused)
//
// Returns:
//   - []metrics.ProcessMemory: always nil
//   - error: always ErrNotSupported
func (c *CachedMemoryCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessMemory, error) {
	// Return not supported error.
	return nil, ErrNotSupported
}

// CollectPressure collects memory pressure metrics.
// Pressure metrics are not cached.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.MemoryPressure: memory pressure statistics
//   - error: nil on success, error if collection fails
func (c *CachedMemoryCollector) CollectPressure(ctx context.Context) (metrics.MemoryPressure, error) {
	// Delegate to regular collector as pressure is not cached.
	return (&MemoryCollector{}).CollectPressure(ctx)
}
