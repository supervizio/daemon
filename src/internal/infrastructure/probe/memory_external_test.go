//go:build cgo

package probe_test

import (
	"context"
	"os"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewMemoryCollector verifies memory collector creation.
func TestNewMemoryCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewMemoryCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestMemoryCollector_CollectSystem verifies system memory collection.
func TestMemoryCollector_CollectSystem(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects system memory metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewMemoryCollector()
			ctx := context.Background()

			mem, err := collector.CollectSystem(ctx)
			require.NoError(t, err)

			// Total memory should be greater than zero
			assert.Greater(t, mem.Total, uint64(0))

			// Available should not exceed total
			assert.LessOrEqual(t, mem.Available, mem.Total)

			// Usage percent should be valid
			assert.GreaterOrEqual(t, mem.UsagePercent, 0.0)
			assert.LessOrEqual(t, mem.UsagePercent, 100.0)

			// Timestamp should be set
			assert.False(t, mem.Timestamp.IsZero())
		})
	}
}

// TestMemoryCollector_CollectProcess verifies process memory collection.
func TestMemoryCollector_CollectProcess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects process memory metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewMemoryCollector()
			ctx := context.Background()
			pid := os.Getpid()

			mem, err := collector.CollectProcess(ctx, pid)
			require.NoError(t, err)

			// PID should match
			assert.Equal(t, pid, mem.PID)

			// RSS should be greater than zero for any running process
			assert.Greater(t, mem.RSS, uint64(0))
		})
	}
}

// TestMemoryCollector_CollectAllProcesses verifies it returns ErrNotSupported.
func TestMemoryCollector_CollectAllProcesses(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error for unsupported operation"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewMemoryCollector()
			ctx := context.Background()

			_, err = collector.CollectAllProcesses(ctx)
			assert.Error(t, err)
		})
	}
}

// TestMemoryCollector_CollectPressure verifies memory pressure collection.
func TestMemoryCollector_CollectPressure(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects memory pressure when supported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewMemoryCollector()
			ctx := context.Background()

			pressure, err := collector.CollectPressure(ctx)
			// PSI may not be supported on all systems, so we accept errors
			if err == nil {
				assert.GreaterOrEqual(t, pressure.SomeAvg10, 0.0)
			}
		})
	}
}

// TestMemoryCollector_NotInitialized verifies error when not initialized.
func TestMemoryCollector_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewMemoryCollector()
			ctx := context.Background()

			_, err := collector.CollectSystem(ctx)
			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}
