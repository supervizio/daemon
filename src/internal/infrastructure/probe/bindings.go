//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// cross-platform system metrics collection.
package probe

/*
#cgo CFLAGS: -I${SRCDIR}/../../../lib/probe/include
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/linux-amd64 -lprobe -lpthread -ldl -lm
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/linux-arm64 -lprobe -lpthread -ldl -lm
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/darwin-amd64 -lprobe -lpthread -ldl -lm -framework CoreFoundation -framework IOKit
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/darwin-arm64 -lprobe -lpthread -ldl -lm -framework CoreFoundation -framework IOKit
#cgo freebsd,amd64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/freebsd-amd64 -lprobe -lpthread -lm

#include "probe.h"
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"sync"
)

// Go-only error code constants matching probe.h.
const (
	// probeCodeOK indicates success.
	probeCodeOK int = 0
	// probeCodeNotSupported indicates the operation is not supported.
	probeCodeNotSupported int = 1
	// probeCodePermission indicates permission denied.
	probeCodePermission int = 2
	// probeCodeNotFound indicates resource not found.
	probeCodeNotFound int = 3
	// probeCodeInvalidParam indicates invalid parameter.
	probeCodeInvalidParam int = 4
	// probeCodeIO indicates I/O error.
	probeCodeIO int = 5
	// probeCodeInternal indicates internal error.
	probeCodeInternal int = 99
)

var (
	// initialized tracks whether the probe library has been initialized.
	initialized bool
	// initMu protects access to the initialized flag.
	initMu sync.Mutex
)

// Init initializes the Rust probe library.
// Must be called once before using any probe functions.
// Safe to call multiple times; subsequent calls are no-ops.
//
// Returns:
//   - error: nil on success, error if initialization fails
func Init() error {
	initMu.Lock()
	defer initMu.Unlock()

	// Check if already initialized.
	if initialized {
		// Return nil for idempotent behavior.
		return nil
	}

	result := C.probe_init()
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return initialization error.
		return err
	}

	initialized = true
	// Return success.
	return nil
}

// Shutdown releases resources held by the Rust probe library.
// Should be called at program exit.
func Shutdown() {
	initMu.Lock()
	defer initMu.Unlock()

	// Check if library is initialized.
	if !initialized {
		// Nothing to do if not initialized.
		return
	}

	C.probe_shutdown()
	initialized = false
}

// IsInitialized returns whether the probe library has been initialized.
//
// Returns:
//   - bool: true if the library is initialized
func IsInitialized() bool {
	initMu.Lock()
	defer initMu.Unlock()
	// Return current initialization state.
	return initialized
}

// Platform returns the current platform name.
//
// Returns:
//   - string: platform identifier (e.g., "linux", "darwin", "freebsd")
func Platform() string {
	cStr := C.probe_get_platform()
	// Convert C string to Go string.
	return C.GoString(cStr)
}

// QuotaSupported returns whether resource quotas are supported on this platform.
//
// Returns:
//   - bool: true if quotas are supported
func QuotaSupported() bool {
	// Delegate to Rust library for platform-specific check.
	return bool(C.probe_quota_is_supported())
}

// resultToError converts a C ProbeResult to a Go error.
//
// Params:
//   - r: the C ProbeResult to convert
//
// Returns:
//   - error: nil on success, appropriate error on failure
func resultToError(r C.ProbeResult) error {
	// Check if the result indicates success.
	if r.success {
		// Return nil for successful operations.
		return nil
	}

	// Map error code to Go error using the Go-only function.
	code := int(r.error_code)
	// Check if the error code maps to a known error.
	if err := mapProbeErrorCode(code); err != nil {
		// Return known error.
		return err
	}

	// Handle unknown error codes with message.
	if r.error_message != nil {
		// Build error with code and message.
		return newProbeError(code, C.GoString(r.error_message))
	}
	// Fallback to generic internal error.
	return ErrInternal
}

// mapProbeErrorCode maps a probe error code to a Go sentinel error.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - code: the error code from the probe library
//
// Returns:
//   - error: the mapped error, or nil if code is unknown
func mapProbeErrorCode(code int) error {
	// Map error code to Go sentinel error.
	switch code {
	// Success case.
	case probeCodeOK:
		// Return nil for success.
		return nil
	// Not supported error.
	case probeCodeNotSupported:
		// Return not supported error.
		return ErrNotSupported
	// Permission error.
	case probeCodePermission:
		// Return permission error.
		return ErrPermission
	// Not found error.
	case probeCodeNotFound:
		// Return not found error.
		return ErrNotFound
	// Invalid parameter error.
	case probeCodeInvalidParam:
		// Return invalid parameter error.
		return ErrInvalidParam
	// I/O error.
	case probeCodeIO:
		// Return I/O error.
		return ErrIO
	// Internal error.
	case probeCodeInternal:
		// Return internal error.
		return ErrInternal
	// Unknown error code.
	default:
		// Return nil for unknown codes.
		return nil
	}
}

// newProbeError creates a new probeError with the given code and message.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - code: the error code
//   - message: the error message
//
// Returns:
//   - error: the constructed error
func newProbeError(code int, message string) error {
	// Construct and return the probe error.
	return &probeError{
		code:    code,
		message: message,
	}
}

// probeError wraps an error code and message from the probe library.
// It implements the error interface for custom error reporting.
type probeError struct {
	code    int
	message string
}

// Error implements the error interface.
//
// Returns:
//   - string: the error message
func (e *probeError) Error() string {
	// Return the error message from the probe library.
	return e.message
}

// checkInitialized verifies the library is initialized.
//
// Returns:
//   - error: nil if initialized, ErrNotInitialized otherwise
func checkInitialized() error {
	initMu.Lock()
	defer initMu.Unlock()
	// Check initialization state.
	if !initialized {
		// Return error if not initialized.
		return ErrNotInitialized
	}
	// Return nil if initialized.
	return nil
}

// checkContext verifies the context has not been cancelled.
// This should be called before expensive FFI operations to allow cancellation.
//
// Params:
//   - ctx: the context to check
//
// Returns:
//   - error: nil if context is active, context error if cancelled/deadline exceeded
func checkContext(ctx context.Context) error {
	// Check if context was cancelled or deadline exceeded.
	select {
	case <-ctx.Done():
		// Return the context error.
		return ctx.Err()
	default:
		// Context is still active.
		return nil
	}
}
