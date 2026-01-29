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

// NewOperationError wraps an error with operation context for debugging.
//
// Params:
//   - op: operation name that failed (e.g., "lookup user", "start process")
//   - err: underlying error to wrap
//
// Returns:
//   - *OperationError: wrapped error with operation context
func NewOperationError(op string, err error) *OperationError {
	// return wrapped error with operation context.
	return &OperationError{Op: op, Err: err}
}

// Error formats the error with operation prefix for context.
//
// Returns:
//   - string: formatted error message "op: err" or just "op" if no underlying error
func (e *OperationError) Error() string {
	// Include underlying error when present.
	if e.Err != nil {
		// return formatted error with operation and underlying error.
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	// operation-only message when no underlying error.
	return e.Op
}

// Unwrap enables errors.Is and errors.As to work with wrapped errors.
//
// Returns:
//   - error: the underlying error for error chain traversal
func (e *OperationError) Unwrap() error { return e.Err }

// WrapError adds operation context to an error, returning nil for nil input.
//
// Params:
//   - op: operation name that failed
//   - err: underlying error (may be nil)
//
// Returns:
//   - error: wrapped error or nil if input was nil
func WrapError(op string, err error) error {
	// Preserve nil to allow clean conditional error handling.
	if err == nil {
		// nil input results in nil output.
		return nil
	}
	// wrap with operation context.
	return &OperationError{Op: op, Err: err}
}
