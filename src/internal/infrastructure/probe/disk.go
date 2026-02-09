//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
#include <stdlib.h>
*/
import "C"

import (
	"context"
	"strings"
	"time"
	"unsafe"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// sectorSize is the standard disk sector size in bytes.
const sectorSize uint64 = 512

// DiskCollector provides disk metrics via the Rust probe library.
// It implements the metrics.DiskCollector interface for disk statistics.
type DiskCollector struct{}

// NewDiskCollector creates a new disk collector.
//
// Returns:
//   - *DiskCollector: new disk collector instance
func NewDiskCollector() *DiskCollector {
	// Return a new empty collector instance.
	return &DiskCollector{}
}

// ListPartitions returns all mounted partitions.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - []metrics.Partition: list of mounted partitions
//   - error: nil on success, error if probe not initialized or collection fails
func (d *DiskCollector) ListPartitions(ctx context.Context) ([]metrics.Partition, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return nil on validation failure.
		return nil, err
	}
	// List partitions from C library.
	var list C.PartitionList
	result := C.probe_list_partitions(&list)
	// Check if listing failed.
	if err := resultToError(result); err != nil {
		// Return nil on listing failure.
		return nil, err
	}
	defer C.probe_free_partition_list(&list)
	// Convert C list to Go slice.
	count := int(list.count)
	partitions := make([]metrics.Partition, 0, count)
	items := unsafe.Slice(list.items, count)
	// Iterate over each partition item.
	for _, item := range items {
		optStr := cCharArrayToString(item.options[:])
		var opts []string
		// Parse mount options if present.
		if optStr != "" {
			opts = strings.Split(optStr, ",")
		}
		partitions = append(partitions, metrics.Partition{
			Device:     cCharArrayToStringCached(item.device[:], true),     // stable: device names don't change
			Mountpoint: cCharArrayToStringCached(item.mount_point[:], true), // stable: mount points don't change
			FSType:     cCharArrayToStringCached(item.fs_type[:], true),     // stable: filesystem types don't change
			Options:    opts,
		})
	}
	// Return the collected partitions.
	return partitions, nil
}

// CollectUsage collects disk usage for a specific path.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//   - path: filesystem path to collect usage for
//
// Returns:
//   - metrics.DiskUsage: disk usage statistics for the path
//   - error: nil on success, error if probe not initialized or collection fails
func (d *DiskCollector) CollectUsage(ctx context.Context, path string) (metrics.DiskUsage, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return empty metrics on validation failure.
		return metrics.DiskUsage{}, err
	}
	// Collect disk usage from C library.
	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))
	var usage C.DiskUsage
	result := C.probe_collect_disk_usage(cPath, &usage)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return empty metrics on collection failure.
		return metrics.DiskUsage{}, err
	}
	// Return collected disk usage with current timestamp.
	return metrics.DiskUsage{
		Path:         path,
		Total:        uint64(usage.total_bytes),
		Used:         uint64(usage.used_bytes),
		Free:         uint64(usage.free_bytes),
		UsagePercent: float64(usage.used_percent),
		InodesTotal:  uint64(usage.inodes_total),
		InodesUsed:   uint64(usage.inodes_used),
		InodesFree:   uint64(usage.inodes_free),
		Timestamp:    time.Now(),
	}, nil
}

// CollectAllUsage collects disk usage for all mounted partitions.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []metrics.DiskUsage: disk usage for all partitions
//   - error: nil on success, error if listing partitions fails
func (d *DiskCollector) CollectAllUsage(ctx context.Context) ([]metrics.DiskUsage, error) {
	partitions, err := d.ListPartitions(ctx)
	// Check if listing partitions failed.
	if err != nil {
		// Return nil slice with error.
		return nil, err
	}

	usages := make([]metrics.DiskUsage, 0, len(partitions))
	// Collect usage for each partition.
	for _, p := range partitions {
		usage, err := d.CollectUsage(ctx, p.Mountpoint)
		// Only include successful collections.
		if err == nil {
			usages = append(usages, usage)
		}
	}

	// Return collected disk usages.
	return usages, nil
}

// CollectIO collects I/O statistics for all block devices.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - []metrics.DiskIOStats: I/O statistics for all block devices
//   - error: nil on success, error if probe not initialized or collection fails
func (d *DiskCollector) CollectIO(ctx context.Context) ([]metrics.DiskIOStats, error) {
	// Validate context and initialization state.
	if err := validateCollectionContext(ctx); err != nil {
		// Return nil on validation failure.
		return nil, err
	}
	// Collect disk I/O from C library.
	var list C.DiskIOStatsList
	result := C.probe_collect_disk_io(&list)
	// Check if collection failed.
	if err := resultToError(result); err != nil {
		// Return nil on collection failure.
		return nil, err
	}
	defer C.probe_free_disk_io_list(&list)
	// Convert C list to Go slice.
	count := int(list.count)
	stats := make([]metrics.DiskIOStats, 0, count)
	items := unsafe.Slice(list.items, count)
	// Iterate over each block device.
	for _, item := range items {
		stats = append(stats, metrics.DiskIOStats{
			Device:         cCharArrayToStringCached(item.device[:], true), // stable: device names don't change
			ReadBytes:      uint64(item.sectors_read) * sectorSize,
			WriteBytes:     uint64(item.sectors_written) * sectorSize,
			ReadCount:      uint64(item.reads_completed),
			WriteCount:     uint64(item.writes_completed),
			ReadTime:       time.Duration(item.read_time_ms) * time.Millisecond,
			WriteTime:      time.Duration(item.write_time_ms) * time.Millisecond,
			IOInProgress:   uint64(item.io_in_progress),
			IOTime:         time.Duration(item.io_time_ms) * time.Millisecond,
			WeightedIOTime: time.Duration(item.weighted_io_time_ms) * time.Millisecond,
			Timestamp:      time.Now(),
		})
	}
	// Return collected I/O statistics.
	return stats, nil
}

// CollectDeviceIO collects I/O statistics for a specific device.
//
// Params:
//   - ctx: context for cancellation
//   - device: device name to collect I/O stats for
//
// Returns:
//   - metrics.DiskIOStats: I/O statistics for the device
//   - error: nil on success, ErrNotFound if device not found
func (d *DiskCollector) CollectDeviceIO(ctx context.Context, device string) (metrics.DiskIOStats, error) {
	all, err := d.CollectIO(ctx)
	// Check if collecting all I/O stats failed.
	if err != nil {
		// Return empty stats with error.
		return metrics.DiskIOStats{}, err
	}

	// Search for the requested device in the collected stats.
	for _, stat := range all {
		// Check if this is the requested device.
		if stat.Device == device {
			// Return the matching device stats.
			return stat, nil
		}
	}

	// Return error if device was not found.
	return metrics.DiskIOStats{}, ErrNotFound
}

// cCharArrayToString converts a C char array to a Go string.
// This is a thin CGO wrapper that delegates to the Go-only bytesToStringWithNull.
//
// Params:
//   - arr: C char array to convert
//
// Returns:
//   - string: the converted Go string
func cCharArrayToString(arr []C.char) string {
	// Convert C char array to byte slice.
	bytes := unsafe.Slice((*byte)(unsafe.Pointer(&arr[0])), len(arr))
	// Delegate to Go-only function.
	return bytesToStringWithNull(bytes)
}
