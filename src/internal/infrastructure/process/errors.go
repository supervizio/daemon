// Package process provides shared error types for process-related operations.
package process

import (
	"errors"
	"fmt"
)

// Sentinel errors for process operations.
var (
	// ErrProcessNotFound indicates that the specified process could not be found.
	ErrProcessNotFound error = errors.New("process not found")
	// ErrPermissionDenied indicates that the operation was denied due to insufficient permissions.
	ErrPermissionDenied error = errors.New("permission denied")
	// ErrNotSupported indicates that the operation is not supported on this platform.
	ErrNotSupported error = errors.New("operation not supported on this platform")
)

// OperationError wraps OS-specific errors with context.
// It provides operation information to help identify where failures occurred.
type OperationError struct {
	// Op is the name of the operation that failed.
	Op string
	// Err is the underlying error that caused the failure.
	Err error
}

// NewOperationError creates a new OperationError with the given operation and error.
//
// Params:
//   - op: the name of the operation that failed
//   - err: the underlying error that caused the failure
//
// Returns:
//   - *OperationError: a new OperationError instance with the provided values
func NewOperationError(op string, err error) *OperationError {
	// Create and return new OperationError with provided values.
	return &OperationError{Op: op, Err: err}
}

// Error returns the string representation of the operation error.
//
// Returns:
//   - string: formatted error message with operation context
func (e *OperationError) Error() string {
	// Check if underlying error is non-nil to format message.
	if e.Err != nil {
		// Return formatted message with operation and error.
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	// Return operation name when no underlying error exists.
	return e.Op
}

// Unwrap returns the underlying error for errors.Is and errors.As support.
//
// Returns:
//   - error: the underlying error
func (e *OperationError) Unwrap() error {
	// Return the underlying error for error chain support.
	return e.Err
}

// WrapError wraps an error with operation context.
//
// Params:
//   - op: the name of the operation that failed
//   - err: the underlying error to wrap
//
// Returns:
//   - error: a wrapped OperationError or nil if err is nil
func WrapError(op string, err error) error {
	// Check if error is nil to avoid wrapping nil errors.
	if err == nil {
		// Return nil when no error to wrap.
		return nil
	}
	// Return wrapped error with operation context.
	return &OperationError{Op: op, Err: err}
}
