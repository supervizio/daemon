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
	"sync"
)

var (
	// initialized tracks whether the probe library has been initialized.
	initialized bool
	// initMu protects access to the initialized flag.
	initMu sync.Mutex
	// probeErrorMap maps C error codes to Go sentinel errors.
	probeErrorMap map[C.int]error = map[C.int]error{
		probeOK:            nil,
		probeErrNotSupport: ErrNotSupported,
		probeErrPermission: ErrPermission,
		probeErrNotFound:   ErrNotFound,
		probeErrInvalidPar: ErrInvalidParam,
		probeErrIO:         ErrIO,
		probeErrInternal:   ErrInternal,
	}
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
	// Return nil for successful operations.
	if r.success {
		// Success case.
		return nil
	}

	// Look up error in map.
	if err, ok := probeErrorMap[r.error_code]; ok {
		// Return known error.
		return err
	}

	// Handle unknown error codes with message.
	return buildUnknownError(r)
}

// buildUnknownError creates an error for unknown error codes.
//
// Params:
//   - r: the C ProbeResult with unknown error code
//
// Returns:
//   - error: the constructed error
func buildUnknownError(r C.ProbeResult) error {
	// Return error with message if available.
	if r.error_message != nil {
		// Build error with code and message.
		return &probeError{
			code:    int(r.error_code),
			message: C.GoString(r.error_message),
		}
	}
	// Fallback to generic internal error.
	return ErrInternal
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
