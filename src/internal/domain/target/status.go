// Package target provides domain entities for external monitoring targets.
// External targets are services, containers, or hosts that supervizio
// monitors but does not manage (no lifecycle control).
package target

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
)

// State represents the health state of an external target.
type State string

// State constants define the possible health states.
const (
	// StateUnknown indicates the target's health is not yet known.
	StateUnknown State = "unknown"

	// StateHealthy indicates the target is responding to probes.
	StateHealthy State = "healthy"

	// StateUnhealthy indicates the target is not responding to probes.
	StateUnhealthy State = "unhealthy"

	// StateDegraded indicates partial health issues.
	StateDegraded State = "degraded"
)

// String returns the string representation of the state.
//
// Returns:
//   - string: the state as a string.
func (s State) String() string {
	// Convert State enum to string for display and serialization.
	return string(s)
}

// IsHealthy checks if the state indicates the target is healthy.
//
// Returns:
//   - bool: true if the state is healthy.
func (s State) IsHealthy() bool {
	// Compare with StateHealthy constant to determine if healthy.
	return s == StateHealthy
}

// Status holds the current health status of an external target.
// It tracks probe results, state transitions, and consecutive
// success/failure counts for threshold-based state management.
type Status struct {
	// TargetID is the unique identifier of the target.
	TargetID string

	// TargetName is the human-readable name.
	TargetName string

	// TargetType is the kind of target.
	TargetType Type

	// State is the current health state.
	State State

	// LastProbeResult is the result of the most recent probe.
	LastProbeResult *health.CheckResult

	// LastProbeTime is when the last probe was executed.
	LastProbeTime time.Time

	// ConsecutiveSuccesses is the count of consecutive successful probes.
	ConsecutiveSuccesses int

	// ConsecutiveFailures is the count of consecutive failed probes.
	ConsecutiveFailures int

	// LastStateChange is when the state last changed.
	LastStateChange time.Time

	// Message provides additional context about the status.
	Message string
}

// NewStatus creates a new status for a target in unknown state.
//
// Params:
//   - target: the external target to create status for.
//
// Returns:
//   - *Status: a new status in unknown state.
func NewStatus(target *ExternalTarget) *Status {
	// Initialize status in unknown state until first probe completes.
	return &Status{
		TargetID:   target.ID,
		TargetName: target.Name,
		TargetType: target.Type,
		State:      StateUnknown,
	}
}

// UpdateFromProbe updates the status based on a probe result.
//
// Params:
//   - result: the probe result.
//   - successThreshold: consecutive successes to mark healthy.
//   - failureThreshold: consecutive failures to mark unhealthy.
func (s *Status) UpdateFromProbe(result health.CheckResult, successThreshold, failureThreshold int) {
	s.LastProbeResult = &result
	s.LastProbeTime = time.Now()

	previousState := s.State

	// Handle successful probe by incrementing success count.
	if result.Success {
		s.ConsecutiveSuccesses++
		s.ConsecutiveFailures = 0
		s.Message = result.Output

		// Transition to healthy if success threshold is met.
		if s.ConsecutiveSuccesses >= successThreshold {
			s.State = StateHealthy
		}
	} else {
		// Handle failed probe by incrementing failure count and extracting error.
		s.ConsecutiveFailures++
		s.ConsecutiveSuccesses = 0

		// Extract error message from probe result if present.
		if result.Error != nil {
			s.Message = result.Error.Error()
		} else {
			// Use probe output as message if no error object.
			s.Message = result.Output
		}

		// Transition to unhealthy if failure threshold is met.
		if s.ConsecutiveFailures >= failureThreshold {
			s.State = StateUnhealthy
		}
	}

	// Record state transition timestamp if state changed.
	if s.State != previousState {
		s.LastStateChange = time.Now()
	}
}

// MarkHealthy sets the status to healthy.
//
// Params:
//   - message: optional message describing the status.
func (s *Status) MarkHealthy(message string) {
	// Record state transition timestamp if changing from non-healthy.
	if s.State != StateHealthy {
		s.LastStateChange = time.Now()
	}
	s.State = StateHealthy
	s.Message = message
	s.ConsecutiveFailures = 0
}

// MarkUnhealthy sets the status to unhealthy.
//
// Params:
//   - message: message describing why the target is unhealthy.
func (s *Status) MarkUnhealthy(message string) {
	// Record state transition timestamp if changing from non-unhealthy.
	if s.State != StateUnhealthy {
		s.LastStateChange = time.Now()
	}
	s.State = StateUnhealthy
	s.Message = message
	s.ConsecutiveSuccesses = 0
}

// Latency returns the latency of the last probe.
//
// Returns:
//   - time.Duration: the probe latency, or 0 if no probe has been executed.
func (s *Status) Latency() time.Duration {
	// Return zero if no probe has been executed yet.
	if s.LastProbeResult == nil {
		// No probe result available yet.
		return 0
	}
	// Return latency from last probe result.
	return s.LastProbeResult.Latency
}

// SinceLastProbe returns the time since the last probe.
//
// Returns:
//   - time.Duration: time since last probe, or 0 if no probe has been executed.
func (s *Status) SinceLastProbe() time.Duration {
	// Return zero if no probe has been executed yet.
	if s.LastProbeTime.IsZero() {
		// Probe time not set yet.
		return 0
	}
	// Calculate time elapsed since last probe.
	return time.Since(s.LastProbeTime)
}

// SinceLastStateChange returns the time since the last state change.
//
// Returns:
//   - time.Duration: time since last state change, or 0 if state never changed.
func (s *Status) SinceLastStateChange() time.Duration {
	// Return zero if state has never changed.
	if s.LastStateChange.IsZero() {
		// State never changed from initial unknown.
		return 0
	}
	// Calculate time elapsed since last state change.
	return time.Since(s.LastStateChange)
}
