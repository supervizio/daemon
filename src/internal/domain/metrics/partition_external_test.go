// Package metrics_test provides black-box tests for the metrics package.
package metrics_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// TestNewPartition tests the NewPartition constructor.
func TestNewPartition(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		device     string
		mountpoint string
		fsType     string
		options    []string
	}{
		{
			name:       "ext4_root_partition",
			device:     "/dev/sda1",
			mountpoint: "/",
			fsType:     "ext4",
			options:    []string{"rw", "noatime"},
		},
		{
			name:       "nvme_home_partition",
			device:     "/dev/nvme0n1p2",
			mountpoint: "/home",
			fsType:     "xfs",
			options:    []string{"rw", "relatime"},
		},
		{
			name:       "zfs_pool_partition",
			device:     "tank/data",
			mountpoint: "/data",
			fsType:     "zfs",
			options:    []string{"rw", "atime", "compression"},
		},
		{
			name:       "empty_options",
			device:     "/dev/sdb1",
			mountpoint: "/mnt/backup",
			fsType:     "btrfs",
			options:    []string{},
		},
		{
			name:       "nil_options",
			device:     "/dev/sdc1",
			mountpoint: "/mnt/external",
			fsType:     "ntfs",
			options:    nil,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Create Partition using constructor.
			partition := metrics.NewPartition(tt.device, tt.mountpoint, tt.fsType, tt.options)

			// Verify pointer is not nil.
			require.NotNil(t, partition)

			// Verify all fields are correctly set.
			assert.Equal(t, tt.device, partition.Device)
			assert.Equal(t, tt.mountpoint, partition.Mountpoint)
			assert.Equal(t, tt.fsType, partition.FSType)
			assert.Equal(t, tt.options, partition.Options)
		})
	}
}

// TestPartition tests Partition value object creation.
func TestPartition(t *testing.T) {
	t.Parallel()

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

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			// Verify partition fields.
			assert.Equal(t, tt.wantDevice, tt.partition.Device)
			assert.Equal(t, tt.wantMount, tt.partition.Mountpoint)
			assert.Equal(t, tt.wantFS, tt.partition.FSType)
		})
	}
}
