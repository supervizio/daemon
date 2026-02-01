//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

// MetricType represents a type of metric for cache configuration.
// It identifies specific metric categories for TTL settings.
type MetricType uint8

// Metric types for cache TTL configuration.
const (
	// MetricCPUSystem is the system CPU metrics type.
	MetricCPUSystem MetricType = 0
	// MetricCPUPressure is the CPU pressure metrics type.
	MetricCPUPressure MetricType = 1
	// MetricMemorySystem is the system memory metrics type.
	MetricMemorySystem MetricType = 2
	// MetricMemoryPressure is the memory pressure metrics type.
	MetricMemoryPressure MetricType = 3
	// MetricLoad is the load average metrics type.
	MetricLoad MetricType = 4
	// MetricDiskPartitions is the disk partitions metrics type.
	MetricDiskPartitions MetricType = 5
	// MetricDiskUsage is the disk usage metrics type.
	MetricDiskUsage MetricType = 6
	// MetricDiskIO is the disk I/O metrics type.
	MetricDiskIO MetricType = 7
	// MetricNetInterfaces is the network interfaces metrics type.
	MetricNetInterfaces MetricType = 8
	// MetricNetStats is the network statistics metrics type.
	MetricNetStats MetricType = 9
	// MetricIOStats is the I/O statistics metrics type.
	MetricIOStats MetricType = 10
	// MetricIOPressure is the I/O pressure metrics type.
	MetricIOPressure MetricType = 11
)
