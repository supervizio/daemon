// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestPartition tests Partition value object creation.
func TestPartition(t *testing.T) {
	tests := []struct {
		name       string
		partition  metrics.Partition
		wantDevice string
		wantMount  string
		wantFS     string
	}{
		{
			name: "ext4_root_partition",
			partition: metrics.Partition{
				Device:     "/dev/sda1",
				Mountpoint: "/",
				FSType:     "ext4",
				Options:    []string{"rw", "noatime"},
			},
			wantDevice: "/dev/sda1",
			wantMount:  "/",
			wantFS:     "ext4",
		},
		{
			name: "nvme_home_partition",
			partition: metrics.Partition{
				Device:     "/dev/nvme0n1p2",
				Mountpoint: "/home",
				FSType:     "xfs",
				Options:    []string{"rw", "relatime"},
			},
			wantDevice: "/dev/nvme0n1p2",
			wantMount:  "/home",
			wantFS:     "xfs",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantDevice, tt.partition.Device)
			assert.Equal(t, tt.wantMount, tt.partition.Mountpoint)
			assert.Equal(t, tt.wantFS, tt.partition.FSType)
		})
	}
}

// TestDiskUsage tests DiskUsage value object methods.
func TestDiskUsage(t *testing.T) {
	tests := []struct {
		name          string
		usage         metrics.DiskUsage
		wantAvailable uint64
	}{
		{
			name: "root_partition_usage",
			usage: metrics.DiskUsage{
				Path:         "/",
				Device:       "/dev/sda1",
				FSType:       "ext4",
				Total:        100 * 1024 * 1024 * 1024, // 100GB
				Used:         40 * 1024 * 1024 * 1024,  // 40GB
				Free:         60 * 1024 * 1024 * 1024,  // 60GB
				UsagePercent: 40.0,
				InodesTotal:  6553600,
				InodesUsed:   100000,
				InodesFree:   6453600,
				Timestamp:    time.Now(),
			},
			wantAvailable: 60 * 1024 * 1024 * 1024,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantAvailable, tt.usage.Available())
		})
	}
}

// TestDiskIOStats tests DiskIOStats value object methods.
func TestDiskIOStats(t *testing.T) {
	tests := []struct {
		name           string
		stats          metrics.DiskIOStats
		wantTotalOps   uint64
		wantTotalBytes uint64
		wantTotalTime  time.Duration
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
			wantTotalOps:   1500,
			wantTotalBytes: 1024 * 1024 * 150,
			wantTotalTime:  8 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantTotalOps, tt.stats.TotalOperations())
			assert.Equal(t, tt.wantTotalBytes, tt.stats.TotalBytes())
			assert.Equal(t, tt.wantTotalTime, tt.stats.TotalTime())
		})
	}
}
