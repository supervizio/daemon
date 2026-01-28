// Package tui provides terminal user interface rendering for superviz.io.
package tui

import "github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"

// SystemMetricsAdapter provides system metrics.
// This is a placeholder that delegates to TUI collectors.
type SystemMetricsAdapter struct {
	// System metrics will be collected via collectors.
	// This is a placeholder that can be extended.
}

// NewSystemMetricsAdapter creates a new system metrics adapter.
//
// Returns:
//   - *SystemMetricsAdapter: the created adapter.
func NewSystemMetricsAdapter() *SystemMetricsAdapter {
	return &SystemMetricsAdapter{}
}

// Metrics implements Metricser.
//
// Returns:
//   - model.SystemMetrics: empty metrics (TUI collectors handle this).
func (a *SystemMetricsAdapter) Metrics() model.SystemMetrics {
	// System metrics are collected by the TUI collectors.
	// This returns empty metrics; the TUI will use collectors instead.
	return model.SystemMetrics{}
}
