// Package probe provides CGO bindings to the Rust probe library for
// cross-platform system metrics collection.
package probe

import "errors"

// Error codes matching probe.h constants.
const (
	probeOK            = 0
	probeErrNotSupport = 1
	probeErrPermission = 2
	probeErrNotFound   = 3
	probeErrInvalidPar = 4
	probeErrIO         = 5
	probeErrInternal   = 99
)

// Sentinel errors for probe operations.
var (
	// ErrNotSupported indicates the operation is not supported on this platform.
	ErrNotSupported = errors.New("operation not supported on this platform")

	// ErrPermission indicates permission was denied for the operation.
	ErrPermission = errors.New("permission denied")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound = errors.New("resource not found")

	// ErrInvalidParam indicates an invalid parameter was provided.
	ErrInvalidParam = errors.New("invalid parameter")

	// ErrIO indicates an I/O error occurred.
	ErrIO = errors.New("I/O error")

	// ErrInternal indicates an internal error occurred.
	ErrInternal = errors.New("internal error")

	// ErrNotInitialized indicates the probe library was not initialized.
	ErrNotInitialized = errors.New("probe library not initialized")
)
