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

// TestState_IsRunning tests the IsRunning method of State type.
//
// Params:
//   - t: the testing context.
func TestState_IsRunning(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state to test.
		state process.State
		// expected is the expected result.
		expected bool
	}{
		{"running returns true", process.StateRunning, true},
		{"stopped returns false", process.StateStopped, false},
		{"starting returns false", process.StateStarting, false},
		{"stopping returns false", process.StateStopping, false},
		{"failed returns false", process.StateFailed, false},
	}

	// Iterate through all IsRunning test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.IsRunning())
		})
	}
}

// TestState_IsStopping tests the IsStopping method of State type.
//
// Params:
//   - t: the testing context.
func TestState_IsStopping(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state to test.
		state process.State
		// expected is the expected result.
		expected bool
	}{
		{"stopping returns true", process.StateStopping, true},
		{"running returns false", process.StateRunning, false},
		{"stopped returns false", process.StateStopped, false},
		{"starting returns false", process.StateStarting, false},
		{"failed returns false", process.StateFailed, false},
	}

	// Iterate through all IsStopping test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.IsStopping())
		})
	}
}

// TestState_IsStarting tests the IsStarting method of State type.
//
// Params:
//   - t: the testing context.
func TestState_IsStarting(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state to test.
		state process.State
		// expected is the expected result.
		expected bool
	}{
		{"starting returns true", process.StateStarting, true},
		{"running returns false", process.StateRunning, false},
		{"stopped returns false", process.StateStopped, false},
		{"stopping returns false", process.StateStopping, false},
		{"failed returns false", process.StateFailed, false},
	}

	// Iterate through all IsStarting test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.IsStarting())
		})
	}
}

// TestState_IsFailed tests the IsFailed method of State type.
//
// Params:
//   - t: the testing context.
func TestState_IsFailed(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state to test.
		state process.State
		// expected is the expected result.
		expected bool
	}{
		{"failed returns true", process.StateFailed, true},
		{"running returns false", process.StateRunning, false},
		{"stopped returns false", process.StateStopped, false},
		{"starting returns false", process.StateStarting, false},
		{"stopping returns false", process.StateStopping, false},
	}

	// Iterate through all IsFailed test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.IsFailed())
		})
	}
}

// TestState_IsStopped tests the IsStopped method of State type.
//
// Params:
//   - t: the testing context.
func TestState_IsStopped(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// state is the process state to test.
		state process.State
		// expected is the expected result.
		expected bool
	}{
		{"stopped returns true", process.StateStopped, true},
		{"running returns false", process.StateRunning, false},
		{"starting returns false", process.StateStarting, false},
		{"stopping returns false", process.StateStopping, false},
		{"failed returns false", process.StateFailed, false},
	}

	// Iterate through all IsStopped test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.state.IsStopped())
		})
	}
}
