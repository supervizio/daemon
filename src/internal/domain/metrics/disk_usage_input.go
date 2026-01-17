// Package metrics provides domain types for system and process metrics collection.
package metrics

// DiskUsageInput contains the input parameters for creating DiskUsage.
//
// This struct groups the parameters needed to construct a DiskUsage value object.
type DiskUsageInput struct {
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
	// InodesTotal is the total number of inodes.
	InodesTotal uint64
	// InodesUsed is the number of used inodes.
	InodesUsed uint64
	// InodesFree is the number of free inodes.
	InodesFree uint64
}
