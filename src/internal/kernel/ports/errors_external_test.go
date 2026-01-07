// Package ports_test provides black-box tests for the ports package.
package ports_test

import (
	"errors"
	"testing"

	"github.com/kodflow/daemon/internal/kernel/ports"
)

// TestSentinelErrors verifies that sentinel errors are properly defined.
//
// Returns:
//   - None (test function)
func TestSentinelErrors(t *testing.T) {
	// Define test cases for each sentinel error.
	tests := []struct {
		name string
		err  error
		msg  string
	}{
		{name: "ErrProcessNotFound", err: ports.ErrProcessNotFound, msg: "process not found"},
		{name: "ErrPermissionDenied", err: ports.ErrPermissionDenied, msg: "permission denied"},
		{name: "ErrSignalNotSupported", err: ports.ErrSignalNotSupported, msg: "signal not supported on this platform"},
		{name: "ErrUserNotFound", err: ports.ErrUserNotFound, msg: "user not found"},
		{name: "ErrGroupNotFound", err: ports.ErrGroupNotFound, msg: "group not found"},
		{name: "ErrNotSupported", err: ports.ErrNotSupported, msg: "operation not supported on this platform"},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Verify error message matches expected value.
			if tt.err.Error() != tt.msg {
				// Report failure if messages don't match.
				t.Errorf("expected %q, got %q", tt.msg, tt.err.Error())
			}
		})
	}
}

// TestNewKernelError verifies that NewKernelError creates errors correctly.
//
// Returns:
//   - None (test function)
func TestNewKernelError(t *testing.T) {
	// Define test cases for NewKernelError.
	tests := []struct {
		name          string
		op            string
		underlyingErr error
		checkOp       bool
		checkErr      bool
	}{
		{
			name:          "with underlying error",
			op:            "test_op",
			underlyingErr: errors.New("underlying error"),
			checkOp:       true,
			checkErr:      true,
		},
		{
			name:          "with nil underlying error",
			op:            "nil_op",
			underlyingErr: nil,
			checkOp:       true,
			checkErr:      false,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create a new KernelError using the constructor.
			kerr := ports.NewKernelError(tt.op, tt.underlyingErr)

			// Verify the error is not nil.
			if kerr == nil {
				// Report failure if error is nil.
				t.Fatal("expected non-nil KernelError")
			}

			// Verify the operation name is set correctly.
			if tt.checkOp && kerr.Op != tt.op {
				// Report failure if operation name is incorrect.
				t.Errorf("expected Op to be %q, got %q", tt.op, kerr.Op)
			}

			// Verify the underlying error is set correctly.
			if tt.checkErr && !errors.Is(kerr.Err, tt.underlyingErr) {
				// Report failure if underlying error is incorrect.
				t.Errorf("expected Err to be %v, got %v", tt.underlyingErr, kerr.Err)
			}
		})
	}
}

// TestKernelErrorError verifies that KernelError.Error returns the correct string.
//
// Returns:
//   - None (test function)
func TestKernelErrorError(t *testing.T) {
	// Define test cases for Error method.
	tests := []struct {
		name     string
		op       string
		err      error
		expected string
	}{
		{
			name:     "with underlying error",
			op:       "read",
			err:      errors.New("file not found"),
			expected: "read: file not found",
		},
		{
			name:     "without underlying error",
			op:       "write",
			err:      nil,
			expected: "write",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create a KernelError using the constructor.
			kerr := ports.NewKernelError(tt.op, tt.err)

			// Get the error string.
			result := kerr.Error()

			// Verify the error string matches expected value.
			if result != tt.expected {
				// Report failure if strings don't match.
				t.Errorf("expected %q, got %q", tt.expected, result)
			}
		})
	}
}

// TestKernelErrorUnwrap verifies that KernelError.Unwrap returns the underlying error.
//
// Returns:
//   - None (test function)
func TestKernelErrorUnwrap(t *testing.T) {
	// Define test cases for Unwrap method.
	tests := []struct {
		name          string
		op            string
		underlyingErr error
	}{
		{
			name:          "unwrap underlying error",
			op:            "test_op",
			underlyingErr: errors.New("underlying error"),
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create a KernelError with the underlying error.
			kerr := ports.NewKernelError(tt.op, tt.underlyingErr)

			// Unwrap the error.
			unwrapped := kerr.Unwrap()

			// Verify the unwrapped error matches the original.
			if !errors.Is(unwrapped, tt.underlyingErr) {
				// Report failure if errors don't match.
				t.Errorf("expected unwrapped error to be %v, got %v", tt.underlyingErr, unwrapped)
			}
		})
	}
}

// TestKernelErrorUnwrapNil verifies that Unwrap returns nil when there is no underlying error.
//
// Returns:
//   - None (test function)
func TestKernelErrorUnwrapNil(t *testing.T) {
	// Define test cases for Unwrap with nil.
	tests := []struct {
		name string
		op   string
	}{
		{
			name: "unwrap nil error",
			op:   "test_op",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Create a KernelError without an underlying error.
			kerr := ports.NewKernelError(tt.op, nil)

			// Unwrap the error.
			unwrapped := kerr.Unwrap()

			// Verify the unwrapped error is nil.
			if unwrapped != nil {
				// Report failure if error is not nil.
				t.Errorf("expected unwrapped error to be nil, got %v", unwrapped)
			}
		})
	}
}

// TestWrapError verifies that WrapError wraps errors correctly.
//
// Returns:
//   - None (test function)
func TestWrapError(t *testing.T) {
	// Define test cases for WrapError.
	tests := []struct {
		name          string
		op            string
		underlyingErr error
		expected      string
	}{
		{
			name:          "wrap underlying error",
			op:            "test_op",
			underlyingErr: errors.New("underlying error"),
			expected:      "test_op: underlying error",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Wrap the error.
			wrapped := ports.WrapError(tt.op, tt.underlyingErr)

			// Verify the wrapped error is not nil.
			if wrapped == nil {
				// Report failure if wrapped error is nil.
				t.Fatal("expected non-nil wrapped error")
			}

			// Verify the error message format.
			if wrapped.Error() != tt.expected {
				// Report failure if error message is incorrect.
				t.Errorf("expected %q, got %q", tt.expected, wrapped.Error())
			}
		})
	}
}

// TestWrapErrorNil verifies that WrapError returns nil for nil errors.
//
// Returns:
//   - None (test function)
func TestWrapErrorNil(t *testing.T) {
	// Define test cases for WrapError with nil.
	tests := []struct {
		name string
		op   string
	}{
		{
			name: "wrap nil error",
			op:   "test_op",
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Wrap a nil error.
			wrapped := ports.WrapError(tt.op, nil)

			// Verify the wrapped error is nil.
			if wrapped != nil {
				// Report failure if wrapped error is not nil.
				t.Errorf("expected nil wrapped error, got %v", wrapped)
			}
		})
	}
}

// TestErrorsIs verifies that errors.Is works with wrapped KernelErrors.
//
// Returns:
//   - None (test function)
func TestErrorsIs(t *testing.T) {
	// Define test cases for errors.Is with wrapped errors.
	tests := []struct {
		name       string
		op         string
		targetErr  error
		shouldFind bool
	}{
		{
			name:       "find wrapped ErrProcessNotFound",
			op:         "test_op",
			targetErr:  ports.ErrProcessNotFound,
			shouldFind: true,
		},
	}

	// Iterate over test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Wrap a sentinel error.
			wrapped := ports.WrapError(tt.op, tt.targetErr)

			// Verify errors.Is can find the sentinel error.
			found := errors.Is(wrapped, tt.targetErr)
			if found != tt.shouldFind {
				// Report failure if errors.Is doesn't work as expected.
				t.Errorf("errors.Is returned %v, expected %v", found, tt.shouldFind)
			}
		})
	}
}
