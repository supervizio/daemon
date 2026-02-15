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
//   - ts: timestamp to use for the metrics.
//
// Returns:
//   - metrics.SystemCPU: the constructed CPU metrics.
func buildCPUMetricsFromRaw(raw *RawCPUData, ts time.Time) metrics.SystemCPU {
	// Compute usage percent from idle percent.
	return metrics.SystemCPU{
		UsagePercent: fullPercentage - raw.IdlePercent,
		Timestamp:    ts,
	}
}

// buildMemoryMetricsFromRaw constructs memory metrics from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw memory data.
//   - ts: timestamp to use for the metrics.
//
// Returns:
//   - metrics.SystemMemory: the constructed memory metrics.
func buildMemoryMetricsFromRaw(raw *RawMemoryData, ts time.Time) metrics.SystemMemory {
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
		Timestamp: ts,
	}
}

// buildLoadMetricsFromRaw constructs load metrics from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw load data.
//   - ts: timestamp to use for the metrics.
//
// Returns:
//   - metrics.LoadAverage: the constructed load metrics.
func buildLoadMetricsFromRaw(raw *RawLoadData, ts time.Time) metrics.LoadAverage {
	// Build load metrics from raw data.
	return metrics.LoadAverage{
		Load1:     raw.Load1Min,
		Load5:     raw.Load5Min,
		Load15:    raw.Load15Min,
		Timestamp: ts,
	}
}

// buildIOStatsMetricsFromRaw constructs I/O stats from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw I/O stats data.
//   - ts: timestamp to use for the metrics.
//
// Returns:
//   - IOStatsSummary: the constructed I/O stats summary.
func buildIOStatsMetricsFromRaw(raw *RawIOStatsData, ts time.Time) IOStatsSummary {
	// Build I/O stats from raw data.
	return IOStatsSummary{
		ReadOps:    raw.ReadOps,
		ReadBytes:  raw.ReadBytes,
		WriteOps:   raw.WriteOps,
		WriteBytes: raw.WriteBytes,
		Timestamp:  ts,
	}
}

// buildPressureMetricsFromRaw constructs pressure metrics from raw data.
// This is a Go-only function that can be tested without CGO.
//
// Params:
//   - raw: pointer to the raw pressure data.
//   - ts: timestamp to use for the metrics.
//
// Returns:
//   - *AllPressure: the constructed pressure metrics.
func buildPressureMetricsFromRaw(raw *RawPressureMetrics, ts time.Time) *AllPressure {
	// Build pressure metrics from raw data.
	return &AllPressure{
		CPU: metrics.CPUPressure{
			SomeAvg10:  raw.CPU.SomeAvg10,
			SomeAvg60:  raw.CPU.SomeAvg60,
			SomeAvg300: raw.CPU.SomeAvg300,
			SomeTotal:  raw.CPU.SomeTotalUs,
			Timestamp:  ts,
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
				Timestamp:  ts,
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
				Timestamp:  ts,
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
// Uses a single timestamp for all metrics to reduce time.Now() calls.
//
// Params:
//   - raw: pointer to the raw AllMetrics data.
//
// Returns:
//   - *AllMetrics: the constructed Go struct.
func buildAllMetricsFromRaw(raw *RawAllMetrics) *AllMetrics {
	// Single timestamp for all metrics (timestamp batching optimization)
	ts := time.Now()

	// Build all metrics from raw data with shared timestamp.
	all := &AllMetrics{
		CPU:           buildCPUMetricsFromRaw(&raw.CPU, ts),
		Memory:        buildMemoryMetricsFromRaw(&raw.Memory, ts),
		Load:          buildLoadMetricsFromRaw(&raw.Load, ts),
		IOStats:       buildIOStatsMetricsFromRaw(&raw.IOStats, ts),
		Timestamp:     time.UnixMicro(raw.TimestampUs),
		Partitions:    buildPartitionsFromRaw(raw.Partitions),
		DiskUsage:     buildDiskUsageFromRaw(raw.DiskUsage),
		DiskIO:        buildDiskIOFromRaw(raw.DiskIO),
		NetInterfaces: buildNetInterfacesFromRaw(raw.NetInterfaces),
		NetStats:      buildNetStatsFromRaw(raw.NetStats),
	}

	// Copy pressure if available.
	if raw.Pressure.Available {
		all.Pressure = buildPressureMetricsFromRaw(&raw.Pressure, ts)
	}

	// Return the constructed metrics.
	return all
}
