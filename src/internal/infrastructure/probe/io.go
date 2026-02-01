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

// IOCollector provides I/O metrics via the Rust probe library.
// It implements the metrics.IOCollector interface for system-wide I/O statistics.
type IOCollector struct{}

// NewIOCollector creates a new I/O collector.
//
// Returns:
//   - *IOCollector: new I/O collector instance
func NewIOCollector() *IOCollector {
	// Return a new empty collector instance.
	return &IOCollector{}
}

// CollectStats collects system-wide I/O statistics.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.IOStats: system-wide I/O statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (i *IOCollector) CollectStats(ctx context.Context) (metrics.IOStats, error) {
	_ = ctx // reserved for future cancellation support
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty stats with initialization error.
		return metrics.IOStats{}, err
	}

	var stats C.IOStats
	result := C.probe_collect_io_stats(&stats)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty stats with collection error.
		return metrics.IOStats{}, err
	}

	// Return collected I/O statistics with current timestamp.
	return metrics.IOStats{
		ReadOpsTotal:    uint64(stats.read_ops),
		ReadBytesTotal:  uint64(stats.read_bytes),
		WriteOpsTotal:   uint64(stats.write_ops),
		WriteBytesTotal: uint64(stats.write_bytes),
		Timestamp:       time.Now(),
	}, nil
}

// CollectPressure collects I/O pressure metrics (PSI).
// Note: PSI is a Linux 4.20+ feature, not available cross-platform.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.IOPressure: I/O pressure metrics
//   - error: nil on success, error if probe not initialized or collection fails
func (i *IOCollector) CollectPressure(ctx context.Context) (metrics.IOPressure, error) {
	_ = ctx // reserved for future cancellation support
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty pressure metrics with initialization error.
		return metrics.IOPressure{}, err
	}

	var pressure C.IOPressure
	result := C.probe_collect_io_pressure(&pressure)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty pressure metrics with collection error.
		return metrics.IOPressure{}, err
	}

	// Return collected I/O pressure metrics with current timestamp.
	return metrics.IOPressure{
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
