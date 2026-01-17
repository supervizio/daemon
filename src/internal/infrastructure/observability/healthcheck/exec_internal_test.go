// Package healthcheck provides internal tests for Exec prober.
package healthcheck

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

// TestExecProber_executeCommand_outputTruncation tests output truncation for large outputs.
func TestExecProber_executeCommand_outputTruncation(t *testing.T) {
	tests := []struct {
		name               string
		command            string
		args               []string
		expectTruncation   bool
		expectOutputSubstr string
	}{
		{
			name:    "large_output_truncated_on_failure",
			command: "sh",
			// Generate 5KB of output (more than maxOutputBytes of 4KB) then fail.
			args:               []string{"-c", "yes | head -c 5000; exit 1"},
			expectTruncation:   true,
			expectOutputSubstr: "[truncated]",
		},
		{
			name:               "small_output_not_truncated_on_failure",
			command:            "sh",
			args:               []string{"-c", "echo 'small error'; exit 1"},
			expectTruncation:   false,
			expectOutputSubstr: "small error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Exec prober.
			prober := NewExecProber(5 * time.Second)
			ctx := context.Background()
			start := time.Now()

			// Call internal executeCommand method.
			result := prober.executeCommand(ctx, tt.command, tt.args, start)

			// Verify failure (all test cases should fail).
			assert.False(t, result.Success)

			// Verify expected output substring.
			assert.Contains(t, result.Output, tt.expectOutputSubstr)

			// Verify truncation indicator.
			if tt.expectTruncation {
				assert.Contains(t, result.Output, "[truncated]")
			} else {
				assert.NotContains(t, result.Output, "[truncated]")
			}
		})
	}
}

// TestExecProber_executeCommand_zeroTimeout tests execution with zero timeout.
func TestExecProber_executeCommand_zeroTimeout(t *testing.T) {
	tests := []struct {
		name          string
		command       string
		args          []string
		expectSuccess bool
	}{
		{
			name:          "zero_timeout_runs_without_context_timeout",
			command:       "true",
			args:          nil,
			expectSuccess: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Exec prober with zero timeout.
			prober := NewExecProber(0)
			ctx := context.Background()
			start := time.Now()

			// Call internal executeCommand method.
			result := prober.executeCommand(ctx, tt.command, tt.args, start)

			// Verify success.
			assert.Equal(t, tt.expectSuccess, result.Success)
		})
	}
}
