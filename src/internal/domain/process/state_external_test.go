// Package process_test provides external tests for process state.
// It tests the public API of State type using black-box testing.
package process_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/process"
)

// TestState_String tests the String method of State type.
//
// Params:
//   - t: the testing context.
func TestState_String(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state to test.
		state process.State
		// want is the expected string representation.
		want string
	}{
		{"stopped", process.StateStopped, "stopped"},
		{"starting", process.StateStarting, "starting"},
		{"running", process.StateRunning, "running"},
		{"stopping", process.StateStopping, "stopping"},
		{"failed", process.StateFailed, "failed"},
		{"unknown", process.State(99), "unknown"},
	}

	// Iterate through all state string test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.state.String())
		})
	}
}

// TestState_IsTerminal tests the IsTerminal method of State type.
//
// Params:
//   - t: the testing context.
func TestState_IsTerminal(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state to test.
		state process.State
		// isTerminal is the expected result.
		isTerminal bool
	}{
		{"stopped is terminal", process.StateStopped, true},
		{"failed is terminal", process.StateFailed, true},
		{"starting is not terminal", process.StateStarting, false},
		{"running is not terminal", process.StateRunning, false},
		{"stopping is not terminal", process.StateStopping, false},
	}

	// Iterate through all terminal state test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isTerminal, tt.state.IsTerminal())
		})
	}
}

// TestState_IsActive tests the IsActive method of State type.
//
// Params:
//   - t: the testing context.
func TestState_IsActive(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state to test.
		state process.State
		// isActive is the expected result.
		isActive bool
	}{
		{"starting is active", process.StateStarting, true},
		{"running is active", process.StateRunning, true},
		{"stopped is not active", process.StateStopped, false},
		{"stopping is not active", process.StateStopping, false},
		{"failed is not active", process.StateFailed, false},
	}

	// Iterate through all active state test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.isActive, tt.state.IsActive())
		})
	}
}
