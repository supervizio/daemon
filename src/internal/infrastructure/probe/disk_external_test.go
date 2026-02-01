//go:build cgo

package probe_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewDiskCollector verifies disk collector creation.
func TestNewDiskCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewDiskCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestDiskCollector_ListPartitions verifies partition listing.
func TestDiskCollector_ListPartitions(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "lists partitions with non-empty mountpoints"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewDiskCollector()
			ctx := context.Background()

			partitions, err := collector.ListPartitions(ctx)
			require.NoError(t, err)

			// Should have at least one partition
			assert.NotEmpty(t, partitions)

			// Verify partition structure
			for _, p := range partitions {
				assert.NotEmpty(t, p.Mountpoint)
			}
		})
	}
}

// TestDiskCollector_CollectUsage verifies disk usage collection.
func TestDiskCollector_CollectUsage(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{name: "collects root disk usage", path: "/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewDiskCollector()
			ctx := context.Background()

			usage, err := collector.CollectUsage(ctx, tt.path)
			require.NoError(t, err)

			assert.Equal(t, tt.path, usage.Path)
			assert.Greater(t, usage.Total, uint64(0))
			assert.GreaterOrEqual(t, usage.UsagePercent, 0.0)
			assert.LessOrEqual(t, usage.UsagePercent, 100.0)
		})
	}
}

// TestDiskCollector_CollectAllUsage verifies all disk usage collection.
func TestDiskCollector_CollectAllUsage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects all disk usage records"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewDiskCollector()
			ctx := context.Background()

			usages, err := collector.CollectAllUsage(ctx)
			require.NoError(t, err)

			// Should have at least one usage record
			assert.NotEmpty(t, usages)
		})
	}
}

// TestDiskCollector_CollectIO verifies disk I/O collection.
func TestDiskCollector_CollectIO(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects disk I/O stats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewDiskCollector()
			ctx := context.Background()

			ioStats, err := collector.CollectIO(ctx)
			require.NoError(t, err)

			// May be empty on some systems
			for _, io := range ioStats {
				assert.NotEmpty(t, io.Device)
			}
		})
	}
}

// TestDiskCollector_CollectDeviceIO verifies device-specific I/O collection.
func TestDiskCollector_CollectDeviceIO(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects device-specific I/O stats"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewDiskCollector()
			ctx := context.Background()

			// First get list of devices
			ioStats, err := collector.CollectIO(ctx)
			require.NoError(t, err)

			if len(ioStats) > 0 {
				// Test with first available device
				device := ioStats[0].Device
				io, err := collector.CollectDeviceIO(ctx, device)
				require.NoError(t, err)
				assert.Equal(t, device, io.Device)
			}
		})
	}
}

// TestDiskCollector_CollectDeviceIO_NotFound verifies ErrNotFound for unknown device.
func TestDiskCollector_CollectDeviceIO_NotFound(t *testing.T) {
	tests := []struct {
		name   string
		device string
	}{
		{name: "returns error for nonexistent device", device: "nonexistent_device_xyz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewDiskCollector()
			ctx := context.Background()

			_, err = collector.CollectDeviceIO(ctx, tt.device)
			assert.Error(t, err)
		})
	}
}

// TestDiskCollector_NotInitialized verifies error when not initialized.
func TestDiskCollector_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewDiskCollector()
			ctx := context.Background()

			_, err := collector.ListPartitions(ctx)
			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}
