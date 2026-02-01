//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

// fullPercentage represents 100% as a constant for percentage calculations.
const fullPercentage float64 = 100.0

// CollectAll collects all system metrics in one call.
// This function extracts C data and delegates building to Go-only functions.
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

	// Build raw data by extracting from C.
	raw := &RawAllMetrics{
		TimestampNs: int64(cAll.timestamp_ns),
		CPU:         RawCPUData{IdlePercent: float64(cAll.cpu.idle_percent)},
		Memory: RawMemoryData{
			TotalBytes:     uint64(cAll.memory.total_bytes),
			AvailableBytes: uint64(cAll.memory.available_bytes),
			UsedBytes:      uint64(cAll.memory.used_bytes),
			CachedBytes:    uint64(cAll.memory.cached_bytes),
			BuffersBytes:   uint64(cAll.memory.buffers_bytes),
			SwapTotalBytes: uint64(cAll.memory.swap_total_bytes),
			SwapUsedBytes:  uint64(cAll.memory.swap_used_bytes),
		},
		Load: RawLoadData{
			Load1Min:  float64(cAll.load.load_1min),
			Load5Min:  float64(cAll.load.load_5min),
			Load15Min: float64(cAll.load.load_15min),
		},
		IOStats: RawIOStatsData{
			ReadOps:    uint64(cAll.io_stats.read_ops),
			ReadBytes:  uint64(cAll.io_stats.read_bytes),
			WriteOps:   uint64(cAll.io_stats.write_ops),
			WriteBytes: uint64(cAll.io_stats.write_bytes),
		},
		Pressure: extractPressureFromC(&cAll),
	}

	// Extract arrays.
	raw.Partitions = extractPartitionsFromC(&cAll)
	raw.DiskUsage = extractDiskUsageFromC(&cAll)
	raw.DiskIO = extractDiskIOFromC(&cAll)
	raw.NetInterfaces = extractNetIfacesFromC(&cAll)
	raw.NetStats = extractNetStatsFromC(&cAll)

	// Delegate to Go-only builder and return.
	return buildAllMetricsFromRaw(raw), nil
}

// extractPressureFromC extracts pressure data from C struct.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - RawPressureMetrics: the extracted pressure data.
func extractPressureFromC(cAll *C.AllMetrics) RawPressureMetrics {
	// Build pressure metrics from C data.
	return RawPressureMetrics{
		Available: bool(cAll.pressure.available),
		CPU: RawCPUPressure{
			SomeAvg10:   float64(cAll.pressure.cpu.some_avg10),
			SomeAvg60:   float64(cAll.pressure.cpu.some_avg60),
			SomeAvg300:  float64(cAll.pressure.cpu.some_avg300),
			SomeTotalUs: uint64(cAll.pressure.cpu.some_total_us),
		},
		Memory: RawMemoryPressure{
			SomeAvg10:   float64(cAll.pressure.memory.some_avg10),
			SomeAvg60:   float64(cAll.pressure.memory.some_avg60),
			SomeAvg300:  float64(cAll.pressure.memory.some_avg300),
			SomeTotalUs: uint64(cAll.pressure.memory.some_total_us),
			FullAvg10:   float64(cAll.pressure.memory.full_avg10),
			FullAvg60:   float64(cAll.pressure.memory.full_avg60),
			FullAvg300:  float64(cAll.pressure.memory.full_avg300),
			FullTotalUs: uint64(cAll.pressure.memory.full_total_us),
		},
		IO: RawIOPressure{
			SomeAvg10:   float64(cAll.pressure.io.some_avg10),
			SomeAvg60:   float64(cAll.pressure.io.some_avg60),
			SomeAvg300:  float64(cAll.pressure.io.some_avg300),
			SomeTotalUs: uint64(cAll.pressure.io.some_total_us),
			FullAvg10:   float64(cAll.pressure.io.full_avg10),
			FullAvg60:   float64(cAll.pressure.io.full_avg60),
			FullAvg300:  float64(cAll.pressure.io.full_avg300),
			FullTotalUs: uint64(cAll.pressure.io.full_total_us),
		},
	}
}

// extractPartitionsFromC extracts partition data from C struct.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []RawPartitionData: the extracted partition data.
func extractPartitionsFromC(cAll *C.AllMetrics) []RawPartitionData {
	count := int(cAll.partition_count)
	result := make([]RawPartitionData, 0, count)
	// Iterate over each C partition.
	for idx := range count {
		pt := cAll.partitions[idx]
		result = append(result, RawPartitionData{
			Device:     C.GoString(&pt.device[0]),
			MountPoint: C.GoString(&pt.mount_point[0]),
			FSType:     C.GoString(&pt.fs_type[0]),
			Options:    C.GoString(&pt.options[0]),
		})
	}
	// Return extracted partitions.
	return result
}

// extractDiskUsageFromC extracts disk usage data from C struct.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []RawDiskUsageData: the extracted disk usage data.
func extractDiskUsageFromC(cAll *C.AllMetrics) []RawDiskUsageData {
	count := int(cAll.disk_usage_count)
	result := make([]RawDiskUsageData, 0, count)
	// Iterate over each C disk usage.
	for idx := range count {
		du := cAll.disk_usage[idx]
		result = append(result, RawDiskUsageData{
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
	// Return extracted disk usage.
	return result
}

// extractDiskIOFromC extracts disk I/O data from C struct.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []RawDiskIOData: the extracted disk I/O data.
func extractDiskIOFromC(cAll *C.AllMetrics) []RawDiskIOData {
	count := int(cAll.disk_io_count)
	result := make([]RawDiskIOData, 0, count)
	// Iterate over each C disk I/O.
	for idx := range count {
		dio := cAll.disk_io[idx]
		result = append(result, RawDiskIOData{
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
	// Return extracted disk I/O.
	return result
}

// extractNetIfacesFromC extracts network interface data from C struct.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []RawNetInterfaceData: the extracted interface data.
func extractNetIfacesFromC(cAll *C.AllMetrics) []RawNetInterfaceData {
	count := int(cAll.net_interface_count)
	result := make([]RawNetInterfaceData, 0, count)
	// Iterate over each C network interface.
	for idx := range count {
		iface := cAll.net_interfaces[idx]
		result = append(result, RawNetInterfaceData{
			Name:       C.GoString(&iface.name[0]),
			MACAddress: C.GoString(&iface.mac_address[0]),
			MTU:        uint32(iface.mtu),
			IsUp:       bool(iface.is_up),
			IsLoopback: bool(iface.is_loopback),
		})
	}
	// Return extracted interfaces.
	return result
}

// extractNetStatsFromC extracts network stats data from C struct.
//
// Params:
//   - cAll: pointer to the C AllMetrics struct.
//
// Returns:
//   - []RawNetStatsData: the extracted network stats.
func extractNetStatsFromC(cAll *C.AllMetrics) []RawNetStatsData {
	count := int(cAll.net_stats_count)
	result := make([]RawNetStatsData, 0, count)
	// Iterate over each C network stat.
	for idx := range count {
		ns := cAll.net_stats[idx]
		result = append(result, RawNetStatsData{
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
	// Return extracted network stats.
	return result
}
