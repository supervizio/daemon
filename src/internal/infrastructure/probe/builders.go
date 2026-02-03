//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// buildCPUMetricsFromRaw constructs CPU metrics from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw CPU data.
//
// Returns:
//   - metrics.SystemCPU: the constructed CPU metrics.
func buildCPUMetricsFromRaw(raw *RawCPUData) metrics.SystemCPU {
	// Compute usage percent from idle percent.
	return metrics.SystemCPU{
		UsagePercent: fullPercentage - raw.IdlePercent,
		Timestamp:    time.Now(),
	}
}

// buildMemoryMetricsFromRaw constructs memory metrics from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw memory data.
//
// Returns:
//   - metrics.SystemMemory: the constructed memory metrics.
func buildMemoryMetricsFromRaw(raw *RawMemoryData) metrics.SystemMemory {
	// Build memory metrics from raw data.
	return metrics.SystemMemory{
		Total:     raw.TotalBytes,
		Available: raw.AvailableBytes,
		Used:      raw.UsedBytes,
		Cached:    raw.CachedBytes,
		Buffers:   raw.BuffersBytes,
		SwapTotal: raw.SwapTotalBytes,
		SwapUsed:  raw.SwapUsedBytes,
		SwapFree:  raw.SwapTotalBytes - raw.SwapUsedBytes,
		Timestamp: time.Now(),
	}
}

// buildLoadMetricsFromRaw constructs load metrics from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw load data.
//
// Returns:
//   - metrics.LoadAverage: the constructed load metrics.
func buildLoadMetricsFromRaw(raw *RawLoadData) metrics.LoadAverage {
	// Build load metrics from raw data.
	return metrics.LoadAverage{
		Load1:     raw.Load1Min,
		Load5:     raw.Load5Min,
		Load15:    raw.Load15Min,
		Timestamp: time.Now(),
	}
}

// buildIOStatsMetricsFromRaw constructs I/O stats from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw I/O stats data.
//
// Returns:
//   - IOStatsSummary: the constructed I/O stats summary.
func buildIOStatsMetricsFromRaw(raw *RawIOStatsData) IOStatsSummary {
	// Build I/O stats from raw data.
	return IOStatsSummary{
		ReadOps:    raw.ReadOps,
		ReadBytes:  raw.ReadBytes,
		WriteOps:   raw.WriteOps,
		WriteBytes: raw.WriteBytes,
		Timestamp:  time.Now(),
	}
}

// buildPressureMetricsFromRaw constructs pressure metrics from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw pressure data.
//
// Returns:
//   - *AllPressure: the constructed pressure metrics.
func buildPressureMetricsFromRaw(raw *RawPressureMetrics) *AllPressure {
	// Build pressure metrics from raw data.
	return &AllPressure{
		CPU: metrics.CPUPressure{
			SomeAvg10:  raw.CPU.SomeAvg10,
			SomeAvg60:  raw.CPU.SomeAvg60,
			SomeAvg300: raw.CPU.SomeAvg300,
			SomeTotal:  raw.CPU.SomeTotalUs,
			Timestamp:  time.Now(),
		},
		Memory: metrics.MemoryPressure{
			Pressure: metrics.Pressure{
				SomeAvg10:  raw.Memory.SomeAvg10,
				SomeAvg60:  raw.Memory.SomeAvg60,
				SomeAvg300: raw.Memory.SomeAvg300,
				SomeTotal:  raw.Memory.SomeTotalUs,
				FullAvg10:  raw.Memory.FullAvg10,
				FullAvg60:  raw.Memory.FullAvg60,
				FullAvg300: raw.Memory.FullAvg300,
				FullTotal:  raw.Memory.FullTotalUs,
				Timestamp:  time.Now(),
			},
		},
		IO: metrics.IOPressure{
			Pressure: metrics.Pressure{
				SomeAvg10:  raw.IO.SomeAvg10,
				SomeAvg60:  raw.IO.SomeAvg60,
				SomeAvg300: raw.IO.SomeAvg300,
				SomeTotal:  raw.IO.SomeTotalUs,
				FullAvg10:  raw.IO.FullAvg10,
				FullAvg60:  raw.IO.FullAvg60,
				FullAvg300: raw.IO.FullAvg300,
				FullTotal:  raw.IO.FullTotalUs,
				Timestamp:  time.Now(),
			},
		},
	}
}

// buildPartitionsFromRaw constructs partition info from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: slice of raw partition data.
//
// Returns:
//   - []PartitionInfo: the constructed partition info list.
func buildPartitionsFromRaw(raw []RawPartitionData) []PartitionInfo {
	partitions := make([]PartitionInfo, 0, len(raw))
	// Convert each raw partition to Go struct.
	for _, pt := range raw {
		partitions = append(partitions, PartitionInfo(pt))
	}
	// Return collected partitions.
	return partitions
}

// buildDiskUsageFromRaw constructs disk usage info from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: slice of raw disk usage data.
//
// Returns:
//   - []DiskUsageInfo: the constructed disk usage info list.
func buildDiskUsageFromRaw(raw []RawDiskUsageData) []DiskUsageInfo {
	usage := make([]DiskUsageInfo, 0, len(raw))
	// Convert each raw disk usage to Go struct.
	for _, du := range raw {
		usage = append(usage, DiskUsageInfo(du))
	}
	// Return collected disk usage.
	return usage
}

// buildDiskIOFromRaw constructs disk I/O info from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: slice of raw disk I/O data.
//
// Returns:
//   - []DiskIOInfo: the constructed disk I/O info list.
func buildDiskIOFromRaw(raw []RawDiskIOData) []DiskIOInfo {
	diskIO := make([]DiskIOInfo, 0, len(raw))
	// Convert each raw disk I/O to Go struct.
	for _, dio := range raw {
		diskIO = append(diskIO, DiskIOInfo(dio))
	}
	// Return collected disk I/O.
	return diskIO
}

// buildNetInterfacesFromRaw constructs network interface info from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: slice of raw network interface data.
//
// Returns:
//   - []NetInterfaceInfo: the constructed network interface info list.
func buildNetInterfacesFromRaw(raw []RawNetInterfaceData) []NetInterfaceInfo {
	interfaces := make([]NetInterfaceInfo, 0, len(raw))
	// Convert each raw network interface to Go struct.
	for _, iface := range raw {
		interfaces = append(interfaces, NetInterfaceInfo(iface))
	}
	// Return collected interfaces.
	return interfaces
}

// buildNetStatsFromRaw constructs network stats from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: slice of raw network stats data.
//
// Returns:
//   - []NetStatsInfo: the constructed network stats info list.
func buildNetStatsFromRaw(raw []RawNetStatsData) []NetStatsInfo {
	stats := make([]NetStatsInfo, 0, len(raw))
	// Convert each raw network stat to Go struct.
	for _, ns := range raw {
		stats = append(stats, NetStatsInfo(ns))
	}
	// Return collected network stats.
	return stats
}

// buildAllMetricsFromRaw constructs an AllMetrics struct from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw AllMetrics data.
//
// Returns:
//   - *AllMetrics: the constructed Go struct.
func buildAllMetricsFromRaw(raw *RawAllMetrics) *AllMetrics {
	// Build all metrics from raw data.
	all := &AllMetrics{
		CPU:           buildCPUMetricsFromRaw(&raw.CPU),
		Memory:        buildMemoryMetricsFromRaw(&raw.Memory),
		Load:          buildLoadMetricsFromRaw(&raw.Load),
		IOStats:       buildIOStatsMetricsFromRaw(&raw.IOStats),
		Timestamp:     time.Unix(0, raw.TimestampNs),
		Partitions:    buildPartitionsFromRaw(raw.Partitions),
		DiskUsage:     buildDiskUsageFromRaw(raw.DiskUsage),
		DiskIO:        buildDiskIOFromRaw(raw.DiskIO),
		NetInterfaces: buildNetInterfacesFromRaw(raw.NetInterfaces),
		NetStats:      buildNetStatsFromRaw(raw.NetStats),
	}

	// Copy pressure if available.
	if raw.Pressure.Available {
		all.Pressure = buildPressureMetricsFromRaw(&raw.Pressure)
	}

	// Return the constructed metrics.
	return all
}
