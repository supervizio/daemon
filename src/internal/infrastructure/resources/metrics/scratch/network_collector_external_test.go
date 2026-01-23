// Package scratch_test provides black-box tests for scratch metrics collectors.
package scratch_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNetworkCollector_ListInterfaces tests interface listing.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_ListInterfaces(t *testing.T) {
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

			collector := scratch.NewNetworkCollector()
			require.NotNil(t, collector)

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.ListInterfaces(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestNetworkCollector_CollectStats tests interface stats collection.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_CollectStats(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// iface is the interface name.
		iface string
		// ctxCanceled indicates if context should be canceled.
		ctxCanceled bool
		// wantErr is the expected error.
		wantErr error
	}{
		{
			name:    "eth0_interface_returns_ErrNotSupported",
			iface:   "eth0",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:    "lo_interface_returns_ErrNotSupported",
			iface:   "lo",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:        "returns_context_error_when_canceled",
			iface:       "eth0",
			ctxCanceled: true,
			wantErr:     context.Canceled,
		},
		{
			name:    "empty_interface_returns_ErrEmptyInterface",
			iface:   "",
			wantErr: scratch.ErrEmptyInterface,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewNetworkCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectStats(ctx, tt.iface)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestNetworkCollector_CollectAllStats tests all interface stats collection.
//
// Params:
//   - t: the testing context
func TestNetworkCollector_CollectAllStats(t *testing.T) {
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

			collector := scratch.NewNetworkCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectAllStats(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// Test_NewNetworkCollector verifies NewNetworkCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewNetworkCollector(t *testing.T) {
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

			collector := scratch.NewNetworkCollector()

			if tt.wantNotNil {
				assert.NotNil(t, collector)
			}
		})
	}
}
