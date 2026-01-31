//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// cross-platform system metrics collection.
package probe

/*
#include "probe.h"
*/
import "C"

import "errors"

// Error codes matching probe.h constants.
const (
	probeOK            C.int = 0
	probeErrNotSupport C.int = 1
	probeErrPermission C.int = 2
	probeErrNotFound   C.int = 3
	probeErrInvalidPar C.int = 4
	probeErrIO         C.int = 5
	probeErrInternal   C.int = 99
)

// Sentinel errors for probe operations.
var (
	// ErrNotSupported indicates the operation is not supported on this platform.
	ErrNotSupported error = errors.New("operation not supported on this platform")

	// ErrPermission indicates permission was denied for the operation.
	ErrPermission error = errors.New("permission denied")

	// ErrNotFound indicates the requested resource was not found.
	ErrNotFound error = errors.New("resource not found")

	// ErrInvalidParam indicates an invalid parameter was provided.
	ErrInvalidParam error = errors.New("invalid parameter")

	// ErrIO indicates an I/O error occurred.
	ErrIO error = errors.New("io error")

	// ErrInternal indicates an internal error occurred.
	ErrInternal error = errors.New("internal error")

	// ErrNotInitialized indicates the probe library was not initialized.
	ErrNotInitialized error = errors.New("probe library not initialized")
)
