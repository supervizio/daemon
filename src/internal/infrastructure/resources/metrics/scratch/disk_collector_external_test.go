// Package scratch_test provides black-box tests for scratch metrics collectors.
package scratch_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/resources/metrics/scratch"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestDiskCollector_ListPartitions tests partition listing.
//
// Params:
//   - t: the testing context
func TestDiskCollector_ListPartitions(t *testing.T) {
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

			collector := scratch.NewDiskCollector()
			require.NotNil(t, collector)

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.ListPartitions(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestDiskCollector_CollectUsage tests disk usage collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectUsage(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// path is the mount path.
		path string
		// ctxCanceled indicates if context should be canceled.
		ctxCanceled bool
		// wantErr is the expected error.
		wantErr error
	}{
		{
			name:    "root_mount_returns_ErrNotSupported",
			path:    "/",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:    "home_mount_returns_ErrNotSupported",
			path:    "/home",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:        "returns_context_error_when_canceled",
			path:        "/",
			ctxCanceled: true,
			wantErr:     context.Canceled,
		},
		{
			name:    "empty_path_returns_ErrEmptyPath",
			path:    "",
			wantErr: scratch.ErrEmptyPath,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewDiskCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectUsage(ctx, tt.path)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestDiskCollector_CollectAllUsage tests all disk usage collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectAllUsage(t *testing.T) {
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

			collector := scratch.NewDiskCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectAllUsage(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestDiskCollector_CollectIO tests I/O stats collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectIO(t *testing.T) {
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

			collector := scratch.NewDiskCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectIO(ctx)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// TestDiskCollector_CollectDeviceIO tests device I/O stats collection.
//
// Params:
//   - t: the testing context
func TestDiskCollector_CollectDeviceIO(t *testing.T) {
	t.Parallel()

	tests := []struct {
		// name is the test case name.
		name string
		// device is the device name.
		device string
		// ctxCanceled indicates if context should be canceled.
		ctxCanceled bool
		// wantErr is the expected error.
		wantErr error
	}{
		{
			name:    "sda_device_returns_ErrNotSupported",
			device:  "sda",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:    "nvme_device_returns_ErrNotSupported",
			device:  "nvme0n1",
			wantErr: scratch.ErrNotSupported,
		},
		{
			name:        "returns_context_error_when_canceled",
			device:      "sda",
			ctxCanceled: true,
			wantErr:     context.Canceled,
		},
		{
			name:    "empty_device_returns_ErrEmptyDevice",
			device:  "",
			wantErr: scratch.ErrEmptyDevice,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			collector := scratch.NewDiskCollector()

			ctx := context.Background()
			if tt.ctxCanceled {
				var cancel context.CancelFunc
				ctx, cancel = context.WithCancel(ctx)
				cancel()
			}

			_, err := collector.CollectDeviceIO(ctx, tt.device)
			assert.ErrorIs(t, err, tt.wantErr)
		})
	}
}

// Test_NewDiskCollector verifies NewDiskCollector creates a valid collector.
//
// Params:
//   - t: testing context for assertions
func Test_NewDiskCollector(t *testing.T) {
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

			collector := scratch.NewDiskCollector()

			if tt.wantNotNil {
				assert.NotNil(t, collector)
			}
		})
	}
}
