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

// ProcessCollector collects per-process metrics via the Rust probe library.
// It implements appmetrics.Collector interface.
type ProcessCollector struct{}

// NewProcessCollector creates a new process metrics collector.
func NewProcessCollector() *ProcessCollector {
	return &ProcessCollector{}
}

// CollectCPU collects CPU metrics for a specific process.
func (c *ProcessCollector) CollectCPU(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	if err := checkInitialized(); err != nil {
		return metrics.ProcessCPU{}, err
	}

	var cProc C.ProcessMetrics
	result := C.probe_collect_process(C.int32_t(pid), &cProc)
	if err := resultToError(result); err != nil {
		return metrics.ProcessCPU{}, err
	}

	return metrics.ProcessCPU{
		PID:          int(cProc.pid),
		UsagePercent: float64(cProc.cpu_percent),
		Timestamp:    time.Now(),
		// Note: Jiffies (User, System, etc.) not available cross-platform.
		// The UsagePercent is calculated by the Rust probe based on delta.
	}, nil
}

// CollectMemory collects memory metrics for a specific process.
func (c *ProcessCollector) CollectMemory(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
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
		// Note: Shared, Swap, Data, Stack not available cross-platform.
	}, nil
}
