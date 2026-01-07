//go:build unix

// Package adapters provides platform-specific implementations of kernel interfaces.
// This file contains internal (white-box) tests for Unix signal functionality.
// It tests error handling paths that cannot be triggered through normal execution.
package adapters

import (
	"os"
	"syscall"
	"testing"
)

// Test_UnixSignalManager_ForwardToSelf tests the Forward method with self-signaling.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none
func Test_UnixSignalManager_ForwardToSelf(t *testing.T) {
	// Define test cases for Forward to self.
	tests := []struct {
		// name is the test case name.
		name string
		// sig is the signal to forward.
		sig os.Signal
		// expectError indicates if an error is expected.
		expectError bool
	}{
		{
			name:        "forward signal 0 to self succeeds",
			sig:         syscall.Signal(0),
			expectError: false,
		},
		{
			name:        "forward SIGCONT to self succeeds",
			sig:         syscall.SIGCONT,
			expectError: false,
		},
	}

	sm := NewUnixSignalManager()

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.Forward(os.Getpid(), tt.sig)
			// Check error expectation.
			if tt.expectError && err == nil {
				t.Error("expected error")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// Test_UnixSignalManager_ForwardErrorPath tests that the Forward method
// properly handles errors from Signal (since FindProcess never fails on Unix).
//
// Params:
//   - t: the testing context
//
// Returns:
//   - none
func Test_UnixSignalManager_ForwardErrorPath(t *testing.T) {
	// Define test cases that trigger Signal errors.
	tests := []struct {
		// name is the test case name.
		name string
		// pid is the process ID to signal.
		pid int
		// sig is the signal to forward.
		sig os.Signal
		// expectError indicates if an error is expected.
		expectError bool
	}{
		{
			name:        "forward to nonexistent large PID fails",
			pid:         999999998,
			sig:         syscall.SIGTERM,
			expectError: true,
		},
		{
			name:        "forward to PID 0 fails",
			pid:         0,
			sig:         syscall.Signal(0),
			expectError: true,
		},
	}

	sm := NewUnixSignalManager()

	// Iterate over test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := sm.Forward(tt.pid, tt.sig)
			// Check error expectation.
			if tt.expectError && err == nil {
				t.Error("expected error")
			}
			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
