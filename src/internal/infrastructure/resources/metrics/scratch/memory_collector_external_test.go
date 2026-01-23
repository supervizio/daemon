// Package scratch_test provides black-box tests for scratch metrics collectors.
package scratch_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestMemoryCollector_CollectSystem tests system memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectSystem(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// ctxCanceled indicates if context should be canceled.
		ctxCanceled bool
		// wantErr is the expected error.
		wantErr error
	}{
		{
			name:    "returns_ErrNotSupported",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:        "returns_context_error_when_canceled",
			ctxCanceled: true,
			wantErr:     context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewMemoryCollector()
			require.NotNil(t, collector)

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectSystem(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestMemoryCollector_CollectProcess tests process memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectProcess(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// pid is the process ID to test.
		pid int
		// ctxCanceled indicates if context should be canceled.
		ctxCanceled bool
		// wantErr is the expected error.
		wantErr error
	}{
		{
			name:    "pid_1_returns_ErrNotSupported",
			pid:     1,
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:    "pid_0_returns_ErrInvalidPID",
			pid:     0,
			wantErr: scratch.ErrInvalidPID,
		},
		{
			name:        "returns_context_error_when_canceled",
			pid:         1,
			ctxCanceled: true,
			wantErr:     context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewMemoryCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectProcess(ctx, tt.pid)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestMemoryCollector_CollectAllProcesses tests all processes memory metrics collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectAllProcesses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// ctxCanceled indicates if context should be canceled.
		ctxCanceled bool
		// wantErr is the expected error.
		wantErr error
	}{
		{
			name:    "returns_ErrNotSupported",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:        "returns_context_error_when_canceled",
			ctxCanceled: true,
			wantErr:     context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewMemoryCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectAllProcesses(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestMemoryCollector_CollectPressure tests memory pressure collection.
//
// Params:
//   - t: the testing context
func TestMemoryCollector_CollectPressure(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// ctxCanceled indicates if context should be canceled.
		ctxCanceled bool
		// wantErr is the expected error.
		wantErr error
	}{
		{
			name:    "returns_ErrNotSupported",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:        "returns_context_error_when_canceled",
			ctxCanceled: true,
			wantErr:     context.Canceled,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewMemoryCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectPressure(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// Test_NewMemoryCollector verifies NewMemoryCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewMemoryCollector(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// wantNotNil indicates if collector should be non-nil.
		wantNotNil bool
	}{
		{
			name:       "returns_valid_collector",
			wantNotNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewMemoryCollector()

			if tt.wantNotNil {
				assert.NotNil(t, collector)
			}
		})
	}
}
