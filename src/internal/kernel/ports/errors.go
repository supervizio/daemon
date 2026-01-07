// Package ports defines the interfaces for OS abstraction.
package ports

import (
	"errors"
	"fmt"
)

// Sentinel errors for kernel operations.
var (
	ErrProcessNotFound    = errors.New("process not found")
	ErrPermissionDenied   = errors.New("permission denied")
	ErrSignalNotSupported = errors.New("signal not supported on this platform")
	ErrUserNotFound       = errors.New("user not found")
	ErrGroupNotFound      = errors.New("group not found")
	ErrNotSupported       = errors.New("operation not supported on this platform")
)

// KernelError wraps OS-specific errors with context.
type KernelError struct {
	Op  string // Operation that failed
	Err error  // Underlying error
}

func (e *KernelError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Op, e.Err)
	}
	return e.Op
}

func (e *KernelError) Unwrap() error {
	return e.Err
}

// WrapError wraps an error with operation context.
func WrapError(op string, err error) error {
	if err == nil {
		return nil
	}
	return &KernelError{Op: op, Err: err}
}
