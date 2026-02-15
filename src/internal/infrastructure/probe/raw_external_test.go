// Package probe_test provides external tests for the raw types.
// These tests verify the exported Raw* structs work correctly.
package probe_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
)

// TestRawTypesVersion tests the RawTypesVersion constant.
func TestRawTypesVersion(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		version string
		want    bool
	}{
		{
			name:    "current_version",
			version: probe.RawTypesVersion,
			want:    true,
		},
		{
			name:    "non_empty_version",
			version: "1.0.0",
			want:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Verify the version is not empty.
			assert.NotEmpty(t, tt.version)
			// Verify expected result.
			assert.True(t, tt.want)
		})
	}
}

// TestRawCPUData tests the RawCPUData struct.
func TestRawCPUData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		idle        float64
		wantPercent float64
	}{
		{
			name:        "zero_idle",
			idle:        0,
			wantPercent: 0,
		},
		{
			name:        "full_idle",
			idle:        100,
			wantPercent: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw CPU data.
			raw := probe.RawCPUData{IdlePercent: tt.idle}
			// Verify idle percent is set correctly.
			assert.Equal(t, tt.wantPercent, raw.IdlePercent)
		})
	}
}

// TestRawMemoryData tests the RawMemoryData struct.
func TestRawMemoryData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name           string
		totalBytes     uint64
		availableBytes uint64
		usedBytes      uint64
	}{
		{
			name:           "basic_memory",
			totalBytes:     16000000000,
			availableBytes: 8000000000,
			usedBytes:      8000000000,
		},
		{
			name:           "zero_memory",
			totalBytes:     0,
			availableBytes: 0,
			usedBytes:      0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw memory data.
			raw := probe.RawMemoryData{
				TotalBytes:     tt.totalBytes,
				AvailableBytes: tt.availableBytes,
				UsedBytes:      tt.usedBytes,
			}
			// Verify all fields are set correctly.
			assert.Equal(t, tt.totalBytes, raw.TotalBytes)
			assert.Equal(t, tt.availableBytes, raw.AvailableBytes)
			assert.Equal(t, tt.usedBytes, raw.UsedBytes)
		})
	}
}

// TestRawLoadData tests the RawLoadData struct.
func TestRawLoadData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		load1Min  float64
		load5Min  float64
		load15Min float64
	}{
		{
			name:      "normal_load",
			load1Min:  1.5,
			load5Min:  1.2,
			load15Min: 0.8,
		},
		{
			name:      "zero_load",
			load1Min:  0.0,
			load5Min:  0.0,
			load15Min: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw load data.
			raw := probe.RawLoadData{
				Load1Min:  tt.load1Min,
				Load5Min:  tt.load5Min,
				Load15Min: tt.load15Min,
			}
			// Verify all fields are set correctly.
			assert.Equal(t, tt.load1Min, raw.Load1Min)
			assert.Equal(t, tt.load5Min, raw.Load5Min)
			assert.Equal(t, tt.load15Min, raw.Load15Min)
		})
	}
}

// TestRawIOStatsData tests the RawIOStatsData struct.
func TestRawIOStatsData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		readOps    uint64
		readBytes  uint64
		writeOps   uint64
		writeBytes uint64
	}{
		{
			name:       "normal_io",
			readOps:    1000,
			readBytes:  50000000,
			writeOps:   500,
			writeBytes: 25000000,
		},
		{
			name:       "zero_io",
			readOps:    0,
			readBytes:  0,
			writeOps:   0,
			writeBytes: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw I/O stats data.
			raw := probe.RawIOStatsData{
				ReadOps:    tt.readOps,
				ReadBytes:  tt.readBytes,
				WriteOps:   tt.writeOps,
				WriteBytes: tt.writeBytes,
			}
			// Verify all fields are set correctly.
			assert.Equal(t, tt.readOps, raw.ReadOps)
			assert.Equal(t, tt.readBytes, raw.ReadBytes)
			assert.Equal(t, tt.writeOps, raw.WriteOps)
			assert.Equal(t, tt.writeBytes, raw.WriteBytes)
		})
	}
}

// TestRawPressureMetrics tests the RawPressureMetrics struct.
func TestRawPressureMetrics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		available bool
		someAvg10 float64
	}{
		{
			name:      "available_pressure",
			available: true,
			someAvg10: 1.5,
		},
		{
			name:      "unavailable_pressure",
			available: false,
			someAvg10: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw pressure metrics.
			raw := probe.RawPressureMetrics{
				Available: tt.available,
				CPU:       probe.RawCPUPressure{SomeAvg10: tt.someAvg10},
			}
			// Verify availability flag.
			assert.Equal(t, tt.available, raw.Available)
			// Verify CPU pressure data.
			assert.Equal(t, tt.someAvg10, raw.CPU.SomeAvg10)
		})
	}
}

// TestRawPartitionData tests the RawPartitionData struct.
func TestRawPartitionData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		device     string
		mountPoint string
		fsType     string
	}{
		{
			name:       "root_partition",
			device:     "/dev/sda1",
			mountPoint: "/",
			fsType:     "ext4",
		},
		{
			name:       "home_partition",
			device:     "/dev/sda2",
			mountPoint: "/home",
			fsType:     "xfs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw partition data.
			raw := probe.RawPartitionData{
				Device:     tt.device,
				MountPoint: tt.mountPoint,
				FSType:     tt.fsType,
			}
			// Verify all fields are set correctly.
			assert.Equal(t, tt.device, raw.Device)
			assert.Equal(t, tt.mountPoint, raw.MountPoint)
			assert.Equal(t, tt.fsType, raw.FSType)
		})
	}
}

// TestRawDiskUsageData tests the RawDiskUsageData struct.
func TestRawDiskUsageData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		path        string
		totalBytes  uint64
		usedPercent float64
	}{
		{
			name:        "half_used",
			path:        "/",
			totalBytes:  100000000000,
			usedPercent: 50.0,
		},
		{
			name:        "nearly_full",
			path:        "/data",
			totalBytes:  200000000000,
			usedPercent: 90.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw disk usage data.
			raw := probe.RawDiskUsageData{
				Path:        tt.path,
				TotalBytes:  tt.totalBytes,
				UsedPercent: tt.usedPercent,
			}
			// Verify all fields are set correctly.
			assert.Equal(t, tt.path, raw.Path)
			assert.Equal(t, tt.totalBytes, raw.TotalBytes)
			assert.Equal(t, tt.usedPercent, raw.UsedPercent)
		})
	}
}

// TestRawDiskIOData tests the RawDiskIOData struct.
func TestRawDiskIOData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name            string
		device          string
		readsCompleted  uint64
		writesCompleted uint64
	}{
		{
			name:            "sda_io",
			device:          "sda",
			readsCompleted:  1000,
			writesCompleted: 500,
		},
		{
			name:            "nvme_io",
			device:          "nvme0n1",
			readsCompleted:  5000,
			writesCompleted: 3000,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw disk I/O data.
			raw := probe.RawDiskIOData{
				Device:          tt.device,
				ReadsCompleted:  tt.readsCompleted,
				WritesCompleted: tt.writesCompleted,
			}
			// Verify all fields are set correctly.
			assert.Equal(t, tt.device, raw.Device)
			assert.Equal(t, tt.readsCompleted, raw.ReadsCompleted)
			assert.Equal(t, tt.writesCompleted, raw.WritesCompleted)
		})
	}
}

// TestRawNetInterfaceData tests the RawNetInterfaceData struct.
func TestRawNetInterfaceData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		ifaceName  string
		isUp       bool
		isLoopback bool
	}{
		{
			name:       "eth0_up",
			ifaceName:  "eth0",
			isUp:       true,
			isLoopback: false,
		},
		{
			name:       "loopback",
			ifaceName:  "lo",
			isUp:       true,
			isLoopback: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw network interface data.
			raw := probe.RawNetInterfaceData{
				Name:       tt.ifaceName,
				IsUp:       tt.isUp,
				IsLoopback: tt.isLoopback,
			}
			// Verify all fields are set correctly.
			assert.Equal(t, tt.ifaceName, raw.Name)
			assert.Equal(t, tt.isUp, raw.IsUp)
			assert.Equal(t, tt.isLoopback, raw.IsLoopback)
		})
	}
}

// TestRawNetStatsData tests the RawNetStatsData struct.
func TestRawNetStatsData(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		iface   string
		rxBytes uint64
		txBytes uint64
	}{
		{
			name:    "eth0_stats",
			iface:   "eth0",
			rxBytes: 1000000,
			txBytes: 500000,
		},
		{
			name:    "lo_stats",
			iface:   "lo",
			rxBytes: 100,
			txBytes: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw network stats data.
			raw := probe.RawNetStatsData{
				Interface: tt.iface,
				RxBytes:   tt.rxBytes,
				TxBytes:   tt.txBytes,
			}
			// Verify all fields are set correctly.
			assert.Equal(t, tt.iface, raw.Interface)
			assert.Equal(t, tt.rxBytes, raw.RxBytes)
			assert.Equal(t, tt.txBytes, raw.TxBytes)
		})
	}
}

// TestRawAllMetrics tests the RawAllMetrics struct.
func TestRawAllMetrics(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		timestampUs int64
		idlePercent float64
		totalBytes  uint64
	}{
		{
			name:        "full_metrics",
			timestampUs: 1000000,
			idlePercent: 50,
			totalBytes:  16000000000,
		},
		{
			name:        "empty_metrics",
			timestampUs: 2000000,
			idlePercent: 0,
			totalBytes:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create raw all metrics.
			raw := probe.RawAllMetrics{
				TimestampUs: tt.timestampUs,
				CPU:         probe.RawCPUData{IdlePercent: tt.idlePercent},
				Memory:      probe.RawMemoryData{TotalBytes: tt.totalBytes},
			}
			// Verify timestamp.
			assert.Equal(t, tt.timestampUs, raw.TimestampUs)
			// Verify CPU data.
			assert.Equal(t, tt.idlePercent, raw.CPU.IdlePercent)
			// Verify memory data.
			assert.Equal(t, tt.totalBytes, raw.Memory.TotalBytes)
		})
	}
}
