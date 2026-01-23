// Package scratch_test provides black-box tests for scratch metrics collectors.
package scratch_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestIOCollector_CollectStats tests I/O stats collection.
//
// Params:
//   - t: the testing context
func TestIOCollector_CollectStats(t *testing.T) {
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

			collector := scratch.NewIOCollector()
			require.NotNil(t, collector)

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectStats(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestIOCollector_CollectPressure tests I/O pressure collection.
//
// Params:
//   - t: the testing context
func TestIOCollector_CollectPressure(t *testing.T) {
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

			collector := scratch.NewIOCollector()

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

// Test_NewIOCollector verifies NewIOCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewIOCollector(t *testing.T) {
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

			collector := scratch.NewIOCollector()

			if tt.wantNotNil {
				assert.NotNil(t, collector)
			}
		})
	}
}
