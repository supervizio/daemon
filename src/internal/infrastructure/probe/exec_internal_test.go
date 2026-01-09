// Package probe provides internal tests for Exec prober.
package probe

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// TestExecProber_internalFields tests internal struct fields.
func TestExecProber_internalFields(t *testing.T) {
	tests := []struct {
		name            string
		timeout         time.Duration
		expectedTimeout time.Duration
	}{
		{
			name:            "timeout_is_stored",
			timeout:         5 * time.Second,
			expectedTimeout: 5 * time.Second,
		},
		{
			name:            "zero_timeout_is_stored",
			timeout:         0,
			expectedTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Exec prober.
			prober := NewExecProber(tt.timeout)

			// Verify internal timeout field.
			assert.Equal(t, tt.expectedTimeout, prober.timeout)
		})
	}
}

// TestProberTypeExec_constant tests the constant value.
func TestProberTypeExec_constant(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "constant_value",
			expected: "exec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify constant matches expected value.
			assert.Equal(t, tt.expected, proberTypeExec)
		})
	}
}


// TestExecProber_executeCommand tests the internal executeCommand method.
func TestExecProber_executeCommand(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		args          []string
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name:          "successful_true_command",
			command:       "true",
			args:          nil,
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name:          "successful_echo_command",
			command:       "echo",
			args:          []string{"test"},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name:          "failure_false_command",
			command:       "false",
			args:          nil,
			timeout:       time.Second,
			expectSuccess: false,
		},
		{
			name:          "failure_nonexistent_command",
			command:       "nonexistent_command_12345",
			args:          nil,
			timeout:       time.Second,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Exec prober.
			prober := NewExecProber(tt.timeout)
			ctx := context.Background()
			start := time.Now()

			// Call internal executeCommand method.
			result := prober.executeCommand(ctx, tt.command, tt.args, start)

			// Verify result.
			if tt.expectSuccess {
				assert.True(t, result.Success)
			} else {
				assert.False(t, result.Success)
			}

			// Latency should always be positive.
			assert.Greater(t, result.Latency, time.Duration(0))
		})
	}
}
