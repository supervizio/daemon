// Package tui provides terminal user interface rendering for superviz.io.
package tui

import (
	"time"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// TUISnapshotData contains service data for TUI display.
type TUISnapshotData struct {
	Name   string
	State  process.State
	PID    int
	Uptime int64
}

// TUISnapshotsProvider provides service snapshots for TUI display.
type TUISnapshotsProvider interface {
	// TUISnapshots returns service data for TUI display.
	TUISnapshots() []TUISnapshotData
}

// SupervisorMetrics provides process metrics from the metrics tracker.
type SupervisorMetrics interface {
	// Get returns metrics for a specific service.
	Get(serviceName string) (domainmetrics.ProcessMetrics, bool)
}

// DynamicServiceProvider queries the supervisor on each call.
type DynamicServiceProvider struct {
	provider TUISnapshotsProvider
	metrics  SupervisorMetrics
}

// NewDynamicServiceProvider creates a new dynamic service provider.
func NewDynamicServiceProvider(provider TUISnapshotsProvider, metrics SupervisorMetrics) *DynamicServiceProvider {
	return &DynamicServiceProvider{
		provider: provider,
		metrics:  metrics,
	}
}

// Services implements ServiceProvider.
func (p *DynamicServiceProvider) Services() []model.ServiceSnapshot {
	if p.provider == nil {
		return nil
	}

	snapshots := p.provider.TUISnapshots()
	result := make([]model.ServiceSnapshot, 0, len(snapshots))

	for _, snap := range snapshots {
		ss := model.ServiceSnapshot{
			Name:   snap.Name,
			State:  snap.State,
			PID:    snap.PID,
			Uptime: time.Duration(snap.Uptime) * time.Second,
		}

		// Add metrics if available.
		if p.metrics != nil {
			if m, ok := p.metrics.Get(snap.Name); ok {
				ss.CPUPercent = m.CPU.UsagePercent
				ss.MemoryRSS = m.Memory.RSS
			}
		}

		result = append(result, ss)
	}

	return result
}

// SystemMetricsAdapter provides system metrics.
type SystemMetricsAdapter struct {
	// System metrics will be collected via collectors.
	// This is a placeholder that can be extended.
}

// NewSystemMetricsAdapter creates a new system metrics adapter.
func NewSystemMetricsAdapter() *SystemMetricsAdapter {
	return &SystemMetricsAdapter{}
}

// SystemMetrics implements MetricsProvider.
func (a *SystemMetricsAdapter) SystemMetrics() model.SystemMetrics {
	// System metrics are collected by the TUI collectors.
	// This returns empty metrics; the TUI will use collectors instead.
	return model.SystemMetrics{}
}

// LogAdapter provides log summary.
type LogAdapter struct {
	// Log summary will be collected from log files.
	// This is a placeholder that can be extended.
}

// NewLogAdapter creates a new log adapter.
func NewLogAdapter() *LogAdapter {
	return &LogAdapter{}
}

// LogSummary implements HealthProvider.
func (a *LogAdapter) LogSummary() model.LogSummary {
	// Log summary is collected by the TUI collectors.
	// This returns empty summary; the TUI will use collectors instead.
	return model.LogSummary{}
}
