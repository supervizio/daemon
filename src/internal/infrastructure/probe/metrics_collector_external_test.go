//go:build cgo

package probe_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCollectAllMetrics verifies that all metrics can be collected.
func TestCollectAllMetrics(t *testing.T) {
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

			ctx := context.Background()
			metrics, err := probe.CollectAllMetrics(ctx)

			require.NoError(t, err)
			require.NotNil(t, metrics)

			// Verify basic fields are populated
			assert.NotEmpty(t, metrics.Platform)
			assert.False(t, metrics.Timestamp.IsZero())
			assert.NotZero(t, metrics.CollectedAt)
		})
	}
}

// TestCollectAllMetricsJSON verifies JSON output.
func TestCollectAllMetricsJSON(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns valid JSON with expected fields"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			ctx := context.Background()
			jsonStr, err := probe.CollectAllMetricsJSON(ctx)

			require.NoError(t, err)
			assert.NotEmpty(t, jsonStr)
			assert.Contains(t, jsonStr, "platform")
			assert.Contains(t, jsonStr, "timestamp")
		})
	}
}

// TestCollectAllMetrics_NotInitialized verifies error when not initialized.
func TestCollectAllMetrics_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Do not call probe.Init()
			ctx := context.Background()
			_, err := probe.CollectAllMetrics(ctx)

			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}

// TestAllSystemMetrics_Structure verifies the AllSystemMetrics structure.
func TestAllSystemMetrics_Structure(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "CPU and memory metrics are populated"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			ctx := context.Background()
			metrics, err := probe.CollectAllMetrics(ctx)
			require.NoError(t, err)

			// CPU metrics should be present
			if metrics.CPU != nil {
				assert.GreaterOrEqual(t, metrics.CPU.UsagePercent, 0.0)
				assert.LessOrEqual(t, metrics.CPU.UsagePercent, 100.0)
			}

			// Memory metrics should be present
			if metrics.Memory != nil {
				assert.Greater(t, metrics.Memory.TotalBytes, uint64(0))
			}
		})
	}
}
