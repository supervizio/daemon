// Package shared provides common domain types used across multiple domain packages.
package shared_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/shared"
)

// TestDuration_Duration verifies the Duration method returns the underlying time.Duration.
//
// Params:
//   - t: testing context for assertions
func TestDuration_Duration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration shared.Duration
		expected time.Duration
	}{
		{
			name:     "30 seconds",
			duration: shared.Seconds(30),
			expected: 30 * time.Second,
		},
		{
			name:     "zero duration",
			duration: shared.Duration(0),
			expected: 0,
		},
		{
			name:     "1 minute",
			duration: shared.Minutes(1),
			expected: 1 * time.Minute,
		},
		{
			name:     "500 milliseconds",
			duration: shared.FromTimeDuration(500 * time.Millisecond),
			expected: 500 * time.Millisecond,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.duration.Duration())
		})
	}
}

// TestDuration_Seconds verifies the Seconds method returns duration in seconds.
//
// Params:
//   - t: testing context for assertions
func TestDuration_Seconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration shared.Duration
		expected float64
	}{
		{
			name:     "30 seconds",
			duration: shared.Seconds(30),
			expected: 30.0,
		},
		{
			name:     "zero duration",
			duration: shared.Duration(0),
			expected: 0.0,
		},
		{
			name:     "90 seconds",
			duration: shared.Seconds(90),
			expected: 90.0,
		},
		{
			name:     "500 milliseconds",
			duration: shared.FromTimeDuration(500 * time.Millisecond),
			expected: 0.5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.duration.Seconds())
		})
	}
}

// TestDuration_Milliseconds verifies the Milliseconds method returns duration in milliseconds.
//
// Params:
//   - t: testing context for assertions
func TestDuration_Milliseconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration shared.Duration
		expected int64
	}{
		{
			name:     "30 seconds",
			duration: shared.Seconds(30),
			expected: 30000,
		},
		{
			name:     "zero duration",
			duration: shared.Duration(0),
			expected: 0,
		},
		{
			name:     "1 minute",
			duration: shared.Minutes(1),
			expected: 60000,
		},
		{
			name:     "500 milliseconds",
			duration: shared.FromTimeDuration(500 * time.Millisecond),
			expected: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.duration.Milliseconds())
		})
	}
}

// TestDuration_String verifies the String method returns human-readable format.
//
// Params:
//   - t: testing context for assertions
func TestDuration_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		duration shared.Duration
		expected string
	}{
		{
			name:     "30 seconds",
			duration: shared.Seconds(30),
			expected: "30s",
		},
		{
			name:     "zero duration",
			duration: shared.Duration(0),
			expected: "0s",
		},
		{
			name:     "90 seconds (1m30s)",
			duration: shared.Seconds(90),
			expected: "1m30s",
		},
		{
			name:     "1 minute",
			duration: shared.Minutes(1),
			expected: "1m0s",
		},
		{
			name:     "500 milliseconds",
			duration: shared.FromTimeDuration(500 * time.Millisecond),
			expected: "500ms",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.duration.String())
		})
	}
}

// TestSeconds verifies that Seconds constructor creates a Duration from seconds.
//
// Params:
//   - t: testing context for assertions
func TestSeconds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int
		expected time.Duration
	}{
		{name: "10 seconds", input: 10, expected: 10 * time.Second},
		{name: "0 seconds", input: 0, expected: 0},
		{name: "60 seconds", input: 60, expected: 60 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := shared.Seconds(tt.input)
			assert.Equal(t, tt.expected, d.Duration())
		})
	}
}

// TestMinutes verifies that Minutes constructor creates a Duration from minutes.
//
// Params:
//   - t: testing context for assertions
func TestMinutes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    int
		expected time.Duration
	}{
		{name: "5 minutes", input: 5, expected: 5 * time.Minute},
		{name: "0 minutes", input: 0, expected: 0},
		{name: "1 minute", input: 1, expected: 1 * time.Minute},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := shared.Minutes(tt.input)
			assert.Equal(t, tt.expected, d.Duration())
		})
	}
}

// TestFromTimeDuration verifies that FromTimeDuration creates a Duration from time.Duration.
//
// Params:
//   - t: testing context for assertions
func TestFromTimeDuration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    time.Duration
		expected time.Duration
	}{
		{name: "15 seconds", input: 15 * time.Second, expected: 15 * time.Second},
		{name: "0 duration", input: 0, expected: 0},
		{name: "1 hour", input: 1 * time.Hour, expected: 1 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			d := shared.FromTimeDuration(tt.input)
			assert.Equal(t, tt.expected, d.Duration())
		})
	}
}
