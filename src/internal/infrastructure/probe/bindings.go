//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// cross-platform system metrics collection.
package probe

/*
#cgo CFLAGS: -I${SRCDIR}/../../../lib/probe/include
#cgo linux,amd64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/linux-amd64 -lprobe -lpthread -ldl -lm
#cgo linux,arm64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/linux-arm64 -lprobe -lpthread -ldl -lm
#cgo linux,arm LDFLAGS: -L${SRCDIR}/../../../../dist/lib/linux-arm -lprobe -lpthread -ldl -lm
#cgo linux,386 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/linux-386 -lprobe -lpthread -ldl -lm
#cgo linux,riscv64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/linux-riscv64 -lprobe -lpthread -ldl -lm
#cgo darwin,amd64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/darwin-amd64 -lprobe -lpthread -ldl -lm -framework CoreFoundation -framework IOKit
#cgo darwin,arm64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/darwin-arm64 -lprobe -lpthread -ldl -lm -framework CoreFoundation -framework IOKit
#cgo freebsd,amd64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/freebsd-amd64 -lprobe -lpthread -lm -ldevstat
#cgo openbsd,amd64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/openbsd-amd64 -lprobe -lpthread -lm -lc++abi
#cgo netbsd,amd64 LDFLAGS: -L${SRCDIR}/../../../../dist/lib/netbsd-amd64 -lprobe -lpthread -lm

#include "probe.h"
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"sync"
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
//   - string: platform identifier (e.g., "linux", "darwin", "freebsd"), "unknown" if unavailable
func Platform() string {
	cStr := C.probe_get_platform()
	// Guard against nil pointer from FFI.
	if cStr == nil {
		return "unknown"
	}
	// Convert C string to Go string.
	return C.GoString(cStr)
}

// OSVersion returns the OS version string (e.g., "Linux 6.12.69", "Darwin 24.6.0").
//
// Returns:
//   - string: OS name and release version from uname, "unknown" if unavailable
func OSVersion() string {
	cStr := C.probe_get_os_version()
	// Guard against nil pointer from FFI.
	if cStr == nil {
		return "unknown"
	}
	// Convert C string to Go string.
	return C.GoString(cStr)
}

// KernelVersion returns the full kernel build string from uname.
//
// Returns:
//   - string: kernel version string (e.g., "#1 SMP PREEMPT_DYNAMIC ..."), "unknown" if unavailable
func KernelVersion() string {
	cStr := C.probe_get_kernel_version()
	// Guard against nil pointer from FFI.
	if cStr == nil {
		return "unknown"
	}
	// Convert C string to Go string.
	return C.GoString(cStr)
}

// Arch returns the machine architecture (e.g., "x86_64", "aarch64", "arm64").
//
// Returns:
//   - string: machine architecture from uname, "unknown" if unavailable
func Arch() string {
	cStr := C.probe_get_arch()
	// Guard against nil pointer from FFI.
	if cStr == nil {
		return "unknown"
	}
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
// This is a thin CGO wrapper that delegates to the Go-only convertResultToError.
//
// Params:
//   - r: the C ProbeResult to convert
//
// Returns:
//   - error: nil on success, appropriate error on failure
func resultToError(r C.ProbeResult) error {
	// Extract message from C string if present.
	var msg string
	// Check if C error message pointer is valid.
	if r.error_message != nil {
		msg = C.GoString(r.error_message)
	}
	// Delegate to Go-only function.
	return convertResultToError(bool(r.success), int(r.error_code), msg)
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

// validateCollectionContext validates context and initialization state.
// Combines checkContext and checkInitialized for collector methods.
//
// Params:
//   - ctx: the context to check
//
// Returns:
//   - error: nil if valid, context error or ErrNotInitialized otherwise
func validateCollectionContext(ctx context.Context) error {
	// Check if context has been cancelled before expensive FFI call.
	if err := checkContext(ctx); err != nil {
		// Return the context error.
		return err
	}
	// Verify probe library is initialized.
	return checkInitialized()
}
