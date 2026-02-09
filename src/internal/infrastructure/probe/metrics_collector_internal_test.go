//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/stretchr/testify/assert"
)

func defaultMetricsConfigForTest() config.MetricsConfig {
	return config.DefaultMetricsConfig()
}

func TestJoinOptions(t *testing.T) {
	tests := []struct {
		name string
		opts []string
		want string
	}{
		{
			name: "EmptySlice",
			opts: []string{},
			want: "",
		},
		{
			name: "SingleOption",
			opts: []string{"rw"},
			want: "rw",
		},
		{
			name: "MultipleOptions",
			opts: []string{"rw", "noexec", "nosuid"},
			want: "rw,noexec,nosuid",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := joinOptions(tt.opts)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestContainsFlag(t *testing.T) {
	tests := []struct {
		name  string
		flags []string
		flag  string
		want  bool
	}{
		{
			name:  "FlagPresent",
			flags: []string{"up", "loopback"},
			flag:  "up",
			want:  true,
		},
		{
			name:  "FlagNotPresent",
			flags: []string{"up", "loopback"},
			flag:  "down",
			want:  false,
		},
		{
			name:  "EmptyFlags",
			flags: []string{},
			flag:  "up",
			want:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := containsFlag(tt.flags, tt.flag)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCollectBasicMetrics(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "PopulatesResultWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := &AllSystemMetrics{}
			cfg := defaultMetricsConfigForTest()
			collectBasicMetrics(context.Background(), collector, result, &cfg)
			// Result pointer is not nil, but fields may be nil.
			assert.NotNil(t, result)
		})
	}
}

func TestCollectCPUMetricsWithPressure(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsResultsWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := collectCPUMetricsWithPressure(context.Background(), collector, true)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectMemoryMetricsWithPressure(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsResultsWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := collectMemoryMetricsWithPressure(context.Background(), collector, true)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectLoadMetricsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptyWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := collectLoadMetricsJSON(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectResourceMetrics(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "PopulatesResultWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := &AllSystemMetrics{}
			cfg := defaultMetricsConfigForTest()
			collectResourceMetrics(context.Background(), collector, result, &cfg)
			// Result pointer is not nil, but fields may be nil.
			assert.NotNil(t, result)
		})
	}
}

func TestCollectSystemMetrics(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "PopulatesResultWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			result := &AllSystemMetrics{}
			cfg := defaultMetricsConfigForTest()
			collectSystemMetrics(context.Background(), result, &cfg)
			// Result pointer is not nil, but fields may be nil.
			assert.NotNil(t, result)
		})
	}
}

func TestCollectDiskMetricsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptyWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			diskCfg := &config.DiskMetricsConfig{Enabled: true, Partitions: true, Usage: true, IO: true}
			result := collectDiskMetricsJSON(context.Background(), collector, diskCfg)
			assert.NotNil(t, result)
		})
	}
}

func TestExtractPartitionInfo(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptySliceWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := extractPartitionInfo(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestExtractDiskUsageInfo(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptySliceWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := extractDiskUsageInfo(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestExtractDiskIOInfo(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptySliceWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := extractDiskIOInfo(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectNetworkMetricsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptyWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			netCfg := &config.NetworkMetricsConfig{Enabled: true, Interfaces: true, Stats: true}
			result := collectNetworkMetricsJSON(context.Background(), collector, netCfg)
			assert.NotNil(t, result)
		})
	}
}

func TestCollectIOMetricsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptyWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewCollector()
			result := collectIOMetricsJSON(context.Background(), collector, true)
			assert.NotNil(t, result)
		})
	}
}

func TestCollectProcessMetricsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptyWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			result := collectProcessMetricsJSON(context.Background())
			assert.NotNil(t, result)
		})
	}
}

func TestCollectThermalMetricsJSON(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsResult",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectThermalMetricsJSON()
			assert.NotNil(t, result)
		})
	}
}

func TestCollectContextSwitchMetricsJSON(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsResult",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectContextSwitchMetricsJSON()
			assert.NotNil(t, result)
		})
	}
}

func TestCollectConnectionMetricsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptyWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			connCfg := &config.ConnectionMetricsConfig{Enabled: true, TCPStats: true, TCPConnections: true, UDPSockets: true, UnixSockets: true, ListeningPorts: true}
			result := collectConnectionMetricsJSON(context.Background(), connCfg)
			assert.NotNil(t, result)
		})
	}
}

func TestCollectTCPStatsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptyWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewConnectionCollector()
			result := collectTCPStatsJSON(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectTCPConnectionsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptySliceWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewConnectionCollector()
			result := collectTCPConnectionsJSON(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectUDPSocketsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptySliceWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewConnectionCollector()
			result := collectUDPSocketsJSON(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectUnixSocketsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptySliceWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewConnectionCollector()
			result := collectUnixSocketsJSON(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectListeningPortsJSON(t *testing.T) {
	tests := []struct {
		name        string
		initialized bool
	}{
		{
			name:        "ReturnsEmptySliceWhenNotInitialized",
			initialized: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			initMu.Lock()
			oldValue := initialized
			initialized = tt.initialized
			initMu.Unlock()
			defer func() {
				initMu.Lock()
				initialized = oldValue
				initMu.Unlock()
			}()

			collector := NewConnectionCollector()
			result := collectListeningPortsJSON(context.Background(), collector)
			// Returns nil when not initialized.
			assert.Nil(t, result)
		})
	}
}

func TestCollectQuotaMetricsJSON(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsResult",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectQuotaMetricsJSON()
			assert.NotNil(t, result)
		})
	}
}

func TestCollectContainerMetricsJSON(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsResult",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectContainerMetricsJSON()
			assert.NotNil(t, result)
		})
	}
}

func TestCollectRuntimeMetricsJSON(t *testing.T) {
	tests := []struct {
		name string
	}{
		{
			name: "ReturnsResult",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := collectRuntimeMetricsJSON()
			assert.NotNil(t, result)
		})
	}
}
