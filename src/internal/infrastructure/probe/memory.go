//go:build cgo

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

// MemoryCollector collects memory metrics via the Rust probe library.
// It implements the metrics.MemoryCollector interface.
type MemoryCollector struct{}

// NewMemoryCollector creates a new memory collector backed by the Rust probe.
func NewMemoryCollector() *MemoryCollector {
	return &MemoryCollector{}
}

// CollectSystem collects system-wide memory metrics.
func (m *MemoryCollector) CollectSystem(_ context.Context) (metrics.SystemMemory, error) {
	if err := checkInitialized(); err != nil {
		return metrics.SystemMemory{}, err
	}

	var cMem C.SystemMemory
	result := C.probe_collect_memory(&cMem)
	if err := resultToError(result); err != nil {
		return metrics.SystemMemory{}, err
	}

	total := uint64(cMem.total_bytes)
	available := uint64(cMem.available_bytes)
	used := uint64(cMem.used_bytes)
	swapTotal := uint64(cMem.swap_total_bytes)
	swapUsed := uint64(cMem.swap_used_bytes)

	var usagePercent float64
	if total > 0 {
		usagePercent = float64(used) / float64(total) * 100.0
	}

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
func (m *MemoryCollector) CollectProcess(_ context.Context, pid int) (metrics.ProcessMemory, error) {
	if err := checkInitialized(); err != nil {
		return metrics.ProcessMemory{}, err
	}

	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	if err := resultToError(result); err != nil {
		return metrics.ProcessMemory{}, err
	}

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
func (m *MemoryCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessMemory, error) {
	// The Rust probe does not support enumerating all processes.
	return nil, ErrNotSupported
}

// CollectPressure collects memory pressure metrics (PSI).
// Note: PSI is a Linux 4.20+ feature, not available cross-platform.
func (m *MemoryCollector) CollectPressure(_ context.Context) (metrics.MemoryPressure, error) {
	if err := checkInitialized(); err != nil {
		return metrics.MemoryPressure{}, err
	}

	var pressure C.MemoryPressure
	result := C.probe_collect_memory_pressure(&pressure)
	if err := resultToError(result); err != nil {
		return metrics.MemoryPressure{}, err
	}

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
