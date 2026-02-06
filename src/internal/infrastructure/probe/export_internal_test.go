//go:build cgo

// Package probe exports internal identifiers for testing purposes.
// This file does not use import "C" to avoid the Go 1.25 restriction
// on CGO in internal test files.
package probe

// ProbeErrorCode constants for testing error mapping (matching C values).
const (
	ExportProbeOK            int = 0
	ExportProbeErrNotSupport int = 1
	ExportProbeErrPermission int = 2
	ExportProbeErrNotFound   int = 3
	ExportProbeErrInvalidPar int = 4
	ExportProbeErrIO         int = 5
	ExportProbeErrInternal   int = 99
)

// ExportCheckInitialized exports checkInitialized for external tests.
var ExportCheckInitialized func() error = checkInitialized

// ExportSetInitialized allows tests to set the initialization state.
func ExportSetInitialized(value bool) {
	initMu.Lock()
	defer initMu.Unlock()
	initialized = value
}

// ExportGetInitialized returns the current initialization state.
func ExportGetInitialized() bool {
	initMu.Lock()
	defer initMu.Unlock()
	return initialized
}

// ExportNewProbeError creates a new probeError for testing.
func ExportNewProbeError(code int, message string) error {
	return &probeError{
		code:    code,
		message: message,
	}
}

// ExportGetProbeErrorMapping returns the error for a given probe error code.
func ExportGetProbeErrorMapping(code int) error {
	// Map code to known errors.
	switch code {
	case ExportProbeOK:
		return nil
	case ExportProbeErrNotSupport:
		return ErrNotSupported
	case ExportProbeErrPermission:
		return ErrPermission
	case ExportProbeErrNotFound:
		return ErrNotFound
	case ExportProbeErrInvalidPar:
		return ErrInvalidParam
	case ExportProbeErrIO:
		return ErrIO
	case ExportProbeErrInternal:
		return ErrInternal
	default:
		return nil
	}
}
