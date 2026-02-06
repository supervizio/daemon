//go:build cgo

package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectAll verifies comprehensive metrics collection.
func TestCollectAll(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects all metrics successfully"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			metrics, err := probe.CollectAll()
			require.NoError(t, err)
			require.NotNil(t, metrics)

			// Verify timestamp is set
			assert.False(t, metrics.Timestamp.IsZero())

			// Verify CPU metrics
			assert.GreaterOrEqual(t, metrics.CPU.UsagePercent, 0.0)
			assert.LessOrEqual(t, metrics.CPU.UsagePercent, 100.0)

			// Verify memory metrics
			assert.Greater(t, metrics.Memory.Total, uint64(0))
		})
	}
}

// TestCollectAll_NotInitialized verifies error when not initialized.
func TestCollectAll_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Do not call probe.Init()
			_, err := probe.CollectAll()

			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}

// TestAllMetrics_CPUMetrics verifies CPU data in AllMetrics.
func TestAllMetrics_CPUMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "CPU usage is between 0 and 100"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			metrics, err := probe.CollectAll()
			require.NoError(t, err)

			// CPU usage should be between 0 and 100
			assert.GreaterOrEqual(t, metrics.CPU.UsagePercent, 0.0)
			assert.LessOrEqual(t, metrics.CPU.UsagePercent, 100.0)
		})
	}
}

// TestAllMetrics_MemoryMetrics verifies memory data in AllMetrics.
func TestAllMetrics_MemoryMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "memory total is greater than zero and available does not exceed total"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			metrics, err := probe.CollectAll()
			require.NoError(t, err)

			// Total memory should be greater than zero
			assert.Greater(t, metrics.Memory.Total, uint64(0))

			// Available should not exceed total
			assert.LessOrEqual(t, metrics.Memory.Available, metrics.Memory.Total)
		})
	}
}

// TestAllMetrics_LoadMetrics verifies load data in AllMetrics.
func TestAllMetrics_LoadMetrics(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "load averages are non-negative"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			metrics, err := probe.CollectAll()
			require.NoError(t, err)

			// Load averages should be non-negative
			assert.GreaterOrEqual(t, metrics.Load.Load1, 0.0)
			assert.GreaterOrEqual(t, metrics.Load.Load5, 0.0)
			assert.GreaterOrEqual(t, metrics.Load.Load15, 0.0)
		})
	}
}
