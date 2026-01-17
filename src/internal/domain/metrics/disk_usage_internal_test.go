// Package metrics contains internal tests for disk usage metrics.
package metrics

import (
	"testing"
)

// TestNewDiskUsage_calculations tests the internal percentage calculations.
func TestNewDiskUsage_calculations(t *testing.T) {
	tests := []struct {
		name                   string
		input                  *DiskUsageInput
		wantUsagePercent       float64
		wantInodesUsagePercent float64
	}{
		{
			name: "typical_50_percent_usage",
			input: &DiskUsageInput{
				Path:        "/",
				Device:      "/dev/sda1",
				FSType:      "ext4",
				Total:       1000,
				Used:        500,
				Free:        500,
				InodesTotal: 1000,
				InodesUsed:  500,
				InodesFree:  500,
			},
			wantUsagePercent:       50.0,
			wantInodesUsagePercent: 50.0,
		},
		{
			name: "zero_total_avoids_division_by_zero",
			input: &DiskUsageInput{
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
			wantUsagePercent:       0.0,
			wantInodesUsagePercent: 0.0,
		},
		{
			name: "full_disk_100_percent",
			input: &DiskUsageInput{
				Path:        "/data",
				Device:      "/dev/nvme0n1p1",
				FSType:      "btrfs",
				Total:       1000,
				Used:        1000,
				Free:        0,
				InodesTotal: 1000,
				InodesUsed:  1000,
				InodesFree:  0,
			},
			wantUsagePercent:       100.0,
			wantInodesUsagePercent: 100.0,
		},
		{
			name: "inodes_zero_disk_non_zero",
			input: &DiskUsageInput{
				Path:        "/special",
				Device:      "/dev/sdc1",
				FSType:      "tmpfs",
				Total:       1000,
				Used:        250,
				Free:        750,
				InodesTotal: 0,
				InodesUsed:  0,
				InodesFree:  0,
			},
			wantUsagePercent:       25.0,
			wantInodesUsagePercent: 0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := NewDiskUsage(tt.input)

			// Validate usage percentage calculation.
			if result.UsagePercent != tt.wantUsagePercent {
				t.Errorf("UsagePercent = %v, want %v", result.UsagePercent, tt.wantUsagePercent)
			}
			// Validate inode usage percentage calculation.
			if result.InodesUsagePercent != tt.wantInodesUsagePercent {
				t.Errorf("InodesUsagePercent = %v, want %v", result.InodesUsagePercent, tt.wantInodesUsagePercent)
			}
			// Validate timestamp is set.
			if result.Timestamp.IsZero() {
				t.Error("Timestamp should not be zero")
			}
		})
	}
}
