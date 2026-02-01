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

// percentMultiplier is used to convert ratios to percentages.
const percentMultiplier float64 = 100.0

// MemoryCollector collects memory metrics via the Rust probe library.
// It implements the metrics.MemoryCollector interface.
type MemoryCollector struct{}

// NewMemoryCollector creates a new memory collector backed by the Rust probe.
//
// Returns:
//   - *MemoryCollector: new memory collector instance
func NewMemoryCollector() *MemoryCollector {
	// Return a new empty collector instance.
	return &MemoryCollector{}
}

// CollectSystem collects system-wide memory metrics.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.SystemMemory: system-wide memory statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (m *MemoryCollector) CollectSystem(ctx context.Context) (metrics.SystemMemory, error) {
	_ = ctx // reserved for future cancellation support
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.SystemMemory{}, err
	}

	var cMem C.SystemMemory
	result := C.probe_collect_memory(&cMem)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.SystemMemory{}, err
	}

	total := uint64(cMem.total_bytes)
	available := uint64(cMem.available_bytes)
	used := uint64(cMem.used_bytes)
	swapTotal := uint64(cMem.swap_total_bytes)
	swapUsed := uint64(cMem.swap_used_bytes)

	var usagePercent float64
	// Calculate usage percentage if total memory is available.
	if total > 0 {
		usagePercent = float64(used) / float64(total) * percentMultiplier
	}

	// Return collected memory metrics with current timestamp.
	return metrics.SystemMemory{
		Total:        total,
		Available:    available,
		Used:         used,
		Free:         available, // Free is approximated as Available
		Cached:       uint64(cMem.cached_bytes),
		Buffers:      uint64(cMem.buffers_bytes),
		SwapTotal:    swapTotal,
		SwapUsed:     swapUsed,
		SwapFree:     swapTotal - swapUsed,
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
		// Note: Shared memory is not available from the Rust probe.
	}, nil
}

// CollectProcess collects memory metrics for a specific process.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessMemory: memory metrics for the process
//   - error: nil on success, error if probe not initialized or collection fails
func (m *MemoryCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	_ = ctx // reserved for future cancellation support
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.ProcessMemory{}, err
	}

	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.ProcessMemory{}, err
	}

	// Return collected process memory metrics with current timestamp.
	return metrics.ProcessMemory{
		PID:          int(cProc.pid),
		RSS:          uint64(cProc.memory_rss_bytes),
		VMS:          uint64(cProc.memory_vms_bytes),
		UsagePercent: float64(cProc.memory_percent),
		Timestamp:    time.Now(),
		// Note: Shared, Swap, Data, Stack are not available cross-platform.
	}, nil
}

// CollectAllProcesses is not implemented by the Rust probe.
// Returns an error indicating the operation is not supported.
//
// Params:
//   - ctx: context for cancellation (unused)
//
// Returns:
//   - []metrics.ProcessMemory: always nil
//   - error: always ErrNotSupported
func (m *MemoryCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessMemory, error) {
	_ = ctx // reserved for future cancellation support
	// The Rust probe does not support enumerating all processes.
	return nil, ErrNotSupported
}

// CollectPressure collects memory pressure metrics (PSI).
// Note: PSI is a Linux 4.20+ feature, not available cross-platform.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.MemoryPressure: memory pressure statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (m *MemoryCollector) CollectPressure(ctx context.Context) (metrics.MemoryPressure, error) {
	_ = ctx // reserved for future cancellation support
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.MemoryPressure{}, err
	}

	var pressure C.MemoryPressure
	result := C.probe_collect_memory_pressure(&pressure)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.MemoryPressure{}, err
	}

	// Return collected memory pressure metrics with current timestamp.
	return metrics.MemoryPressure{
		SomeAvg10:  float64(pressure.some_avg10),
		SomeAvg60:  float64(pressure.some_avg60),
		SomeAvg300: float64(pressure.some_avg300),
		SomeTotal:  uint64(pressure.some_total_us),
		FullAvg10:  float64(pressure.full_avg10),
		FullAvg60:  float64(pressure.full_avg60),
		FullAvg300: float64(pressure.full_avg300),
		FullTotal:  uint64(pressure.full_total_us),
		Timestamp:  time.Now(),
	}, nil
}
