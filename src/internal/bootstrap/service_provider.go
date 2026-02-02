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

// supervisorServiceLister wraps a supervisor to provide services to TUI.
type supervisorServiceLister struct {
	sup ServiceSnapshotsForTUIer
}

// ListServices implements tui.ServiceLister.
// Services are returned in definition order (as configured), not sorted.
//
// Returns:
//   - []model.ServiceSnapshot: list of service snapshots for TUI.
func (p *supervisorServiceLister) ListServices() []model.ServiceSnapshot {
	snapshots := p.sup.ServiceSnapshotsForTUI()
	result := make([]model.ServiceSnapshot, 0, len(snapshots))

	// Pre-calculate total listener count to avoid allocations in hot loop (KTN-VAR-HOTLOOP).
	totalListeners := int(0)
	// calculate total listener count across all services
	for i := range snapshots {
		totalListeners += len(snapshots[i].Listeners)
	}

	allListeners := make([]model.ListenerSnapshot, 0, totalListeners)

	// convert supervisor snapshots to model snapshots
	for i := range snapshots {
		snap := &snapshots[i]
		listenerStart := len(allListeners)

		// convert each listener to model format
		for j := range snap.Listeners {
			l := &snap.Listeners[j]
			// append converted listener
			allListeners = append(allListeners, model.ListenerSnapshot{
				Name:      l.Name,
				Port:      l.Port,
				Protocol:  l.Protocol,
				Exposed:   l.Exposed,
				Listening: l.Listening,
				Status:    model.PortStatus(l.StatusInt),
			})
		}

		listeners := allListeners[listenerStart:]

		// append service snapshot with all fields
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

	// return all converted snapshots
	return result
}
