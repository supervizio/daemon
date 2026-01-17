package process_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/infrastructure/process"
)

func TestOperationError(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		err      error
		expected string
	}{
		{
			name:     "with error",
			op:       "test operation",
			err:      errors.New("test error"),
			expected: "test operation: test error",
		},
		{
			name:     "without error",
			op:       "test operation",
			err:      nil,
			expected: "test operation",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opErr := process.NewOperationError(tt.op, tt.err)
			assert.Equal(t, tt.expected, opErr.Error())
		})
	}
}

func TestOperationError_Unwrap(t *testing.T) {
	tests := []struct {
		name        string
		op          string
		err         error
		shouldMatch bool
	}{
		{
			name:        "unwraps to original error",
			op:          "test op",
			err:         errors.New("original error"),
			shouldMatch: true,
		},
		{
			name:        "nil error",
			op:          "test op",
			err:         nil,
			shouldMatch: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			opErr := process.NewOperationError(tt.op, tt.err)
			unwrapped := opErr.Unwrap()

			// Check if unwrapped error matches expected error
			if tt.shouldMatch {
				assert.Equal(t, tt.err, unwrapped)
			} else {
				assert.Nil(t, unwrapped)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name     string
		op       string
		err      error
		wantNil  bool
		expected string
	}{
		{
			name:     "wraps error",
			op:       "test operation",
			err:      errors.New("test error"),
			wantNil:  false,
			expected: "test operation: test error",
		},
		{
			name:    "nil error returns nil",
			op:      "test operation",
			err:     nil,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			wrapped := process.WrapError(tt.op, tt.err)

			// Check if error is nil when expected
			if tt.wantNil {
				assert.Nil(t, wrapped)
			} else {
				require.NotNil(t, wrapped)
				assert.Equal(t, tt.expected, wrapped.Error())
			}
		})
	}
}

func TestSentinelErrors(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrProcessNotFound",
			err:      process.ErrProcessNotFound,
			expected: "process not found",
		},
		{
			name:     "ErrPermissionDenied",
			err:      process.ErrPermissionDenied,
			expected: "permission denied",
		},
		{
			name:     "ErrNotSupported",
			err:      process.ErrNotSupported,
			expected: "operation not supported on this platform",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}
