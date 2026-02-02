//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCheckInitialized(t *testing.T) {
	tests := []struct {
		name           string
		setInitialized bool
		wantErr        bool
	}{
		{
			name:           "ReturnsErrorWhenNotInitialized",
			setInitialized: false,
			wantErr:        true,
		},
		{
			name:           "ReturnsNilWhenInitialized",
			setInitialized: true,
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.setInitialized
			initMu.Unlock()

			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			err := checkInitialized()
			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, ErrNotInitialized)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestCheckContext(t *testing.T) {
	tests := []struct {
		name    string
		ctx     context.Context
		wantErr bool
	}{
		{
			name:    "ReturnsNilForActiveContext",
			ctx:     context.Background(),
			wantErr: false,
		},
		{
			name: "ReturnsErrorForCancelledContext",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := checkContext(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestProbeError_Error(t *testing.T) {
	tests := []struct {
		name    string
		err     *probeError
		wantMsg string
	}{
		{
			name:    "ReturnsErrorMessage",
			err:     &probeError{code: 1, message: "test error message"},
			wantMsg: "test error message",
		},
		{
			name:    "ReturnsEmptyForEmptyMessage",
			err:     &probeError{code: 2, message: ""},
			wantMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			assert.Equal(t, tt.wantMsg, got)
		})
	}
}

func TestMapProbeErrorCode(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		wantErr error
	}{
		{
			name:    "ReturnsNilForSuccess",
			code:    0,
			wantErr: nil,
		},
		{
			name:    "ReturnsNotSupportedFor1",
			code:    1,
			wantErr: ErrNotSupported,
		},
		{
			name:    "ReturnsPermissionFor2",
			code:    2,
			wantErr: ErrPermission,
		},
		{
			name:    "ReturnsNotFoundFor3",
			code:    3,
			wantErr: ErrNotFound,
		},
		{
			name:    "ReturnsInvalidParamFor4",
			code:    4,
			wantErr: ErrInvalidParam,
		},
		{
			name:    "ReturnsIOErrorFor5",
			code:    5,
			wantErr: ErrIO,
		},
		{
			name:    "ReturnsInternalFor99",
			code:    99,
			wantErr: ErrInternal,
		},
		{
			name:    "ReturnsNilForUnknownCode",
			code:    999,
			wantErr: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := mapProbeErrorCode(tt.code)
			if tt.wantErr == nil {
				assert.NoError(t, err)
			} else {
				assert.ErrorIs(t, err, tt.wantErr)
			}
		})
	}
}

func TestNewProbeError(t *testing.T) {
	tests := []struct {
		name    string
		code    int
		message string
		wantMsg string
	}{
		{
			name:    "CreatesErrorWithMessage",
			code:    1,
			message: "test error",
			wantMsg: "test error",
		},
		{
			name:    "CreatesErrorWithEmptyMessage",
			code:    2,
			message: "",
			wantMsg: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := newProbeError(tt.code, tt.message)
			assert.Error(t, err)
			assert.Equal(t, tt.wantMsg, err.Error())
		})
	}
}

// TestValidateCollectionContext tests the validateCollectionContext function.
func TestValidateCollectionContext(t *testing.T) {
	tests := []struct {
		name           string
		ctx            context.Context
		setInitialized bool
		wantErr        bool
	}{
		{
			name:           "ReturnsNilWhenValidAndInitialized",
			ctx:            context.Background(),
			setInitialized: true,
			wantErr:        false,
		},
		{
			name: "ReturnsErrorWhenContextCancelled",
			ctx: func() context.Context {
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				return ctx
			}(),
			setInitialized: true,
			wantErr:        true,
		},
		{
			name:           "ReturnsErrorWhenNotInitialized",
			ctx:            context.Background(),
			setInitialized: false,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.setInitialized
			initMu.Unlock()

			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			err := validateCollectionContext(tt.ctx)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// TestResultToError tests the resultToError function exists.
func TestResultToError(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
	}{
		{name: "function exists and compiles"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// resultToError requires C.ProbeResult, tested via integration tests.
			assert.NotNil(t, resultToError)
		})
	}
}
