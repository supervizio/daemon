//go:build cgo

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

// DiskCollector provides disk metrics via the Rust probe library.
type DiskCollector struct{}

// NewDiskCollector creates a new disk collector.
func NewDiskCollector() *DiskCollector {
	return &DiskCollector{}
}

// ListPartitions returns all mounted partitions.
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func (d *DiskCollector) ListPartitions(_ context.Context) ([]metrics.Partition, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var list C.PartitionList
	result := C.probe_list_partitions(&list)
	if err := resultToError(result); err != nil {
		return nil, err
	}
	defer C.probe_free_partition_list(&list)

	partitions := make([]metrics.Partition, list.count)
	items := unsafe.Slice(list.items, list.count)
	for i, item := range items {
		optStr := cCharArrayToString(item.options[:])
		var opts []string
		if optStr != "" {
			opts = strings.Split(optStr, ",")
		}
		partitions[i] = metrics.Partition{
			Device:     cCharArrayToString(item.device[:]),
			Mountpoint: cCharArrayToString(item.mount_point[:]),
			FSType:     cCharArrayToString(item.fs_type[:]),
			Options:    opts,
		}
	}

	return partitions, nil
}

// CollectUsage collects disk usage for a specific path.
func (d *DiskCollector) CollectUsage(_ context.Context, path string) (metrics.DiskUsage, error) {
	if err := checkInitialized(); err != nil {
		return metrics.DiskUsage{}, err
	}

	cPath := C.CString(path)
	defer C.free(unsafe.Pointer(cPath))

	var usage C.DiskUsage
	result := C.probe_collect_disk_usage(cPath, &usage)
	if err := resultToError(result); err != nil {
		return metrics.DiskUsage{}, err
	}

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
func (d *DiskCollector) CollectAllUsage(ctx context.Context) ([]metrics.DiskUsage, error) {
	partitions, err := d.ListPartitions(ctx)
	if err != nil {
		return nil, err
	}

	usages := make([]metrics.DiskUsage, 0, len(partitions))
	for _, p := range partitions {
		usage, err := d.CollectUsage(ctx, p.Mountpoint)
		if err == nil {
			usages = append(usages, usage)
		}
	}

	return usages, nil
}

// CollectIO collects I/O statistics for all block devices.
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func (d *DiskCollector) CollectIO(_ context.Context) ([]metrics.DiskIOStats, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var list C.DiskIOStatsList
	result := C.probe_collect_disk_io(&list)
	if err := resultToError(result); err != nil {
		return nil, err
	}
	defer C.probe_free_disk_io_list(&list)

	stats := make([]metrics.DiskIOStats, list.count)
	items := unsafe.Slice(list.items, list.count)
	for i, item := range items {
		stats[i] = metrics.DiskIOStats{
			Device:         cCharArrayToString(item.device[:]),
			ReadBytes:      uint64(item.sectors_read) * 512,
			WriteBytes:     uint64(item.sectors_written) * 512,
			ReadCount:      uint64(item.reads_completed),
			WriteCount:     uint64(item.writes_completed),
			ReadTime:       time.Duration(item.read_time_ms) * time.Millisecond,
			WriteTime:      time.Duration(item.write_time_ms) * time.Millisecond,
			IOInProgress:   uint64(item.io_in_progress),
			IOTime:         time.Duration(item.io_time_ms) * time.Millisecond,
			WeightedIOTime: time.Duration(item.weighted_io_time_ms) * time.Millisecond,
			Timestamp:      time.Now(),
		}
	}

	return stats, nil
}

// CollectDeviceIO collects I/O statistics for a specific device.
func (d *DiskCollector) CollectDeviceIO(ctx context.Context, device string) (metrics.DiskIOStats, error) {
	all, err := d.CollectIO(ctx)
	if err != nil {
		return metrics.DiskIOStats{}, err
	}

	for _, stat := range all {
		if stat.Device == device {
			return stat, nil
		}
	}

	return metrics.DiskIOStats{}, ErrNotFound
}

// cCharArrayToString converts a C char array to a Go string.
func cCharArrayToString(arr []C.char) string {
	for i, c := range arr {
		if c == 0 {
			return string(unsafe.Slice((*byte)(unsafe.Pointer(&arr[0])), i))
		}
	}
	return string(unsafe.Slice((*byte)(unsafe.Pointer(&arr[0])), len(arr)))
}
