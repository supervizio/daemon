// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawPartitionData holds raw partition data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawPartitionData struct {
	// Device is the device path.
	Device string
	// MountPoint is the mount location.
	MountPoint string
	// FSType is the filesystem type.
	FSType string
	// Options are the mount options.
	Options string
}
