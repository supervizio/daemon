// Package process provides shared error types for process-related operations.
package process

import (
	"errors"
	"fmt"
)

// Sentinel errors for process operations.
var (
	// ErrProcessNotFound indicates that the specified process could not be found.
	ErrProcessNotFound = errors.New("process not found")
	// ErrPermissionDenied indicates that the operation was denied due to insufficient permissions.
	ErrPermissionDenied = errors.New("permission denied")
	// ErrNotSupported indicates that the operation is not supported on this platform.
	ErrNotSupported = errors.New("operation not supported on this platform")
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
	return &OperationError{Op: op, Err: err}
}

// Error returns the string representation of the operation error.
//
// Returns:
//   - string: formatted error message with operation context
func (e *OperationError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Op
}

// Unwrap returns the underlying error for errors.Is and errors.As support.
//
// Returns:
//   - error: the underlying error
func (e *OperationError) Unwrap() error {
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
	if err == nil {
		return nil
	}
	return &OperationError{Op: op, Err: err}
}
