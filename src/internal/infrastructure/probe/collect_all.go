//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"
import (
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// fullPercentage represents 100% as a constant for percentage calculations.
const fullPercentage float64 = 100.0

// CollectAll collects all system metrics in one call.
// This is more efficient than calling each collector individually
// and provides a consistent snapshot of all metrics at approximately
// the same point in time.
//
// Params:
//   - none
//
// Returns:
//   - *AllMetrics: collected metrics snapshot
//   - error: nil on success, error if probe not initialized or collection fails
//
//nolint:cyclop,funlen // Data marshaling function requires sequential field mapping
func CollectAll() (*AllMetrics, error) {
	// Check if probe is initialized before collecting
	if err := checkInitialized(); err != nil {
		// Return early if probe is not initialized
		return nil, err
	}

	var cAll C.AllMetrics
	result := C.probe_collect_all(&cAll)
	// Check if collection succeeded
	if err := resultToError(result); err != nil {
		// Return early on collection failure
		return nil, err
	}

	// Build the AllMetrics struct from C data
	all := &AllMetrics{
		CPU: metrics.SystemCPU{
			UsagePercent: fullPercentage - float64(cAll.cpu.idle_percent),
			Timestamp:    time.Now(),
		},
		Memory: metrics.SystemMemory{
			Total:     uint64(cAll.memory.total_bytes),
			Available: uint64(cAll.memory.available_bytes),
			Used:      uint64(cAll.memory.used_bytes),
			Cached:    uint64(cAll.memory.cached_bytes),
			Buffers:   uint64(cAll.memory.buffers_bytes),
			SwapTotal: uint64(cAll.memory.swap_total_bytes),
			SwapUsed:  uint64(cAll.memory.swap_used_bytes),
			SwapFree:  uint64(cAll.memory.swap_total_bytes) - uint64(cAll.memory.swap_used_bytes),
			Timestamp: time.Now(),
		},
		Load: metrics.LoadAverage{
			Load1:     float64(cAll.load.load_1min),
			Load5:     float64(cAll.load.load_5min),
			Load15:    float64(cAll.load.load_15min),
			Timestamp: time.Now(),
		},
		IOStats: IOStatsSummary{
			ReadOps:    uint64(cAll.io_stats.read_ops),
			ReadBytes:  uint64(cAll.io_stats.read_bytes),
			WriteOps:   uint64(cAll.io_stats.write_ops),
			WriteBytes: uint64(cAll.io_stats.write_bytes),
			Timestamp:  time.Now(),
		},
		Timestamp: time.Unix(0, int64(cAll.timestamp_ns)),
	}

	// Copy pressure if available (Linux only)
	if bool(cAll.pressure.available) {
		all.Pressure = &AllPressure{
			CPU: metrics.CPUPressure{
				SomeAvg10:  float64(cAll.pressure.cpu.some_avg10),
				SomeAvg60:  float64(cAll.pressure.cpu.some_avg60),
				SomeAvg300: float64(cAll.pressure.cpu.some_avg300),
				SomeTotal:  uint64(cAll.pressure.cpu.some_total_us),
				Timestamp:  time.Now(),
			},
			Memory: metrics.MemoryPressure{
				SomeAvg10:  float64(cAll.pressure.memory.some_avg10),
				SomeAvg60:  float64(cAll.pressure.memory.some_avg60),
				SomeAvg300: float64(cAll.pressure.memory.some_avg300),
				SomeTotal:  uint64(cAll.pressure.memory.some_total_us),
				FullAvg10:  float64(cAll.pressure.memory.full_avg10),
				FullAvg60:  float64(cAll.pressure.memory.full_avg60),
				FullAvg300: float64(cAll.pressure.memory.full_avg300),
				FullTotal:  uint64(cAll.pressure.memory.full_total_us),
				Timestamp:  time.Now(),
			},
			IO: metrics.IOPressure{
				SomeAvg10:  float64(cAll.pressure.io.some_avg10),
				SomeAvg60:  float64(cAll.pressure.io.some_avg60),
				SomeAvg300: float64(cAll.pressure.io.some_avg300),
				SomeTotal:  uint64(cAll.pressure.io.some_total_us),
				FullAvg10:  float64(cAll.pressure.io.full_avg10),
				FullAvg60:  float64(cAll.pressure.io.full_avg60),
				FullAvg300: float64(cAll.pressure.io.full_avg300),
				FullTotal:  uint64(cAll.pressure.io.full_total_us),
				Timestamp:  time.Now(),
			},
		}
	}

	// Copy partitions from C array
	partitionCount := int(cAll.partition_count)
	all.Partitions = make([]PartitionInfo, 0, partitionCount)
	// Iterate over each partition entry
	for idx := range partitionCount {
		pt := cAll.partitions[idx]
		all.Partitions = append(all.Partitions, PartitionInfo{
			Device:     C.GoString(&pt.device[0]),
			MountPoint: C.GoString(&pt.mount_point[0]),
			FSType:     C.GoString(&pt.fs_type[0]),
			Options:    C.GoString(&pt.options[0]),
		})
	}

	// Copy disk usage from C array
	usageCount := int(cAll.disk_usage_count)
	all.DiskUsage = make([]DiskUsageInfo, 0, usageCount)
	// Iterate over each disk usage entry
	for idx := range usageCount {
		du := cAll.disk_usage[idx]
		all.DiskUsage = append(all.DiskUsage, DiskUsageInfo{
			Path:        C.GoString(&du.path[0]),
			TotalBytes:  uint64(du.total_bytes),
			UsedBytes:   uint64(du.used_bytes),
			FreeBytes:   uint64(du.free_bytes),
			UsedPercent: float64(du.used_percent),
			InodesTotal: uint64(du.inodes_total),
			InodesUsed:  uint64(du.inodes_used),
			InodesFree:  uint64(du.inodes_free),
		})
	}

	// Copy disk I/O from C array
	ioCount := int(cAll.disk_io_count)
	all.DiskIO = make([]DiskIOInfo, 0, ioCount)
	// Iterate over each disk I/O entry
	for idx := range ioCount {
		dio := cAll.disk_io[idx]
		all.DiskIO = append(all.DiskIO, DiskIOInfo{
			Device:           C.GoString(&dio.device[0]),
			ReadsCompleted:   uint64(dio.reads_completed),
			SectorsRead:      uint64(dio.sectors_read),
			ReadTimeMs:       uint64(dio.read_time_ms),
			WritesCompleted:  uint64(dio.writes_completed),
			SectorsWritten:   uint64(dio.sectors_written),
			WriteTimeMs:      uint64(dio.write_time_ms),
			IOInProgress:     uint64(dio.io_in_progress),
			IOTimeMs:         uint64(dio.io_time_ms),
			WeightedIOTimeMs: uint64(dio.weighted_io_time_ms),
		})
	}

	// Copy network interfaces from C array
	ifaceCount := int(cAll.net_interface_count)
	all.NetInterfaces = make([]NetInterfaceInfo, 0, ifaceCount)
	// Iterate over each network interface entry
	for idx := range ifaceCount {
		iface := cAll.net_interfaces[idx]
		all.NetInterfaces = append(all.NetInterfaces, NetInterfaceInfo{
			Name:       C.GoString(&iface.name[0]),
			MACAddress: C.GoString(&iface.mac_address[0]),
			MTU:        uint32(iface.mtu),
			IsUp:       bool(iface.is_up),
			IsLoopback: bool(iface.is_loopback),
		})
	}

	// Copy network stats from C array
	statsCount := int(cAll.net_stats_count)
	all.NetStats = make([]NetStatsInfo, 0, statsCount)
	// Iterate over each network stats entry
	for idx := range statsCount {
		ns := cAll.net_stats[idx]
		all.NetStats = append(all.NetStats, NetStatsInfo{
			Interface: C.GoString(&ns._interface[0]),
			RxBytes:   uint64(ns.rx_bytes),
			RxPackets: uint64(ns.rx_packets),
			RxErrors:  uint64(ns.rx_errors),
			RxDrops:   uint64(ns.rx_drops),
			TxBytes:   uint64(ns.tx_bytes),
			TxPackets: uint64(ns.tx_packets),
			TxErrors:  uint64(ns.tx_errors),
			TxDrops:   uint64(ns.tx_drops),
		})
	}

	// Return the collected metrics
	return all, nil
}
