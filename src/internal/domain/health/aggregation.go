// Package health provides domain entities and value objects for health checking.
package health

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/kodflow/daemon/internal/domain/process"
)

// AggregatedHealth represents the combined health status of a service.
// It aggregates process state, listener states, and custom status.
type AggregatedHealth struct {
	// ProcessState is the current state of the process.
	ProcessState process.State

	// Listeners contains the status of each listener.
	Listeners []ListenerStatus

	// CustomStatus is an optional custom status string.
	// Examples: "DRAINING", "DEGRADED", "MAINTENANCE".
	// Empty string or "HEALTHY" indicates normal operation.
	CustomStatus string

	// LastCheck is the timestamp of the last health check.
	LastCheck time.Time

	// Latency is the latest probe latency measurement.
	Latency time.Duration
}

// ListenerStatus represents the health status of a single listener.
type ListenerStatus struct {
	// Name is the listener name.
	Name string

	// State is the listener's current state.
	State listener.State

	// LastProbeResult contains the result of the last probe.
	LastProbeResult *Result

	// ConsecutiveSuccesses is the count of consecutive successful probes.
	ConsecutiveSuccesses int

	// ConsecutiveFailures is the count of consecutive failed probes.
	ConsecutiveFailures int
}

// NewAggregatedHealth creates a new aggregated health status.
//
// Params:
//   - processState: the current process state.
//
// Returns:
//   - *AggregatedHealth: a new aggregated health status.
func NewAggregatedHealth(processState process.State) *AggregatedHealth {
	// Return new aggregated health with process state.
	return &AggregatedHealth{
		ProcessState: processState,
		Listeners:    make([]ListenerStatus, 0),
		LastCheck:    time.Now(),
	}
}

// AddListener adds a listener status.
//
// Params:
//   - name: the listener name.
//   - state: the listener's current state.
func (h *AggregatedHealth) AddListener(name string, state listener.State) {
	// Append listener status.
	h.Listeners = append(h.Listeners, ListenerStatus{
		Name:  name,
		State: state,
	})
}

// SetCustomStatus sets a custom status string.
//
// Params:
//   - status: the custom status string.
func (h *AggregatedHealth) SetCustomStatus(status string) {
	// Set custom status.
	h.CustomStatus = status
	// Update last check time.
	h.LastCheck = time.Now()
}

// SetLatency sets the latency measurement.
//
// Params:
//   - latency: the latency duration.
func (h *AggregatedHealth) SetLatency(latency time.Duration) {
	// Set latency.
	h.Latency = latency
}

// Status computes the overall health status.
// The status is determined by:
// 1. Process must be running
// 2. All listeners must be ready
// 3. CustomStatus must be empty or "HEALTHY"
//
// Returns:
//   - Status: the computed overall status.
func (h *AggregatedHealth) Status() Status {
	// Check if process is running.
	if !h.ProcessState.IsRunning() {
		// Return unhealthy if process is not running.
		return StatusUnhealthy
	}

	// Check all listeners are ready.
	allReady := true
	for _, ls := range h.Listeners {
		if !ls.State.IsReady() {
			allReady = false
			break
		}
	}

	// If not all listeners are ready.
	if !allReady {
		// Check if any listeners are listening (degraded).
		anyListening := false
		for _, ls := range h.Listeners {
			if ls.State.IsListening() {
				anyListening = true
				break
			}
		}
		// Return degraded if some are listening.
		if anyListening {
			return StatusDegraded
		}
		// Return unhealthy if none are listening.
		return StatusUnhealthy
	}

	// Check custom status.
	if h.CustomStatus != "" && h.CustomStatus != "HEALTHY" {
		// Return degraded for non-healthy custom status.
		return StatusDegraded
	}

	// All checks passed.
	return StatusHealthy
}

// IsHealthy returns true if the overall status is healthy.
//
// Returns:
//   - bool: true if healthy, false otherwise.
func (h *AggregatedHealth) IsHealthy() bool {
	// Return true if status is healthy.
	return h.Status() == StatusHealthy
}

// IsDegraded returns true if the overall status is degraded.
//
// Returns:
//   - bool: true if degraded, false otherwise.
func (h *AggregatedHealth) IsDegraded() bool {
	// Return true if status is degraded.
	return h.Status() == StatusDegraded
}

// IsUnhealthy returns true if the overall status is unhealthy.
//
// Returns:
//   - bool: true if unhealthy, false otherwise.
func (h *AggregatedHealth) IsUnhealthy() bool {
	// Return true if status is unhealthy.
	return h.Status() == StatusUnhealthy
}

// AllListenersReady returns true if all listeners are ready.
//
// Returns:
//   - bool: true if all listeners are in Ready state.
func (h *AggregatedHealth) AllListenersReady() bool {
	// Check each listener.
	for _, ls := range h.Listeners {
		if !ls.State.IsReady() {
			// Return false if any listener is not ready.
			return false
		}
	}
	// Return true if all are ready.
	return true
}

// ReadyListenerCount returns the count of ready listeners.
//
// Returns:
//   - int: the number of listeners in Ready state.
func (h *AggregatedHealth) ReadyListenerCount() int {
	count := 0
	// Count ready listeners.
	for _, ls := range h.Listeners {
		if ls.State.IsReady() {
			count++
		}
	}
	// Return count.
	return count
}

// TotalListenerCount returns the total count of listeners.
//
// Returns:
//   - int: the total number of listeners.
func (h *AggregatedHealth) TotalListenerCount() int {
	// Return length of listeners slice.
	return len(h.Listeners)
}
