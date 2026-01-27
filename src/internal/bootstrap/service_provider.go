// Package bootstrap provides dependency injection wiring using Google Wire.
package bootstrap

import (
	"time"

	appsupervisor "github.com/kodflow/daemon/internal/application/supervisor"
	domainhealth "github.com/kodflow/daemon/internal/domain/health"
	domainprocess "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// ServiceSnapshotsForTUIer provides service snapshots for TUI.
type ServiceSnapshotsForTUIer interface {
	ServiceSnapshotsForTUI() []appsupervisor.ServiceSnapshotForTUI
}

// supervisorServiceProvider wraps a supervisor to provide services to TUI.
type supervisorServiceProvider struct {
	sup ServiceSnapshotsForTUIer
}

// Services implements tui.ServiceProvider.
// Services are returned in definition order (as configured), not sorted.
//
// Returns:
//   - []model.ServiceSnapshot: list of service snapshots for TUI.
func (p *supervisorServiceProvider) Services() []model.ServiceSnapshot {
	snapshots := p.sup.ServiceSnapshotsForTUI()
	result := make([]model.ServiceSnapshot, 0, len(snapshots))

	// Pre-calculate total listener count to avoid allocations in hot loop (KTN-VAR-HOTLOOP).
	totalListeners := int(0)
	// Count total listeners across all services.
	for i := range snapshots {
		totalListeners += len(snapshots[i].Listeners)
	}

	// Allocate all listeners at once outside the loop.
	allListeners := make([]model.ListenerSnapshot, 0, totalListeners)

	// Iterate over each service snapshot to convert to TUI model.
	for _, snap := range snapshots {
		// Record starting index for this service's listeners.
		listenerStart := len(allListeners)

		// Convert each listener to TUI model format.
		for _, l := range snap.Listeners {
			allListeners = append(allListeners, model.ListenerSnapshot{
				Name:      l.Name,
				Port:      l.Port,
				Protocol:  l.Protocol,
				Exposed:   l.Exposed,
				Listening: l.Listening,
				Status:    model.PortStatus(l.StatusInt),
			})
		}

		// Slice the listeners for this service from the pre-allocated array.
		listeners := allListeners[listenerStart:]

		result = append(result, model.ServiceSnapshot{
			Name:            snap.Name,
			State:           domainprocess.State(snap.StateInt),
			Health:          domainhealth.Status(snap.HealthInt),
			HasHealthChecks: snap.HasHealthChecks,
			PID:             snap.PID,
			Uptime:          time.Duration(snap.UptimeSecs) * time.Second,
			CPUPercent:      snap.CPUPercent,
			MemoryRSS:       snap.MemoryRSS,
			RestartCount:    snap.RestartCount,
			Ports:           snap.Ports,
			Listeners:       listeners,
		})
	}

	// Keep definition order - no sorting.
	return result
}
