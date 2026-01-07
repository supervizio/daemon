// Package health_test provides external tests for result.go.
// It tests the public API of Result type and factory functions using black-box testing.
package health_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestNewHealthyResult tests the NewHealthyResult factory function.
//
// Params:
//   - t: the testing context.
func TestNewHealthyResult(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// message is the health check message.
		message string
		// duration is the health check duration.
		duration time.Duration
	}{
		{
			name:     "short_duration",
			message:  "check passed",
			duration: 100 * time.Millisecond,
		},
		{
			name:     "long_duration",
			message:  "health check succeeded",
			duration: 5 * time.Second,
		},
		{
			name:     "zero_duration",
			message:  "instant check",
			duration: 0,
		},
		{
			name:     "empty_message",
			message:  "",
			duration: 50 * time.Millisecond,
		},
	}

	// Iterate through all healthy result test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			result := health.NewHealthyResult(tt.message, tt.duration)
			after := time.Now()

			assert.Equal(t, health.StatusHealthy, result.Status)
			assert.Equal(t, tt.message, result.Message)
			assert.Equal(t, tt.duration, result.Duration)
			assert.Nil(t, result.Error)
			assert.True(t, result.Timestamp.After(before) || result.Timestamp.Equal(before))
			assert.True(t, result.Timestamp.Before(after) || result.Timestamp.Equal(after))
		})
	}
}

// TestNewUnhealthyResult tests the NewUnhealthyResult factory function.
//
// Params:
//   - t: the testing context.
func TestNewUnhealthyResult(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// message is the health check message.
		message string
		// duration is the health check duration.
		duration time.Duration
		// err is the error associated with the unhealthy check.
		err error
	}{
		{
			name:     "connection_refused",
			message:  "check failed",
			duration: 5 * time.Second,
			err:      errors.New("connection refused"),
		},
		{
			name:     "timeout_error",
			message:  "health check timed out",
			duration: 30 * time.Second,
			err:      errors.New("timeout"),
		},
		{
			name:     "nil_error",
			message:  "unhealthy without error",
			duration: 100 * time.Millisecond,
			err:      nil,
		},
		{
			name:     "empty_message_with_error",
			message:  "",
			duration: 1 * time.Second,
			err:      errors.New("unknown failure"),
		},
	}

	// Iterate through all unhealthy result test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			before := time.Now()
			result := health.NewUnhealthyResult(tt.message, tt.duration, tt.err)
			after := time.Now()

			assert.Equal(t, health.StatusUnhealthy, result.Status)
			assert.Equal(t, tt.message, result.Message)
			assert.Equal(t, tt.duration, result.Duration)
			assert.Equal(t, tt.err, result.Error)
			assert.True(t, result.Timestamp.After(before) || result.Timestamp.Equal(before))
			assert.True(t, result.Timestamp.Before(after) || result.Timestamp.Equal(after))
		})
	}
}

// TestResult_Fields tests direct field access on Result struct.
//
// Params:
//   - t: the testing context.
func TestResult_Fields(t *testing.T) {
	now := time.Now()

	tests := []struct {
		// name is the test case name.
		name string
		// result is the Result struct to test.
		result health.Result
		// expectedStatus is the expected Status value.
		expectedStatus health.Status
		// expectedMessage is the expected Message value.
		expectedMessage string
		// expectedDuration is the expected Duration value.
		expectedDuration time.Duration
		// expectedTimestamp is the expected Timestamp value.
		expectedTimestamp time.Time
		// expectedError is the expected Error value.
		expectedError error
	}{
		{
			name: "degraded_status",
			result: health.Result{
				Status:    health.StatusDegraded,
				Message:   "partial failure",
				Duration:  200 * time.Millisecond,
				Timestamp: now,
				Error:     nil,
			},
			expectedStatus:    health.StatusDegraded,
			expectedMessage:   "partial failure",
			expectedDuration:  200 * time.Millisecond,
			expectedTimestamp: now,
			expectedError:     nil,
		},
		{
			name: "healthy_status",
			result: health.Result{
				Status:    health.StatusHealthy,
				Message:   "all good",
				Duration:  50 * time.Millisecond,
				Timestamp: now,
				Error:     nil,
			},
			expectedStatus:    health.StatusHealthy,
			expectedMessage:   "all good",
			expectedDuration:  50 * time.Millisecond,
			expectedTimestamp: now,
			expectedError:     nil,
		},
		{
			name: "unhealthy_with_error",
			result: health.Result{
				Status:    health.StatusUnhealthy,
				Message:   "service down",
				Duration:  10 * time.Second,
				Timestamp: now,
				Error:     errors.New("connection failed"),
			},
			expectedStatus:    health.StatusUnhealthy,
			expectedMessage:   "service down",
			expectedDuration:  10 * time.Second,
			expectedTimestamp: now,
			expectedError:     errors.New("connection failed"),
		},
	}

	// Iterate through all result field test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expectedStatus, tt.result.Status)
			assert.Equal(t, tt.expectedMessage, tt.result.Message)
			assert.Equal(t, tt.expectedDuration, tt.result.Duration)
			assert.Equal(t, tt.expectedTimestamp, tt.result.Timestamp)

			// Check error field - compare error messages for non-nil errors.
			if tt.expectedError != nil {
				assert.NotNil(t, tt.result.Error)
				assert.Equal(t, tt.expectedError.Error(), tt.result.Error.Error())
			} else {
				assert.Nil(t, tt.result.Error)
			}
		})
	}
}
