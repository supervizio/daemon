// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawDiskIOData holds raw disk I/O data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawDiskIOData struct {
	// Device is the device name.
	Device string
	// ReadsCompleted is completed read operations.
	ReadsCompleted uint64
	// ReadBytes is bytes read.
	ReadBytes uint64
	// ReadTimeUs is time spent reading in us.
	ReadTimeUs uint64
	// WritesCompleted is completed write operations.
	WritesCompleted uint64
	// WriteBytes is bytes written.
	WriteBytes uint64
	// WriteTimeUs is time spent writing in us.
	WriteTimeUs uint64
	// IOInProgress is current I/O operations.
	IOInProgress uint64
	// IOTimeUs is total I/O time in us.
	IOTimeUs uint64
	// WeightedIOTimeUs is weighted I/O time in us.
	WeightedIOTimeUs uint64
}
