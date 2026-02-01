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
//
// Returns:
//   - *AllMetrics: collected metrics snapshot
//   - error: nil on success, error if collection fails
func CollectAll() (*AllMetrics, error) {
	// Check if probe is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Probe not initialized.
		return nil, err
	}

	var cAll C.AllMetrics
	result := C.probe_collect_all(&cAll)
	if err := resultToError(result); err != nil {
		// Collection failed.
		return nil, err
	}

	// Build and return the AllMetrics struct from C data.
	return buildAllMetrics(&cAll), nil
}

// buildAllMetrics constructs an AllMetrics struct from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct
//
// Returns:
//   - *AllMetrics: the constructed Go struct
func buildAllMetrics(cAll *C.AllMetrics) *AllMetrics {
	all := &AllMetrics{
		CPU:           buildCPUMetrics(cAll),
		Memory:        buildMemoryMetrics(cAll),
		Load:          buildLoadMetrics(cAll),
		IOStats:       buildIOStatsMetrics(cAll),
		Timestamp:     time.Unix(0, int64(cAll.timestamp_ns)),
		Partitions:    buildPartitions(cAll),
		DiskUsage:     buildDiskUsage(cAll),
		DiskIO:        buildDiskIO(cAll),
		NetInterfaces: buildNetInterfaces(cAll),
		NetStats:      buildNetStats(cAll),
	}

	// Copy pressure if available (Linux only).
	if bool(cAll.pressure.available) {
		all.Pressure = buildPressureMetrics(cAll)
	}

	// Return the constructed metrics.
	return all
}

// buildCPUMetrics constructs CPU metrics from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - metrics.SystemCPU: the constructed CPU metrics.
func buildCPUMetrics(cAll *C.AllMetrics) metrics.SystemCPU {
	return metrics.SystemCPU{
		UsagePercent: fullPercentage - float64(cAll.cpu.idle_percent),
		Timestamp:    time.Now(),
	}
}

// buildMemoryMetrics constructs memory metrics from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - metrics.SystemMemory: the constructed memory metrics.
func buildMemoryMetrics(cAll *C.AllMetrics) metrics.SystemMemory {
	return metrics.SystemMemory{
		Total:     uint64(cAll.memory.total_bytes),
		Available: uint64(cAll.memory.available_bytes),
		Used:      uint64(cAll.memory.used_bytes),
		Cached:    uint64(cAll.memory.cached_bytes),
		Buffers:   uint64(cAll.memory.buffers_bytes),
		SwapTotal: uint64(cAll.memory.swap_total_bytes),
		SwapUsed:  uint64(cAll.memory.swap_used_bytes),
		SwapFree:  uint64(cAll.memory.swap_total_bytes) - uint64(cAll.memory.swap_used_bytes),
		Timestamp: time.Now(),
	}
}

// buildLoadMetrics constructs load average metrics from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - metrics.LoadAverage: the constructed load average metrics.
func buildLoadMetrics(cAll *C.AllMetrics) metrics.LoadAverage {
	return metrics.LoadAverage{
		Load1:     float64(cAll.load.load_1min),
		Load5:     float64(cAll.load.load_5min),
		Load15:    float64(cAll.load.load_15min),
		Timestamp: time.Now(),
	}
}

// buildIOStatsMetrics constructs I/O stats from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - IOStatsSummary: the constructed I/O stats summary.
func buildIOStatsMetrics(cAll *C.AllMetrics) IOStatsSummary {
	return IOStatsSummary{
		ReadOps:    uint64(cAll.io_stats.read_ops),
		ReadBytes:  uint64(cAll.io_stats.read_bytes),
		WriteOps:   uint64(cAll.io_stats.write_ops),
		WriteBytes: uint64(cAll.io_stats.write_bytes),
		Timestamp:  time.Now(),
	}
}

// buildPressureMetrics constructs pressure metrics from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - *AllPressure: the constructed pressure metrics.
func buildPressureMetrics(cAll *C.AllMetrics) *AllPressure {
	return &AllPressure{
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

// buildPartitions constructs partition info from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []PartitionInfo: the constructed partition info list.
func buildPartitions(cAll *C.AllMetrics) []PartitionInfo {
	count := int(cAll.partition_count)
	partitions := make([]PartitionInfo, 0, count)
	// Convert each C partition to Go struct.
	for idx := range count {
		pt := cAll.partitions[idx]
		partitions = append(partitions, PartitionInfo{
			Device:     C.GoString(&pt.device[0]),
			MountPoint: C.GoString(&pt.mount_point[0]),
			FSType:     C.GoString(&pt.fs_type[0]),
			Options:    C.GoString(&pt.options[0]),
		})
	}
	// Return collected partitions.
	return partitions
}

// buildDiskUsage constructs disk usage info from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []DiskUsageInfo: the constructed disk usage info list.
func buildDiskUsage(cAll *C.AllMetrics) []DiskUsageInfo {
	count := int(cAll.disk_usage_count)
	usage := make([]DiskUsageInfo, 0, count)
	// Convert each C disk usage to Go struct.
	for idx := range count {
		du := cAll.disk_usage[idx]
		usage = append(usage, DiskUsageInfo{
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
	// Return collected disk usage.
	return usage
}

// buildDiskIO constructs disk I/O info from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []DiskIOInfo: the constructed disk I/O info list.
func buildDiskIO(cAll *C.AllMetrics) []DiskIOInfo {
	count := int(cAll.disk_io_count)
	diskIO := make([]DiskIOInfo, 0, count)
	// Convert each C disk I/O to Go struct.
	for idx := range count {
		dio := cAll.disk_io[idx]
		diskIO = append(diskIO, DiskIOInfo{
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
	// Return collected disk I/O.
	return diskIO
}

// buildNetInterfaces constructs network interface info from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []NetInterfaceInfo: the constructed network interface info list.
func buildNetInterfaces(cAll *C.AllMetrics) []NetInterfaceInfo {
	count := int(cAll.net_interface_count)
	interfaces := make([]NetInterfaceInfo, 0, count)
	// Convert each C network interface to Go struct.
	for idx := range count {
		iface := cAll.net_interfaces[idx]
		interfaces = append(interfaces, NetInterfaceInfo{
			Name:       C.GoString(&iface.name[0]),
			MACAddress: C.GoString(&iface.mac_address[0]),
			MTU:        uint32(iface.mtu),
			IsUp:       bool(iface.is_up),
			IsLoopback: bool(iface.is_loopback),
		})
	}
	// Return collected interfaces.
	return interfaces
}

// buildNetStats constructs network stats from C data.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []NetStatsInfo: the constructed network stats info list.
func buildNetStats(cAll *C.AllMetrics) []NetStatsInfo {
	count := int(cAll.net_stats_count)
	stats := make([]NetStatsInfo, 0, count)
	// Convert each C network stat to Go struct.
	for idx := range count {
		ns := cAll.net_stats[idx]
		stats = append(stats, NetStatsInfo{
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
	// Return collected network stats.
	return stats
}
