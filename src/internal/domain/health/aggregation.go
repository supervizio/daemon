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
		LastCheck:    time.Now(),
	}
}

// AddListener adds a listener status.
//
// Params:
//   - name: the listener name.
//   - state: the listener's current state.
func (h *AggregatedHealth) AddListener(name string, state listener.State) {
	// Append listener status using constructor.
	h.Listeners = append(h.Listeners, NewListenerStatus(name, state))
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

// computeListenerStatus determines the health status based on listener states.
// It returns StatusHealthy if all listeners are ready (or no listeners exist),
// StatusDegraded if some are listening but not all ready, and StatusUnhealthy
// if none are listening.
//
// Returns:
//   - Status: the computed listener status.
func (h *AggregatedHealth) computeListenerStatus() Status {
	// No listeners configured means no listener-based health concerns.
	// Return healthy since there are no listeners to fail.
	if len(h.Listeners) == 0 {
		return StatusHealthy
	}

	// Check if all listeners are ready.
	allReady := h.AllListenersReady()

	// Return healthy if all listeners are ready.
	if allReady {
		// All listeners are in ready state.
		return StatusHealthy
	}

	// Check if any listeners are listening (degraded state).
	anyListening := h.hasAnyListenerListening()

	// Return degraded if some listeners are listening.
	if anyListening {
		// Some listeners are ready/listening but not all.
		return StatusDegraded
	}

	// No listeners are ready or listening.
	return StatusUnhealthy
}

// hasAnyListenerListening checks if any listener is in listening state.
//
// Returns:
//   - bool: true if any listener is listening, false otherwise.
func (h *AggregatedHealth) hasAnyListenerListening() bool {
	// Iterate through all listeners to find one in listening state.
	for _, ls := range h.Listeners {
		// Check if this listener is listening.
		if ls.State.IsListening() {
			// Found a listening listener.
			return true
		}
	}

	// No listeners are listening.
	return false
}

// hasNonHealthyCustomStatus checks if custom status indicates non-healthy state.
//
// Returns:
//   - bool: true if custom status is set and not "HEALTHY", false otherwise.
func (h *AggregatedHealth) hasNonHealthyCustomStatus() bool {
	// Return true if custom status is set and not "HEALTHY".
	return h.CustomStatus != "" && h.CustomStatus != "HEALTHY"
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

	// Compute status based on listener states.
	listenerStatus := h.computeListenerStatus()

	// Return unhealthy or degraded if listeners are not all ready.
	if listenerStatus != StatusHealthy {
		// Return the computed listener status.
		return listenerStatus
	}

	// Check custom status for non-healthy state.
	if h.hasNonHealthyCustomStatus() {
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
	// Check each listener state.
	for _, ls := range h.Listeners {
		// Check if this listener is not ready.
		if !ls.State.IsReady() {
			// Return false if any listener is not ready.
			return false
		}
	}

	// Return true if all listeners are ready.
	return true
}

// ReadyListenerCount returns the count of ready listeners.
//
// Returns:
//   - int: the number of listeners in Ready state.
func (h *AggregatedHealth) ReadyListenerCount() int {
	count := 0

	// Count ready listeners by iterating through all listeners.
	for _, ls := range h.Listeners {
		// Check if this listener is ready.
		if ls.State.IsReady() {
			// Increment count for ready listener.
			count++
		}
	}

	// Return count of ready listeners.
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
