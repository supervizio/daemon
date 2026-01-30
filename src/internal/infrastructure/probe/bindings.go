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
	"unsafe"
)

var (
	initialized bool
	initMu      sync.Mutex
)

// Init initializes the Rust probe library.
// Must be called once before using any probe functions.
// Safe to call multiple times; subsequent calls are no-ops.
func Init() error {
	initMu.Lock()
	defer initMu.Unlock()

	if initialized {
		return nil
	}

	result := C.probe_init()
	if err := resultToError(result); err != nil {
		return err
	}

	initialized = true
	return nil
}

// Shutdown releases resources held by the Rust probe library.
// Should be called at program exit.
func Shutdown() {
	initMu.Lock()
	defer initMu.Unlock()

	if !initialized {
		return
	}

	C.probe_shutdown()
	initialized = false
}

// IsInitialized returns whether the probe library has been initialized.
func IsInitialized() bool {
	initMu.Lock()
	defer initMu.Unlock()
	return initialized
}

// Platform returns the current platform name.
func Platform() string {
	cStr := C.probe_get_platform()
	return C.GoString(cStr)
}

// QuotaSupported returns whether resource quotas are supported on this platform.
func QuotaSupported() bool {
	return bool(C.probe_quota_is_supported())
}

// resultToError converts a C ProbeResult to a Go error.
func resultToError(r C.ProbeResult) error {
	if r.success {
		return nil
	}

	switch r.error_code {
	case probeOK:
		return nil
	case probeErrNotSupport:
		return ErrNotSupported
	case probeErrPermission:
		return ErrPermission
	case probeErrNotFound:
		return ErrNotFound
	case probeErrInvalidPar:
		return ErrInvalidParam
	case probeErrIO:
		return ErrIO
	case probeErrInternal:
		return ErrInternal
	default:
		if r.error_message != nil {
			return &probeError{
				code:    int(r.error_code),
				message: C.GoString(r.error_message),
			}
		}
		return ErrInternal
	}
}

// probeError wraps an error code and message from the probe library.
type probeError struct {
	code    int
	message string
}

// Error implements the error interface.
func (e *probeError) Error() string {
	return e.message
}

// checkInitialized verifies the library is initialized.
func checkInitialized() error {
	initMu.Lock()
	defer initMu.Unlock()
	if !initialized {
		return ErrNotInitialized
	}
	return nil
}

// cStringFree frees a C string allocated by Go.
func cStringFree(s *C.char) {
	C.free(unsafe.Pointer(s))
}
