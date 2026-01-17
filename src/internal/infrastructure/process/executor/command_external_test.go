//go:build unix

// Package executor_test provides black-box tests for the infrastructure executor package.
// It tests the TrustedCommand function for creating exec.Cmd from trusted sources.
package executor_test

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/process/executor"
)

// TestTrustedCommand tests the TrustedCommand function with various command configurations.
//
// Params:
//   - t: the testing context.
//
// Returns:
//   - (none, test function)
func TestTrustedCommand(t *testing.T) {
	// Define test cases for TrustedCommand.
	tests := []struct {
		// name is the test case name.
		name string
		// cmdName is the command executable name.
		cmdName string
		// args are the command arguments.
		args []string
		// expectedPath is the expected path in the command.
		expectedPath string
		// expectedArgs is the expected number of arguments.
		expectedArgs int
	}{
		{
			name:         "simple command without args",
			cmdName:      "echo",
			args:         nil,
			expectedPath: "echo",
			expectedArgs: 0,
		},
		{
			name:         "command with single argument",
			cmdName:      "echo",
			args:         []string{"hello"},
			expectedPath: "echo",
			expectedArgs: 1,
		},
		{
			name:         "command with multiple arguments",
			cmdName:      "echo",
			args:         []string{"hello", "world"},
			expectedPath: "echo",
			expectedArgs: 2,
		},
		{
			name:         "command with absolute path",
			cmdName:      "/bin/sh",
			args:         []string{"-c", "echo test"},
			expectedPath: "/bin/sh",
			expectedArgs: 2,
		},
		{
			name:         "command with nil args slice",
			cmdName:      "true",
			args:         nil,
			expectedPath: "true",
			expectedArgs: 0,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()

			// Call TrustedCommand
			cmd := executor.TrustedCommand(ctx, tt.cmdName, tt.args...)

			// Verify command is not nil
			require.NotNil(t, cmd, "TrustedCommand should return non-nil command")

			// Verify path contains expected command
			assert.Contains(t, cmd.Path, tt.expectedPath, "command path should contain expected name")

			// Verify args length (first arg is always the command itself)
			assert.Equal(t, tt.expectedArgs+1, len(cmd.Args), "args should include command name plus provided args")
		})
	}
}

// TestTrustedCommand_ContextCancellation tests that TrustedCommand respects context cancellation.
//
// Params:
//   - t: the testing context.
//
// Returns:
//   - (none, test function)
func TestTrustedCommand_ContextCancellation(t *testing.T) {
	// Define test cases for context cancellation.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{name: "command respects cancelled context"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())

			// Create a long-running command
			cmd := executor.TrustedCommand(ctx, "sleep", "60")

			// Start the command
			err := cmd.Start()
			require.NoError(t, err, "command should start successfully")

			// Cancel the context
			cancel()

			// Wait for the command to be terminated
			err = cmd.Wait()

			// Verify command was terminated due to context cancellation
			assert.Error(t, err, "command should error when context is cancelled")
		})
	}
}
