//go:build !windows

// Package probe_test provides black-box tests for the probe package.
// This file contains Unix-specific tests for the exec prober.
package probe_test

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domainprobe "github.com/kodflow/daemon/internal/domain/probe"
	"github.com/kodflow/daemon/internal/infrastructure/probe"
)

// TestNewExecProber tests Exec prober creation.
func TestNewExecProber(t *testing.T) {
	tests := []struct {
		name    string
		timeout time.Duration
	}{
		{
			name:    "standard_timeout",
			timeout: 5 * time.Second,
		},
		{
			name:    "short_timeout",
			timeout: 100 * time.Millisecond,
		},
		{
			name:    "zero_timeout",
			timeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Exec prober with specified timeout.
			prober := probe.NewExecProber(tt.timeout)

			// Verify prober is created.
			require.NotNil(t, prober)
		})
	}
}

// TestExecProber_Type tests the Type method.
func TestExecProber_Type(t *testing.T) {
	tests := []struct {
		name     string
		expected string
	}{
		{
			name:     "returns_exec",
			expected: "exec",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Exec prober.
			prober := probe.NewExecProber(time.Second)

			// Verify type identifier.
			assert.Equal(t, tt.expected, prober.Type())
		})
	}
}

// TestExecProber_Probe tests Exec probing functionality.
func TestExecProber_Probe(t *testing.T) {
	tests := []struct {
		name          string
		target        domainprobe.Target
		timeout       time.Duration
		expectSuccess bool
	}{
		{
			name: "successful_true_command",
			target: domainprobe.Target{
				Command: "true",
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "successful_echo_command",
			target: domainprobe.Target{
				Command: "echo",
				Args:    []string{"hello"},
			},
			timeout:       time.Second,
			expectSuccess: true,
		},
		{
			name: "failure_command_with_whitespace_requires_args",
			target: domainprobe.Target{
				Command: "echo hello world",
			},
			timeout:       time.Second,
			expectSuccess: false,
		},
		{
			name: "failure_false_command",
			target: domainprobe.Target{
				Command: "false",
			},
			timeout:       time.Second,
			expectSuccess: false,
		},
		{
			name: "failure_nonexistent_command",
			target: domainprobe.Target{
				Command: "nonexistent_command_12345",
			},
			timeout:       time.Second,
			expectSuccess: false,
		},
		{
			name: "failure_empty_command",
			target: domainprobe.Target{
				Command: "",
			},
			timeout:       time.Second,
			expectSuccess: false,
		},
		{
			name: "failure_whitespace_only_command",
			target: domainprobe.Target{
				Command: "   ",
			},
			timeout:       time.Second,
			expectSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Exec prober.
			prober := probe.NewExecProber(tt.timeout)
			ctx := context.Background()

			// Perform probe.
			result := prober.Probe(ctx, tt.target)

			// Verify result based on expected outcome.
			if tt.expectSuccess {
				assert.True(t, result.Success)
				assert.NoError(t, result.Error)
			} else {
				assert.False(t, result.Success)
			}

			// Latency should always be measured.
			assert.Greater(t, result.Latency, time.Duration(0))
		})
	}
}

// TestExecProber_Probe_Timeout tests command timeout.
func TestExecProber_Probe_Timeout(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "command_times_out",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create prober with short timeout.
			prober := probe.NewExecProber(50 * time.Millisecond)

			target := domainprobe.Target{
				Command: "sleep",
				Args:    []string{"10"},
			}

			// Probe should fail due to timeout.
			result := prober.Probe(context.Background(), target)
			assert.False(t, result.Success)
		})
	}
}

// TestExecProber_Probe_ContextCancellation tests context cancellation.
func TestExecProber_Probe_ContextCancellation(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "cancelled_context",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create prober.
			prober := probe.NewExecProber(10 * time.Second)

			// Create already cancelled context.
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			target := domainprobe.Target{
				Command: "sleep",
				Args:    []string{"10"},
			}

			// Probe should fail due to cancelled context.
			result := prober.Probe(ctx, target)
			assert.False(t, result.Success)
		})
	}
}

// TestExecProber_Probe_OutputCapture tests output capture.
func TestExecProber_Probe_OutputCapture(t *testing.T) {
	tests := []struct {
		name           string
		command        string
		args           []string
		expectedOutput string
	}{
		{
			name:           "captures_stdout",
			command:        "echo",
			args:           []string{"test output"},
			expectedOutput: "test output",
		},
		{
			name:           "captures_multiline",
			command:        "printf",
			args:           []string{"line1\\nline2"},
			expectedOutput: "line1\nline2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create Exec prober.
			prober := probe.NewExecProber(time.Second)

			target := domainprobe.Target{
				Command: tt.command,
				Args:    tt.args,
			}

			// Perform probe.
			result := prober.Probe(context.Background(), target)

			// Verify output captured.
			assert.True(t, result.Success)
			assert.Contains(t, result.Output, tt.expectedOutput)
		})
	}
}
