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

// ProcessCollector collects per-process metrics via the Rust probe library.
// It implements appmetrics.Collector interface.
type ProcessCollector struct{}

// NewProcessCollector creates a new process metrics collector.
//
// Returns:
//   - *ProcessCollector: new collector instance
func NewProcessCollector() *ProcessCollector {
	// Return a new empty collector instance.
	return &ProcessCollector{}
}

// CollectCPU collects CPU metrics for a specific process.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessCPU: CPU metrics for the process
//   - error: nil on success, error if probe not initialized or collection fails
func (c *ProcessCollector) CollectCPU(_ context.Context, pid int) (metrics.ProcessCPU, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.ProcessCPU{}, err
	}

	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.ProcessCPU{}, err
	}

	// Return collected CPU metrics with current timestamp.
	return metrics.ProcessCPU{
		PID:          int(cProc.pid),
		UsagePercent: float64(cProc.cpu_percent),
		Timestamp:    time.Now(),
		// Note: Jiffies (User, System, etc.) not available cross-platform.
		// The UsagePercent is calculated by the Rust probe based on delta.
	}, nil
}

// CollectMemory collects memory metrics for a specific process.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessMemory: memory metrics for the process
//   - error: nil on success, error if probe not initialized or collection fails
func (c *ProcessCollector) CollectMemory(_ context.Context, pid int) (metrics.ProcessMemory, error) {
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

	// Return collected memory metrics with current timestamp.
	return metrics.ProcessMemory{
		PID:          int(cProc.pid),
		RSS:          uint64(cProc.memory_rss_bytes),
		VMS:          uint64(cProc.memory_vms_bytes),
		UsagePercent: float64(cProc.memory_percent),
		Timestamp:    time.Now(),
		// Note: Shared, Swap, Data, Stack not available cross-platform.
	}, nil
}

// CollectFDs collects file descriptor count for a specific process.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//   - pid: process ID to collect metrics for
//
// Returns:
//   - ProcessFDs: file descriptor metrics for the process
//   - error: nil on success, error if probe not initialized or collection fails
func (c *ProcessCollector) CollectFDs(_ context.Context, pid int) (ProcessFDs, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return ProcessFDs{}, err
	}

	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return ProcessFDs{}, err
	}

	// Return collected file descriptor count.
	return ProcessFDs{
		PID:   int(cProc.pid),
		Count: uint32(cProc.num_fds),
	}, nil
}

// CollectIO collects I/O statistics for a specific process.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//   - pid: process ID to collect metrics for
//
// Returns:
//   - ProcessIO: I/O statistics for the process
//   - error: nil on success, error if probe not initialized or collection fails
func (c *ProcessCollector) CollectIO(_ context.Context, pid int) (ProcessIO, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return ProcessIO{}, err
	}

	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return ProcessIO{}, err
	}

	// Return collected I/O statistics.
	return ProcessIO{
		PID:              int(cProc.pid),
		ReadBytesPerSec:  uint64(cProc.read_bytes_per_sec),
		WriteBytesPerSec: uint64(cProc.write_bytes_per_sec),
	}, nil
}
