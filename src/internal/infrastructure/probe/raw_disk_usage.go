// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawDiskUsageData holds raw disk usage data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawDiskUsageData struct {
	// Path is the mount path.
	Path string
	// TotalBytes is total disk space.
	TotalBytes uint64
	// UsedBytes is used disk space.
	UsedBytes uint64
	// FreeBytes is free disk space.
	FreeBytes uint64
	// UsedPercent is the usage percentage.
	UsedPercent float64
	// InodesTotal is total inodes.
	InodesTotal uint64
	// InodesUsed is used inodes.
	InodesUsed uint64
	// InodesFree is free inodes.
	InodesFree uint64
}
