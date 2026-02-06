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

// fullPercent represents 100% for percentage calculations.
const fullPercent float64 = 100.0

// CPUCollector collects CPU metrics via the Rust probe library.
// It implements the metrics.CPUCollector interface.
type CPUCollector struct{}

// NewCPUCollector creates a new CPU collector backed by the Rust probe.
//
// Returns:
//   - *CPUCollector: new CPU collector instance
func NewCPUCollector() *CPUCollector {
	// Return a new empty collector instance.
	return &CPUCollector{}
}

// CollectSystem collects system-wide CPU metrics.
// Note: The Rust probe provides percentages directly, not raw jiffies.
// Fields like IRQ, SoftIRQ, Guest, GuestNice will be zero as they are
// Linux-specific and not available cross-platform.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.SystemCPU: system-wide CPU statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CPUCollector) CollectSystem(ctx context.Context) (metrics.SystemCPU, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return empty metrics on validation failure.
		return metrics.SystemCPU{}, err
	}
	// Collect CPU metrics from C library.
	var cCPU C.SystemCPU
	result := C.probe_collect_cpu(&cCPU)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return empty metrics on collection failure.
		return metrics.SystemCPU{}, err
	}
	// Calculate usage percent from idle percent.
	usagePercent := fullPercent - float64(cCPU.idle_percent)
	// Clamp negative values to zero.
	if usagePercent < 0 {
		usagePercent = 0
	}
	// Return collected CPU metrics with current timestamp.
	return metrics.SystemCPU{
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
	}, nil
}

// CollectProcess collects CPU metrics for a specific process.
// Note: The Rust probe provides a percentage, not raw jiffies.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessCPU: CPU metrics for the process
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CPUCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
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
	// Return collected process CPU metrics with current timestamp.
	return metrics.ProcessCPU{
		PID:          int(cProc.pid),
		UsagePercent: float64(cProc.cpu_percent),
		Timestamp:    time.Now(),
	}, nil
}

// CollectAllProcesses is not implemented by the Rust probe.
// Returns an empty slice and no error.
//
// Params:
//   - ctx: context for cancellation (unused)
//
// Returns:
//   - []metrics.ProcessCPU: always nil
//   - error: always ErrNotSupported
func (c *CPUCollector) CollectAllProcesses(ctx context.Context) ([]metrics.ProcessCPU, error) {
	// Check if context has been cancelled.
	if err := checkContext(ctx); err != nil {
		// Return empty metrics with context error.
		return nil, err
	}
	// The Rust probe does not support enumerating all processes.
	// This would require iterating /proc on Linux, which is platform-specific.
	return nil, ErrNotSupported
}

// CollectLoadAverage collects system load average.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.LoadAverage: system load averages (1, 5, 15 minutes)
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CPUCollector) CollectLoadAverage(ctx context.Context) (metrics.LoadAverage, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return empty metrics on validation failure.
		return metrics.LoadAverage{}, err
	}
	// Collect load average from C library.
	var cLoad C.LoadAverage
	result := C.probe_collect_load(&cLoad)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return empty metrics on collection failure.
		return metrics.LoadAverage{}, err
	}
	// Return collected load average with current timestamp.
	return metrics.LoadAverage{
		Load1:     float64(cLoad.load_1min),
		Load5:     float64(cLoad.load_5min),
		Load15:    float64(cLoad.load_15min),
		Timestamp: time.Now(),
	}, nil
}

// CollectPressure collects CPU pressure metrics (PSI).
// Note: PSI is a Linux 4.20+ feature, not available cross-platform.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.CPUPressure: CPU pressure statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CPUCollector) CollectPressure(ctx context.Context) (metrics.CPUPressure, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return empty metrics on validation failure.
		return metrics.CPUPressure{}, err
	}
	// Collect CPU pressure from C library.
	var pressure C.CPUPressure
	result := C.probe_collect_cpu_pressure(&pressure)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return empty metrics on collection failure.
		return metrics.CPUPressure{}, err
	}
	// Return collected CPU pressure metrics with current timestamp.
	return metrics.CPUPressure{
		SomeAvg10:  float64(pressure.some_avg10),
		SomeAvg60:  float64(pressure.some_avg60),
		SomeAvg300: float64(pressure.some_avg300),
		SomeTotal:  uint64(pressure.some_total_us),
		Timestamp:  time.Now(),
	}, nil
}
