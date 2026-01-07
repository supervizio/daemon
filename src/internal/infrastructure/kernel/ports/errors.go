// Package ports defines the interfaces for OS abstraction.
package ports

import (
	"errors"
	"fmt"
)

// Sentinel errors for kernel operations.
var (
	// ErrProcessNotFound indicates that the specified process could not be found.
	ErrProcessNotFound error = errors.New("process not found")
	// ErrPermissionDenied indicates that the operation was denied due to insufficient permissions.
	ErrPermissionDenied error = errors.New("permission denied")
	// ErrSignalNotSupported indicates that the signal is not supported on this platform.
	ErrSignalNotSupported error = errors.New("signal not supported on this platform")
	// ErrUserNotFound indicates that the specified user could not be found.
	ErrUserNotFound error = errors.New("user not found")
	// ErrGroupNotFound indicates that the specified group could not be found.
	ErrGroupNotFound error = errors.New("group not found")
	// ErrNotSupported indicates that the operation is not supported on this platform.
	ErrNotSupported error = errors.New("operation not supported on this platform")
)

// KernelError wraps OS-specific errors with context.
// It provides operation information to help identify where failures occurred.
type KernelError struct {
	// Op is the name of the operation that failed.
	Op string
	// Err is the underlying error that caused the failure.
	Err error
}

// NewKernelError creates a new KernelError with the given operation and error.
//
// Params:
//   - op: the name of the operation that failed
//   - err: the underlying error that caused the failure
//
// Returns:
//   - *KernelError: a new KernelError instance with the provided values
func NewKernelError(op string, err error) *KernelError {
	// Return a new KernelError with the provided operation and error.
	return &KernelError{Op: op, Err: err}
}

// Error returns the string representation of the kernel error.
//
// Returns:
//   - string: formatted error message with operation context
func (e *KernelError) Error() string {
	// Check if the underlying error is not nil to include it in the message.
	if e.Err != nil {
		// Return formatted error with operation context and underlying error.
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	// Return only the operation name when no underlying error exists.
	return e.Op
}

// Unwrap returns the underlying error for errors.Is and errors.As support.
//
// Returns:
//   - error: the underlying error
func (e *KernelError) Unwrap() error {
	// Return the wrapped error for unwrapping support.
	return e.Err
}

// WrapError wraps an error with operation context.
//
// Params:
//   - op: the name of the operation that failed
//   - err: the underlying error to wrap
//
// Returns:
//   - error: a wrapped KernelError or nil if err is nil
func WrapError(op string, err error) error {
	// Check if the error is nil to avoid wrapping nil errors.
	if err == nil {
		// Return nil for nil errors to preserve nil error semantics.
		return nil
	}
	// Return a new KernelError wrapping the original error with operation context.
	return &KernelError{Op: op, Err: err}
}
