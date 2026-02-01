// Package probe provides internal tests for the builder functions.
package probe

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestBuildCPUMetricsFromRaw tests the buildCPUMetricsFromRaw function.
func TestBuildCPUMetricsFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name         string
		raw          *RawCPUData
		wantUsage    float64
	}{
		{
			name:      "zero_idle",
			raw:       &RawCPUData{IdlePercent: 0},
			wantUsage: 100.0,
		},
		{
			name:      "full_idle",
			raw:       &RawCPUData{IdlePercent: 100},
			wantUsage: 0.0,
		},
		{
			name:      "half_idle",
			raw:       &RawCPUData{IdlePercent: 50},
			wantUsage: 50.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build CPU metrics from raw data.
			result := buildCPUMetricsFromRaw(tt.raw)
			// Verify usage percent.
			assert.Equal(t, tt.wantUsage, result.UsagePercent)
			// Verify timestamp is set.
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestBuildMemoryMetricsFromRaw tests the buildMemoryMetricsFromRaw function.
func TestBuildMemoryMetricsFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  *RawMemoryData
	}{
		{
			name: "basic_memory",
			raw: &RawMemoryData{
				TotalBytes:     16000000000,
				AvailableBytes: 8000000000,
				UsedBytes:      8000000000,
				CachedBytes:    2000000000,
				BuffersBytes:   500000000,
				SwapTotalBytes: 4000000000,
				SwapUsedBytes:  1000000000,
			},
		},
		{
			name: "zero_swap",
			raw: &RawMemoryData{
				TotalBytes:     8000000000,
				AvailableBytes: 4000000000,
				UsedBytes:      4000000000,
				SwapTotalBytes: 0,
				SwapUsedBytes:  0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build memory metrics from raw data.
			result := buildMemoryMetricsFromRaw(tt.raw)
			// Verify memory values.
			assert.Equal(t, tt.raw.TotalBytes, result.Total)
			assert.Equal(t, tt.raw.AvailableBytes, result.Available)
			assert.Equal(t, tt.raw.UsedBytes, result.Used)
			// Verify swap free calculation.
			assert.Equal(t, tt.raw.SwapTotalBytes-tt.raw.SwapUsedBytes, result.SwapFree)
		})
	}
}

// TestBuildLoadMetricsFromRaw tests the buildLoadMetricsFromRaw function.
func TestBuildLoadMetricsFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  *RawLoadData
	}{
		{
			name: "normal_load",
			raw: &RawLoadData{
				Load1Min:  1.5,
				Load5Min:  1.2,
				Load15Min: 0.8,
			},
		},
		{
			name: "zero_load",
			raw: &RawLoadData{
				Load1Min:  0.0,
				Load5Min:  0.0,
				Load15Min: 0.0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build load metrics from raw data.
			result := buildLoadMetricsFromRaw(tt.raw)
			// Verify load values.
			assert.Equal(t, tt.raw.Load1Min, result.Load1)
			assert.Equal(t, tt.raw.Load5Min, result.Load5)
			assert.Equal(t, tt.raw.Load15Min, result.Load15)
		})
	}
}

// TestBuildIOStatsMetricsFromRaw tests the buildIOStatsMetricsFromRaw function.
func TestBuildIOStatsMetricsFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  *RawIOStatsData
	}{
		{
			name: "normal_io",
			raw: &RawIOStatsData{
				ReadOps:    1000,
				ReadBytes:  50000000,
				WriteOps:   500,
				WriteBytes: 25000000,
			},
		},
		{
			name: "zero_io",
			raw: &RawIOStatsData{
				ReadOps:    0,
				ReadBytes:  0,
				WriteOps:   0,
				WriteBytes: 0,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build I/O stats from raw data.
			result := buildIOStatsMetricsFromRaw(tt.raw)
			// Verify I/O values.
			assert.Equal(t, tt.raw.ReadOps, result.ReadOps)
			assert.Equal(t, tt.raw.ReadBytes, result.ReadBytes)
			assert.Equal(t, tt.raw.WriteOps, result.WriteOps)
			assert.Equal(t, tt.raw.WriteBytes, result.WriteBytes)
		})
	}
}

// TestBuildPressureMetricsFromRaw tests the buildPressureMetricsFromRaw function.
func TestBuildPressureMetricsFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  *RawPressureMetrics
	}{
		{
			name: "normal_pressure",
			raw: &RawPressureMetrics{
				Available: true,
				CPU: RawCPUPressure{
					SomeAvg10:   1.5,
					SomeAvg60:   1.2,
					SomeAvg300:  0.8,
					SomeTotalUs: 1000000,
				},
				Memory: RawMemoryPressure{
					SomeAvg10:   0.5,
					FullAvg10:   0.2,
					SomeTotalUs: 500000,
					FullTotalUs: 200000,
				},
				IO: RawIOPressure{
					SomeAvg10:   0.3,
					FullAvg10:   0.1,
					SomeTotalUs: 300000,
					FullTotalUs: 100000,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build pressure metrics from raw data.
			result := buildPressureMetricsFromRaw(tt.raw)
			// Verify CPU pressure.
			assert.Equal(t, tt.raw.CPU.SomeAvg10, result.CPU.SomeAvg10)
			assert.Equal(t, tt.raw.CPU.SomeTotalUs, result.CPU.SomeTotal)
			// Verify memory pressure.
			assert.Equal(t, tt.raw.Memory.SomeAvg10, result.Memory.SomeAvg10)
			assert.Equal(t, tt.raw.Memory.FullAvg10, result.Memory.FullAvg10)
			// Verify I/O pressure.
			assert.Equal(t, tt.raw.IO.SomeAvg10, result.IO.SomeAvg10)
			assert.Equal(t, tt.raw.IO.FullAvg10, result.IO.FullAvg10)
		})
	}
}

// TestBuildPartitionsFromRaw tests the buildPartitionsFromRaw function.
func TestBuildPartitionsFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  []RawPartitionData
		want int
	}{
		{
			name: "single_partition",
			raw: []RawPartitionData{
				{Device: "/dev/sda1", MountPoint: "/", FSType: "ext4", Options: "rw,noatime"},
			},
			want: 1,
		},
		{
			name: "multiple_partitions",
			raw: []RawPartitionData{
				{Device: "/dev/sda1", MountPoint: "/", FSType: "ext4", Options: "rw"},
				{Device: "/dev/sda2", MountPoint: "/home", FSType: "ext4", Options: "rw"},
			},
			want: 2,
		},
		{
			name: "empty_partitions",
			raw:  []RawPartitionData{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build partitions from raw data.
			result := buildPartitionsFromRaw(tt.raw)
			// Verify count.
			assert.Len(t, result, tt.want)
			// Verify data if not empty.
			if tt.want > 0 {
				assert.Equal(t, tt.raw[0].Device, result[0].Device)
				assert.Equal(t, tt.raw[0].MountPoint, result[0].MountPoint)
			}
		})
	}
}

// TestBuildDiskUsageFromRaw tests the buildDiskUsageFromRaw function.
func TestBuildDiskUsageFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  []RawDiskUsageData
		want int
	}{
		{
			name: "single_usage",
			raw: []RawDiskUsageData{
				{Path: "/", TotalBytes: 100000000000, UsedBytes: 50000000000, FreeBytes: 50000000000, UsedPercent: 50.0},
			},
			want: 1,
		},
		{
			name: "empty_usage",
			raw:  []RawDiskUsageData{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build disk usage from raw data.
			result := buildDiskUsageFromRaw(tt.raw)
			// Verify count.
			assert.Len(t, result, tt.want)
			// Verify data if not empty.
			if tt.want > 0 {
				assert.Equal(t, tt.raw[0].Path, result[0].Path)
				assert.Equal(t, tt.raw[0].TotalBytes, result[0].TotalBytes)
			}
		})
	}
}

// TestBuildDiskIOFromRaw tests the buildDiskIOFromRaw function.
func TestBuildDiskIOFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  []RawDiskIOData
		want int
	}{
		{
			name: "single_disk_io",
			raw: []RawDiskIOData{
				{Device: "sda", ReadsCompleted: 1000, WritesCompleted: 500},
			},
			want: 1,
		},
		{
			name: "empty_disk_io",
			raw:  []RawDiskIOData{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build disk I/O from raw data.
			result := buildDiskIOFromRaw(tt.raw)
			// Verify count.
			assert.Len(t, result, tt.want)
			// Verify data if not empty.
			if tt.want > 0 {
				assert.Equal(t, tt.raw[0].Device, result[0].Device)
				assert.Equal(t, tt.raw[0].ReadsCompleted, result[0].ReadsCompleted)
			}
		})
	}
}

// TestBuildNetInterfacesFromRaw tests the buildNetInterfacesFromRaw function.
func TestBuildNetInterfacesFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  []RawNetInterfaceData
		want int
	}{
		{
			name: "single_interface",
			raw: []RawNetInterfaceData{
				{Name: "eth0", MACAddress: "00:11:22:33:44:55", MTU: 1500, IsUp: true, IsLoopback: false},
			},
			want: 1,
		},
		{
			name: "multiple_interfaces",
			raw: []RawNetInterfaceData{
				{Name: "eth0", IsUp: true},
				{Name: "lo", IsLoopback: true},
			},
			want: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build net interfaces from raw data.
			result := buildNetInterfacesFromRaw(tt.raw)
			// Verify count.
			assert.Len(t, result, tt.want)
			// Verify data if not empty.
			if tt.want > 0 {
				assert.Equal(t, tt.raw[0].Name, result[0].Name)
			}
		})
	}
}

// TestBuildNetStatsFromRaw tests the buildNetStatsFromRaw function.
func TestBuildNetStatsFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		raw  []RawNetStatsData
		want int
	}{
		{
			name: "single_stats",
			raw: []RawNetStatsData{
				{Interface: "eth0", RxBytes: 1000000, TxBytes: 500000},
			},
			want: 1,
		},
		{
			name: "empty_stats",
			raw:  []RawNetStatsData{},
			want: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build net stats from raw data.
			result := buildNetStatsFromRaw(tt.raw)
			// Verify count.
			assert.Len(t, result, tt.want)
			// Verify data if not empty.
			if tt.want > 0 {
				assert.Equal(t, tt.raw[0].Interface, result[0].Interface)
				assert.Equal(t, tt.raw[0].RxBytes, result[0].RxBytes)
			}
		})
	}
}

// TestBuildAllMetricsFromRaw tests the buildAllMetricsFromRaw function.
func TestBuildAllMetricsFromRaw(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name             string
		raw              *RawAllMetrics
		wantPressure     bool
	}{
		{
			name: "full_metrics",
			raw: &RawAllMetrics{
				TimestampNs: 1000000000,
				CPU:         RawCPUData{IdlePercent: 50},
				Memory:      RawMemoryData{TotalBytes: 16000000000},
				Load:        RawLoadData{Load1Min: 1.0},
				IOStats:     RawIOStatsData{ReadOps: 100},
				Pressure:    RawPressureMetrics{Available: true},
				Partitions:  []RawPartitionData{{Device: "/dev/sda1"}},
				DiskUsage:   []RawDiskUsageData{{Path: "/"}},
				DiskIO:      []RawDiskIOData{{Device: "sda"}},
				NetInterfaces: []RawNetInterfaceData{{Name: "eth0"}},
				NetStats:    []RawNetStatsData{{Interface: "eth0"}},
			},
			wantPressure: true,
		},
		{
			name: "no_pressure",
			raw: &RawAllMetrics{
				TimestampNs: 2000000000,
				CPU:         RawCPUData{IdlePercent: 75},
				Memory:      RawMemoryData{TotalBytes: 8000000000},
				Load:        RawLoadData{Load1Min: 0.5},
				Pressure:    RawPressureMetrics{Available: false},
			},
			wantPressure: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Build all metrics from raw data.
			result := buildAllMetricsFromRaw(tt.raw)
			// Verify CPU metrics.
			assert.Equal(t, fullPercentage-tt.raw.CPU.IdlePercent, result.CPU.UsagePercent)
			// Verify memory metrics.
			assert.Equal(t, tt.raw.Memory.TotalBytes, result.Memory.Total)
			// Verify pressure presence.
			if tt.wantPressure {
				assert.NotNil(t, result.Pressure)
			} else {
				assert.Nil(t, result.Pressure)
			}
		})
	}
}

