// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewDiskUsage tests the NewDiskUsage constructor.
func TestNewDiskUsage(t *testing.T) {
	tests := []struct {
		name                   string
		input                  metrics.DiskUsageInput
		wantPath               string
		wantDevice             string
		wantFSType             string
		wantUsagePercent       float64
		wantInodesUsagePercent float64
	}{
		{
			name: "root_partition_50_percent",
			input: metrics.DiskUsageInput{
				Path:        "/",
				Device:      "/dev/sda1",
				FSType:      "ext4",
				Total:       100 * 1024 * 1024 * 1024,
				Used:        50 * 1024 * 1024 * 1024,
				Free:        50 * 1024 * 1024 * 1024,
				InodesTotal: 1000000,
				InodesUsed:  250000,
				InodesFree:  750000,
			},
			wantPath:               "/",
			wantDevice:             "/dev/sda1",
			wantFSType:             "ext4",
			wantUsagePercent:       50.0,
			wantInodesUsagePercent: 25.0,
		},
		{
			name: "zero_total_disk",
			input: metrics.DiskUsageInput{
				Path:        "/mnt/empty",
				Device:      "/dev/sdb1",
				FSType:      "xfs",
				Total:       0,
				Used:        0,
				Free:        0,
				InodesTotal: 0,
				InodesUsed:  0,
				InodesFree:  0,
			},
			wantPath:               "/mnt/empty",
			wantDevice:             "/dev/sdb1",
			wantFSType:             "xfs",
			wantUsagePercent:       0.0,
			wantInodesUsagePercent: 0.0,
		},
		{
			name: "full_disk",
			input: metrics.DiskUsageInput{
				Path:        "/data",
				Device:      "/dev/nvme0n1p1",
				FSType:      "btrfs",
				Total:       500 * 1024 * 1024 * 1024,
				Used:        500 * 1024 * 1024 * 1024,
				Free:        0,
				InodesTotal: 5000000,
				InodesUsed:  5000000,
				InodesFree:  0,
			},
			wantPath:               "/data",
			wantDevice:             "/dev/nvme0n1p1",
			wantFSType:             "btrfs",
			wantUsagePercent:       100.0,
			wantInodesUsagePercent: 100.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := metrics.NewDiskUsage(&tt.input)

			assert.Equal(t, tt.wantPath, result.Path)
			assert.Equal(t, tt.wantDevice, result.Device)
			assert.Equal(t, tt.wantFSType, result.FSType)
			assert.Equal(t, tt.wantUsagePercent, result.UsagePercent)
			assert.Equal(t, tt.wantInodesUsagePercent, result.InodesUsagePercent)
			assert.False(t, result.Timestamp.IsZero())
		})
	}
}

// TestDiskUsage_Available tests the Available method.
func TestDiskUsage_Available(t *testing.T) {
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
		{
			name: "no_free_space",
			usage: metrics.DiskUsage{
				Free: 0,
			},
			wantAvailable: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.wantAvailable, tt.usage.Available())
		})
	}
}
