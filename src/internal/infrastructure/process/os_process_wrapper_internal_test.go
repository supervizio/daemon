//go:build unix

// Package process provides internal white-box tests for the os_process_wrapper.
// These tests verify the behavior of osProcessWrapper and defaultFindProcess.
package process

import (
	"os"
	"os/exec"
	"syscall"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test_osProcessWrapper_Signal tests the osProcessWrapper Signal method.
//
// Params:
//   - t: the testing context.
//
// Returns:
//   - (none, test function)
func Test_osProcessWrapper_Signal(t *testing.T) {
	// Define test cases for Signal method.
	tests := []struct {
		// name is the test case name.
		name string
		// signal is the signal to send.
		signal syscall.Signal
		// wantErr indicates if an error is expected.
		wantErr bool
	}{
		{
			name:    "Signal 0 checks process existence",
			signal:  syscall.Signal(0),
			wantErr: false,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Get current process for testing
			proc, err := os.FindProcess(os.Getpid())
			require.NoError(t, err)

			wrapper := &osProcessWrapper{proc: proc}

			// Test Signal
			err = wrapper.Signal(tt.signal)

			// Check error expectation.
			if tt.wantErr {
				// Verify error is returned when expected
				assert.Error(t, err, "Signal should return error")
			} else {
				// Verify no error for successful signal
				assert.NoError(t, err, "Signal should succeed")
			}
		})
	}
}

// Test_osProcessWrapper_Kill tests the osProcessWrapper Kill method.
//
// Params:
//   - t: the testing context.
//
// Returns:
//   - (none, test function)
func Test_osProcessWrapper_Kill(t *testing.T) {
	// Define test cases for Kill method.
	tests := []struct {
		// name is the test case name.
		name string
	}{
		{name: "Kill delegates to underlying process"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
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
	}
}

// Test_osProcessWrapper_Wait tests the osProcessWrapper Wait method.
//
// Params:
//   - t: the testing context.
//
// Returns:
//   - (none, test function)
func Test_osProcessWrapper_Wait(t *testing.T) {
	// Define test cases for Wait method.
	tests := []struct {
		// name is the test case name.
		name string
		// command is the command to run.
		command string
	}{
		{
			name:    "Wait returns process state after exit",
			command: "true",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Start a child process that exits quickly
			cmd := exec.Command(tt.command)
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
}

// Test_defaultFindProcess tests the defaultFindProcess function.
//
// Params:
//   - t: the testing context.
//
// Returns:
//   - (none, test function)
func Test_defaultFindProcess(t *testing.T) {
	// Define test cases for defaultFindProcess.
	tests := []struct {
		// name is the test case name.
		name string
		// getPid returns the PID to find.
		getPid func() int
		// wantErr indicates if an error is expected.
		wantErr bool
	}{
		{
			name:    "finds current process",
			getPid:  os.Getpid,
			wantErr: false,
		},
		{
			name:    "handles arbitrary PID",
			getPid:  func() int { return 1 },
			wantErr: false,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			pid := tt.getPid()
			proc, err := defaultFindProcess(pid)

			// Check error expectation.
			if tt.wantErr {
				// Verify error is returned when expected
				assert.Error(t, err, "should return error")
			} else {
				// Verify no error
				assert.NoError(t, err, "should not return error")
				// Verify process is not nil
				assert.NotNil(t, proc, "process should not be nil")
			}
		})
	}
}
