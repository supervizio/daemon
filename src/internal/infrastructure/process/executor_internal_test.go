//go:build unix

// Package process provides internal white-box tests for the infrastructure process package.
// These tests verify the behavior of private functions in the UnixExecutor implementation.
package process

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/kernel"
	"github.com/kodflow/daemon/internal/kernel/ports"
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
		{
			name:    "valid numeric UID succeeds",
			user:    "0",
			group:   "",
			wantErr: false,
		},
		{
			name:    "valid numeric GID succeeds",
			user:    "",
			group:   "0",
			wantErr: false,
		},
		{
			name:    "valid numeric UID and GID succeeds",
			user:    "0",
			group:   "0",
			wantErr: false,
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

// Test_UnixExecutor_buildCommand_Stderr tests buildCommand with stderr configuration.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_buildCommand_Stderr(t *testing.T) {
	// Define test cases for stderr configuration.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{name: "configures stderr writer"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := NewUnixExecutor()
			ctx := context.Background()

			// Create a stderr writer
			stderr := &mockWriter{}
			spec := domain.Spec{
				Command: "echo test",
				Stderr:  stderr,
			}

			// Call the private function
			cmd, err := executor.buildCommand(ctx, spec)

			// Verify no error
			require.NoError(t, err, "should not return error")
			// Verify stderr is set
			assert.Equal(t, stderr, cmd.Stderr, "stderr should be set")
		})
	}
}

// mockWriter is a mock implementation of io.Writer for testing.
type mockWriter struct{}

// Write implements io.Writer.
//
// Params:
//   - p: bytes to write
//
// Returns:
//   - int: number of bytes written
//   - error: always nil
func (m *mockWriter) Write(p []byte) (int, error) {
	// Return successful write
	return len(p), nil
}

// mockCredentialManager is a mock implementation of CredentialManager for testing.
type mockCredentialManager struct {
	// resolveErr is the error to return from ResolveCredentials.
	resolveErr error
	// applyErr is the error to return from ApplyCredentials.
	applyErr error
	// resolvedUID is the UID to return from ResolveCredentials.
	resolvedUID uint32
	// resolvedGID is the GID to return from ResolveCredentials.
	resolvedGID uint32
}

// LookupUser implements CredentialManager.
//
// Params:
//   - nameOrID: the username or UID to look up
//
// Returns:
//   - *ports.User: nil (not implemented for this mock)
//   - error: always nil
func (m *mockCredentialManager) LookupUser(nameOrID string) (*ports.User, error) {
	// Return nil for this mock
	return nil, nil
}

// LookupGroup implements CredentialManager.
//
// Params:
//   - nameOrID: the group name or GID to look up
//
// Returns:
//   - *ports.Group: nil (not implemented for this mock)
//   - error: always nil
func (m *mockCredentialManager) LookupGroup(nameOrID string) (*ports.Group, error) {
	// Return nil for this mock
	return nil, nil
}

// ResolveCredentials implements CredentialManager.
//
// Params:
//   - username: the username to resolve
//   - groupname: the group name to resolve
//
// Returns:
//   - uid: the configured resolved UID
//   - gid: the configured resolved GID
//   - err: the configured resolve error
func (m *mockCredentialManager) ResolveCredentials(username, groupname string) (uid, gid uint32, err error) {
	// Return configured values
	return m.resolvedUID, m.resolvedGID, m.resolveErr
}

// ApplyCredentials implements CredentialManager.
//
// Params:
//   - cmd: the command to apply credentials to
//   - uid: the user ID to set
//   - gid: the group ID to set
//
// Returns:
//   - error: the configured apply error
func (m *mockCredentialManager) ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error {
	// Return configured error
	return m.applyErr
}

// Test_NewUnixExecutorWithKernel tests the NewUnixExecutorWithKernel constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_NewUnixExecutorWithKernel(t *testing.T) {
	// Define test cases for NewUnixExecutorWithKernel.
	tests := []struct {
		name string
	}{
		{name: "creates executor with custom kernel"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			customKernel := kernel.New()
			executor := NewUnixExecutorWithKernel(customKernel)
			// Verify executor is not nil.
			assert.NotNil(t, executor, "should return non-nil executor")
			// Verify kernel is set correctly.
			assert.Equal(t, customKernel, executor.kernel, "kernel should be set")
		})
	}
}

// Test_UnixExecutor_configureCredentials_ApplyError tests configureCredentials with ApplyCredentials error.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_configureCredentials_ApplyError(t *testing.T) {
	// Define test cases for ApplyCredentials error.
	tests := []struct {
		// name is the test case name.
		name string
		// applyErr is the error to return from ApplyCredentials.
		applyErr error
	}{
		{
			name:     "returns error when ApplyCredentials fails",
			applyErr: errors.New("apply credentials failed"),
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create mock credential manager that returns error on apply
			mockCreds := &mockCredentialManager{
				resolvedUID: 1000,
				resolvedGID: 1000,
				applyErr:    tt.applyErr,
			}

			// Create custom kernel with mock credentials
			customKernel := &kernel.Kernel{
				Credentials: mockCreds,
				Process:     kernel.Default.Process,
				Signals:     kernel.Default.Signals,
				Reaper:      kernel.Default.Reaper,
			}

			executor := NewUnixExecutorWithKernel(customKernel)
			ctx := context.Background()

			// Build a command to configure
			cmd, err := executor.buildCommand(ctx, domain.Spec{Command: "echo test"})
			require.NoError(t, err)

			// Call the private function with valid user (triggers apply)
			err = executor.configureCredentials(cmd, "testuser", "")

			// Verify error is returned
			assert.Error(t, err, "should return error when ApplyCredentials fails")
			// Verify error message contains expected text
			assert.Contains(t, err.Error(), "applying credentials", "error should mention applying credentials")
		})
	}
}

// Test_UnixExecutor_Start_ApplyCredentialsError tests Start with ApplyCredentials failure.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_Start_ApplyCredentialsError(t *testing.T) {
	// Define test cases for ApplyCredentials error during Start.
	tests := []struct {
		name string
	}{
		{name: "returns error when ApplyCredentials fails during Start"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create mock credential manager that returns error on apply
			mockCreds := &mockCredentialManager{
				resolvedUID: 1000,
				resolvedGID: 1000,
				applyErr:    errors.New("mock apply error"),
			}

			// Create custom kernel with mock credentials
			customKernel := &kernel.Kernel{
				Credentials: mockCreds,
				Process:     kernel.Default.Process,
				Signals:     kernel.Default.Signals,
				Reaper:      kernel.Default.Reaper,
			}

			executor := NewUnixExecutorWithKernel(customKernel)
			ctx := context.Background()

			spec := domain.Spec{
				Command: "echo hello",
				User:    "testuser",
			}

			pid, wait, err := executor.Start(ctx, spec)

			// Verify error is returned
			assert.Error(t, err, "should return error")
			// Verify PID is zero
			assert.Equal(t, 0, pid, "PID should be zero")
			// Verify wait channel is nil
			assert.Nil(t, wait, "wait channel should be nil")
		})
	}
}

// mockProcess is a mock implementation of Process for testing.
type mockProcess struct {
	// signalErr is the error to return from Signal.
	signalErr error
	// killErr is the error to return from Kill.
	killErr error
	// waitCh is a channel that blocks Wait until closed.
	waitCh chan struct{}
}

// Signal implements Process.
//
// Params:
//   - sig: the signal to send
//
// Returns:
//   - error: the configured signal error
func (m *mockProcess) Signal(sig os.Signal) error {
	// Return configured error
	return m.signalErr
}

// Kill implements Process.
//
// Returns:
//   - error: the configured kill error
func (m *mockProcess) Kill() error {
	// Return configured error
	return m.killErr
}

// Wait implements Process.
//
// Returns:
//   - *os.ProcessState: nil (not used in tests)
//   - error: nil after wait channel is closed
func (m *mockProcess) Wait() (*os.ProcessState, error) {
	// Wait on channel if provided
	if m.waitCh != nil {
		<-m.waitCh
	}
	// Return nil
	return nil, nil
}

// Test_NewUnixExecutorWithOptions tests the NewUnixExecutorWithOptions constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_NewUnixExecutorWithOptions(t *testing.T) {
	// Define test cases for NewUnixExecutorWithOptions.
	tests := []struct {
		name string
	}{
		{name: "creates executor with custom options"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			customKernel := kernel.New()
			customFinder := func(pid int) (Process, error) {
				return nil, errors.New("mock finder")
			}
			executor := NewUnixExecutorWithOptions(customKernel, customFinder)
			// Verify executor is not nil.
			assert.NotNil(t, executor, "should return non-nil executor")
			// Verify kernel is set correctly.
			assert.Equal(t, customKernel, executor.kernel, "kernel should be set")
			// Verify findProcess is set correctly.
			assert.NotNil(t, executor.findProcess, "findProcess should be set")
		})
	}
}

// Test_UnixExecutor_Stop_FindProcessError tests Stop with findProcess error.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_Stop_FindProcessError(t *testing.T) {
	// Define test cases for findProcess error.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{name: "returns error when findProcess fails"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create mock finder that returns error
			mockFinder := func(pid int) (Process, error) {
				return nil, errors.New("process not found")
			}

			executor := NewUnixExecutorWithOptions(kernel.Default, mockFinder)

			// Call Stop with any PID
			err := executor.Stop(12345, time.Second)

			// Verify error is returned
			assert.Error(t, err, "should return error")
			// Verify error message contains expected text
			assert.Contains(t, err.Error(), "finding process", "error should mention finding process")
		})
	}
}

// Test_UnixExecutor_Stop_KillError tests Stop with Kill error during timeout.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_Stop_KillError(t *testing.T) {
	// Define test cases for Kill error.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{name: "returns error when Kill fails during timeout"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create wait channel that will block forever
			waitCh := make(chan struct{})

			// Create mock process that returns error on Kill
			mockProc := &mockProcess{
				signalErr: nil,
				killErr:   errors.New("kill failed"),
				waitCh:    waitCh,
			}

			// Create mock finder that returns mock process
			mockFinder := func(pid int) (Process, error) {
				return mockProc, nil
			}

			executor := NewUnixExecutorWithOptions(kernel.Default, mockFinder)

			// Call Stop with short timeout (will trigger Kill)
			err := executor.Stop(12345, 10*time.Millisecond)

			// Close wait channel to unblock goroutine
			close(waitCh)

			// Verify error is returned
			assert.Error(t, err, "should return error")
			// Verify error message contains expected text
			assert.Contains(t, err.Error(), "killing process", "error should mention killing process")
		})
	}
}

// Test_UnixExecutor_Signal_FindProcessError tests Signal with findProcess error.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_UnixExecutor_Signal_FindProcessError(t *testing.T) {
	// Define test cases for findProcess error.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{name: "returns error when findProcess fails"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create mock finder that returns error
			mockFinder := func(pid int) (Process, error) {
				return nil, errors.New("process not found")
			}

			executor := NewUnixExecutorWithOptions(kernel.Default, mockFinder)

			// Call Signal with any PID
			err := executor.Signal(12345, syscall.SIGTERM)

			// Verify error is returned
			assert.Error(t, err, "should return error")
			// Verify error message contains expected text
			assert.Contains(t, err.Error(), "finding process", "error should mention finding process")
		})
	}
}

// Test_osProcessWrapper tests the osProcessWrapper methods.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_osProcessWrapper(t *testing.T) {
	// Test Signal method
	t.Run("Signal delegates to underlying process", func(t *testing.T) {
		// Get current process for testing
		proc, err := os.FindProcess(os.Getpid())
		require.NoError(t, err)

		wrapper := &osProcessWrapper{proc: proc}

		// Test Signal (send signal 0 which checks if process exists)
		err = wrapper.Signal(syscall.Signal(0))
		assert.NoError(t, err, "Signal should succeed for current process")
	})

	// Test Kill method
	t.Run("Kill delegates to underlying process", func(t *testing.T) {
		// Start a child process that we can kill
		cmd := exec.Command("sleep", "60")
		err := cmd.Start()
		require.NoError(t, err)

		// Get process wrapper
		wrapper := &osProcessWrapper{proc: cmd.Process}

		// Kill the process
		err = wrapper.Kill()
		assert.NoError(t, err, "Kill should succeed")

		// Wait for the process to avoid zombie
		_, _ = wrapper.Wait()
	})

	// Test Wait method
	t.Run("Wait delegates to underlying process", func(t *testing.T) {
		// Start a child process that exits quickly
		cmd := exec.Command("true")
		err := cmd.Start()
		require.NoError(t, err)

		// Get process wrapper
		wrapper := &osProcessWrapper{proc: cmd.Process}

		// Wait for the process
		state, err := wrapper.Wait()
		assert.NoError(t, err, "Wait should succeed")
		assert.NotNil(t, state, "ProcessState should not be nil")
	})
}

// Test_defaultFindProcess tests the defaultFindProcess function.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func Test_defaultFindProcess(t *testing.T) {
	// Test finding current process
	t.Run("finds current process", func(t *testing.T) {
		proc, err := defaultFindProcess(os.Getpid())
		assert.NoError(t, err, "should find current process")
		assert.NotNil(t, proc, "process should not be nil")
	})
}
