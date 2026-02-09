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
	result := make([]model.ServiceSnapshot, len(snapshots))

	// Pre-calculate total listener count to avoid allocations in hot loop.
	totalListeners := countTotalListeners(snapshots)
	allListeners := make([]model.ListenerSnapshot, totalListeners)

	// convert supervisor snapshots to model snapshots
	listenerIdx := 0
	for i := range snapshots {
		snap := &snapshots[i]
		listenerStart := listenerIdx

		// Convert and append all listeners for this service.
		listenerIdx = convertListenersAt(allListeners, listenerIdx, snap.Listeners)
		listeners := allListeners[listenerStart:listenerIdx]

		// assign service snapshot with all fields
		result[i] = model.ServiceSnapshot{
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
		}
	}

	// return all converted snapshots
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
	// sum listener counts across all services
	for i := range snapshots {
		total += len(snapshots[i].Listeners)
	}
	// return computed total
	return total
}

// convertListenersAt converts listeners into dest slice at startIdx and returns next index.
//
// Params:
//   - dest: pre-allocated destination slice.
//   - startIdx: starting index in dest.
//   - listeners: source listeners to convert.
//
// Returns:
//   - int: next index after converted listeners.
func convertListenersAt(dest []model.ListenerSnapshot, startIdx int, listeners []appsupervisor.ListenerSnapshotForTUI) int {
	// convert each listener to model format with indexed assignment
	for j := range listeners {
		lsn := &listeners[j]
		// assign converted listener with all fields
		dest[startIdx+j] = model.ListenerSnapshot{
			Name:      lsn.Name,
			Port:      lsn.Port,
			Protocol:  lsn.Protocol,
			Exposed:   lsn.Exposed,
			Listening: lsn.Listening,
			Status:    model.PortStatus(lsn.StatusInt),
		}
	}
	// return next index after all converted listeners
	return startIdx + len(listeners)
}
