// Package metrics provides domain types for system and process metrics collection.
package probe

import "time"

// Partition represents a disk partition or mount point.
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

// DiskUsage represents disk space usage for a mount point.
type DiskUsage struct {
	// Path is the mount point path.
	Path string
	// Device is the device name backing this mount.
	Device string
	// FSType is the filesystem type.
	FSType string
	// Total is the total disk space in bytes.
	Total uint64
	// Used is the used disk space in bytes.
	Used uint64
	// Free is the free disk space in bytes.
	Free uint64
	// UsagePercent is the usage percentage (0-100).
	UsagePercent float64
	// InodesTotal is the total number of inodes.
	InodesTotal uint64
	// InodesUsed is the number of used inodes.
	InodesUsed uint64
	// InodesFree is the number of free inodes.
	InodesFree uint64
	// InodesUsagePercent is the inode usage percentage (0-100).
	InodesUsagePercent float64
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// Available returns the available space in bytes.
// This is typically Free but may differ on some filesystems with reserved space.
func (d DiskUsage) Available() uint64 {
	return d.Free
}

// DiskIOStats represents I/O statistics for a block device.
type DiskIOStats struct {
	// Device is the device name (e.g., "sda", "nvme0n1").
	Device string
	// ReadBytes is the total number of bytes read.
	ReadBytes uint64
	// WriteBytes is the total number of bytes written.
	WriteBytes uint64
	// ReadCount is the number of read operations completed.
	ReadCount uint64
	// WriteCount is the number of write operations completed.
	WriteCount uint64
	// ReadTime is the total time spent reading.
	ReadTime time.Duration
	// WriteTime is the total time spent writing.
	WriteTime time.Duration
	// IOInProgress is the number of I/O operations currently in progress.
	IOInProgress uint64
	// IOTime is the total time spent doing I/O operations.
	IOTime time.Duration
	// WeightedIOTime is the weighted time spent doing I/O (time * queue depth).
	WeightedIOTime time.Duration
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// TotalOperations returns the total number of I/O operations.
func (d DiskIOStats) TotalOperations() uint64 {
	return d.ReadCount + d.WriteCount
}

// TotalBytes returns the total number of bytes transferred.
func (d DiskIOStats) TotalBytes() uint64 {
	return d.ReadBytes + d.WriteBytes
}

// TotalTime returns the total time spent on I/O operations.
func (d DiskIOStats) TotalTime() time.Duration {
	return d.ReadTime + d.WriteTime
}
