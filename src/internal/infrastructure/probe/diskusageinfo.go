//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// DiskUsageInfo contains disk usage statistics.
// Used for JSON output in the --probe command.
type DiskUsageInfo struct {
	// Path is the mount point path.
	Path string `dto:"out,api,pub" json:"path"`
	// TotalBytes is the total size in bytes.
	TotalBytes uint64 `dto:"out,api,pub" json:"total_bytes"`
	// UsedBytes is the used space in bytes.
	UsedBytes uint64 `dto:"out,api,pub" json:"used_bytes"`
	// FreeBytes is the free space in bytes.
	FreeBytes uint64 `dto:"out,api,pub" json:"free_bytes"`
	// UsedPercent is the percentage of space used.
	UsedPercent float64 `dto:"out,api,pub" json:"used_percent"`
	// InodesTotal is the total number of inodes.
	InodesTotal uint64 `dto:"out,api,pub" json:"inodes_total"`
	// InodesUsed is the number of used inodes.
	InodesUsed uint64 `dto:"out,api,pub" json:"inodes_used"`
	// InodesFree is the number of free inodes.
	InodesFree uint64 `dto:"out,api,pub" json:"inodes_free"`
}
