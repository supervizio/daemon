//go:build cgo

package probe

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestFullPercentageConstant verifies the percentage constant value.
func TestFullPercentageConstant(t *testing.T) {
	tests := []struct {
		name     string
		expected float64
	}{
		{
			name:     "fullPercentage equals 100",
			expected: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, fullPercentage)
		})
	}
}

// TestCollectAll verifies the CollectAll function.
func TestCollectAll(t *testing.T) {
	tests := []struct {
		name        string
		initProbe   bool
		expectError bool
	}{
		{
			name:        "with initialized probe",
			initProbe:   true,
			expectError: false,
		},
		{
			name:        "without initialized probe",
			initProbe:   false,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.initProbe {
				err := Init()
				require.NoError(t, err)
				defer Shutdown()
			}

			metrics, err := CollectAll()

			if tt.expectError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotNil(t, metrics)
				assert.False(t, metrics.Timestamp.IsZero())
			}
		})
	}
}

// TestBuildAllMetrics verifies the buildAllMetrics function.
func TestBuildAllMetrics(t *testing.T) {
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

			var cAll C.AllMetrics
			cAll.timestamp_ns = C.uint64_t(time.Now().UnixNano())

			result := buildAllMetrics(&cAll)

			assert.NotNil(t, result)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestBuildCPUMetrics verifies the buildCPUMetrics function.
func TestBuildCPUMetrics(t *testing.T) {
	tests := []struct {
		name        string
		idlePercent float32
		expectUsage float64
	}{
		{
			name:        "0 percent idle",
			idlePercent: 0.0,
			expectUsage: 100.0,
		},
		{
			name:        "50 percent idle",
			idlePercent: 50.0,
			expectUsage: 50.0,
		},
		{
			name:        "100 percent idle",
			idlePercent: 100.0,
			expectUsage: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.cpu.idle_percent = C.float(tt.idlePercent)

			result := buildCPUMetrics(&cAll)

			assert.InDelta(t, tt.expectUsage, result.UsagePercent, 0.01)
		})
	}
}

// TestBuildMemoryMetrics verifies the buildMemoryMetrics function.
func TestBuildMemoryMetrics(t *testing.T) {
	tests := []struct {
		name       string
		total      uint64
		available  uint64
		used       uint64
		expectUsed uint64
	}{
		{
			name:       "basic memory values",
			total:      1024,
			available:  512,
			used:       512,
			expectUsed: 512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.memory.total_bytes = C.uint64_t(tt.total)
			cAll.memory.available_bytes = C.uint64_t(tt.available)
			cAll.memory.used_bytes = C.uint64_t(tt.used)

			result := buildMemoryMetrics(&cAll)

			assert.Equal(t, tt.total, result.Total)
			assert.Equal(t, tt.available, result.Available)
			assert.Equal(t, tt.expectUsed, result.Used)
		})
	}
}

// TestBuildLoadMetrics verifies the buildLoadMetrics function.
func TestBuildLoadMetrics(t *testing.T) {
	tests := []struct {
		name  string
		load1 float32
		load5 float32
		load15 float32
	}{
		{
			name:  "normal load",
			load1: 1.0,
			load5: 0.5,
			load15: 0.25,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.load.load_1min = C.float(tt.load1)
			cAll.load.load_5min = C.float(tt.load5)
			cAll.load.load_15min = C.float(tt.load15)

			result := buildLoadMetrics(&cAll)

			assert.InDelta(t, float64(tt.load1), result.Load1, 0.01)
			assert.InDelta(t, float64(tt.load5), result.Load5, 0.01)
			assert.InDelta(t, float64(tt.load15), result.Load15, 0.01)
		})
	}
}

// TestBuildIOStatsMetrics verifies the buildIOStatsMetrics function.
func TestBuildIOStatsMetrics(t *testing.T) {
	tests := []struct {
		name       string
		readOps    uint64
		readBytes  uint64
		writeOps   uint64
		writeBytes uint64
	}{
		{
			name:       "basic io values",
			readOps:    100,
			readBytes:  1024,
			writeOps:   50,
			writeBytes: 512,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.io_stats.read_ops = C.uint64_t(tt.readOps)
			cAll.io_stats.read_bytes = C.uint64_t(tt.readBytes)
			cAll.io_stats.write_ops = C.uint64_t(tt.writeOps)
			cAll.io_stats.write_bytes = C.uint64_t(tt.writeBytes)

			result := buildIOStatsMetrics(&cAll)

			assert.Equal(t, tt.readOps, result.ReadOps)
			assert.Equal(t, tt.readBytes, result.ReadBytes)
			assert.Equal(t, tt.writeOps, result.WriteOps)
			assert.Equal(t, tt.writeBytes, result.WriteBytes)
		})
	}
}

// TestBuildPressureMetrics verifies the buildPressureMetrics function.
func TestBuildPressureMetrics(t *testing.T) {
	tests := []struct {
		name       string
		cpuAvg10   float32
		memAvg10   float32
		ioAvg10    float32
	}{
		{
			name:       "normal pressure",
			cpuAvg10:   5.0,
			memAvg10:   10.0,
			ioAvg10:    2.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.pressure.cpu.some_avg10 = C.float(tt.cpuAvg10)
			cAll.pressure.memory.some_avg10 = C.float(tt.memAvg10)
			cAll.pressure.io.some_avg10 = C.float(tt.ioAvg10)

			result := buildPressureMetrics(&cAll)

			assert.InDelta(t, float64(tt.cpuAvg10), result.CPU.SomeAvg10, 0.01)
			assert.InDelta(t, float64(tt.memAvg10), result.Memory.SomeAvg10, 0.01)
			assert.InDelta(t, float64(tt.ioAvg10), result.IO.SomeAvg10, 0.01)
		})
	}
}

// TestBuildPartitions verifies the buildPartitions function.
func TestBuildPartitions(t *testing.T) {
	tests := []struct {
		name      string
		count     uint32
		expectLen int
	}{
		{
			name:      "empty partitions",
			count:     0,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.partition_count = C.uint32_t(tt.count)

			result := buildPartitions(&cAll)

			assert.Len(t, result, tt.expectLen)
		})
	}
}

// TestBuildDiskUsage verifies the buildDiskUsage function.
func TestBuildDiskUsage(t *testing.T) {
	tests := []struct {
		name      string
		count     uint32
		expectLen int
	}{
		{
			name:      "empty disk usage",
			count:     0,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.disk_usage_count = C.uint32_t(tt.count)

			result := buildDiskUsage(&cAll)

			assert.Len(t, result, tt.expectLen)
		})
	}
}

// TestBuildDiskIO verifies the buildDiskIO function.
func TestBuildDiskIO(t *testing.T) {
	tests := []struct {
		name      string
		count     uint32
		expectLen int
	}{
		{
			name:      "empty disk io",
			count:     0,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.disk_io_count = C.uint32_t(tt.count)

			result := buildDiskIO(&cAll)

			assert.Len(t, result, tt.expectLen)
		})
	}
}

// TestBuildNetInterfaces verifies the buildNetInterfaces function.
func TestBuildNetInterfaces(t *testing.T) {
	tests := []struct {
		name      string
		count     uint32
		expectLen int
	}{
		{
			name:      "empty interfaces",
			count:     0,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.net_interface_count = C.uint32_t(tt.count)

			result := buildNetInterfaces(&cAll)

			assert.Len(t, result, tt.expectLen)
		})
	}
}

// TestBuildNetStats verifies the buildNetStats function.
func TestBuildNetStats(t *testing.T) {
	tests := []struct {
		name      string
		count     uint32
		expectLen int
	}{
		{
			name:      "empty net stats",
			count:     0,
			expectLen: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var cAll C.AllMetrics
			cAll.net_stats_count = C.uint32_t(tt.count)

			result := buildNetStats(&cAll)

			assert.Len(t, result, tt.expectLen)
		})
	}
}
