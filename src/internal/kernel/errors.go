// Package kernel provides OS abstraction for the daemon.
package kernel

import (
	"github.com/kodflow/daemon/internal/kernel/ports"
)

// Re-export errors from ports for convenience.
var (
	ErrProcessNotFound    = ports.ErrProcessNotFound
	ErrPermissionDenied   = ports.ErrPermissionDenied
	ErrSignalNotSupported = ports.ErrSignalNotSupported
	ErrUserNotFound       = ports.ErrUserNotFound
	ErrGroupNotFound      = ports.ErrGroupNotFound
	ErrNotSupported       = ports.ErrNotSupported
)

// WrapError wraps an error with operation context.
var WrapError = ports.WrapError
