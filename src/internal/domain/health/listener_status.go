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
	// create subject status from snapshot
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
	// create subject status from name and state
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
	// store last probe result
	ss.LastProbeResult = result
}

// IncrementSuccesses increments the consecutive success count and resets failures.
func (ss *SubjectStatus) IncrementSuccesses() {
	// increment successes and reset failures
	ss.ConsecutiveSuccesses++
	ss.ConsecutiveFailures = 0
}

// IncrementFailures increments the consecutive failure count and resets successes.
func (ss *SubjectStatus) IncrementFailures() {
	// increment failures and reset successes
	ss.ConsecutiveFailures++
	ss.ConsecutiveSuccesses = 0
}

// SetState updates the subject state.
//
// Params:
//   - state: the new state.
func (ss *SubjectStatus) SetState(state SubjectState) {
	// update subject state
	ss.State = state
}

// IsReady returns true if the subject state is ready.
//
// Returns:
//   - bool: true if subject is ready, false otherwise.
func (ss *SubjectStatus) IsReady() bool {
	// delegate to state check
	return ss.State.IsReady()
}

// IsListening returns true if the subject state is listening.
//
// Returns:
//   - bool: true if subject is listening, false otherwise.
func (ss *SubjectStatus) IsListening() bool {
	// delegate to state check
	return ss.State.IsListening()
}

// EvaluateProbeResult evaluates a probe result WITHOUT mutating state.
// This is a PURE function - it only computes what should happen.
// Call ApplyProbeEvaluation to apply the result after confirming
// that any external state (e.g., Listener) accepts the transition.
//
// Params:
//   - success: whether the probe succeeded.
//   - successThreshold: consecutive successes needed for Ready.
//   - failureThreshold: consecutive failures needed for Listening.
//
// Returns:
//   - ProbeEvaluation: computed next state and counters.
func (ss *SubjectStatus) EvaluateProbeResult(success bool, successThreshold, failureThreshold int) ProbeEvaluation {
	currentSuccesses := ss.ConsecutiveSuccesses
	currentFailures := ss.ConsecutiveFailures

	// handle successful probe
	if success {
		newSuccesses := currentSuccesses + 1
		// check if threshold met for ready transition
		if newSuccesses >= successThreshold {
			// transition to ready state
			return ProbeEvaluation{
				ShouldTransition: true,
				TargetState:      SubjectReady,
				NewSuccessCount:  newSuccesses,
				NewFailureCount:  0,
			}
		}
		// increment successes but no transition
		return ProbeEvaluation{
			ShouldTransition: false,
			TargetState:      ss.State,
			NewSuccessCount:  newSuccesses,
			NewFailureCount:  0,
		}
	}

	newFailures := currentFailures + 1
	// check if threshold met for listening transition
	if newFailures >= failureThreshold {
		// transition to listening state
		return ProbeEvaluation{
			ShouldTransition: true,
			TargetState:      SubjectListening,
			NewSuccessCount:  0,
			NewFailureCount:  newFailures,
		}
	}
	// increment failures but no transition
	return ProbeEvaluation{
		ShouldTransition: false,
		TargetState:      ss.State,
		NewSuccessCount:  0,
		NewFailureCount:  newFailures,
	}
}

// ApplyProbeEvaluation applies a previously computed evaluation.
// Call this ONLY after confirming any external state accepts the transition.
// This maintains consistency between SubjectStatus and external state.
//
// Params:
//   - eval: the evaluation result from EvaluateProbeResult.
func (ss *SubjectStatus) ApplyProbeEvaluation(eval ProbeEvaluation) {
	// update success and failure counters
	ss.ConsecutiveSuccesses = eval.NewSuccessCount
	ss.ConsecutiveFailures = eval.NewFailureCount
	// apply state transition if required
	if eval.ShouldTransition {
		ss.State = eval.TargetState
	}
}

// ResetCounters resets both consecutive success and failure counts to zero.
// Use this when external state rejects a transition to avoid counter drift.
func (ss *SubjectStatus) ResetCounters() {
	// reset both counters to zero
	ss.ConsecutiveSuccesses = 0
	ss.ConsecutiveFailures = 0
}

// ListenerStatus is an alias for SubjectStatus for backward compatibility.
//
// Deprecated: Use SubjectStatus instead.
type ListenerStatus = SubjectStatus

// NewListenerStatus creates a new listener status (backward compatibility).
//
// Params:
//   - name: the listener name.
//   - state: the listener's current state.
//
// Returns:
//   - SubjectStatus: a new subject status for the listener.
//
// Deprecated: Use NewSubjectStatusFromState instead.
func NewListenerStatus(name string, state SubjectState) SubjectStatus {
	// delegate to new constructor
	return NewSubjectStatusFromState(name, state)
}
