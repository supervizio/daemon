//go:build cgo

package probe

/*
#include "probe.h"
*/
import "C"

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestCheckInitialized verifies initialization checking.
func TestCheckInitialized(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
		expectErr   bool
		expectedErr error
	}{
		{
			name:        "not initialized returns error",
			initialized: false,
			expectErr:   true,
			expectedErr: ErrNotInitialized,
		},
		{
			name:        "initialized returns nil",
			initialized: true,
			expectErr:   false,
			expectedErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Ensure clean state
			initMu.Lock()
			wasInitialized := initialized
			initialized = tt.initialized
			initMu.Unlock()

			defer func() {
				initMu.Lock()
				initialized = wasInitialized
				initMu.Unlock()
			}()

			err := checkInitialized()
			if tt.expectErr {
				assert.Error(t, err)
				assert.Equal(t, tt.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestResultToError verifies C result to Go error conversion.
func TestResultToError(t *testing.T) {
	tests := []struct {
		name      string
		success   bool
		errorCode C.int
		expectNil bool
		expectErr error
	}{
		{
			name:      "success returns nil",
			success:   true,
			errorCode: probeOK,
			expectNil: true,
		},
		{
			name:      "not supported error",
			success:   false,
			errorCode: probeErrNotSupport,
			expectErr: ErrNotSupported,
		},
		{
			name:      "permission error",
			success:   false,
			errorCode: probeErrPermission,
			expectErr: ErrPermission,
		},
		{
			name:      "not found error",
			success:   false,
			errorCode: probeErrNotFound,
			expectErr: ErrNotFound,
		},
		{
			name:      "invalid param error",
			success:   false,
			errorCode: probeErrInvalidPar,
			expectErr: ErrInvalidParam,
		},
		{
			name:      "io error",
			success:   false,
			errorCode: probeErrIO,
			expectErr: ErrIO,
		},
		{
			name:      "internal error",
			success:   false,
			errorCode: probeErrInternal,
			expectErr: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := C.ProbeResult{
				success:    C.bool(tt.success),
				error_code: tt.errorCode,
			}

			err := resultToError(result)
			if tt.expectNil {
				assert.NoError(t, err)
			} else {
				assert.Equal(t, tt.expectErr, err)
			}
		})
	}
}

// TestBuildUnknownError verifies unknown error construction.
func TestBuildUnknownError(t *testing.T) {
	tests := []struct {
		name         string
		errorCode    C.int
		hasMessage   bool
		errorMessage string
		expectMsg    string
	}{
		{
			name:         "unknown code without message",
			errorCode:    C.int(999),
			hasMessage:   false,
			errorMessage: "",
			expectMsg:    ErrInternal.Error(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := C.ProbeResult{
				success:       C.bool(false),
				error_code:    tt.errorCode,
				error_message: nil,
			}

			err := buildUnknownError(result)
			assert.Error(t, err)
			assert.Equal(t, tt.expectMsg, err.Error())
		})
	}
}

// TestProbeError verifies probeError error implementation.
func TestProbeError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		expected string
	}{
		{
			name:     "returns message",
			code:     42,
			message:  "test error message",
			expected: "test error message",
		},
		{
			name:     "empty message",
			code:     0,
			message:  "",
			expected: "",
		},
		{
			name:     "complex error message",
			code:     100,
			message:  "failed to collect metrics: resource unavailable",
			expected: "failed to collect metrics: resource unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := &probeError{
				code:    tt.code,
				message: tt.message,
			}

			assert.Equal(t, tt.expected, err.Error())
		})
	}
}

// TestError verifies the Error method of probeError.
func TestError(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		message  string
		expected string
	}{
		{
			name:     "simple message",
			code:     1,
			message:  "simple error",
			expected: "simple error",
		},
		{
			name:     "detailed message",
			code:     99,
			message:  "detailed error: context timeout exceeded",
			expected: "detailed error: context timeout exceeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			pe := &probeError{
				code:    tt.code,
				message: tt.message,
			}

			result := pe.Error()
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestProbeErrorMap verifies error code mapping.
func TestProbeErrorMap(t *testing.T) {
	tests := []struct {
		name     string
		code     int
		expected error
	}{
		{
			name:     "probeOK maps to nil",
			code:     probeOK,
			expected: nil,
		},
		{
			name:     "probeErrNotSupport maps to ErrNotSupported",
			code:     probeErrNotSupport,
			expected: ErrNotSupported,
		},
		{
			name:     "probeErrPermission maps to ErrPermission",
			code:     probeErrPermission,
			expected: ErrPermission,
		},
		{
			name:     "probeErrNotFound maps to ErrNotFound",
			code:     probeErrNotFound,
			expected: ErrNotFound,
		},
		{
			name:     "probeErrInvalidPar maps to ErrInvalidParam",
			code:     probeErrInvalidPar,
			expected: ErrInvalidParam,
		},
		{
			name:     "probeErrIO maps to ErrIO",
			code:     probeErrIO,
			expected: ErrIO,
		},
		{
			name:     "probeErrInternal maps to ErrInternal",
			code:     probeErrInternal,
			expected: ErrInternal,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, probeErrorMap[tt.code])
		})
	}
}
