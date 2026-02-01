//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// DiskIOInfo contains disk I/O statistics.
// Used for JSON output in the --probe command.
type DiskIOInfo struct {
	// Device is the device name (e.g., "sda", "nvme0n1").
	Device string `dto:"out,api,pub" json:"device"`
	// ReadsCompleted is the number of completed read operations.
	ReadsCompleted uint64 `dto:"out,api,pub" json:"reads_completed"`
	// SectorsRead is the number of sectors read.
	SectorsRead uint64 `dto:"out,api,pub" json:"sectors_read"`
	// ReadTimeMs is the time spent reading in milliseconds.
	ReadTimeMs uint64 `dto:"out,api,pub" json:"read_time_ms"`
	// WritesCompleted is the number of completed write operations.
	WritesCompleted uint64 `dto:"out,api,pub" json:"writes_completed"`
	// SectorsWritten is the number of sectors written.
	SectorsWritten uint64 `dto:"out,api,pub" json:"sectors_written"`
	// WriteTimeMs is the time spent writing in milliseconds.
	WriteTimeMs uint64 `dto:"out,api,pub" json:"write_time_ms"`
	// IOInProgress is the number of I/O operations in progress.
	IOInProgress uint64 `dto:"out,api,pub" json:"io_in_progress"`
	// IOTimeMs is the total time spent on I/O in milliseconds.
	IOTimeMs uint64 `dto:"out,api,pub" json:"io_time_ms"`
	// WeightedIOTimeMs is the weighted time spent on I/O.
	WeightedIOTimeMs uint64 `dto:"out,api,pub" json:"weighted_io_time_ms"`
}
