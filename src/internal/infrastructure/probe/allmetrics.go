//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// AllMetrics contains all system metrics collected in one call.
// This struct aggregates CPU, memory, load, I/O, disk, and network metrics
// for efficient batch collection.
type AllMetrics struct {
	// CPU contains system CPU metrics.
	CPU metrics.SystemCPU `dto:"out,api,pub" json:"cpu"`
	// Memory contains system memory metrics.
	Memory metrics.SystemMemory `dto:"out,api,pub" json:"memory"`
	// Load contains system load averages.
	Load metrics.LoadAverage `dto:"out,api,pub" json:"load"`
	// IOStats contains aggregated I/O statistics.
	IOStats IOStatsSummary `dto:"out,api,pub" json:"ioStats"`
	// Pressure contains PSI pressure metrics (Linux only).
	Pressure *AllPressure `dto:"out,api,pub" json:"pressure,omitempty"`
	// Partitions contains mounted partition information.
	Partitions []PartitionInfo `dto:"out,api,pub" json:"partitions"`
	// DiskUsage contains disk usage for all partitions.
	DiskUsage []DiskUsageInfo `dto:"out,api,pub" json:"diskUsage"`
	// DiskIO contains disk I/O statistics.
	DiskIO []DiskIOInfo `dto:"out,api,pub" json:"diskIO"`
	// NetInterfaces contains network interface information.
	NetInterfaces []NetInterfaceInfo `dto:"out,api,pub" json:"netInterfaces"`
	// NetStats contains network statistics.
	NetStats []NetStatsInfo `dto:"out,api,pub" json:"netStats"`
	// Timestamp is when the metrics were collected.
	Timestamp time.Time `dto:"out,api,pub" json:"timestamp"`
}
