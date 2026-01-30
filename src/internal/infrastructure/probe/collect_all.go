//go:build cgo

package probe

/*
#include "probe.h"
*/
import "C"
import (
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// AllMetrics contains all system metrics collected in one call.
type AllMetrics struct {
	// CPU metrics.
	CPU metrics.SystemCPU
	// Memory metrics.
	Memory metrics.SystemMemory
	// Load average.
	Load metrics.LoadAverage
	// I/O statistics.
	IOStats IOStatsSummary
	// Pressure metrics (nil on non-Linux platforms).
	Pressure *AllPressure
	// Timestamp when metrics were collected.
	Timestamp time.Time

	// Disk partitions.
	Partitions []PartitionInfo
	// Disk usage for all partitions.
	DiskUsage []DiskUsageInfo
	// Disk I/O statistics.
	DiskIO []DiskIOInfo
	// Network interfaces.
	NetInterfaces []NetInterfaceInfo
	// Network statistics.
	NetStats []NetStatsInfo
}

// AllPressure contains all pressure metrics (Linux PSI).
type AllPressure struct {
	CPU    metrics.CPUPressure
	Memory metrics.MemoryPressure
	IO     metrics.IOPressure
}

// IOStatsSummary contains system-wide I/O statistics.
type IOStatsSummary struct {
	ReadOps    uint64
	ReadBytes  uint64
	WriteOps   uint64
	WriteBytes uint64
	Timestamp  time.Time
}

// PartitionInfo contains partition information.
type PartitionInfo struct {
	Device     string
	MountPoint string
	FSType     string
	Options    string
}

// DiskUsageInfo contains disk usage information.
type DiskUsageInfo struct {
	Path        string
	TotalBytes  uint64
	UsedBytes   uint64
	FreeBytes   uint64
	UsedPercent float64
	InodesTotal uint64
	InodesUsed  uint64
	InodesFree  uint64
}

// DiskIOInfo contains disk I/O statistics.
type DiskIOInfo struct {
	Device           string
	ReadsCompleted   uint64
	SectorsRead      uint64
	ReadTimeMs       uint64
	WritesCompleted  uint64
	SectorsWritten   uint64
	WriteTimeMs      uint64
	IOInProgress     uint64
	IOTimeMs         uint64
	WeightedIOTimeMs uint64
}

// NetInterfaceInfo contains network interface information.
type NetInterfaceInfo struct {
	Name       string
	MACAddress string
	MTU        uint32
	IsUp       bool
	IsLoopback bool
}

// NetStatsInfo contains network interface statistics.
type NetStatsInfo struct {
	Interface string
	RxBytes   uint64
	RxPackets uint64
	RxErrors  uint64
	RxDrops   uint64
	TxBytes   uint64
	TxPackets uint64
	TxErrors  uint64
	TxDrops   uint64
}

// CollectAll collects all system metrics in one call.
//
// This is more efficient than calling each collector individually
// and provides a consistent snapshot of all metrics at approximately
// the same point in time.
func CollectAll() (*AllMetrics, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var cAll C.AllMetrics
	result := C.probe_collect_all(&cAll)
	if err := resultToError(result); err != nil {
		return nil, err
	}

	all := &AllMetrics{
		CPU: metrics.SystemCPU{
			UsagePercent: 100.0 - float64(cAll.cpu.idle_percent),
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

	// Copy partitions
	partitionCount := int(cAll.partition_count)
	all.Partitions = make([]PartitionInfo, partitionCount)
	for i := 0; i < partitionCount; i++ {
		p := cAll.partitions[i]
		all.Partitions[i] = PartitionInfo{
			Device:     C.GoString(&p.device[0]),
			MountPoint: C.GoString(&p.mount_point[0]),
			FSType:     C.GoString(&p.fs_type[0]),
			Options:    C.GoString(&p.options[0]),
		}
	}

	// Copy disk usage
	usageCount := int(cAll.disk_usage_count)
	all.DiskUsage = make([]DiskUsageInfo, usageCount)
	for i := 0; i < usageCount; i++ {
		u := cAll.disk_usage[i]
		all.DiskUsage[i] = DiskUsageInfo{
			Path:        C.GoString(&u.path[0]),
			TotalBytes:  uint64(u.total_bytes),
			UsedBytes:   uint64(u.used_bytes),
			FreeBytes:   uint64(u.free_bytes),
			UsedPercent: float64(u.used_percent),
			InodesTotal: uint64(u.inodes_total),
			InodesUsed:  uint64(u.inodes_used),
			InodesFree:  uint64(u.inodes_free),
		}
	}

	// Copy disk I/O
	ioCount := int(cAll.disk_io_count)
	all.DiskIO = make([]DiskIOInfo, ioCount)
	for i := 0; i < ioCount; i++ {
		io := cAll.disk_io[i]
		all.DiskIO[i] = DiskIOInfo{
			Device:           C.GoString(&io.device[0]),
			ReadsCompleted:   uint64(io.reads_completed),
			SectorsRead:      uint64(io.sectors_read),
			ReadTimeMs:       uint64(io.read_time_ms),
			WritesCompleted:  uint64(io.writes_completed),
			SectorsWritten:   uint64(io.sectors_written),
			WriteTimeMs:      uint64(io.write_time_ms),
			IOInProgress:     uint64(io.io_in_progress),
			IOTimeMs:         uint64(io.io_time_ms),
			WeightedIOTimeMs: uint64(io.weighted_io_time_ms),
		}
	}

	// Copy network interfaces
	ifaceCount := int(cAll.net_interface_count)
	all.NetInterfaces = make([]NetInterfaceInfo, ifaceCount)
	for i := 0; i < ifaceCount; i++ {
		iface := cAll.net_interfaces[i]
		all.NetInterfaces[i] = NetInterfaceInfo{
			Name:       C.GoString(&iface.name[0]),
			MACAddress: C.GoString(&iface.mac_address[0]),
			MTU:        uint32(iface.mtu),
			IsUp:       bool(iface.is_up),
			IsLoopback: bool(iface.is_loopback),
		}
	}

	// Copy network stats
	statsCount := int(cAll.net_stats_count)
	all.NetStats = make([]NetStatsInfo, statsCount)
	for i := 0; i < statsCount; i++ {
		s := cAll.net_stats[i]
		all.NetStats[i] = NetStatsInfo{
			Interface: C.GoString(&s._interface[0]),
			RxBytes:   uint64(s.rx_bytes),
			RxPackets: uint64(s.rx_packets),
			RxErrors:  uint64(s.rx_errors),
			RxDrops:   uint64(s.rx_drops),
			TxBytes:   uint64(s.tx_bytes),
			TxPackets: uint64(s.tx_packets),
			TxErrors:  uint64(s.tx_errors),
			TxDrops:   uint64(s.tx_drops),
		}
	}

	return all, nil
}

// ToPartition converts PartitionInfo to domain Partition.
func (p *PartitionInfo) ToPartition() metrics.Partition {
	options := strings.Split(p.Options, ",")
	return metrics.Partition{
		Device:     p.Device,
		Mountpoint: p.MountPoint,
		FSType:     p.FSType,
		Options:    options,
	}
}

// ToNetStats converts NetStatsInfo to domain NetStats.
func (n *NetStatsInfo) ToNetStats() metrics.NetStats {
	return metrics.NetStats{
		Interface:   n.Interface,
		BytesSent:   n.TxBytes,
		BytesRecv:   n.RxBytes,
		PacketsSent: n.TxPackets,
		PacketsRecv: n.RxPackets,
		ErrorsIn:    n.RxErrors,
		ErrorsOut:   n.TxErrors,
		DropsIn:     n.RxDrops,
		DropsOut:    n.TxDrops,
		Timestamp:   time.Now(),
	}
}
