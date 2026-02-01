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

// fullPercentCache represents 100% for cache percentage calculations.
const fullPercentCache float64 = 100.0

// CachedCPUCollector provides CPU metrics collection with caching support.
// It implements the metrics.CPUCollector interface.
type CachedCPUCollector struct{}

// NewCachedCPUCollector creates a new cached CPU collector.
//
// Returns:
//   - *CachedCPUCollector: new cached CPU collector instance
func NewCachedCPUCollector() *CachedCPUCollector {
	// Return a new empty collector instance.
	return &CachedCPUCollector{}
}

// CollectSystem collects system-wide CPU metrics with caching.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.SystemCPU: cached system-wide CPU statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CachedCPUCollector) CollectSystem(_ context.Context) (metrics.SystemCPU, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.SystemCPU{}, err
	}

	var cCPU C.SystemCPU
	result := C.probe_collect_cpu_cached(&cCPU)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.SystemCPU{}, err
	}

	usagePercent := fullPercentCache - float64(cCPU.idle_percent)
	// Clamp negative values to zero.
	if usagePercent < 0 {
		usagePercent = 0
	}

	// Return cached CPU metrics with current timestamp.
	return metrics.SystemCPU{
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
	}, nil
}

// CollectProcess collects CPU metrics for a specific process.
// Process metrics are not cached.
//
// Params:
//   - ctx: context for cancellation
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessCPU: CPU metrics for the process
//   - error: nil on success, error if collection fails
func (c *CachedCPUCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	// Process metrics are not cached, delegate to regular collector.
	return (&CPUCollector{}).CollectProcess(ctx, pid)
}

// CollectAllProcesses is not implemented.
//
// Params:
//   - ctx: context for cancellation (unused)
//
// Returns:
//   - []metrics.ProcessCPU: always nil
//   - error: always ErrNotSupported
func (c *CachedCPUCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessCPU, error) {
	// Return not supported error.
	return nil, ErrNotSupported
}

// CollectLoadAverage collects system load average with caching.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.LoadAverage: cached system load averages
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CachedCPUCollector) CollectLoadAverage(_ context.Context) (metrics.LoadAverage, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.LoadAverage{}, err
	}

	var cLoad C.LoadAverage
	result := C.probe_collect_load_cached(&cLoad)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.LoadAverage{}, err
	}

	// Return cached load average with current timestamp.
	return metrics.LoadAverage{
		Load1:     float64(cLoad.load_1min),
		Load5:     float64(cLoad.load_5min),
		Load15:    float64(cLoad.load_15min),
		Timestamp: time.Now(),
	}, nil
}

// CollectPressure collects CPU pressure metrics.
// Pressure metrics are not cached.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.CPUPressure: CPU pressure statistics
//   - error: nil on success, error if collection fails
func (c *CachedCPUCollector) CollectPressure(ctx context.Context) (metrics.CPUPressure, error) {
	// Delegate to regular collector as pressure is not cached.
	return (&CPUCollector{}).CollectPressure(ctx)
}
