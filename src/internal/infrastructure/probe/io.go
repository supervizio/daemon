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

// IOCollector provides I/O metrics via the Rust probe library.
type IOCollector struct{}

// NewIOCollector creates a new I/O collector.
func NewIOCollector() *IOCollector {
	return &IOCollector{}
}

// CollectStats collects system-wide I/O statistics.
func (i *IOCollector) CollectStats(_ context.Context) (metrics.IOStats, error) {
	if err := checkInitialized(); err != nil {
		return metrics.IOStats{}, err
	}

	var stats C.IOStats
	result := C.probe_collect_io_stats(&stats)
	if err := resultToError(result); err != nil {
		return metrics.IOStats{}, err
	}

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
func (i *IOCollector) CollectPressure(_ context.Context) (metrics.IOPressure, error) {
	if err := checkInitialized(); err != nil {
		return metrics.IOPressure{}, err
	}

	var pressure C.IOPressure
	result := C.probe_collect_io_pressure(&pressure)
	if err := resultToError(result); err != nil {
		return metrics.IOPressure{}, err
	}

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
