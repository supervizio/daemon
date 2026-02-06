// Package metrics provides domain types for system and process metrics collection.
package metrics

// Partition represents a disk partition or mount point.
//
// This value object describes a mounted filesystem, including its device path,
// mount location, filesystem type, and mount options.
type Partition struct {
	// Device is the device name (e.g., "/dev/sda1", "/dev/nvme0n1p1").
	Device string
	// Mountpoint is the mount path (e.g., "/", "/home").
	Mountpoint string
	// FSType is the filesystem type (e.g., "ext4", "xfs", "zfs", "apfs").
	FSType string
	// Options are mount options (e.g., "rw", "noatime").
	Options []string
}

// NewPartition creates a new Partition instance.
//
// Params:
//   - device: device name (e.g., "/dev/sda1")
//   - mountpoint: mount path (e.g., "/")
//   - fsType: filesystem type (e.g., "ext4")
//   - options: mount options slice
//
// Returns:
//   - *Partition: initialized partition struct
func NewPartition(device, mountpoint, fsType string, options []string) *Partition {
	// initialize with all partition fields
	return &Partition{
		Device:     device,
		Mountpoint: mountpoint,
		FSType:     fsType,
		Options:    options,
	}
}
