//go:build unix

// Package executor_test provides black-box tests for the infrastructure executor package.
// It tests the Executor implementation of the domain.Executor interface.
package executor_test

import (
	"context"
	"fmt"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/domain/shared"
	"github.com/kodflow/daemon/internal/infrastructure/process/executor"
)

// TestNewExecutor tests the NewExecutor constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestNewExecutor(t *testing.T) {
	// Define test cases for NewExecutor.
	tests := []struct {
		name string
	}{
		{name: "returns non-nil executor"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			exec := executor.NewExecutor()
			// Verify executor is not nil.
			assert.NotNil(t, exec, "NewExecutor should return a non-nil instance")
		})
	}
}

// TestNew tests the New constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestNew(t *testing.T) {
	// Define test cases for New.
	tests := []struct {
		name string
	}{
		{name: "returns non-nil executor"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			exec := executor.New()
			// Verify executor is not nil.
			assert.NotNil(t, exec, "New should return a non-nil instance")
		})
	}
}

// TestNewWithDeps tests the NewWithDeps constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestNewWithDeps(t *testing.T) {
	// Define test cases for NewWithDeps.
	tests := []struct {
		name string
	}{
		{name: "returns non-nil executor with dependencies"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			exec := executor.NewWithDeps(nil, nil)
			// Verify executor is not nil.
			assert.NotNil(t, exec, "NewWithDeps should return a non-nil instance")
		})
	}
}

// TestNewWithOptions tests the NewWithOptions constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestNewWithOptions(t *testing.T) {
	// Define test cases for NewWithOptions.
	tests := []struct {
		name string
	}{
		{name: "returns non-nil executor with custom options"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			exec := executor.NewWithOptions(nil, nil, nil)
			// Verify executor is not nil.
			assert.NotNil(t, exec, "NewWithOptions should return a non-nil instance")
		})
	}
}

// TestExecutor_Start tests the Start method with valid commands.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Start(t *testing.T) {
	// Define test cases for Start.
	tests := []struct {
		// name is the test case name.
		name string
		// spec is the process specification to test.
		spec domain.Spec
		// wantErr indicates if an error is expected.
		wantErr bool
	}{
		{
			name: "simple echo command",
			spec: domain.Spec{
				Command: "echo hello",
			},
			wantErr: false,
		},
		{
			name: "command with args",
			spec: domain.Spec{
				Command: "echo",
				Args:    []string{"hello", "world"},
			},
			wantErr: false,
		},
		{
			name: "command with environment",
			spec: domain.Spec{
				Command: "env",
				Env:     map[string]string{"TEST_VAR": "test_value"},
			},
			wantErr: false,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			pid, wait, err := executor.Start(ctx, tt.spec)

			// Check error expectation.
			if tt.wantErr {
				// Verify error is returned when expected.
				assert.Error(t, err)
				return
			}

			// Verify no error for successful cases.
			require.NoError(t, err)
			// Verify PID is positive.
			assert.Greater(t, pid, 0, "PID should be positive")
			// Verify wait channel is not nil.
			assert.NotNil(t, wait, "wait channel should not be nil")

			// Wait for process to complete.
			result := <-wait
			// Verify process exited successfully.
			assert.Equal(t, 0, result.Code, "exit code should be 0")
		})
	}
}

// TestExecutor_Start_EmptyCommand tests Start with an empty command.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Start_EmptyCommand(t *testing.T) {
	// Define test cases for empty command.
	tests := []struct {
		name string
	}{
		{name: "returns error for empty command"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			spec := domain.Spec{
				Command: "",
			}

			_, _, err := executor.Start(ctx, spec)

			// Verify error is returned for empty command.
			assert.Error(t, err)
			// Verify the specific error type.
			assert.ErrorIs(t, err, shared.ErrEmptyCommand)
		})
	}
}

// TestExecutor_Start_Success tests Start with a simple command.
// Note: I/O capture is handled at infrastructure level, not via domain Spec.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Start_Success(t *testing.T) {
	// Define test cases for basic execution.
	tests := []struct {
		// name is the test case name.
		name string
		// command is the command to execute.
		command string
	}{
		{
			name:    "executes simple command",
			command: "echo hello",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			spec := domain.Spec{
				Command: tt.command,
			}

			pid, wait, err := executor.Start(ctx, spec)

			// Verify no error.
			require.NoError(t, err)
			// Verify PID is positive.
			assert.Greater(t, pid, 0)

			// Wait for process to complete.
			result := <-wait

			// Verify successful exit.
			assert.Equal(t, 0, result.Code)
		})
	}
}

// TestExecutor_Signal tests the Signal method.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Signal(t *testing.T) {
	// Define test cases for Signal.
	tests := []struct {
		name string
	}{
		{name: "can send signal to running process"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			// Start a long-running process.
			spec := domain.Spec{
				Command: "sleep 10",
			}

			pid, wait, err := executor.Start(ctx, spec)
			require.NoError(t, err)

			// Give process time to start.
			time.Sleep(50 * time.Millisecond)

			// Send SIGTERM signal.
			err = executor.Signal(pid, syscall.SIGTERM)
			// Verify signal was sent successfully.
			assert.NoError(t, err)

			// Wait for process to exit.
			result := <-wait
			// Verify process was terminated by signal.
			assert.NotEqual(t, 0, result.Code)
		})
	}
}

// TestExecutor_Stop tests the Stop method.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Stop(t *testing.T) {
	// Define test cases for Stop.
	tests := []struct {
		// name is the test case name.
		name string
		// timeout is the stop timeout.
		timeout time.Duration
	}{
		{
			name:    "graceful stop with timeout",
			timeout: 5 * time.Second,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			// Start a long-running process.
			spec := domain.Spec{
				Command: "sleep 60",
			}

			pid, _, err := executor.Start(ctx, spec)
			require.NoError(t, err)

			// Give process time to start.
			time.Sleep(50 * time.Millisecond)

			// Stop the process.
			err = executor.Stop(pid, tt.timeout)
			// Verify stop completed without error.
			assert.NoError(t, err)
		})
	}
}

// TestExecutor_Start_WithWorkingDirectory tests Start with working directory.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Start_WithWorkingDirectory(t *testing.T) {
	// Define test cases for working directory.
	tests := []struct {
		// name is the test case name.
		name string
		// dir is the working directory.
		dir string
	}{
		{
			name: "runs in specified directory",
			dir:  "/tmp",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			// Create a marker file in the working directory to verify.
			markerFile := fmt.Sprintf("executor_test_%d", time.Now().UnixNano())
			spec := domain.Spec{
				Command: "touch",
				Args:    []string{markerFile},
				Dir:     tt.dir,
			}

			pid, wait, err := executor.Start(ctx, spec)
			require.NoError(t, err)
			// Verify PID is positive.
			assert.Greater(t, pid, 0)

			// Wait for process to complete.
			result := <-wait

			// Verify successful exit.
			assert.Equal(t, 0, result.Code)

			// Verify marker file was created in working directory.
			markerPath := tt.dir + "/" + markerFile
			_, err = os.Stat(markerPath)
			assert.NoError(t, err, "marker file should exist in working directory")

			// Cleanup marker file.
			_ = os.Remove(markerPath)
		})
	}
}

// TestExecutor_Start_InvalidCommand tests Start with an invalid command.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Start_InvalidCommand(t *testing.T) {
	// Define test cases for invalid commands.
	tests := []struct {
		// name is the test case name.
		name string
		// command is the invalid command.
		command string
	}{
		{
			name:    "non-existent command",
			command: "/nonexistent/command/path",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			spec := domain.Spec{
				Command: tt.command,
			}

			_, _, err := executor.Start(ctx, spec)
			// Verify error is returned for invalid command.
			assert.Error(t, err)
		})
	}
}

// TestExecutor_Start_NonZeroExit tests Start with a command that exits non-zero.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Start_NonZeroExit(t *testing.T) {
	// Define test cases for non-zero exit.
	tests := []struct {
		// name is the test case name.
		name string
		// command is the command that exits non-zero.
		command string
		// args contains additional arguments.
		args []string
		// expectedCode is the expected exit code.
		expectedCode int
	}{
		{
			name:         "exit code 1",
			command:      "sh",
			args:         []string{"-c", "exit 1"},
			expectedCode: 1,
		},
		{
			name:         "exit code 42",
			command:      "sh",
			args:         []string{"-c", "exit 42"},
			expectedCode: 42,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			spec := domain.Spec{
				Command: tt.command,
				Args:    tt.args,
			}

			pid, wait, err := executor.Start(ctx, spec)
			require.NoError(t, err)
			// Verify PID is positive.
			assert.Greater(t, pid, 0)

			// Wait for process to complete.
			result := <-wait
			// Verify exit code matches expected.
			assert.Equal(t, tt.expectedCode, result.Code)
		})
	}
}

// TestExecutor_Start_WithCredentialsError tests Start with invalid credentials.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Start_WithCredentialsError(t *testing.T) {
	// Define test cases for credential errors.
	tests := []struct {
		// name is the test case name.
		name string
		// user is the invalid user.
		user string
		// group is the invalid group.
		group string
	}{
		{
			name: "invalid user returns error",
			user: "nonexistent_user_xyz123",
		},
		{
			name:  "invalid group returns error",
			group: "nonexistent_group_xyz123",
		},
		{
			name:  "both invalid returns error",
			user:  "nonexistent_user_xyz123",
			group: "nonexistent_group_xyz123",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			spec := domain.Spec{
				Command: "echo hello",
				User:    tt.user,
				Group:   tt.group,
			}

			pid, wait, err := executor.Start(ctx, spec)

			// Verify error is returned.
			assert.Error(t, err)
			// Verify PID is zero.
			assert.Equal(t, 0, pid)
			// Verify wait channel is nil.
			assert.Nil(t, wait)
		})
	}
}

// TestExecutor_Stop_Timeout tests Stop with timeout forcing kill.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Stop_Timeout(t *testing.T) {
	// Define test cases for stop timeout.
	tests := []struct {
		// name is the test case name.
		name string
		// timeout is the stop timeout.
		timeout time.Duration
	}{
		{
			name:    "kills process after timeout",
			timeout: 100 * time.Millisecond,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			// Start a process that ignores SIGTERM (traps it and continues).
			spec := domain.Spec{
				Command: "sh",
				Args:    []string{"-c", "trap '' TERM; sleep 60"},
			}

			pid, _, err := executor.Start(ctx, spec)
			require.NoError(t, err)

			// Give process time to start and set up trap.
			time.Sleep(100 * time.Millisecond)

			// Stop the process with short timeout (should force kill).
			err = executor.Stop(pid, tt.timeout)
			// Verify stop completed without error.
			assert.NoError(t, err)
		})
	}
}

// TestExecutor_Stop_AlreadyExited tests Stop when process has already exited.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Stop_AlreadyExited(t *testing.T) {
	// Define test cases for already exited process.
	tests := []struct {
		name string
	}{
		{name: "returns error for already exited process"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			// Start a fast-exiting process.
			spec := domain.Spec{
				Command: "true",
			}

			pid, wait, err := executor.Start(ctx, spec)
			require.NoError(t, err)

			// Wait for process to complete.
			<-wait

			// Give OS time to clean up process.
			time.Sleep(50 * time.Millisecond)

			// Try to stop already-exited process.
			err = executor.Stop(pid, time.Second)
			// Verify error is returned (process no longer exists).
			assert.Error(t, err)
		})
	}
}

// TestExecutor_Signal_AlreadyExited tests Signal when process has already exited.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestExecutor_Signal_AlreadyExited(t *testing.T) {
	// Define test cases for signaling exited process.
	tests := []struct {
		name string
	}{
		{name: "returns error for already exited process"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := executor.New()
			ctx := context.Background()

			// Start a fast-exiting process.
			spec := domain.Spec{
				Command: "true",
			}

			pid, wait, err := executor.Start(ctx, spec)
			require.NoError(t, err)

			// Wait for process to complete.
			<-wait

			// Give OS time to clean up process.
			time.Sleep(50 * time.Millisecond)

			// Try to signal already-exited process.
			err = executor.Signal(pid, syscall.SIGTERM)
			// Verify error is returned (process no longer exists).
			assert.Error(t, err)
		})
	}
}
