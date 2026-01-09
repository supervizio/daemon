// Package health_test provides black-box tests for the health package.
package health_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/listener"
	"github.com/stretchr/testify/assert"
)

// TestNewListenerStatus tests ListenerStatus creation.
func TestNewListenerStatus(t *testing.T) {
	tests := []struct {
		name          string
		listenerName  string
		listenerState listener.State
	}{
		{
			name:          "closed_listener",
			listenerName:  "http",
			listenerState: listener.Closed,
		},
		{
			name:          "listening_listener",
			listenerName:  "grpc",
			listenerState: listener.Listening,
		},
		{
			name:          "ready_listener",
			listenerName:  "admin",
			listenerState: listener.Ready,
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
		listenerState  listener.State
		result         health.Result
		expectedStatus health.Status
	}{
		{
			name:           "set_healthy_result",
			listenerName:   "http",
			listenerState:  listener.Ready,
			result:         health.NewHealthyResult("OK", 100),
			expectedStatus: health.StatusHealthy,
		},
		{
			name:           "set_unhealthy_result",
			listenerName:   "grpc",
			listenerState:  listener.Listening,
			result:         health.NewUnhealthyResult("Connection refused", 100, nil),
			expectedStatus: health.StatusUnhealthy,
		},
		{
			name:           "set_healthy_result_on_closed_listener",
			listenerName:   "admin",
			listenerState:  listener.Closed,
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
			ls := health.NewListenerStatus("http", listener.Listening)
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
			ls := health.NewListenerStatus("http", listener.Ready)
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
		state    listener.State
		expected bool
	}{
		{
			name:     "closed_not_ready",
			state:    listener.Closed,
			expected: false,
		},
		{
			name:     "listening_not_ready",
			state:    listener.Listening,
			expected: false,
		},
		{
			name:     "ready_is_ready",
			state:    listener.Ready,
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
		state    listener.State
		expected bool
	}{
		{
			name:     "closed_not_listening",
			state:    listener.Closed,
			expected: false,
		},
		{
			name:     "listening_is_listening",
			state:    listener.Listening,
			expected: true,
		},
		{
			name:     "ready_is_also_listening",
			state:    listener.Ready,
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
