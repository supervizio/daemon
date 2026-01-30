//go:build cgo

package probe

import (
	appmetrics "github.com/kodflow/daemon/internal/application/metrics"
	"github.com/kodflow/daemon/internal/domain/metrics"
)

// NewSystemCollector creates a SystemCollector using the Rust probe.
// This is the main entry point for system metrics collection.
// It provides a unified cross-platform implementation.
//
// Returns:
//   - metrics.SystemCollector: cross-platform metrics collector
func NewSystemCollector() metrics.SystemCollector {
	return NewCollector()
}

// NewAppProcessCollector creates an appmetrics.Collector using the Rust probe.
// This is used by the application layer for process metrics tracking.
//
// Returns:
//   - appmetrics.Collector: process metrics collector for application layer
func NewAppProcessCollector() appmetrics.Collector {
	return NewProcessCollector()
}

// DetectedPlatform returns a string describing the detected platform.
// This is useful for logging and diagnostics.
//
// Returns:
//   - string: platform identifier from the Rust probe
func DetectedPlatform() string {
	return Platform()
}
