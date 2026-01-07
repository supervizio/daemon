// Package process provides domain entities and value objects for process lifecycle management.
package process_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/process"
)

// TestEventType_String verifies string representation of event types.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that each EventType constant returns its expected
// string representation, including an "unknown" fallback for undefined values.
func TestEventType_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		eventType process.EventType
		want      string
	}{
		{"started", process.EventStarted, "started"},
		{"stopped", process.EventStopped, "stopped"},
		{"failed", process.EventFailed, "failed"},
		{"restarting", process.EventRestarting, "restarting"},
		{"healthy", process.EventHealthy, "healthy"},
		{"unhealthy", process.EventUnhealthy, "unhealthy"},
		{"unknown", process.EventType(99), "unknown"},
	}

	// Iterate through all test cases to verify string conversion
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.want, tt.eventType.String())
		})
	}
}

// TestNewEvent verifies the NewEvent constructor creates events correctly.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that NewEvent properly initializes all event fields
// including type, process name, PID, exit code, timestamp, and error.
func TestNewEvent(t *testing.T) {
	t.Parallel()

	testErr := errors.New("test error")

	tests := []struct {
		name        string
		eventType   process.EventType
		processName string
		pid         int
		exitCode    int
		err         error
	}{
		{
			name:        "creates event with all fields",
			eventType:   process.EventStarted,
			processName: "test-service",
			pid:         1234,
			exitCode:    0,
			err:         nil,
		},
		{
			name:        "creates event with error",
			eventType:   process.EventFailed,
			processName: "test-service",
			pid:         0,
			exitCode:    1,
			err:         testErr,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before := time.Now()
			event := process.NewEvent(tt.eventType, tt.processName, tt.pid, tt.exitCode, tt.err)
			after := time.Now()

			assert.Equal(t, tt.eventType, event.Type)
			assert.Equal(t, tt.processName, event.Process)
			assert.Equal(t, tt.pid, event.PID)
			assert.Equal(t, tt.exitCode, event.ExitCode)
			assert.Equal(t, tt.err, event.Error)
			assert.True(t, event.Timestamp.After(before) || event.Timestamp.Equal(before))
			assert.True(t, event.Timestamp.Before(after) || event.Timestamp.Equal(after))
		})
	}
}

// TestEvent_Fields verifies direct field access on Event struct.
//
// Params:
//   - t: testing context for assertions
//
// This test validates that Event struct fields can be accessed directly
// and contain the expected values after initialization.
func TestEvent_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    process.Event
		wantType process.EventType
		wantProc string
		wantPID  int
		wantCode int
	}{
		{
			name: "stopped event fields",
			event: process.Event{
				Type:      process.EventStopped,
				Process:   "my-service",
				PID:       5678,
				ExitCode:  0,
				Timestamp: time.Now(),
				Error:     nil,
			},
			wantType: process.EventStopped,
			wantProc: "my-service",
			wantPID:  5678,
			wantCode: 0,
		},
		{
			name: "failed event fields",
			event: process.Event{
				Type:      process.EventFailed,
				Process:   "other-service",
				PID:       1234,
				ExitCode:  1,
				Timestamp: time.Now(),
				Error:     nil,
			},
			wantType: process.EventFailed,
			wantProc: "other-service",
			wantPID:  1234,
			wantCode: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantType, tt.event.Type)
			assert.Equal(t, tt.wantProc, tt.event.Process)
			assert.Equal(t, tt.wantPID, tt.event.PID)
			assert.Equal(t, tt.wantCode, tt.event.ExitCode)
		})
	}
}
