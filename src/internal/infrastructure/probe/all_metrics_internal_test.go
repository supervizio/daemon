//go:build cgo

package probe

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestJoinOptions verifies option string joining.
func TestJoinOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		opts     []string
		expected string
	}{
		{
			name:     "empty slice",
			opts:     []string{},
			expected: "",
		},
		{
			name:     "single option",
			opts:     []string{"rw"},
			expected: "rw",
		},
		{
			name:     "multiple options",
			opts:     []string{"rw", "noatime", "nodiratime"},
			expected: "rw,noatime,nodiratime",
		},
		{
			name:     "nil slice",
			opts:     nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := joinOptions(tt.opts)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestContainsFlag verifies flag checking.
func TestContainsFlag(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    []string
		flag     string
		expected bool
	}{
		{
			name:     "flag present",
			flags:    []string{"up", "loopback", "running"},
			flag:     "up",
			expected: true,
		},
		{
			name:     "flag absent",
			flags:    []string{"up", "running"},
			flag:     "loopback",
			expected: false,
		},
		{
			name:     "empty slice",
			flags:    []string{},
			flag:     "up",
			expected: false,
		},
		{
			name:     "nil slice",
			flags:    nil,
			flag:     "up",
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := containsFlag(tt.flags, tt.flag)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// TestCollectBasicMetrics verifies basic metrics collection.
func TestCollectBasicMetrics(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		expectCPU bool
		expectMem bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
			expectCPU: true,
			expectMem: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()
			result := &AllSystemMetrics{}

			collectBasicMetrics(ctx, collector, result)

			if tt.expectCPU {
				assert.NotNil(t, result.CPU)
			}
			if tt.expectMem {
				assert.NotNil(t, result.Memory)
			}
		})
	}
}

// TestCollectCPUMetricsWithPressure verifies CPU metrics with pressure collection.
func TestCollectCPUMetricsWithPressure(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectNil   bool
		expectUsage bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectNil:   false,
			expectUsage: true,
		},
		{
			name:      "without initialized probe",
			initProbe: false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := collectCPUMetricsWithPressure(ctx, collector)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if tt.expectUsage {
					assert.GreaterOrEqual(t, result.UsagePercent, 0.0)
					assert.LessOrEqual(t, result.UsagePercent, 100.0)
				}
			}
		})
	}
}

// TestCollectMemoryMetricsWithPressure verifies memory metrics with pressure collection.
func TestCollectMemoryMetricsWithPressure(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectNil   bool
		expectTotal bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectNil:   false,
			expectTotal: true,
		},
		{
			name:      "without initialized probe",
			initProbe: false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := collectMemoryMetricsWithPressure(ctx, collector)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				if tt.expectTotal {
					assert.Greater(t, result.TotalBytes, uint64(0))
				}
			}
		})
	}
}

// TestCollectLoadMetricsJSON verifies load average metrics collection.
func TestCollectLoadMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		expectNil bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
			expectNil: false,
		},
		{
			name:      "without initialized probe",
			initProbe: false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := collectLoadMetricsJSON(ctx, collector)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
				assert.GreaterOrEqual(t, result.Load1Min, 0.0)
			}
		})
	}
}

// TestCollectResourceMetrics verifies resource metrics collection.
func TestCollectResourceMetrics(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()
			result := &AllSystemMetrics{}

			collectResourceMetrics(ctx, collector, result)

			// Disk metrics should be populated
			assert.NotNil(t, result.Disk)
		})
	}
}

// TestCollectSystemMetrics verifies system metrics collection.
func TestCollectSystemMetrics(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			result := &AllSystemMetrics{}

			collectSystemMetrics(ctx, result)

			// Process metrics should be populated
			assert.NotNil(t, result.Process)
		})
	}
}

// TestCollectDiskMetricsJSON verifies disk metrics JSON collection.
func TestCollectDiskMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		expectNil bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
			expectNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := collectDiskMetricsJSON(ctx, collector)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestExtractPartitionInfo verifies partition info extraction.
func TestExtractPartitionInfo(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		expectNil bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
			expectNil: false,
		},
		{
			name:      "without initialized probe",
			initProbe: false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := extractPartitionInfo(ctx, collector)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				// On a running system we should have at least one partition
				assert.NotNil(t, result)
			}
		})
	}
}

// TestExtractDiskUsageInfo verifies disk usage info extraction.
func TestExtractDiskUsageInfo(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		expectNil bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
			expectNil: false,
		},
		{
			name:      "without initialized probe",
			initProbe: false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := extractDiskUsageInfo(ctx, collector)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				assert.NotNil(t, result)
			}
		})
	}
}

// TestExtractDiskIOInfo verifies disk I/O info extraction.
func TestExtractDiskIOInfo(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
		expectNil bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
			expectNil: false,
		},
		{
			name:      "without initialized probe",
			initProbe: false,
			expectNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := extractDiskIOInfo(ctx, collector)

			if tt.expectNil {
				assert.Nil(t, result)
			} else {
				// I/O info may be empty on some systems but should not be nil
				assert.NotNil(t, result)
			}
		})
	}
}

// TestCollectNetworkMetricsJSON verifies network metrics JSON collection.
func TestCollectNetworkMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := collectNetworkMetricsJSON(ctx, collector)

			assert.NotNil(t, result)
		})
	}
}

// TestCollectIOMetricsJSON verifies I/O metrics JSON collection.
func TestCollectIOMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewCollector()

			result := collectIOMetricsJSON(ctx, collector)

			assert.NotNil(t, result)
		})
	}
}

// TestCollectProcessMetricsJSON verifies process metrics JSON collection.
func TestCollectProcessMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()

			result := collectProcessMetricsJSON(ctx)

			assert.NotNil(t, result)
			assert.Greater(t, result.CurrentPID, int32(0))
		})
	}
}

// TestCollectThermalMetricsJSON verifies thermal metrics JSON collection.
func TestCollectThermalMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			result := collectThermalMetricsJSON()

			assert.NotNil(t, result)
			// Supported field should be set based on platform
		})
	}
}

// TestCollectContextSwitchMetricsJSON verifies context switch metrics JSON collection.
func TestCollectContextSwitchMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			result := collectContextSwitchMetricsJSON()

			assert.NotNil(t, result)
		})
	}
}

// TestCollectConnectionMetricsJSON verifies connection metrics JSON collection.
func TestCollectConnectionMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()

			result := collectConnectionMetricsJSON(ctx)

			assert.NotNil(t, result)
		})
	}
}

// TestCollectTCPStatsJSON verifies TCP stats JSON collection.
func TestCollectTCPStatsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewConnectionCollector()

			result := collectTCPStatsJSON(ctx, collector)

			// May be nil if TCP stats collection fails
			if result != nil {
				assert.GreaterOrEqual(t, result.Total, uint32(0))
			}
		})
	}
}

// TestCollectTCPConnectionsJSON verifies TCP connections JSON collection.
func TestCollectTCPConnectionsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewConnectionCollector()

			result := collectTCPConnectionsJSON(ctx, collector)

			// Result may be nil or empty slice depending on system state
			if result != nil {
				for _, conn := range result {
					assert.NotEmpty(t, conn.Family)
				}
			}
		})
	}
}

// TestCollectUDPSocketsJSON verifies UDP sockets JSON collection.
func TestCollectUDPSocketsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewConnectionCollector()

			result := collectUDPSocketsJSON(ctx, collector)

			// Result may be nil or empty slice depending on system state
			if result != nil {
				for _, sock := range result {
					assert.NotEmpty(t, sock.Family)
				}
			}
		})
	}
}

// TestCollectUnixSocketsJSON verifies Unix sockets JSON collection.
func TestCollectUnixSocketsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewConnectionCollector()

			result := collectUnixSocketsJSON(ctx, collector)

			// Result may be nil or empty slice depending on system state
			_ = result // Just verify no panic
		})
	}
}

// TestCollectListeningPortsJSON verifies listening ports JSON collection.
func TestCollectListeningPortsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			ctx := context.Background()
			collector := NewConnectionCollector()

			result := collectListeningPortsJSON(ctx, collector)

			// Result may be nil or empty slice depending on system state
			if result != nil {
				for _, port := range result {
					assert.Equal(t, "tcp", port.Protocol)
				}
			}
		})
	}
}

// TestCollectQuotaMetricsJSON verifies quota metrics JSON collection.
func TestCollectQuotaMetricsJSON(t *testing.T) {
	tests := []struct {
		name      string
		initProbe bool
	}{
		{
			name:      "with initialized probe",
			initProbe: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			result := collectQuotaMetricsJSON()

			assert.NotNil(t, result)
		})
	}
}
