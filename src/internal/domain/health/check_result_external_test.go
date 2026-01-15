// Package health_test provides black-box tests for the health package.
package health_test

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/health"
)

// TestNewSuccessCheckResult tests successful result creation.
func TestNewSuccessCheckResult(t *testing.T) {
	tests := []struct {
		name            string
		latency         time.Duration
		output          string
		expectedSuccess bool
	}{
		{
			name:            "with_output",
			latency:         100 * time.Millisecond,
			output:          "OK",
			expectedSuccess: true,
		},
		{
			name:            "without_output",
			latency:         50 * time.Millisecond,
			output:          "",
			expectedSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create result.
			result := health.NewSuccessCheckResult(tt.latency, tt.output)

			// Verify fields.
			assert.Equal(t, tt.expectedSuccess, result.Success)
			assert.Equal(t, tt.latency, result.Latency)
			assert.Equal(t, tt.output, result.Output)
			assert.Nil(t, result.Error)
		})
	}
}

// TestNewFailureCheckResult tests failed result creation.
func TestNewFailureCheckResult(t *testing.T) {
	tests := []struct {
		name            string
		latency         time.Duration
		output          string
		err             error
		expectedSuccess bool
	}{
		{
			name:            "with_error",
			latency:         100 * time.Millisecond,
			output:          "",
			err:             errors.New("connection refused"),
			expectedSuccess: false,
		},
		{
			name:            "timeout_error",
			latency:         5 * time.Second,
			output:          "",
			err:             health.ErrProbeTimeout,
			expectedSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create result.
			result := health.NewFailureCheckResult(tt.latency, tt.output, tt.err)

			// Verify fields.
			assert.Equal(t, tt.expectedSuccess, result.Success)
			assert.Equal(t, tt.latency, result.Latency)
			assert.Equal(t, tt.output, result.Output)
			assert.Equal(t, tt.err, result.Error)
		})
	}
}

// TestCheckResult_IsSuccess tests the IsSuccess method.
func TestCheckResult_IsSuccess(t *testing.T) {
	tests := []struct {
		name     string
		result   health.CheckResult
		expected bool
	}{
		{
			name:     "success",
			result:   health.NewSuccessCheckResult(time.Millisecond, ""),
			expected: true,
		},
		{
			name:     "failure",
			result:   health.NewFailureCheckResult(time.Millisecond, "", errors.New("fail")),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify IsSuccess.
			assert.Equal(t, tt.expected, tt.result.IsSuccess())
		})
	}
}

// TestCheckResult_IsFailure tests the IsFailure method.
func TestCheckResult_IsFailure(t *testing.T) {
	tests := []struct {
		name     string
		result   health.CheckResult
		expected bool
	}{
		{
			name:     "success",
			result:   health.NewSuccessCheckResult(time.Millisecond, ""),
			expected: false,
		},
		{
			name:     "failure",
			result:   health.NewFailureCheckResult(time.Millisecond, "", errors.New("fail")),
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify IsFailure.
			assert.Equal(t, tt.expected, tt.result.IsFailure())
		})
	}
}
