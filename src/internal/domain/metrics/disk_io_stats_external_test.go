// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewDiskIOStats tests the NewDiskIOStats constructor.
func TestNewDiskIOStats(t *testing.T) {
	tests := []struct {
		name       string
		input      metrics.DiskIOStatsInput
		wantDevice string
	}{
		{
			name: "sda_active_device",
			input: metrics.DiskIOStatsInput{
				Device:         "sda",
				ReadBytes:      1024 * 1024 * 100,
				WriteBytes:     1024 * 1024 * 50,
				ReadCount:      1000,
				WriteCount:     500,
				ReadTime:       5 * time.Second,
				WriteTime:      3 * time.Second,
				IOInProgress:   2,
				IOTime:         8 * time.Second,
				WeightedIOTime: 10 * time.Second,
			},
			wantDevice: "sda",
		},
		{
			name: "nvme_idle_device",
			input: metrics.DiskIOStatsInput{
				Device:         "nvme0n1",
				ReadBytes:      0,
				WriteBytes:     0,
				ReadCount:      0,
				WriteCount:     0,
				ReadTime:       0,
				WriteTime:      0,
				IOInProgress:   0,
				IOTime:         0,
				WeightedIOTime: 0,
			},
			wantDevice: "nvme0n1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metrics.NewDiskIOStats(&tt.input)

			assert.Equal(t, tt.wantDevice, result.Device)
			assert.Equal(t, tt.input.ReadBytes, result.ReadBytes)
			assert.Equal(t, tt.input.WriteBytes, result.WriteBytes)
			assert.Equal(t, tt.input.ReadCount, result.ReadCount)
			assert.Equal(t, tt.input.WriteCount, result.WriteCount)
			assert.Equal(t, tt.input.ReadTime, result.ReadTime)
			assert.Equal(t, tt.input.WriteTime, result.WriteTime)
			assert.Equal(t, tt.input.IOInProgress, result.IOInProgress)
			assert.Equal(t, tt.input.IOTime, result.IOTime)
			assert.Equal(t, tt.input.WeightedIOTime, result.WeightedIOTime)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestDiskIOStats_TotalOperations tests the TotalOperations method.
func TestDiskIOStats_TotalOperations(t *testing.T) {
	tests := []struct {
		name         string
		stats        metrics.DiskIOStats
		wantTotalOps uint64
	}{
		{
			name: "mixed_operations",
			stats: metrics.DiskIOStats{
				ReadCount:  1000,
				WriteCount: 500,
			},
			wantTotalOps: 1500,
		},
		{
			name: "read_only",
			stats: metrics.DiskIOStats{
				ReadCount:  2000,
				WriteCount: 0,
			},
			wantTotalOps: 2000,
		},
		{
			name: "no_operations",
			stats: metrics.DiskIOStats{
				ReadCount:  0,
				WriteCount: 0,
			},
			wantTotalOps: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTotalOps, tt.stats.TotalOperations())
		})
	}
}

// TestDiskIOStats_TotalBytes tests the TotalBytes method.
func TestDiskIOStats_TotalBytes(t *testing.T) {
	tests := []struct {
		name           string
		stats          metrics.DiskIOStats
		wantTotalBytes uint64
	}{
		{
			name: "mixed_bytes",
			stats: metrics.DiskIOStats{
				ReadBytes:  1024 * 1024 * 100,
				WriteBytes: 1024 * 1024 * 50,
			},
			wantTotalBytes: 1024 * 1024 * 150,
		},
		{
			name: "read_only_bytes",
			stats: metrics.DiskIOStats{
				ReadBytes:  1024 * 1024 * 200,
				WriteBytes: 0,
			},
			wantTotalBytes: 1024 * 1024 * 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTotalBytes, tt.stats.TotalBytes())
		})
	}
}

// TestDiskIOStats_TotalTime tests the TotalTime method.
func TestDiskIOStats_TotalTime(t *testing.T) {
	tests := []struct {
		name          string
		stats         metrics.DiskIOStats
		wantTotalTime time.Duration
	}{
		{
			name: "sda_io_stats",
			stats: metrics.DiskIOStats{
				Device:         "sda",
				ReadBytes:      1024 * 1024 * 100, // 100MB read
				WriteBytes:     1024 * 1024 * 50,  // 50MB written
				ReadCount:      1000,
				WriteCount:     500,
				ReadTime:       5 * time.Second,
				WriteTime:      3 * time.Second,
				IOInProgress:   2,
				IOTime:         8 * time.Second,
				WeightedIOTime: 10 * time.Second,
				Timestamp:      time.Now(),
			},
			wantTotalTime: 8 * time.Second,
		},
		{
			name: "no_time",
			stats: metrics.DiskIOStats{
				ReadTime:  0,
				WriteTime: 0,
			},
			wantTotalTime: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTotalTime, tt.stats.TotalTime())
		})
	}
}
