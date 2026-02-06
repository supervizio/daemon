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
func (c *ProcessCollector) CollectCPU(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return empty metrics on validation failure.
		return metrics.ProcessCPU{}, err
	}
	// Collect process metrics from C library.
	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return empty metrics on collection failure.
		return metrics.ProcessCPU{}, err
	}
	// Return collected CPU metrics with current timestamp.
	return metrics.ProcessCPU{
		PID:          int(cProc.pid),
		UsagePercent: float64(cProc.cpu_percent),
		Timestamp:    time.Now(),
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
func (c *ProcessCollector) CollectMemory(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return empty metrics on validation failure.
		return metrics.ProcessMemory{}, err
	}
	// Collect process metrics from C library.
	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return empty metrics on collection failure.
		return metrics.ProcessMemory{}, err
	}
	// Return collected memory metrics with current timestamp.
	return metrics.ProcessMemory{
		PID:          int(cProc.pid),
		RSS:          uint64(cProc.memory_rss_bytes),
		VMS:          uint64(cProc.memory_vms_bytes),
		UsagePercent: float64(cProc.memory_percent),
		Timestamp:    time.Now(),
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
func (c *ProcessCollector) CollectFDs(ctx context.Context, pid int) (ProcessFDs, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return empty metrics on validation failure.
		return ProcessFDs{}, err
	}
	// Collect process metrics from C library.
	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return empty metrics on collection failure.
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
func (c *ProcessCollector) CollectIO(ctx context.Context, pid int) (ProcessIO, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return empty I/O stats on validation failure.
		return ProcessIO{}, err
	}
	// Collect process metrics from C library.
	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return empty I/O stats on collection failure.
		return ProcessIO{}, err
	}
	// Return collected I/O statistics.
	return ProcessIO{
		PID:              int(cProc.pid),
		ReadBytesPerSec:  uint64(cProc.read_bytes_per_sec),
		WriteBytesPerSec: uint64(cProc.write_bytes_per_sec),
	}, nil
}
