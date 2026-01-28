//go:build linux

// Package metrics provides platform detection and factory for metrics collectors.
package metrics

import (
	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/linux"
)

// NewProcessCollector creates a ProcessCollector appropriate for the current platform.
// On Linux, it uses the /proc filesystem for accurate CPU and memory metrics.
//
// Returns:
//   - appmetrics.Collector: platform-specific process metrics collector
func NewProcessCollector() appmetrics.Collector {
	// Return Linux process collector that uses /proc filesystem.
	return linux.NewProcessCollector()
}
