// Package metrics provides domain types for system and process metrics collection.
package metrics

import "time"

// DiskIOStats represents I/O statistics for a block device.
//
// This value object captures cumulative I/O counters and timing information
// from /proc/diskstats or /sys/block/*/stat.
type DiskIOStats struct {
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
	// Timestamp is when this sample was taken.
	Timestamp time.Time
}

// NewDiskIOStats creates a new DiskIOStats instance.
//
// Params:
//   - input: pointer to DiskIOStatsInput containing all disk I/O parameters.
//
// Returns:
//   - *DiskIOStats: initialized disk I/O statistics struct.
func NewDiskIOStats(input *DiskIOStatsInput) *DiskIOStats {
	// initialize with all disk I/O fields
	return &DiskIOStats{
		Device:         input.Device,
		ReadBytes:      input.ReadBytes,
		WriteBytes:     input.WriteBytes,
		ReadCount:      input.ReadCount,
		WriteCount:     input.WriteCount,
		ReadTime:       input.ReadTime,
		WriteTime:      input.WriteTime,
		IOInProgress:   input.IOInProgress,
		IOTime:         input.IOTime,
		WeightedIOTime: input.WeightedIOTime,
		Timestamp:      time.Now(),
	}
}

// TotalOperations returns the total number of I/O operations.
//
// Returns:
//   - uint64: sum of read and write operation counts.
func (d *DiskIOStats) TotalOperations() uint64 {
	// sum read and write operations
	return d.ReadCount + d.WriteCount
}

// TotalBytes returns the total number of bytes transferred.
//
// Returns:
//   - uint64: sum of bytes read and written.
func (d *DiskIOStats) TotalBytes() uint64 {
	// sum read and write bytes
	return d.ReadBytes + d.WriteBytes
}

// TotalTime returns the total time spent on I/O operations.
//
// Returns:
//   - time.Duration: sum of read and write time.
func (d *DiskIOStats) TotalTime() time.Duration {
	// sum read and write durations
	return d.ReadTime + d.WriteTime
}
