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

// CPUCollector collects CPU metrics via the Rust probe library.
// It implements the metrics.CPUCollector interface.
type CPUCollector struct{}

// NewCPUCollector creates a new CPU collector backed by the Rust probe.
func NewCPUCollector() *CPUCollector {
	return &CPUCollector{}
}

// CollectSystem collects system-wide CPU metrics.
// Note: The Rust probe provides percentages directly, not raw jiffies.
// Fields like IRQ, SoftIRQ, Guest, GuestNice will be zero as they are
// Linux-specific and not available cross-platform.
func (c *CPUCollector) CollectSystem(_ context.Context) (metrics.SystemCPU, error) {
	if err := checkInitialized(); err != nil {
		return metrics.SystemCPU{}, err
	}

	var cCPU C.SystemCPU
	result := C.probe_collect_cpu(&cCPU)
	if err := resultToError(result); err != nil {
		return metrics.SystemCPU{}, err
	}

	// The Rust probe provides percentages, not jiffies.
	// We set UsagePercent directly and leave jiffies fields as zero.
	// The UsagePercent is calculated as 100 - idle_percent.
	usagePercent := 100.0 - float64(cCPU.idle_percent)
	if usagePercent < 0 {
		usagePercent = 0
	}

	return metrics.SystemCPU{
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
		// Note: Jiffies fields (User, Nice, System, etc.) are not available
		// cross-platform. Use Linux-specific collector for detailed metrics.
	}, nil
}

// CollectProcess collects CPU metrics for a specific process.
// Note: The Rust probe provides a percentage, not raw jiffies.
func (c *CPUCollector) CollectProcess(_ context.Context, pid int) (metrics.ProcessCPU, error) {
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
		// Note: Jiffies fields not available cross-platform.
	}, nil
}

// CollectAllProcesses is not implemented by the Rust probe.
// Returns an empty slice and no error.
func (c *CPUCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessCPU, error) {
	// The Rust probe does not support enumerating all processes.
	// This would require iterating /proc on Linux, which is platform-specific.
	return nil, ErrNotSupported
}

// CollectLoadAverage collects system load average.
func (c *CPUCollector) CollectLoadAverage(_ context.Context) (metrics.LoadAverage, error) {
	if err := checkInitialized(); err != nil {
		return metrics.LoadAverage{}, err
	}

	var cLoad C.LoadAverage
	result := C.probe_collect_load(&cLoad)
	if err := resultToError(result); err != nil {
		return metrics.LoadAverage{}, err
	}

	return metrics.LoadAverage{
		Load1:     float64(cLoad.load_1min),
		Load5:     float64(cLoad.load_5min),
		Load15:    float64(cLoad.load_15min),
		Timestamp: time.Now(),
		// Note: Process counts (RunningProcesses, TotalProcesses, LastPID)
		// are not available cross-platform from the Rust probe.
	}, nil
}

// CollectPressure collects CPU pressure metrics (PSI).
// Note: PSI is a Linux 4.20+ feature, not available cross-platform.
func (c *CPUCollector) CollectPressure(_ context.Context) (metrics.CPUPressure, error) {
	if err := checkInitialized(); err != nil {
		return metrics.CPUPressure{}, err
	}

	var pressure C.CPUPressure
	result := C.probe_collect_cpu_pressure(&pressure)
	if err := resultToError(result); err != nil {
		return metrics.CPUPressure{}, err
	}

	return metrics.CPUPressure{
		SomeAvg10:  float64(pressure.some_avg10),
		SomeAvg60:  float64(pressure.some_avg60),
		SomeAvg300: float64(pressure.some_avg300),
		SomeTotal:  uint64(pressure.some_total_us),
		Timestamp:  time.Now(),
	}, nil
}
