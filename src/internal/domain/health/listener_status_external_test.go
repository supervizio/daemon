// Package health_test provides black-box tests for the health package.
package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestNewSubjectStatus tests SubjectStatus creation from snapshot.
func TestNewSubjectStatus(t *testing.T) {
	tests := []struct {
		name     string
		snapshot health.SubjectSnapshot
	}{
		{
			name: "from_listener_snapshot",
			snapshot: health.SubjectSnapshot{
				Name:  "http",
				Kind:  "listener",
				State: health.SubjectListening,
			},
		},
		{
			name: "from_process_snapshot",
			snapshot: health.SubjectSnapshot{
				Name:  "worker",
				Kind:  "process",
				State: health.SubjectRunning,
			},
		},
		{
			name: "from_ready_snapshot",
			snapshot: health.SubjectSnapshot{
				Name:  "admin",
				Kind:  "listener",
				State: health.SubjectReady,
			},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status from snapshot.
			ss := health.NewSubjectStatus(tt.snapshot)

			// Verify fields.
			assert.Equal(t, tt.snapshot.Name, ss.Name)
			assert.Equal(t, tt.snapshot.State, ss.State)
			assert.Nil(t, ss.LastProbeResult)
			assert.Zero(t, ss.ConsecutiveSuccesses)
			assert.Zero(t, ss.ConsecutiveFailures)
		})
	}
}

// TestNewSubjectStatusFromState tests SubjectStatus creation from name and state.
func TestNewSubjectStatusFromState(t *testing.T) {
	tests := []struct {
		name        string
		subjectName string
		state       health.SubjectState
	}{
		{
			name:        "closed_subject",
			subjectName: "http",
			state:       health.SubjectClosed,
		},
		{
			name:        "listening_subject",
			subjectName: "grpc",
			state:       health.SubjectListening,
		},
		{
			name:        "ready_subject",
			subjectName: "admin",
			state:       health.SubjectReady,
		},
		{
			name:        "running_process",
			subjectName: "worker",
			state:       health.SubjectRunning,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status from state.
			ss := health.NewSubjectStatusFromState(tt.subjectName, tt.state)

			// Verify fields.
			assert.Equal(t, tt.subjectName, ss.Name)
			assert.Equal(t, tt.state, ss.State)
			assert.Nil(t, ss.LastProbeResult)
			assert.Zero(t, ss.ConsecutiveSuccesses)
			assert.Zero(t, ss.ConsecutiveFailures)
		})
	}
}

// TestSubjectStatus_SetState tests SetState method.
func TestSubjectStatus_SetState(t *testing.T) {
	tests := []struct {
		name         string
		initialState health.SubjectState
		newState     health.SubjectState
	}{
		{
			name:         "closed_to_listening",
			initialState: health.SubjectClosed,
			newState:     health.SubjectListening,
		},
		{
			name:         "listening_to_ready",
			initialState: health.SubjectListening,
			newState:     health.SubjectReady,
		},
		{
			name:         "ready_to_closed",
			initialState: health.SubjectReady,
			newState:     health.SubjectClosed,
		},
		{
			name:         "running_to_stopped",
			initialState: health.SubjectRunning,
			newState:     health.SubjectStopped,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with initial state.
			ss := health.NewSubjectStatusFromState("subject", tt.initialState)

			// Set new state.
			ss.SetState(tt.newState)

			// Verify state was updated.
			assert.Equal(t, tt.newState, ss.State)
		})
	}
}

// TestSubjectStatus_SetLastProbeResult tests SetLastProbeResult method.
func TestSubjectStatus_SetLastProbeResult(t *testing.T) {
	tests := []struct {
		name           string
		subjectName    string
		state          health.SubjectState
		result         health.Result
		expectedStatus health.Status
	}{
		{
			name:           "set_healthy_result",
			subjectName:    "http",
			state:          health.SubjectReady,
			result:         health.NewHealthyResult("OK", 100),
			expectedStatus: health.StatusHealthy,
		},
		{
			name:           "set_unhealthy_result",
			subjectName:    "grpc",
			state:          health.SubjectListening,
			result:         health.NewUnhealthyResult("Connection refused", 100, nil),
			expectedStatus: health.StatusUnhealthy,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status.
			ss := health.NewSubjectStatusFromState(tt.subjectName, tt.state)

			// Set last probe result.
			result := tt.result
			ss.SetLastProbeResult(&result)

			// Verify result was set.
			assert.NotNil(t, ss.LastProbeResult)
			assert.Equal(t, tt.expectedStatus, ss.LastProbeResult.Status)
		})
	}
}

// TestSubjectStatus_IncrementSuccesses tests IncrementSuccesses method.
func TestSubjectStatus_IncrementSuccesses(t *testing.T) {
	tests := []struct {
		name                       string
		initialFailures            int
		initialSuccesses           int
		incrementCount             int
		expectedSuccesses          int
		expectedFailuresAfterReset int
	}{
		{
			name:                       "increment_from_zero",
			initialFailures:            0,
			initialSuccesses:           0,
			incrementCount:             1,
			expectedSuccesses:          1,
			expectedFailuresAfterReset: 0,
		},
		{
			name:                       "increment_resets_failures",
			initialFailures:            3,
			initialSuccesses:           0,
			incrementCount:             1,
			expectedSuccesses:          1,
			expectedFailuresAfterReset: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with initial state.
			ss := health.NewSubjectStatusFromState("subject", health.SubjectListening)
			ss.ConsecutiveFailures = tt.initialFailures
			ss.ConsecutiveSuccesses = tt.initialSuccesses

			// Increment successes.
			for range tt.incrementCount {
				ss.IncrementSuccesses()
			}

			// Verify expected state.
			assert.Equal(t, tt.expectedSuccesses, ss.ConsecutiveSuccesses)
			assert.Equal(t, tt.expectedFailuresAfterReset, ss.ConsecutiveFailures)
		})
	}
}

// TestSubjectStatus_IncrementFailures tests IncrementFailures method.
func TestSubjectStatus_IncrementFailures(t *testing.T) {
	tests := []struct {
		name                        string
		initialSuccesses            int
		initialFailures             int
		incrementCount              int
		expectedFailures            int
		expectedSuccessesAfterReset int
	}{
		{
			name:                        "increment_from_zero",
			initialSuccesses:            0,
			initialFailures:             0,
			incrementCount:              1,
			expectedFailures:            1,
			expectedSuccessesAfterReset: 0,
		},
		{
			name:                        "increment_resets_successes",
			initialSuccesses:            5,
			initialFailures:             0,
			incrementCount:              1,
			expectedFailures:            1,
			expectedSuccessesAfterReset: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with initial state.
			ss := health.NewSubjectStatusFromState("subject", health.SubjectReady)
			ss.ConsecutiveSuccesses = tt.initialSuccesses
			ss.ConsecutiveFailures = tt.initialFailures

			// Increment failures.
			for range tt.incrementCount {
				ss.IncrementFailures()
			}

			// Verify expected state.
			assert.Equal(t, tt.expectedFailures, ss.ConsecutiveFailures)
			assert.Equal(t, tt.expectedSuccessesAfterReset, ss.ConsecutiveSuccesses)
		})
	}
}

// TestSubjectStatus_IsReady tests IsReady method on SubjectStatus.
func TestSubjectStatus_IsReady(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{
			name:     "closed_not_ready",
			state:    health.SubjectClosed,
			expected: false,
		},
		{
			name:     "listening_not_ready",
			state:    health.SubjectListening,
			expected: false,
		},
		{
			name:     "ready_is_ready",
			state:    health.SubjectReady,
			expected: true,
		},
		{
			name:     "running_is_ready",
			state:    health.SubjectRunning,
			expected: true,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status.
			ss := health.NewSubjectStatusFromState("subject", tt.state)

			// Check if ready.
			assert.Equal(t, tt.expected, ss.IsReady())
		})
	}
}

// TestSubjectStatus_IsListening tests IsListening method on SubjectStatus.
func TestSubjectStatus_IsListening(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{
			name:     "closed_not_listening",
			state:    health.SubjectClosed,
			expected: false,
		},
		{
			name:     "listening_is_listening",
			state:    health.SubjectListening,
			expected: true,
		},
		{
			name:     "ready_is_also_listening",
			state:    health.SubjectReady,
			expected: true,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status.
			ss := health.NewSubjectStatusFromState("subject", tt.state)

			// Check if listening.
			assert.Equal(t, tt.expected, ss.IsListening())
		})
	}
}

// TestNewListenerStatus tests ListenerStatus creation.
func TestNewListenerStatus(t *testing.T) {
	tests := []struct {
		name          string
		listenerName  string
		listenerState health.SubjectState
	}{
		{
			name:          "closed_listener",
			listenerName:  "http",
			listenerState: health.SubjectClosed,
		},
		{
			name:          "listening_listener",
			listenerName:  "grpc",
			listenerState: health.SubjectListening,
		},
		{
			name:          "ready_listener",
			listenerName:  "admin",
			listenerState: health.SubjectReady,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener status.
			ls := health.NewListenerStatus(tt.listenerName, tt.listenerState)

			// Verify fields.
			assert.Equal(t, tt.listenerName, ls.Name)
			assert.Equal(t, tt.listenerState, ls.State)
			assert.Nil(t, ls.LastProbeResult)
			assert.Zero(t, ls.ConsecutiveSuccesses)
			assert.Zero(t, ls.ConsecutiveFailures)
		})
	}
}

// TestListenerStatus_SetLastProbeResult tests SetLastProbeResult method.
func TestListenerStatus_SetLastProbeResult(t *testing.T) {
	tests := []struct {
		name           string
		listenerName   string
		listenerState  health.SubjectState
		result         health.Result
		expectedStatus health.Status
	}{
		{
			name:           "set_healthy_result",
			listenerName:   "http",
			listenerState:  health.SubjectReady,
			result:         health.NewHealthyResult("OK", 100),
			expectedStatus: health.StatusHealthy,
		},
		{
			name:           "set_unhealthy_result",
			listenerName:   "grpc",
			listenerState:  health.SubjectListening,
			result:         health.NewUnhealthyResult("Connection refused", 100, nil),
			expectedStatus: health.StatusUnhealthy,
		},
		{
			name:           "set_healthy_result_on_closed_listener",
			listenerName:   "admin",
			listenerState:  health.SubjectClosed,
			result:         health.NewHealthyResult("OK", 50),
			expectedStatus: health.StatusHealthy,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener status.
			ls := health.NewListenerStatus(tt.listenerName, tt.listenerState)

			// Set last probe result.
			result := tt.result
			ls.SetLastProbeResult(&result)

			// Verify result was set.
			assert.NotNil(t, ls.LastProbeResult)
			assert.Equal(t, tt.expectedStatus, ls.LastProbeResult.Status)
		})
	}
}

// TestListenerStatus_IncrementSuccesses tests IncrementSuccesses method.
func TestListenerStatus_IncrementSuccesses(t *testing.T) {
	tests := []struct {
		name                       string
		initialFailures            int
		initialSuccesses           int
		incrementCount             int
		expectedSuccesses          int
		expectedFailuresAfterReset int
	}{
		{
			name:                       "increment_from_zero",
			initialFailures:            0,
			initialSuccesses:           0,
			incrementCount:             1,
			expectedSuccesses:          1,
			expectedFailuresAfterReset: 0,
		},
		{
			name:                       "increment_resets_failures",
			initialFailures:            3,
			initialSuccesses:           0,
			incrementCount:             1,
			expectedSuccesses:          1,
			expectedFailuresAfterReset: 0,
		},
		{
			name:                       "multiple_increments",
			initialFailures:            0,
			initialSuccesses:           0,
			incrementCount:             3,
			expectedSuccesses:          3,
			expectedFailuresAfterReset: 0,
		},
		{
			name:                       "increment_with_existing_successes",
			initialFailures:            0,
			initialSuccesses:           2,
			incrementCount:             2,
			expectedSuccesses:          4,
			expectedFailuresAfterReset: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener status with initial state.
			ls := health.NewListenerStatus("http", health.SubjectListening)
			ls.ConsecutiveFailures = tt.initialFailures
			ls.ConsecutiveSuccesses = tt.initialSuccesses

			// Increment successes the specified number of times.
			for range tt.incrementCount {
				ls.IncrementSuccesses()
			}

			// Verify expected state.
			assert.Equal(t, tt.expectedSuccesses, ls.ConsecutiveSuccesses)
			assert.Equal(t, tt.expectedFailuresAfterReset, ls.ConsecutiveFailures)
		})
	}
}

// TestListenerStatus_IncrementFailures tests IncrementFailures method.
func TestListenerStatus_IncrementFailures(t *testing.T) {
	tests := []struct {
		name                        string
		initialSuccesses            int
		initialFailures             int
		incrementCount              int
		expectedFailures            int
		expectedSuccessesAfterReset int
	}{
		{
			name:                        "increment_from_zero",
			initialSuccesses:            0,
			initialFailures:             0,
			incrementCount:              1,
			expectedFailures:            1,
			expectedSuccessesAfterReset: 0,
		},
		{
			name:                        "increment_resets_successes",
			initialSuccesses:            5,
			initialFailures:             0,
			incrementCount:              1,
			expectedFailures:            1,
			expectedSuccessesAfterReset: 0,
		},
		{
			name:                        "multiple_increments",
			initialSuccesses:            0,
			initialFailures:             0,
			incrementCount:              3,
			expectedFailures:            3,
			expectedSuccessesAfterReset: 0,
		},
		{
			name:                        "increment_with_existing_failures",
			initialSuccesses:            0,
			initialFailures:             2,
			incrementCount:              2,
			expectedFailures:            4,
			expectedSuccessesAfterReset: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener status with initial state.
			ls := health.NewListenerStatus("http", health.SubjectReady)
			ls.ConsecutiveSuccesses = tt.initialSuccesses
			ls.ConsecutiveFailures = tt.initialFailures

			// Increment failures the specified number of times.
			for range tt.incrementCount {
				ls.IncrementFailures()
			}

			// Verify expected state.
			assert.Equal(t, tt.expectedFailures, ls.ConsecutiveFailures)
			assert.Equal(t, tt.expectedSuccessesAfterReset, ls.ConsecutiveSuccesses)
		})
	}
}

// TestListenerStatus_IsReady tests IsReady method.
func TestListenerStatus_IsReady(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{
			name:     "closed_not_ready",
			state:    health.SubjectClosed,
			expected: false,
		},
		{
			name:     "listening_not_ready",
			state:    health.SubjectListening,
			expected: false,
		},
		{
			name:     "ready_is_ready",
			state:    health.SubjectReady,
			expected: true,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener status.
			ls := health.NewListenerStatus("http", tt.state)

			// Check if ready.
			assert.Equal(t, tt.expected, ls.IsReady())
		})
	}
}

// TestListenerStatus_IsListening tests IsListening method.
func TestListenerStatus_IsListening(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{
			name:     "closed_not_listening",
			state:    health.SubjectClosed,
			expected: false,
		},
		{
			name:     "listening_is_listening",
			state:    health.SubjectListening,
			expected: true,
		},
		{
			name:     "ready_is_also_listening",
			state:    health.SubjectReady,
			expected: true,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create listener status.
			ls := health.NewListenerStatus("http", tt.state)

			// Check if listening.
			assert.Equal(t, tt.expected, ls.IsListening())
		})
	}
}

// TestSubjectStatus_EvaluateProbeResult tests EvaluateProbeResult pure function.
func TestSubjectStatus_EvaluateProbeResult(t *testing.T) {
	// Define test cases for purity verification.
	tests := []struct {
		name             string
		initialState     health.SubjectState
		initialSuccesses int
		initialFailures  int
		success          bool
		successThreshold int
		failureThreshold int
	}{
		{
			name:             "success_with_failures",
			initialState:     health.SubjectListening,
			initialSuccesses: 0,
			initialFailures:  2,
			success:          true,
			successThreshold: 1,
			failureThreshold: 3,
		},
		{
			name:             "failure_with_successes",
			initialState:     health.SubjectReady,
			initialSuccesses: 3,
			initialFailures:  0,
			success:          false,
			successThreshold: 1,
			failureThreshold: 1,
		},
		{
			name:             "success_threshold_met",
			initialState:     health.SubjectListening,
			initialSuccesses: 2,
			initialFailures:  0,
			success:          true,
			successThreshold: 3,
			failureThreshold: 3,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with known state.
			ss := health.NewSubjectStatusFromState("test", tt.initialState)
			ss.ConsecutiveSuccesses = tt.initialSuccesses
			ss.ConsecutiveFailures = tt.initialFailures

			// Call pure evaluation.
			_ = ss.EvaluateProbeResult(tt.success, tt.successThreshold, tt.failureThreshold)

			// Verify no mutations occurred.
			assert.Equal(t, tt.initialSuccesses, ss.ConsecutiveSuccesses, "EvaluateProbeResult should not mutate ConsecutiveSuccesses")
			assert.Equal(t, tt.initialFailures, ss.ConsecutiveFailures, "EvaluateProbeResult should not mutate ConsecutiveFailures")
			assert.Equal(t, tt.initialState, ss.State, "EvaluateProbeResult should not mutate State")
		})
	}
}

// TestSubjectStatus_EvaluateProbeResult_SuccessThreshold tests success threshold logic.
func TestSubjectStatus_EvaluateProbeResult_SuccessThreshold(t *testing.T) {
	tests := []struct {
		name             string
		initialSuccesses int
		initialFailures  int
		success          bool
		successThreshold int
		failureThreshold int
		wantTransition   bool
		wantTargetState  health.SubjectState
		wantSuccessCount int
		wantFailureCount int
	}{
		{
			name:             "success_below_threshold",
			initialSuccesses: 0,
			initialFailures:  0,
			success:          true,
			successThreshold: 3,
			failureThreshold: 3,
			wantTransition:   false,
			wantTargetState:  health.SubjectListening,
			wantSuccessCount: 1,
			wantFailureCount: 0,
		},
		{
			name:             "success_meets_threshold",
			initialSuccesses: 2,
			initialFailures:  0,
			success:          true,
			successThreshold: 3,
			failureThreshold: 3,
			wantTransition:   true,
			wantTargetState:  health.SubjectReady,
			wantSuccessCount: 3,
			wantFailureCount: 0,
		},
		{
			name:             "success_exceeds_threshold",
			initialSuccesses: 5,
			initialFailures:  0,
			success:          true,
			successThreshold: 3,
			failureThreshold: 3,
			wantTransition:   true,
			wantTargetState:  health.SubjectReady,
			wantSuccessCount: 6,
			wantFailureCount: 0,
		},
		{
			name:             "success_resets_failures",
			initialSuccesses: 0,
			initialFailures:  2,
			success:          true,
			successThreshold: 3,
			failureThreshold: 3,
			wantTransition:   false,
			wantTargetState:  health.SubjectListening,
			wantSuccessCount: 1,
			wantFailureCount: 0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with initial state.
			ss := health.NewSubjectStatusFromState("test", health.SubjectListening)
			ss.ConsecutiveSuccesses = tt.initialSuccesses
			ss.ConsecutiveFailures = tt.initialFailures

			// Evaluate probe result.
			eval := ss.EvaluateProbeResult(tt.success, tt.successThreshold, tt.failureThreshold)

			// Verify evaluation result.
			assert.Equal(t, tt.wantTransition, eval.ShouldTransition, "ShouldTransition mismatch")
			assert.Equal(t, tt.wantTargetState, eval.TargetState, "TargetState mismatch")
			assert.Equal(t, tt.wantSuccessCount, eval.NewSuccessCount, "NewSuccessCount mismatch")
			assert.Equal(t, tt.wantFailureCount, eval.NewFailureCount, "NewFailureCount mismatch")
		})
	}
}

// TestSubjectStatus_EvaluateProbeResult_FailureThreshold tests failure threshold logic.
func TestSubjectStatus_EvaluateProbeResult_FailureThreshold(t *testing.T) {
	tests := []struct {
		name             string
		initialSuccesses int
		initialFailures  int
		success          bool
		successThreshold int
		failureThreshold int
		wantTransition   bool
		wantTargetState  health.SubjectState
		wantSuccessCount int
		wantFailureCount int
	}{
		{
			name:             "failure_below_threshold",
			initialSuccesses: 0,
			initialFailures:  0,
			success:          false,
			successThreshold: 3,
			failureThreshold: 3,
			wantTransition:   false,
			wantTargetState:  health.SubjectReady,
			wantSuccessCount: 0,
			wantFailureCount: 1,
		},
		{
			name:             "failure_meets_threshold",
			initialSuccesses: 0,
			initialFailures:  2,
			success:          false,
			successThreshold: 3,
			failureThreshold: 3,
			wantTransition:   true,
			wantTargetState:  health.SubjectListening,
			wantSuccessCount: 0,
			wantFailureCount: 3,
		},
		{
			name:             "failure_exceeds_threshold",
			initialSuccesses: 0,
			initialFailures:  5,
			success:          false,
			successThreshold: 3,
			failureThreshold: 3,
			wantTransition:   true,
			wantTargetState:  health.SubjectListening,
			wantSuccessCount: 0,
			wantFailureCount: 6,
		},
		{
			name:             "failure_resets_successes",
			initialSuccesses: 2,
			initialFailures:  0,
			success:          false,
			successThreshold: 3,
			failureThreshold: 3,
			wantTransition:   false,
			wantTargetState:  health.SubjectReady,
			wantSuccessCount: 0,
			wantFailureCount: 1,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with initial state.
			ss := health.NewSubjectStatusFromState("test", health.SubjectReady)
			ss.ConsecutiveSuccesses = tt.initialSuccesses
			ss.ConsecutiveFailures = tt.initialFailures

			// Evaluate probe result.
			eval := ss.EvaluateProbeResult(tt.success, tt.successThreshold, tt.failureThreshold)

			// Verify evaluation result.
			assert.Equal(t, tt.wantTransition, eval.ShouldTransition, "ShouldTransition mismatch")
			assert.Equal(t, tt.wantTargetState, eval.TargetState, "TargetState mismatch")
			assert.Equal(t, tt.wantSuccessCount, eval.NewSuccessCount, "NewSuccessCount mismatch")
			assert.Equal(t, tt.wantFailureCount, eval.NewFailureCount, "NewFailureCount mismatch")
		})
	}
}

// TestSubjectStatus_ApplyProbeEvaluation tests ApplyProbeEvaluation method.
func TestSubjectStatus_ApplyProbeEvaluation(t *testing.T) {
	tests := []struct {
		name              string
		initialState      health.SubjectState
		eval              health.ProbeEvaluation
		expectedState     health.SubjectState
		expectedSuccesses int
		expectedFailures  int
	}{
		{
			name:         "apply_transition_to_ready",
			initialState: health.SubjectListening,
			eval: health.ProbeEvaluation{
				ShouldTransition: true,
				TargetState:      health.SubjectReady,
				NewSuccessCount:  3,
				NewFailureCount:  0,
			},
			expectedState:     health.SubjectReady,
			expectedSuccesses: 3,
			expectedFailures:  0,
		},
		{
			name:         "apply_transition_to_listening",
			initialState: health.SubjectReady,
			eval: health.ProbeEvaluation{
				ShouldTransition: true,
				TargetState:      health.SubjectListening,
				NewSuccessCount:  0,
				NewFailureCount:  3,
			},
			expectedState:     health.SubjectListening,
			expectedSuccesses: 0,
			expectedFailures:  3,
		},
		{
			name:         "apply_no_transition",
			initialState: health.SubjectListening,
			eval: health.ProbeEvaluation{
				ShouldTransition: false,
				TargetState:      health.SubjectListening,
				NewSuccessCount:  1,
				NewFailureCount:  0,
			},
			expectedState:     health.SubjectListening,
			expectedSuccesses: 1,
			expectedFailures:  0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with initial state.
			ss := health.NewSubjectStatusFromState("test", tt.initialState)

			// Apply evaluation.
			ss.ApplyProbeEvaluation(tt.eval)

			// Verify final state.
			assert.Equal(t, tt.expectedState, ss.State, "State mismatch")
			assert.Equal(t, tt.expectedSuccesses, ss.ConsecutiveSuccesses, "ConsecutiveSuccesses mismatch")
			assert.Equal(t, tt.expectedFailures, ss.ConsecutiveFailures, "ConsecutiveFailures mismatch")
		})
	}
}

// TestSubjectStatus_ResetCounters tests ResetCounters method.
func TestSubjectStatus_ResetCounters(t *testing.T) {
	tests := []struct {
		name             string
		initialSuccesses int
		initialFailures  int
	}{
		{
			name:             "reset_both_nonzero",
			initialSuccesses: 5,
			initialFailures:  3,
		},
		{
			name:             "reset_only_successes",
			initialSuccesses: 10,
			initialFailures:  0,
		},
		{
			name:             "reset_only_failures",
			initialSuccesses: 0,
			initialFailures:  7,
		},
		{
			name:             "reset_already_zero",
			initialSuccesses: 0,
			initialFailures:  0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with initial counters.
			ss := health.NewSubjectStatusFromState("test", health.SubjectListening)
			ss.ConsecutiveSuccesses = tt.initialSuccesses
			ss.ConsecutiveFailures = tt.initialFailures

			// Reset counters.
			ss.ResetCounters()

			// Verify both counters are zero.
			assert.Equal(t, 0, ss.ConsecutiveSuccesses, "ConsecutiveSuccesses should be zero")
			assert.Equal(t, 0, ss.ConsecutiveFailures, "ConsecutiveFailures should be zero")
		})
	}
}

// TestSubjectStatus_EvaluateAndApply_Integration tests the full evaluate-then-apply workflow.
func TestSubjectStatus_EvaluateAndApply_Integration(t *testing.T) {
	tests := []struct {
		name               string
		initialState       health.SubjectState
		probeResults       []bool
		successThreshold   int
		failureThreshold   int
		expectedFinalState health.SubjectState
	}{
		{
			name:               "three_successes_triggers_ready",
			initialState:       health.SubjectListening,
			probeResults:       []bool{true, true, true},
			successThreshold:   3,
			failureThreshold:   3,
			expectedFinalState: health.SubjectReady,
		},
		{
			name:               "three_failures_triggers_listening",
			initialState:       health.SubjectReady,
			probeResults:       []bool{false, false, false},
			successThreshold:   3,
			failureThreshold:   3,
			expectedFinalState: health.SubjectListening,
		},
		{
			name:               "mixed_results_no_transition",
			initialState:       health.SubjectListening,
			probeResults:       []bool{true, true, false, true, true},
			successThreshold:   3,
			failureThreshold:   3,
			expectedFinalState: health.SubjectListening,
		},
		{
			name:               "recovery_after_failures",
			initialState:       health.SubjectListening,
			probeResults:       []bool{false, false, true, true, true},
			successThreshold:   3,
			failureThreshold:   3,
			expectedFinalState: health.SubjectReady,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create subject status with initial state.
			ss := health.NewSubjectStatusFromState("test", tt.initialState)

			// Process each probe result.
			for _, success := range tt.probeResults {
				eval := ss.EvaluateProbeResult(success, tt.successThreshold, tt.failureThreshold)
				ss.ApplyProbeEvaluation(eval)
			}

			// Verify final state.
			assert.Equal(t, tt.expectedFinalState, ss.State, "Final state mismatch")
		})
	}
}
