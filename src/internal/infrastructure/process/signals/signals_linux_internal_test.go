//go:build linux

// Package signals provides platform-specific implementations of kernel interfaces.
// This file contains internal (white-box) tests for Linux-specific signal functionality.
package signals

import (
	"os"
	"syscall"
	"testing"
)

// TestPlatformInit tests that platform-specific signals are registered during initialization.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestPlatformInit(t *testing.T) {
	// Define test cases for platformInit.
	tests := []struct {
		name       string
		signalName string
		wantSignal syscall.Signal
	}{
		{
			name:       "SIGPWR is registered",
			signalName: "SIGPWR",
			wantSignal: syscall.SIGPWR,
		},
		{
			name:       "SIGSTKFLT is registered",
			signalName: "SIGSTKFLT",
			wantSignal: syscall.SIGSTKFLT,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create a new manager (which calls platformInit).
			m := New()

			// Look up the signal by name.
			sig, ok := m.SignalByName(tt.signalName)
			if !ok {
				t.Errorf("signal %s not registered", tt.signalName)
				return
			}

			// Verify the signal value.
			if sig != tt.wantSignal {
				t.Errorf("signal %s = %v; want %v", tt.signalName, sig, tt.wantSignal)
			}
		})
	}
}

// TestPrctlSubreaper tests the prctlSubreaper function internally.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestPrctlSubreaper(t *testing.T) {
	// Define test cases for prctlSubreaper.
	tests := []struct {
		name string
		flag int
	}{
		{name: "enable subreaper", flag: 1},
		{name: "disable subreaper", flag: 0},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			err := prctlSubreaper(tt.flag)
			// Check if no error occurred.
			if err != nil {
				t.Errorf("prctlSubreaper(%d) returned error: %v", tt.flag, err)
			}
		})
	}
}

// TestPrctlSubreaperError tests error handling in prctlSubreaper.
//
// This test injects syscall errors to verify error path coverage.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestPrctlSubreaperError(t *testing.T) {
	// Save the original syscall function.
	originalPrctlSyscall := prctlSyscall
	// Restore it after the test.
	defer func() { prctlSyscall = originalPrctlSyscall }()

	// Define test cases.
	tests := []struct {
		name    string
		errno   syscall.Errno
		flag    int
		wantErr bool
	}{
		{
			name:    "EINVAL error",
			errno:   syscall.EINVAL,
			flag:    1,
			wantErr: true,
		},
		{
			name:    "EPERM error",
			errno:   syscall.EPERM,
			flag:    0,
			wantErr: true,
		},
		{
			name:    "EFAULT error",
			errno:   syscall.EFAULT,
			flag:    1,
			wantErr: true,
		},
		{
			name:    "ENOSYS error",
			errno:   syscall.ENOSYS,
			flag:    0,
			wantErr: true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Override prctlSyscall to return an error.
			prctlSyscall = func(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) (uintptr, uintptr, syscall.Errno) {
				return 0, 0, tt.errno
			}

			// Call prctlSubreaper.
			err := prctlSubreaper(tt.flag)

			// Verify error is returned.
			if (err != nil) != tt.wantErr {
				t.Errorf("prctlSubreaper() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify error is properly wrapped.
			if err != nil {
				errStr := err.Error()
				if errStr == "" {
					t.Error("expected non-empty error message")
				}
				// Verify error is a SyscallError.
				if _, ok := err.(*os.SyscallError); !ok {
					t.Errorf("expected error to be *os.SyscallError, got %T", err)
				}
			}
		})
	}
}

// TestIsSubreaperError tests error handling in IsSubreaper.
//
// This test injects syscall errors to verify error path coverage.
//
// Params:
//   - t: the testing context
//
// Returns:
//   - (none, test function)
func TestIsSubreaperError(t *testing.T) {
	// Save the original syscall function.
	originalPrctlSyscall := prctlSyscall
	// Restore it after the test.
	defer func() { prctlSyscall = originalPrctlSyscall }()

	// Define test cases.
	tests := []struct {
		name     string
		errno    syscall.Errno
		wantErr  bool
		wantBool bool
	}{
		{
			name:     "EINVAL error",
			errno:    syscall.EINVAL,
			wantErr:  true,
			wantBool: false,
		},
		{
			name:     "EPERM error",
			errno:    syscall.EPERM,
			wantErr:  true,
			wantBool: false,
		},
		{
			name:     "EFAULT error",
			errno:    syscall.EFAULT,
			wantErr:  true,
			wantBool: false,
		},
		{
			name:     "ENOSYS error",
			errno:    syscall.ENOSYS,
			wantErr:  true,
			wantBool: false,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Override prctlSyscall to return an error.
			prctlSyscall = func(trap uintptr, a1 uintptr, a2 uintptr, a3 uintptr) (uintptr, uintptr, syscall.Errno) {
				return 0, 0, tt.errno
			}

			// Create a manager.
			m := New()

			// Call IsSubreaper.
			isSubreaper, err := m.IsSubreaper()

			// Verify error is returned.
			if (err != nil) != tt.wantErr {
				t.Errorf("IsSubreaper() error = %v, wantErr %v", err, tt.wantErr)
			}

			// Verify bool return value.
			if isSubreaper != tt.wantBool {
				t.Errorf("IsSubreaper() bool = %v, want %v", isSubreaper, tt.wantBool)
			}

			// Verify error is properly wrapped.
			if err != nil {
				errStr := err.Error()
				if errStr == "" {
					t.Error("expected non-empty error message")
				}
				// Verify error is a SyscallError.
				if _, ok := err.(*os.SyscallError); !ok {
					t.Errorf("expected error to be *os.SyscallError, got %T", err)
				}
			}
		})
	}
}
