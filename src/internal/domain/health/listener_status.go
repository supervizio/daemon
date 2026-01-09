// Package health provides domain entities and value objects for health checking.
package health

import "github.com/kodflow/daemon/internal/domain/listener"

// ListenerStatus represents the health status of a single listener.
// It tracks the listener's name, state, probe results, and consecutive success/failure counts.
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

// NewListenerStatus creates a new listener status.
//
// Params:
//   - name: the listener name.
//   - state: the listener's current state.
//
// Returns:
//   - ListenerStatus: a new listener status.
func NewListenerStatus(name string, state listener.State) ListenerStatus {
	// Return new listener status with name and state.
	return ListenerStatus{
		Name:  name,
		State: state,
	}
}

// SetLastProbeResult sets the last probe result.
//
// Params:
//   - result: the probe result.
func (ls *ListenerStatus) SetLastProbeResult(result *Result) {
	// Set last probe result.
	ls.LastProbeResult = result
}

// IncrementSuccesses increments the consecutive success count and resets failures.
func (ls *ListenerStatus) IncrementSuccesses() {
	// Increment consecutive successes.
	ls.ConsecutiveSuccesses++
	// Reset consecutive failures.
	ls.ConsecutiveFailures = 0
}

// IncrementFailures increments the consecutive failure count and resets successes.
func (ls *ListenerStatus) IncrementFailures() {
	// Increment consecutive failures.
	ls.ConsecutiveFailures++
	// Reset consecutive successes.
	ls.ConsecutiveSuccesses = 0
}

// IsReady returns true if the listener state is ready.
//
// Returns:
//   - bool: true if listener is ready, false otherwise.
func (ls *ListenerStatus) IsReady() bool {
	// Return true if state is ready.
	return ls.State.IsReady()
}

// IsListening returns true if the listener state is listening.
//
// Returns:
//   - bool: true if listener is listening, false otherwise.
func (ls *ListenerStatus) IsListening() bool {
	// Return true if state is listening.
	return ls.State.IsListening()
}
