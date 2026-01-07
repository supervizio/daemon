// Package process provides internal white-box tests for the infrastructure process package.
// These tests verify the behavior of private functions in the UnixExecutor implementation.
package process

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/kodflow/daemon/internal/domain/process"
)

// mockCmdWaiter is a mock implementation of cmdWaiter for testing.
type mockCmdWaiter struct {
	// waitErr is the error to return from Wait.
	waitErr error
}

// Wait implements the cmdWaiter interface.
//
// Params:
//   - None
//
// Returns:
//   - error: the configured wait error
func (m *mockCmdWaiter) Wait() error {
	// Return the configured error
	return m.waitErr
}

// Test_UnixExecutor_waitForProcess tests waitForProcess with various outcomes.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_waitForProcess(t *testing.T) {
	// Define test cases for successful waitForProcess.
	tests := []struct {
		// name is the test case name.
		name string
		// waitErr is the error the mock should return.
		waitErr error
		// expectedCode is the expected exit code.
		expectedCode int
		// expectedErr indicates if result should have error.
		expectedErr bool
	}{
		{
			name:         "process exits successfully",
			waitErr:      nil,
			expectedCode: 0,
			expectedErr:  false,
		},
		{
			name:         "process exits with generic error",
			waitErr:      errors.New("command failed"),
			expectedCode: -1,
			expectedErr:  true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := NewUnixExecutor()
			mockCmd := &mockCmdWaiter{waitErr: tt.waitErr}
			waitCh := make(chan domain.ExitResult, 1)

			// Call the private function
			executor.waitForProcess(mockCmd, waitCh)

			// Read result from channel
			result := <-waitCh

			// Verify exit code
			assert.Equal(t, tt.expectedCode, result.Code, "exit code should match expected")

			// Verify error presence
			if tt.expectedErr {
				// Check error is set when expected
				assert.NotNil(t, result.Error, "error should be set")
			} else {
				// Check error is nil when not expected
				assert.Nil(t, result.Error, "error should be nil")
			}

			// Verify channel is closed
			_, ok := <-waitCh
			assert.False(t, ok, "channel should be closed")
		})
	}
}

// Test_UnixExecutor_buildCommand tests buildCommand with various specifications.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_buildCommand(t *testing.T) {
	// Define test cases for buildCommand.
	tests := []struct {
		// name is the test case name.
		name string
		// spec is the process specification.
		spec domain.Spec
		// wantErr indicates if an error is expected.
		wantErr bool
		// expectedErr is the expected error type.
		expectedErr error
		// expectedPath is the expected command path (if no error).
		expectedPath string
		// expectedEnvVar is an expected environment variable (if any).
		expectedEnvVar string
	}{
		{
			name:        "empty string returns error",
			spec:        domain.Spec{Command: ""},
			wantErr:     true,
			expectedErr: domain.ErrEmptyCommand,
		},
		{
			name:        "whitespace only returns error",
			spec:        domain.Spec{Command: "   "},
			wantErr:     true,
			expectedErr: domain.ErrEmptyCommand,
		},
		{
			name:         "simple command",
			spec:         domain.Spec{Command: "echo hello"},
			wantErr:      false,
			expectedPath: "echo",
		},
		{
			name:         "command with additional args",
			spec:         domain.Spec{Command: "echo", Args: []string{"hello", "world"}},
			wantErr:      false,
			expectedPath: "echo",
		},
		{
			name:         "command with dir",
			spec:         domain.Spec{Command: "pwd", Dir: "/tmp"},
			wantErr:      false,
			expectedPath: "pwd",
		},
		{
			name:           "single environment variable",
			spec:           domain.Spec{Command: "env", Env: map[string]string{"TEST_VAR": "test_value"}},
			wantErr:        false,
			expectedPath:   "env",
			expectedEnvVar: "TEST_VAR=test_value",
		},
		{
			name:           "environment variable with equals",
			spec:           domain.Spec{Command: "env", Env: map[string]string{"CONFIG": "key=value"}},
			wantErr:        false,
			expectedPath:   "env",
			expectedEnvVar: "CONFIG=key=value",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := NewUnixExecutor()
			ctx := context.Background()

			// Call the private function
			cmd, err := executor.buildCommand(ctx, tt.spec)

			// Check error expectation.
			if tt.wantErr {
				// Verify error is returned
				assert.Error(t, err, "should return error")
				// Verify it's the specific error if expected
				if tt.expectedErr != nil {
					assert.ErrorIs(t, err, tt.expectedErr, "should return expected error")
				}
				// Verify cmd is nil
				assert.Nil(t, cmd, "cmd should be nil on error")
				return
			}

			// Verify no error
			require.NoError(t, err, "should not return error")
			// Verify cmd is not nil
			require.NotNil(t, cmd, "cmd should not be nil")
			// Verify path contains expected command
			assert.Contains(t, cmd.Path, tt.expectedPath, "path should contain expected command")
			// Verify dir if specified
			if tt.spec.Dir != "" {
				assert.Equal(t, tt.spec.Dir, cmd.Dir, "dir should match")
			}
			// Verify environment variable if specified
			if tt.expectedEnvVar != "" {
				assert.Contains(t, cmd.Env, tt.expectedEnvVar, "env should contain expected variable")
			}
		})
	}
}

// Test_UnixExecutor_configureCredentials tests configureCredentials with various user/group combinations.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_configureCredentials(t *testing.T) {
	// Define test cases for credentials configuration.
	tests := []struct {
		// name is the test case name.
		name string
		// user is the user to configure.
		user string
		// group is the group to configure.
		group string
		// wantErr indicates if an error is expected.
		wantErr bool
	}{
		{
			name:    "empty user and group skips configuration",
			user:    "",
			group:   "",
			wantErr: false,
		},
		{
			name:    "empty user with empty group skips configuration",
			user:    "",
			group:   "",
			wantErr: false,
		},
		{
			name:    "invalid user returns error",
			user:    "nonexistent_user_12345",
			group:   "",
			wantErr: true,
		},
		{
			name:    "invalid group returns error",
			user:    "",
			group:   "nonexistent_group_12345",
			wantErr: true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := NewUnixExecutor()
			ctx := context.Background()

			// Build a command to configure
			cmd, err := executor.buildCommand(ctx, domain.Spec{Command: "echo test"})
			require.NoError(t, err)

			// Call the private function
			err = executor.configureCredentials(cmd, tt.user, tt.group)

			// Check error expectation.
			if tt.wantErr {
				// Verify error is returned
				assert.Error(t, err, "should return error for invalid credentials")
			} else {
				// Verify no error for valid/empty credentials
				assert.NoError(t, err, "should not return error")
			}
		})
	}
}
