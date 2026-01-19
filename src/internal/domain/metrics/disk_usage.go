// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// DiskUsage represents disk space usage for a mount point.
//
// This value object captures both block and inode usage statistics for a filesystem.
// Use UsagePercent and InodesUsagePercent to monitor capacity.
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

// NewDiskUsage creates a new DiskUsage instance with calculated fields.
//
// Params:
//   - input: pointer to DiskUsageInput containing all disk usage parameters.
//
// Returns:
//   - *DiskUsage: initialized disk usage metrics with calculated percentages.
func NewDiskUsage(input *DiskUsageInput) *DiskUsage {
	// Calculate usage percentage, handling zero total case.
	var usagePercent float64
	// Check if total is non-zero to avoid division by zero.
	if input.Total > 0 {
		usagePercent = float64(input.Used) / float64(input.Total) * percentMultiplier
	}
	// Calculate inode usage percentage, handling zero total case.
	var inodesUsagePercent float64
	// Check if inodesTotal is non-zero to avoid division by zero.
	if input.InodesTotal > 0 {
		inodesUsagePercent = float64(input.InodesUsed) / float64(input.InodesTotal) * percentMultiplier
	}
	// Return initialized disk usage metrics struct.
	return &DiskUsage{
		Path:               input.Path,
		Device:             input.Device,
		FSType:             input.FSType,
		Total:              input.Total,
		Used:               input.Used,
		Free:               input.Free,
		UsagePercent:       usagePercent,
		InodesTotal:        input.InodesTotal,
		InodesUsed:         input.InodesUsed,
		InodesFree:         input.InodesFree,
		InodesUsagePercent: inodesUsagePercent,
		Timestamp:          time.Now(),
	}
}

// Available returns the available space in bytes.
//
// Returns:
//   - uint64: available space, which may differ from Free on filesystems with reserved space.
func (d *DiskUsage) Available() uint64 {
	// Return free space (on most filesystems this equals available space).
	return d.Free
}
