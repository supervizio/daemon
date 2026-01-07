//go:build unix

// Package process_test provides black-box tests for the infrastructure process package.
// It tests the UnixExecutor implementation of the domain.Executor interface.
package process_test

import (
	"bytes"
	"context"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	domain "github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/process"
)

// TestNewUnixExecutor tests the NewUnixExecutor constructor.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestNewUnixExecutor(t *testing.T) {
	// Define test cases for NewUnixExecutor.
	tests := []struct {
		name string
	}{
		{name: "returns non-nil executor"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := process.NewUnixExecutor()
			// Verify executor is not nil.
			assert.NotNil(t, executor, "NewUnixExecutor should return a non-nil instance")
		})
	}
}

// TestUnixExecutor_Start tests the Start method with valid commands.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestUnixExecutor_Start(t *testing.T) {
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
			executor := process.NewUnixExecutor()
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

// TestUnixExecutor_Start_EmptyCommand tests Start with an empty command.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestUnixExecutor_Start_EmptyCommand(t *testing.T) {
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
			executor := process.NewUnixExecutor()
			ctx := context.Background()

			spec := domain.Spec{
				Command: "",
			}

			_, _, err := executor.Start(ctx, spec)

			// Verify error is returned for empty command.
			assert.Error(t, err)
			// Verify the specific error type.
			assert.ErrorIs(t, err, domain.ErrEmptyCommand)
		})
	}
}

// TestUnixExecutor_Start_WithOutput tests Start with output capture.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestUnixExecutor_Start_WithOutput(t *testing.T) {
	// Define test cases for output capture.
	tests := []struct {
		// name is the test case name.
		name string
		// command is the command to execute.
		command string
		// expected is the expected output.
		expected string
	}{
		{
			name:     "captures stdout",
			command:  "echo hello",
			expected: "hello\n",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			executor := process.NewUnixExecutor()
			ctx := context.Background()

			var stdout bytes.Buffer
			spec := domain.Spec{
				Command: tt.command,
				Stdout:  &stdout,
			}

			pid, wait, err := executor.Start(ctx, spec)

			// Verify no error.
			require.NoError(t, err)
			// Verify PID is positive.
			assert.Greater(t, pid, 0)

			// Wait for process to complete.
			<-wait

			// Verify captured output.
			assert.Equal(t, tt.expected, stdout.String())
		})
	}
}

// TestUnixExecutor_Signal tests the Signal method.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestUnixExecutor_Signal(t *testing.T) {
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
			executor := process.NewUnixExecutor()
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

// TestUnixExecutor_Stop tests the Stop method.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestUnixExecutor_Stop(t *testing.T) {
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
			executor := process.NewUnixExecutor()
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

// TestUnixExecutor_Start_WithWorkingDirectory tests Start with working directory.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestUnixExecutor_Start_WithWorkingDirectory(t *testing.T) {
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
			executor := process.NewUnixExecutor()
			ctx := context.Background()

			var stdout bytes.Buffer
			spec := domain.Spec{
				Command: "pwd",
				Dir:     tt.dir,
				Stdout:  &stdout,
			}

			pid, wait, err := executor.Start(ctx, spec)
			require.NoError(t, err)
			// Verify PID is positive.
			assert.Greater(t, pid, 0)

			// Wait for process to complete.
			<-wait

			// Verify working directory in output.
			assert.Contains(t, stdout.String(), tt.dir)
		})
	}
}

// TestUnixExecutor_Start_InvalidCommand tests Start with an invalid command.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestUnixExecutor_Start_InvalidCommand(t *testing.T) {
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
			executor := process.NewUnixExecutor()
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

// TestUnixExecutor_Start_NonZeroExit tests Start with a command that exits non-zero.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestUnixExecutor_Start_NonZeroExit(t *testing.T) {
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
			executor := process.NewUnixExecutor()
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
