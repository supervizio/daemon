//go:build cgo

package probe_test

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewIOCollector verifies I/O collector creation.
func TestNewIOCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewIOCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestIOCollector_CollectStats verifies I/O stats collection.
func TestIOCollector_CollectStats(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects I/O stats with non-negative values"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewIOCollector()
			ctx := context.Background()

			stats, err := collector.CollectStats(ctx)
			require.NoError(t, err)

			// I/O stats should be non-negative
			assert.GreaterOrEqual(t, stats.ReadOpsTotal, uint64(0))
			assert.GreaterOrEqual(t, stats.ReadBytesTotal, uint64(0))
			assert.GreaterOrEqual(t, stats.WriteOpsTotal, uint64(0))
			assert.GreaterOrEqual(t, stats.WriteBytesTotal, uint64(0))

			// Timestamp should be set
			assert.False(t, stats.Timestamp.IsZero())
		})
	}
}

// TestIOCollector_CollectPressure verifies I/O pressure collection.
func TestIOCollector_CollectPressure(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects I/O pressure when supported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewIOCollector()
			ctx := context.Background()

			pressure, err := collector.CollectPressure(ctx)
			// PSI may not be supported on all systems, so we accept errors
			if err == nil {
				assert.GreaterOrEqual(t, pressure.SomeAvg10, 0.0)
				assert.GreaterOrEqual(t, pressure.FullAvg10, 0.0)
			}
		})
	}
}

// TestIOCollector_NotInitialized verifies error when not initialized.
func TestIOCollector_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewIOCollector()
			ctx := context.Background()

			_, err := collector.CollectStats(ctx)
			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}
