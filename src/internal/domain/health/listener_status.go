// Package health provides domain entities and value objects for health checking.
package health

// SubjectStatus represents the health status of a monitored subject.
// It tracks the subject's snapshot, probe results, and consecutive success/failure counts.
// This type uses SubjectSnapshot to avoid coupling to concrete listener/process types.
type SubjectStatus struct {
	// Name is the subject name.
	Name string

	// State is the subject's current state (domain-owned type).
	State SubjectState

	// LastProbeResult contains the result of the last probe.
	LastProbeResult *Result

	// ConsecutiveSuccesses is the count of consecutive successful probes.
	ConsecutiveSuccesses int

	// ConsecutiveFailures is the count of consecutive failed probes.
	ConsecutiveFailures int
}

// NewSubjectStatus creates a new subject status from a snapshot.
//
// Params:
//   - snapshot: the subject snapshot.
//
// Returns:
//   - SubjectStatus: a new subject status.
func NewSubjectStatus(snapshot SubjectSnapshot) SubjectStatus {
	// Return new subject status from snapshot.
	return SubjectStatus{
		Name:  snapshot.Name,
		State: snapshot.State,
	}
}

// NewSubjectStatusFromState creates a new subject status from name and state.
//
// Params:
//   - name: the subject name.
//   - state: the subject's current state.
//
// Returns:
//   - SubjectStatus: a new subject status.
func NewSubjectStatusFromState(name string, state SubjectState) SubjectStatus {
	// Return new subject status with name and state.
	return SubjectStatus{
		Name:  name,
		State: state,
	}
}

// SetLastProbeResult sets the last probe result.
//
// Params:
//   - result: the probe result.
func (ss *SubjectStatus) SetLastProbeResult(result *Result) {
	// Set last probe result.
	ss.LastProbeResult = result
}

// IncrementSuccesses increments the consecutive success count and resets failures.
func (ss *SubjectStatus) IncrementSuccesses() {
	// Increment consecutive successes.
	ss.ConsecutiveSuccesses++
	// Reset consecutive failures.
	ss.ConsecutiveFailures = 0
}

// IncrementFailures increments the consecutive failure count and resets successes.
func (ss *SubjectStatus) IncrementFailures() {
	// Increment consecutive failures.
	ss.ConsecutiveFailures++
	// Reset consecutive successes.
	ss.ConsecutiveSuccesses = 0
}

// SetState updates the subject state.
//
// Params:
//   - state: the new state.
func (ss *SubjectStatus) SetState(state SubjectState) {
	// Update the state.
	ss.State = state
}

// IsReady returns true if the subject state is ready.
//
// Returns:
//   - bool: true if subject is ready, false otherwise.
func (ss *SubjectStatus) IsReady() bool {
	// Delegate to SubjectState method.
	return ss.State.IsReady()
}

// IsListening returns true if the subject state is listening.
//
// Returns:
//   - bool: true if subject is listening, false otherwise.
func (ss *SubjectStatus) IsListening() bool {
	// Delegate to SubjectState method.
	return ss.State.IsListening()
}

// ListenerStatus is an alias for SubjectStatus for backward compatibility.
// Deprecated: Use SubjectStatus instead.
type ListenerStatus = SubjectStatus

// NewListenerStatus creates a new listener status (backward compatibility).
// Deprecated: Use NewSubjectStatusFromState instead.
//
// Params:
//   - name: the listener name.
//   - state: the listener's current state.
//
// Returns:
//   - SubjectStatus: a new subject status for the listener.
func NewListenerStatus(name string, state SubjectState) SubjectStatus {
	// Delegate to new function.
	return NewSubjectStatusFromState(name, state)
}
