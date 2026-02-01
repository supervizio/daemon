//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
//
//nolint:ktn-struct-onefile // Metric output types are logically grouped for JSON serialization
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

// AllPressure contains all PSI pressure metrics.
// Only available on Linux with PSI support.
type AllPressure struct {
	// CPU contains CPU pressure metrics.
	CPU metrics.CPUPressure `dto:"out,api,pub" json:"cpu"`
	// Memory contains memory pressure metrics.
	Memory metrics.MemoryPressure `dto:"out,api,pub" json:"memory"`
	// IO contains I/O pressure metrics.
	IO metrics.IOPressure `dto:"out,api,pub" json:"io"`
}

// IOStatsSummary contains aggregated I/O statistics.
// This provides a summary of all disk I/O operations.
type IOStatsSummary struct {
	// ReadOps is the total number of read operations.
	ReadOps uint64 `dto:"out,api,pub" json:"readOps"`
	// ReadBytes is the total number of bytes read.
	ReadBytes uint64 `dto:"out,api,pub" json:"readBytes"`
	// WriteOps is the total number of write operations.
	WriteOps uint64 `dto:"out,api,pub" json:"writeOps"`
	// WriteBytes is the total number of bytes written.
	WriteBytes uint64 `dto:"out,api,pub" json:"writeBytes"`
	// Timestamp is when the stats were collected.
	Timestamp time.Time `dto:"out,api,pub" json:"timestamp"`
}

// PartitionInfo contains information about a mounted partition.
// Used for JSON output in the --probe command.
type PartitionInfo struct {
	// Device is the device name (e.g., "/dev/sda1").
	Device string `dto:"out,api,pub" json:"device"`
	// MountPoint is the mount path (e.g., "/home").
	MountPoint string `dto:"out,api,pub" json:"mount_point"`
	// FSType is the filesystem type (e.g., "ext4", "xfs").
	FSType string `dto:"out,api,pub" json:"fs_type"`
	// Options is the mount options as a comma-separated string.
	Options string `dto:"out,api,pub" json:"options,omitempty"`
}

// DiskUsageInfo contains disk usage statistics.
// Used for JSON output in the --probe command.
type DiskUsageInfo struct {
	// Path is the mount point path.
	Path string `dto:"out,api,pub" json:"path"`
	// TotalBytes is the total size in bytes.
	TotalBytes uint64 `dto:"out,api,pub" json:"total_bytes"`
	// UsedBytes is the used space in bytes.
	UsedBytes uint64 `dto:"out,api,pub" json:"used_bytes"`
	// FreeBytes is the free space in bytes.
	FreeBytes uint64 `dto:"out,api,pub" json:"free_bytes"`
	// UsedPercent is the percentage of space used.
	UsedPercent float64 `dto:"out,api,pub" json:"used_percent"`
	// InodesTotal is the total number of inodes.
	InodesTotal uint64 `dto:"out,api,pub" json:"inodes_total"`
	// InodesUsed is the number of used inodes.
	InodesUsed uint64 `dto:"out,api,pub" json:"inodes_used"`
	// InodesFree is the number of free inodes.
	InodesFree uint64 `dto:"out,api,pub" json:"inodes_free"`
}

// DiskIOInfo contains disk I/O statistics.
// Used for JSON output in the --probe command.
type DiskIOInfo struct {
	// Device is the device name (e.g., "sda", "nvme0n1").
	Device string `dto:"out,api,pub" json:"device"`
	// ReadsCompleted is the number of completed read operations.
	ReadsCompleted uint64 `dto:"out,api,pub" json:"reads_completed"`
	// SectorsRead is the number of sectors read.
	SectorsRead uint64 `dto:"out,api,pub" json:"sectors_read"`
	// ReadTimeMs is the time spent reading in milliseconds.
	ReadTimeMs uint64 `dto:"out,api,pub" json:"read_time_ms"`
	// WritesCompleted is the number of completed write operations.
	WritesCompleted uint64 `dto:"out,api,pub" json:"writes_completed"`
	// SectorsWritten is the number of sectors written.
	SectorsWritten uint64 `dto:"out,api,pub" json:"sectors_written"`
	// WriteTimeMs is the time spent writing in milliseconds.
	WriteTimeMs uint64 `dto:"out,api,pub" json:"write_time_ms"`
	// IOInProgress is the number of I/O operations in progress.
	IOInProgress uint64 `dto:"out,api,pub" json:"io_in_progress"`
	// IOTimeMs is the total time spent on I/O in milliseconds.
	IOTimeMs uint64 `dto:"out,api,pub" json:"io_time_ms"`
	// WeightedIOTimeMs is the weighted time spent on I/O.
	WeightedIOTimeMs uint64 `dto:"out,api,pub" json:"weighted_io_time_ms"`
}

// NetInterfaceInfo contains network interface information.
// Used for JSON output in the --probe command.
type NetInterfaceInfo struct {
	// Name is the interface name (e.g., "eth0", "en0").
	Name string `dto:"out,api,pub" json:"name"`
	// MACAddress is the hardware MAC address.
	MACAddress string `dto:"out,api,pub" json:"mac_address"`
	// MTU is the maximum transmission unit.
	MTU uint32 `dto:"out,api,pub" json:"mtu"`
	// IsUp indicates if the interface is up.
	IsUp bool `dto:"out,api,pub" json:"is_up"`
	// IsLoopback indicates if this is a loopback interface.
	IsLoopback bool `dto:"out,api,pub" json:"is_loopback"`
}

// NetStatsInfo contains network interface statistics.
// Used for JSON output in the --probe command.
type NetStatsInfo struct {
	// Interface is the interface name.
	Interface string `dto:"out,api,pub" json:"interface"`
	// RxBytes is the number of bytes received.
	RxBytes uint64 `dto:"out,api,pub" json:"rx_bytes"`
	// RxPackets is the number of packets received.
	RxPackets uint64 `dto:"out,api,pub" json:"rx_packets"`
	// RxErrors is the number of receive errors.
	RxErrors uint64 `dto:"out,api,pub" json:"rx_errors"`
	// RxDrops is the number of received packets dropped.
	RxDrops uint64 `dto:"out,api,pub" json:"rx_drops"`
	// TxBytes is the number of bytes transmitted.
	TxBytes uint64 `dto:"out,api,pub" json:"tx_bytes"`
	// TxPackets is the number of packets transmitted.
	TxPackets uint64 `dto:"out,api,pub" json:"tx_packets"`
	// TxErrors is the number of transmit errors.
	TxErrors uint64 `dto:"out,api,pub" json:"tx_errors"`
	// TxDrops is the number of transmitted packets dropped.
	TxDrops uint64 `dto:"out,api,pub" json:"tx_drops"`
}
