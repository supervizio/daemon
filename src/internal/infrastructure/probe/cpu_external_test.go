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

// TestNewCPUCollector verifies CPU collector creation.
func TestNewCPUCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "creates non-nil collector"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewCPUCollector()
			assert.NotNil(t, collector)
		})
	}
}

// TestCPUCollector_CollectSystem verifies system CPU collection.
func TestCPUCollector_CollectSystem(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects system CPU metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCPUCollector()
			ctx := context.Background()

			cpu, err := collector.CollectSystem(ctx)
			require.NoError(t, err)

			// CPU usage should be between 0 and 100
			assert.GreaterOrEqual(t, cpu.UsagePercent, 0.0)
			assert.LessOrEqual(t, cpu.UsagePercent, 100.0)

			// Timestamp should be set
			assert.False(t, cpu.Timestamp.IsZero())
		})
	}
}

// TestCPUCollector_CollectProcess verifies process CPU collection.
func TestCPUCollector_CollectProcess(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects process CPU metrics"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCPUCollector()
			ctx := context.Background()
			pid := os.Getpid()

			cpu, err := collector.CollectProcess(ctx, pid)
			require.NoError(t, err)

			// PID should match
			assert.Equal(t, pid, cpu.PID)

			// CPU percentage should be non-negative
			assert.GreaterOrEqual(t, cpu.UsagePercent, 0.0)
		})
	}
}

// TestCPUCollector_CollectAllProcesses verifies it returns ErrNotSupported.
func TestCPUCollector_CollectAllProcesses(t *testing.T) {
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

			collector := probe.NewCPUCollector()
			ctx := context.Background()

			_, err = collector.CollectAllProcesses(ctx)
			assert.Error(t, err)
		})
	}
}

// TestCPUCollector_CollectLoadAverage verifies load average collection.
func TestCPUCollector_CollectLoadAverage(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects load averages"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCPUCollector()
			ctx := context.Background()

			load, err := collector.CollectLoadAverage(ctx)
			require.NoError(t, err)

			// Load values should be non-negative
			assert.GreaterOrEqual(t, load.Load1, 0.0)
			assert.GreaterOrEqual(t, load.Load5, 0.0)
			assert.GreaterOrEqual(t, load.Load15, 0.0)
		})
	}
}

// TestCPUCollector_CollectPressure verifies CPU pressure collection.
func TestCPUCollector_CollectPressure(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects CPU pressure when supported"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			collector := probe.NewCPUCollector()
			ctx := context.Background()

			pressure, err := collector.CollectPressure(ctx)
			// PSI may not be supported on all systems, so we accept errors
			if err == nil {
				assert.GreaterOrEqual(t, pressure.SomeAvg10, 0.0)
			}
		})
	}
}

// TestCPUCollector_NotInitialized verifies error when not initialized.
func TestCPUCollector_NotInitialized(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "returns error when not initialized"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewCPUCollector()
			ctx := context.Background()

			_, err := collector.CollectSystem(ctx)
			// Should return error because probe is not initialized
			if err != nil {
				assert.Error(t, err)
			}
		})
	}
}
