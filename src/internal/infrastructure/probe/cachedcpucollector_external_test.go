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

func TestNewCachedCPUCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsNonNilCollector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewCachedCPUCollector()
			assert.NotNil(t, collector)
		})
	}
}

func TestCachedCPUCollector_CollectSystem(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		wantErr   bool
	}{
		{
			name:      "CollectsWhenInitialized",
			initProbe: true,
			wantErr:   false,
		},
		{
			name:      "FailsWhenNotInitialized",
			initProbe: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe.Shutdown()

			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewCachedCPUCollector()
			cpu, err := collector.CollectSystem(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, cpu.UsagePercent, float64(0))
				assert.LessOrEqual(t, cpu.UsagePercent, float64(100))
				assert.False(t, cpu.Timestamp.IsZero())
			}
		})
	}
}

func TestCachedCPUCollector_CollectProcess(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		pid       int
		wantErr   bool
	}{
		{
			name:      "CollectsValidPID",
			initProbe: true,
			pid:       os.Getpid(),
			wantErr:   false,
		},
		{
			name:      "FailsWhenNotInitialized",
			initProbe: false,
			pid:       os.Getpid(),
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe.Shutdown()

			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewCachedCPUCollector()
			cpu, err := collector.CollectProcess(context.Background(), tt.pid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pid, cpu.PID)
			}
		})
	}
}

func TestCachedCPUCollector_CollectAllProcesses(t *testing.T) {
	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "ReturnsNotSupported",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewCachedCPUCollector()
			_, err := collector.CollectAllProcesses(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, probe.ErrNotSupported)
			}
		})
	}
}

func TestCachedCPUCollector_CollectLoadAverage(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		wantErr   bool
	}{
		{
			name:      "CollectsWhenInitialized",
			initProbe: true,
			wantErr:   false,
		},
		{
			name:      "FailsWhenNotInitialized",
			initProbe: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe.Shutdown()

			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewCachedCPUCollector()
			load, err := collector.CollectLoadAverage(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.GreaterOrEqual(t, load.Load1, float64(0))
				assert.GreaterOrEqual(t, load.Load5, float64(0))
				assert.GreaterOrEqual(t, load.Load15, float64(0))
			}
		})
	}
}

func TestCachedCPUCollector_CollectPressure(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		wantErr   bool
	}{
		{
			name:      "CollectsWhenInitialized",
			initProbe: true,
			wantErr:   false,
		},
		{
			name:      "FailsWhenNotInitialized",
			initProbe: false,
			wantErr:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe.Shutdown()

			if tt.initProbe {
				err := probe.Init()
				require.NoError(t, err)
				defer probe.Shutdown()
			}

			collector := probe.NewCachedCPUCollector()
			_, err := collector.CollectPressure(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// Pressure may not be supported on all platforms
				// so we just check it doesn't panic
				t.Logf("CollectPressure error: %v", err)
			}
		})
	}
}
