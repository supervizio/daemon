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
	// Pre-allocate result with capacity.
	result := make([]model.ServiceSnapshot, 0, len(snapshots))

	// Pre-calculate total listener count to avoid allocations in hot loop.
	totalListeners := countTotalListeners(snapshots)
	// Pre-allocate all listeners with capacity for append.
	allListeners := make([]model.ListenerSnapshot, 0, totalListeners)

	// Convert supervisor snapshots to model snapshots.
	for i := range snapshots {
		snap := &snapshots[i]

		// Convert and append all listeners for this service.
		listenerStart := len(allListeners)
		allListeners = appendConvertedListeners(allListeners, snap.Listeners)
		listeners := allListeners[listenerStart:]

		// Append service snapshot with all fields.
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

	// Return all converted snapshots.
	return result
}

// countTotalListeners counts total listeners across all services.
//
// Params:
//   - snapshots: service snapshots to count.
//
// Returns:
//   - int: total listener count.
func countTotalListeners(snapshots []appsupervisor.ServiceSnapshotForTUI) int {
	total := 0
	// Sum listener counts across all services.
	for i := range snapshots {
		total += len(snapshots[i].Listeners)
	}
	// Return computed total.
	return total
}

// appendConvertedListeners converts and appends listeners to dest slice.
//
// Params:
//   - dest: destination slice to append to.
//   - listeners: source listeners to convert.
//
// Returns:
//   - []model.ListenerSnapshot: dest slice with converted listeners appended.
func appendConvertedListeners(dest []model.ListenerSnapshot, listeners []appsupervisor.ListenerSnapshotForTUI) []model.ListenerSnapshot {
	// Convert each listener to model format and append.
	for j := range listeners {
		lsn := &listeners[j]
		// Append converted listener with all fields.
		dest = append(dest, model.ListenerSnapshot{
			Name:      lsn.Name,
			Port:      lsn.Port,
			Protocol:  lsn.Protocol,
			Exposed:   lsn.Exposed,
			Listening: lsn.Listening,
			Status:    model.PortStatus(lsn.StatusInt),
		})
	}
	// Return dest with appended listeners.
	return dest
}
