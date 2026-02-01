// Package health provides domain entities and value objects for health checking.
package health

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/process"
)

// AggregatedHealth represents the combined health status of a service.
// It aggregates process state, subject states, and custom status.
// Uses SubjectState (domain-owned type) to avoid coupling to listener/process packages.
type AggregatedHealth struct {
	// ProcessState is the current state of the process.
	ProcessState process.State

	// Subjects contains the status of each monitored subject (listeners, processes).
	// Backward compatible alias: Listeners.
	Subjects []SubjectStatus

	// CustomStatus is an optional custom status string.
	// Examples: "DRAINING", "DEGRADED", "MAINTENANCE".
	// Empty string or "HEALTHY" indicates normal operation.
	CustomStatus string

	// LastCheck is the timestamp of the last health check.
	LastCheck time.Time

	// Latency is the latest probe latency measurement.
	Latency time.Duration
}

// Listeners returns the subjects slice for backward compatibility.
//
// Returns:
//   - []SubjectStatus: the list of subject statuses.
//
// Deprecated: Use Subjects field directly instead.
func (h *AggregatedHealth) Listeners() []SubjectStatus {
	// return subjects for backward compatibility
	return h.Subjects
}

// NewAggregatedHealth creates a new aggregated health status.
//
// Params:
//   - processState: the current process state.
//
// Returns:
//   - *AggregatedHealth: a new aggregated health status.
func NewAggregatedHealth(processState process.State) *AggregatedHealth {
	// initialize with current timestamp
	return &AggregatedHealth{
		ProcessState: processState,
		LastCheck:    time.Now(),
	}
}

// AddSubject adds a subject status from a snapshot.
//
// Params:
//   - snapshot: the subject snapshot.
func (h *AggregatedHealth) AddSubject(snapshot SubjectSnapshot) {
	// append new subject status
	h.Subjects = append(h.Subjects, NewSubjectStatus(snapshot))
}

// AddListener adds a listener status (backward compatibility).
//
// Params:
//   - name: the listener name.
//   - state: the subject's current state.
//
// Deprecated: Use AddSubject instead.
func (h *AggregatedHealth) AddListener(name string, state SubjectState) {
	// append listener for backward compatibility
	h.Subjects = append(h.Subjects, NewSubjectStatusFromState(name, state))
}

// SetCustomStatus sets a custom status string.
//
// Params:
//   - status: the custom status string.
func (h *AggregatedHealth) SetCustomStatus(status string) {
	// update custom status and timestamp
	h.CustomStatus = status
	h.LastCheck = time.Now()
}

// SetLatency sets the latency measurement.
//
// Params:
//   - latency: the latency duration.
func (h *AggregatedHealth) SetLatency(latency time.Duration) {
	// update latency measurement
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
	// no listeners means healthy
	if len(h.Subjects) == 0 {
		// no subjects to check
		return StatusHealthy
	}

	allReady := h.AllListenersReady()
	// all listeners ready means healthy
	if allReady {
		// all subjects ready
		return StatusHealthy
	}

	anyListening := h.hasAnyListenerListening()
	// some listening but not all ready means degraded
	if anyListening {
		// partial availability
		return StatusDegraded
	}

	// none listening means unhealthy
	return StatusUnhealthy
}

// hasAnyListenerListening checks if any listener is in listening state.
//
// Returns:
//   - bool: true if any listener is listening, false otherwise.
func (h *AggregatedHealth) hasAnyListenerListening() bool {
	// check each subject for listening state
	for _, ls := range h.Subjects {
		// found a listening subject
		if ls.State.IsListening() {
			// found listening subject
			return true
		}
	}
	// no listening subjects found
	return false
}

// hasNonHealthyCustomStatus checks if custom status indicates non-healthy state.
//
// Returns:
//   - bool: true if custom status is set and not "HEALTHY", false otherwise.
func (h *AggregatedHealth) hasNonHealthyCustomStatus() bool {
	// check for non-healthy custom status
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
	// process not running means unhealthy
	if !h.ProcessState.IsRunning() {
		// process down
		return StatusUnhealthy
	}

	listenerStatus := h.computeListenerStatus()
	// check listener health
	if listenerStatus != StatusHealthy {
		// propagate listener status
		return listenerStatus
	}

	// check custom status degradation
	if h.hasNonHealthyCustomStatus() {
		// custom degradation
		return StatusDegraded
	}

	// all checks passed
	return StatusHealthy
}

// IsHealthy returns true if the overall status is healthy.
//
// Returns:
//   - bool: true if healthy, false otherwise.
func (h *AggregatedHealth) IsHealthy() bool {
	// check for healthy status
	return h.Status() == StatusHealthy
}

// IsDegraded returns true if the overall status is degraded.
//
// Returns:
//   - bool: true if degraded, false otherwise.
func (h *AggregatedHealth) IsDegraded() bool {
	// check for degraded status
	return h.Status() == StatusDegraded
}

// IsUnhealthy returns true if the overall status is unhealthy.
//
// Returns:
//   - bool: true if unhealthy, false otherwise.
func (h *AggregatedHealth) IsUnhealthy() bool {
	// check for unhealthy status
	return h.Status() == StatusUnhealthy
}

// AllListenersReady returns true if all listeners are ready.
//
// Returns:
//   - bool: true if all listeners are in Ready state.
func (h *AggregatedHealth) AllListenersReady() bool {
	// check each subject for ready state
	for _, ls := range h.Subjects {
		// found a non-ready subject
		if !ls.State.IsReady() {
			// subject not ready
			return false
		}
	}
	// all subjects are ready
	return true
}

// ReadyListenerCount returns the count of ready listeners.
//
// Returns:
//   - int: the number of listeners in Ready state.
func (h *AggregatedHealth) ReadyListenerCount() int {
	count := 0
	// count ready subjects
	for _, ls := range h.Subjects {
		// increment count for ready subjects
		if ls.State.IsReady() {
			count++
		}
	}
	// return total ready count
	return count
}

// TotalListenerCount returns the total count of listeners.
//
// Returns:
//   - int: the total number of listeners.
func (h *AggregatedHealth) TotalListenerCount() int {
	// return total subject count
	return len(h.Subjects)
}
