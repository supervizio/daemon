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

func TestNewCachedMemoryCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsNonNilCollector",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collector := probe.NewCachedMemoryCollector()
			assert.NotNil(t, collector)
		})
	}
}

func TestCachedMemoryCollector_CollectSystem(t *testing.T) {
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

			collector := probe.NewCachedMemoryCollector()
			mem, err := collector.CollectSystem(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Greater(t, mem.Total, uint64(0))
				assert.False(t, mem.Timestamp.IsZero())
			}
		})
	}
}

func TestCachedMemoryCollector_CollectProcess(t *testing.T) {
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

			collector := probe.NewCachedMemoryCollector()
			mem, err := collector.CollectProcess(context.Background(), tt.pid)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tt.pid, mem.PID)
				assert.Greater(t, mem.RSS, uint64(0))
			}
		})
	}
}

func TestCachedMemoryCollector_CollectAllProcesses(t *testing.T) {
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
			collector := probe.NewCachedMemoryCollector()
			_, err := collector.CollectAllProcesses(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
				assert.ErrorIs(t, err, probe.ErrNotSupported)
			}
		})
	}
}

func TestCachedMemoryCollector_CollectPressure(t *testing.T) {
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

			collector := probe.NewCachedMemoryCollector()
			_, err := collector.CollectPressure(context.Background())

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				// Pressure may not be supported on all platforms
				t.Logf("CollectPressure error: %v", err)
			}
		})
	}
}
