// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// DiskIOStatsInput contains the input parameters for creating DiskIOStats.
//
// This struct groups the parameters needed to construct a DiskIOStats value object.
type DiskIOStatsInput struct {
	// Device is the device name (e.g., "sda", "nvme0n1").
	Device string
	// ReadBytes is the total number of bytes read.
	ReadBytes uint64
	// WriteBytes is the total number of bytes written.
	WriteBytes uint64
	// ReadCount is the number of read operations completed.
	ReadCount uint64
	// WriteCount is the number of write operations completed.
	WriteCount uint64
	// ReadTime is the total time spent reading.
	ReadTime time.Duration
	// WriteTime is the total time spent writing.
	WriteTime time.Duration
	// IOInProgress is the number of I/O operations currently in progress.
	IOInProgress uint64
	// IOTime is the total time spent doing I/O operations.
	IOTime time.Duration
	// WeightedIOTime is the weighted time spent doing I/O (time * queue depth).
	WeightedIOTime time.Duration
}
