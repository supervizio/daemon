// Package health provides domain entities and value objects for health checking.
package health

// ProbeEvaluation represents the result of evaluating a probe outcome.
// This is a pure value - no side effects during creation.
// Use EvaluateProbeResult to create, then ApplyProbeEvaluation to mutate.
type ProbeEvaluation struct {
	// ShouldTransition indicates whether a state change is warranted.
	ShouldTransition bool

	// TargetState is the state to transition to if ShouldTransition is true.
	TargetState SubjectState

	// NewSuccessCount is the computed consecutive success count.
	NewSuccessCount int

	// NewFailureCount is the computed consecutive failure count.
	NewFailureCount int
}
