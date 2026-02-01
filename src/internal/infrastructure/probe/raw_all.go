// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// RawAllMetrics holds all raw metrics data for Go-only building.
// This struct mirrors C data without requiring CGO for testing.
type RawAllMetrics struct {
	// TimestampNs is the timestamp in nanoseconds.
	TimestampNs int64
	// CPU holds raw CPU data.
	CPU RawCPUData
	// Memory holds raw memory data.
	Memory RawMemoryData
	// Load holds raw load data.
	Load RawLoadData
	// IOStats holds raw I/O stats.
	IOStats RawIOStatsData
	// Pressure holds raw pressure data.
	Pressure RawPressureMetrics
	// Partitions holds raw partition data.
	Partitions []RawPartitionData
	// DiskUsage holds raw disk usage data.
	DiskUsage []RawDiskUsageData
	// DiskIO holds raw disk I/O data.
	DiskIO []RawDiskIOData
	// NetInterfaces holds raw network interface data.
	NetInterfaces []RawNetInterfaceData
	// NetStats holds raw network stats data.
	NetStats []RawNetStatsData
}
