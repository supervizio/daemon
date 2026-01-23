// Package scratch_test provides black-box tests for scratch metrics collectors.
package scratch_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCPUCollector_CollectSystem tests system CPU metrics collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectSystem(t *testing.T) {
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
			name:        "returns_ErrNotSupported",
			ctxCanceled: false,
			wantErr:     scratch.ErrNotSupported,
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

			collector := scratch.NewCPUCollector()
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

// TestCPUCollector_CollectProcess tests process CPU metrics collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectProcess(t *testing.T) {
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

			collector := scratch.NewCPUCollector()

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

// TestCPUCollector_CollectAllProcesses tests all processes CPU metrics collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectAllProcesses(t *testing.T) {
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

			collector := scratch.NewCPUCollector()

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

// TestCPUCollector_CollectLoadAverage tests load average collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectLoadAverage(t *testing.T) {
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

			collector := scratch.NewCPUCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectLoadAverage(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestCPUCollector_CollectPressure tests CPU pressure collection.
//
// Params:
//   - t: the testing context
func TestCPUCollector_CollectPressure(t *testing.T) {
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

			collector := scratch.NewCPUCollector()

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

// Test_NewCPUCollector verifies NewCPUCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewCPUCollector(t *testing.T) {
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

			collector := scratch.NewCPUCollector()

			if tt.wantNotNil {
				assert.NotNil(t, collector)
			}
		})
	}
}
