// Package health_test provides black-box tests for the health package.
package health_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestNewSubjectSnapshot tests the NewSubjectSnapshot constructor.
//
// Params:
//   - t: the testing context.
func TestNewSubjectSnapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		subjName string
		kind     string
		state    health.SubjectState
	}{
		{
			name:     "listener_ready",
			subjName: "http",
			kind:     "listener",
			state:    health.SubjectReady,
		},
		{
			name:     "process_running",
			subjName: "worker",
			kind:     "process",
			state:    health.SubjectRunning,
		},
		{
			name:     "listener_listening",
			subjName: "grpc",
			kind:     "listener",
			state:    health.SubjectListening,
		},
		{
			name:     "process_stopped",
			subjName: "daemon",
			kind:     "process",
			state:    health.SubjectStopped,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create snapshot using constructor.
			snapshot := health.NewSubjectSnapshot(tt.subjName, tt.kind, tt.state)

			// Verify fields.
			assert.Equal(t, tt.subjName, snapshot.Name)
			assert.Equal(t, tt.kind, snapshot.Kind)
			assert.Equal(t, tt.state, snapshot.State)
		})
	}
}

// TestSubjectState_IsReady tests the IsReady method on SubjectState.
func TestSubjectState_IsReady(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"ready_is_ready", health.SubjectReady, true},
		{"running_is_ready", health.SubjectRunning, true},
		{"listening_not_ready", health.SubjectListening, false},
		{"closed_not_ready", health.SubjectClosed, false},
		{"stopped_not_ready", health.SubjectStopped, false},
		{"failed_not_ready", health.SubjectFailed, false},
		{"unknown_not_ready", health.SubjectUnknown, false},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Check IsReady directly on SubjectState.
			assert.Equal(t, tt.expected, tt.state.IsReady())
		})
	}
}

// TestSubjectState_IsListening tests the IsListening method on SubjectState.
func TestSubjectState_IsListening(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"listening_is_listening", health.SubjectListening, true},
		{"ready_is_also_listening", health.SubjectReady, true},
		{"closed_not_listening", health.SubjectClosed, false},
		{"stopped_not_listening", health.SubjectStopped, false},
		{"failed_not_listening", health.SubjectFailed, false},
		{"running_not_listening", health.SubjectRunning, false},
		{"unknown_not_listening", health.SubjectUnknown, false},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Check IsListening directly on SubjectState.
			assert.Equal(t, tt.expected, tt.state.IsListening())
		})
	}
}

// TestSubjectState_IsClosed tests the IsClosed method on SubjectState.
func TestSubjectState_IsClosed(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"closed_is_closed", health.SubjectClosed, true},
		{"stopped_is_closed", health.SubjectStopped, true},
		{"failed_is_closed", health.SubjectFailed, true},
		{"ready_not_closed", health.SubjectReady, false},
		{"running_not_closed", health.SubjectRunning, false},
		{"listening_not_closed", health.SubjectListening, false},
		{"unknown_not_closed", health.SubjectUnknown, false},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Check IsClosed directly on SubjectState.
			assert.Equal(t, tt.expected, tt.state.IsClosed())
		})
	}
}

// TestSubjectSnapshot_IsReady tests the IsReady method on SubjectSnapshot.
func TestSubjectSnapshot_IsReady(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"ready", health.SubjectReady, true},
		{"running", health.SubjectRunning, true},
		{"listening", health.SubjectListening, false},
		{"closed", health.SubjectClosed, false},
		{"unknown", health.SubjectUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := health.NewSubjectSnapshot("test", "listener", tt.state)
			assert.Equal(t, tt.expected, s.IsReady())
		})
	}
}

func TestSubjectSnapshot_IsListening(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"listening", health.SubjectListening, true},
		{"ready_is_also_listening", health.SubjectReady, true},
		{"closed_not_listening", health.SubjectClosed, false},
		{"stopped_not_listening", health.SubjectStopped, false},
		{"unknown_not_listening", health.SubjectUnknown, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := health.NewSubjectSnapshot("test", "listener", tt.state)
			assert.Equal(t, tt.expected, s.IsListening())
		})
	}
}

func TestSubjectSnapshot_IsClosed(t *testing.T) {
	tests := []struct {
		name     string
		state    health.SubjectState
		expected bool
	}{
		{"closed", health.SubjectClosed, true},
		{"stopped", health.SubjectStopped, true},
		{"failed", health.SubjectFailed, true},
		{"ready", health.SubjectReady, false},
		{"running", health.SubjectRunning, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := health.NewSubjectSnapshot("test", "process", tt.state)
			assert.Equal(t, tt.expected, s.IsClosed())
		})
	}
}
