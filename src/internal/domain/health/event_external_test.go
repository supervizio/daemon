// Package health_test provides black-box tests for the health domain.
package health_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestNewEvent tests the NewEvent constructor function.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestNewEvent(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name     string
		checker  string
		status   health.Status
		message  string
		duration time.Duration
	}{
		{
			name:     "healthy_status",
			checker:  "http-checker",
			status:   health.StatusHealthy,
			message:  "OK",
			duration: 50 * time.Millisecond,
		},
		{
			name:     "unhealthy_status",
			checker:  "tcp-checker",
			status:   health.StatusUnhealthy,
			message:  "connection refused",
			duration: 100 * time.Millisecond,
		},
		{
			name:     "unknown_status",
			checker:  "cmd-checker",
			status:   health.StatusUnknown,
			message:  "timeout",
			duration: 5 * time.Second,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a health result for testing.
			result := health.NewHealthyResult(tt.message, tt.duration)

			// Capture time bounds for timestamp validation.
			before := time.Now()
			event := health.NewEvent(tt.checker, tt.status, result)
			after := time.Now()

			// Assert event fields match expected values.
			assert.Equal(t, tt.checker, event.Checker)
			assert.Equal(t, tt.status, event.Status)
			assert.Equal(t, result.Message, event.Result.Message)
			assert.True(t, event.Timestamp.After(before) || event.Timestamp.Equal(before))
			assert.True(t, event.Timestamp.Before(after) || event.Timestamp.Equal(after))
		})
	}
}

// TestEvent_Fields tests the Event struct field assignments.
//
// Params:
//   - t: testing context for assertions and error reporting
func TestEvent_Fields(t *testing.T) {
	// Define test cases for table-driven testing.
	tests := []struct {
		name    string
		checker string
		status  health.Status
		message string
	}{
		{
			name:    "tcp_unhealthy",
			checker: "tcp-checker",
			status:  health.StatusUnhealthy,
			message: "timeout",
		},
		{
			name:    "http_healthy",
			checker: "http-checker",
			status:  health.StatusHealthy,
			message: "OK",
		},
		{
			name:    "cmd_unknown",
			checker: "cmd-checker",
			status:  health.StatusUnknown,
			message: "script not found",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create timestamp for testing.
			now := time.Now()

			// Create result with test data.
			result := health.Result{
				Status:    tt.status,
				Message:   tt.message,
				Duration:  10 * time.Second,
				Timestamp: now,
			}

			// Create event with test data.
			event := health.Event{
				Checker:   tt.checker,
				Status:    tt.status,
				Result:    result,
				Timestamp: now,
			}

			// Assert all fields are correctly assigned.
			assert.Equal(t, tt.checker, event.Checker)
			assert.Equal(t, tt.status, event.Status)
			assert.Equal(t, result, event.Result)
			assert.Equal(t, now, event.Timestamp)
		})
	}
}
