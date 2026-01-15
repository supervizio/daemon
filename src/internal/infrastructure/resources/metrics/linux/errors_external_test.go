//go:build linux

package linux_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/linux"
)

// TestInvalidPIDError_Error verifies error message formatting.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestInvalidPIDError_Error(t *testing.T) {
	tests := []struct {
		name     string
		pid      int
		expected string
	}{
		{
			name:     "negative PID",
			pid:      -5,
			expected: "invalid pid: -5",
		},
		{
			name:     "zero PID",
			pid:      0,
			expected: "invalid pid: 0",
		},
		{
			name:     "large negative PID",
			pid:      -99999,
			expected: "invalid pid: -99999",
		},
	}

	// Test each scenario.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create error with constructor.
			err := linux.NewInvalidPIDError(tt.pid)

			// Verify error message format.
			if err.Error() != tt.expected {
				t.Errorf("expected %q, got %q", tt.expected, err.Error())
			}
		})
	}
}

// TestNewInvalidPIDError verifies constructor behavior.
//
// Params:
//   - t: testing context
//
// Returns:
//   - none
func TestNewInvalidPIDError(t *testing.T) {
	tests := []struct {
		name string
		pid  int
	}{
		{
			name: "negative PID",
			pid:  -1,
		},
		{
			name: "zero PID",
			pid:  0,
		},
	}

	// Test each scenario.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create error with constructor.
			err := linux.NewInvalidPIDError(tt.pid)

			// Verify error is not nil.
			if err == nil {
				t.Error("expected non-nil error")
				return
			}

			// Verify PID is set correctly.
			if err.PID != tt.pid {
				t.Errorf("expected PID=%d, got %d", tt.pid, err.PID)
			}
		})
	}
}
