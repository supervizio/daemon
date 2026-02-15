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
	// ReadBytes is the number of bytes read.
	ReadBytes uint64 `dto:"out,api,pub" json:"read_bytes"`
	// ReadTimeUs is the time spent reading in microseconds.
	ReadTimeUs uint64 `dto:"out,api,pub" json:"read_time_us"`
	// WritesCompleted is the number of completed write operations.
	WritesCompleted uint64 `dto:"out,api,pub" json:"writes_completed"`
	// WriteBytes is the number of bytes written.
	WriteBytes uint64 `dto:"out,api,pub" json:"write_bytes"`
	// WriteTimeUs is the time spent writing in microseconds.
	WriteTimeUs uint64 `dto:"out,api,pub" json:"write_time_us"`
	// IOInProgress is the number of I/O operations in progress.
	IOInProgress uint64 `dto:"out,api,pub" json:"io_in_progress"`
	// IOTimeUs is the total time spent on I/O in microseconds.
	IOTimeUs uint64 `dto:"out,api,pub" json:"io_time_us"`
	// WeightedIOTimeUs is the weighted time spent on I/O in microseconds.
	WeightedIOTimeUs uint64 `dto:"out,api,pub" json:"weighted_io_time_us"`
}
