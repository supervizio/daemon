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
	// SectorsRead is sectors read.
	SectorsRead uint64
	// ReadTimeMs is time spent reading in ms.
	ReadTimeMs uint64
	// WritesCompleted is completed write operations.
	WritesCompleted uint64
	// SectorsWritten is sectors written.
	SectorsWritten uint64
	// WriteTimeMs is time spent writing in ms.
	WriteTimeMs uint64
	// IOInProgress is current I/O operations.
	IOInProgress uint64
	// IOTimeMs is total I/O time in ms.
	IOTimeMs uint64
	// WeightedIOTimeMs is weighted I/O time in ms.
	WeightedIOTimeMs uint64
}
