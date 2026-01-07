// Package kernel provides OS abstraction for the daemon.
package kernel

import (
	"github.com/kodflow/daemon/internal/infrastructure/kernel/ports"
)

// Re-export errors from ports for convenience.
var (
	// ErrProcessNotFound indicates that the specified process could not be found.
	ErrProcessNotFound error = ports.ErrProcessNotFound
	// ErrPermissionDenied indicates that the operation was denied due to insufficient permissions.
	ErrPermissionDenied error = ports.ErrPermissionDenied
	// ErrSignalNotSupported indicates that the signal is not supported on this platform.
	ErrSignalNotSupported error = ports.ErrSignalNotSupported
	// ErrUserNotFound indicates that the specified user could not be found.
	ErrUserNotFound error = ports.ErrUserNotFound
	// ErrGroupNotFound indicates that the specified group could not be found.
	ErrGroupNotFound error = ports.ErrGroupNotFound
	// ErrNotSupported indicates that the operation is not supported on this platform.
	ErrNotSupported error = ports.ErrNotSupported
	// WrapError wraps an error with operation context.
	WrapError func(op string, err error) error = ports.WrapError
)
