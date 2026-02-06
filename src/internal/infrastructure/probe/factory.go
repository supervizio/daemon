//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
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
	// Create and return the unified collector implementation.
	return NewCollector()
}

// NewAppProcessCollector creates an appmetrics.Collector using the Rust probe.
// This is used by the application layer for process metrics tracking.
//
// Returns:
//   - appmetrics.Collector: process metrics collector for application layer
func NewAppProcessCollector() appmetrics.Collector {
	// Create and return the process metrics collector.
	return NewProcessCollector()
}

// DetectedPlatform returns a string describing the detected platform.
// This is useful for logging and diagnostics.
//
// Returns:
//   - string: platform identifier from the Rust probe
func DetectedPlatform() string {
	// Return platform identifier from the native probe.
	return Platform()
}
