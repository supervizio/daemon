// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawIOStatsData holds raw I/O stats for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawIOStatsData struct {
	// ReadOps represents read operation count.
	ReadOps uint64
	// ReadBytes represents read bytes.
	ReadBytes uint64
	// WriteOps represents write operation count.
	WriteOps uint64
	// WriteBytes represents write bytes.
	WriteBytes uint64
}
