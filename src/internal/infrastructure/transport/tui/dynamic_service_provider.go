// Package tui provides terminal user interface rendering for superviz.io.
package tui

import (
	"time"

	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// TUISnapshotser provides service snapshots for TUI display.
type TUISnapshotser interface {
	// TUISnapshots returns service data for TUI display.
	TUISnapshots() []TUISnapshotData
}

// ProcessMetricsProvider provides process metrics from the metrics tracker.
// It wraps a metrics tracker to provide TUI-compatible process metrics.
type ProcessMetricsProvider interface {
	// Get returns metrics for a specific service.
	Get(serviceName string) (domainmetrics.ProcessMetrics, bool)
	// Has checks if metrics exist for a service.
	Has(serviceName string) bool
}

// DynamicServiceProvider queries the supervisor on each call.
// It bridges the supervisor's metric tracking with the TUI's display model.
type DynamicServiceProvider struct {
	provider TUISnapshotser
	metrics  ProcessMetricsProvider
}

// NewDynamicServiceProvider creates a new dynamic service provider.
//
// Params:
//   - provider: the TUI snapshots provider.
//   - metrics: the supervisor metrics tracker.
//
// Returns:
//   - *DynamicServiceProvider: the created provider.
func NewDynamicServiceProvider(provider TUISnapshotser, metrics ProcessMetricsProvider) *DynamicServiceProvider {
	// return computed result.
	return &DynamicServiceProvider{
		provider: provider,
		metrics:  metrics,
	}
}

// ListServices implements ServiceLister.
//
// Returns:
//   - []model.ServiceSnapshot: the service snapshots.
func (p *DynamicServiceProvider) ListServices() []model.ServiceSnapshot {
	// handle nil condition.
	if p.provider == nil {
		// return nil to indicate no error.
		return nil
	}

	snapshots := p.provider.TUISnapshots()
	result := make([]model.ServiceSnapshot, 0, len(snapshots))

	// iterate over collection.
	for _, snap := range snapshots {
		ss := model.ServiceSnapshot{
			Name:   snap.Name,
			State:  snap.State,
			PID:    snap.PID,
			Uptime: time.Duration(snap.Uptime) * time.Second,
		}

		// handle non-nil condition.
		if p.metrics != nil {
			// evaluate condition.
			if m, ok := p.metrics.Get(snap.Name); ok {
				ss.CPUPercent = m.CPU.UsagePercent
				ss.MemoryRSS = m.Memory.RSS
			}
		}

		result = append(result, ss)
	}

	// return computed result.
	return result
}
