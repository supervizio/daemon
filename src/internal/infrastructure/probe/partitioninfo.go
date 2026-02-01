//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// PartitionInfo contains information about a mounted partition.
// Used for JSON output in the --probe command.
type PartitionInfo struct {
	// Device is the device name (e.g., "/dev/sda1").
	Device string `dto:"out,api,pub" json:"device"`
	// MountPoint is the mount path (e.g., "/home").
	MountPoint string `dto:"out,api,pub" json:"mount_point"`
	// FSType is the filesystem type (e.g., "ext4", "xfs").
	FSType string `dto:"out,api,pub" json:"fs_type"`
	// Options is the mount options as a comma-separated string.
	Options string `dto:"out,api,pub" json:"options,omitempty"`
}
