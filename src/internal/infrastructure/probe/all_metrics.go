//go:build cgo

// doc://
// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
//
//nolint:ktn-struct-onefile // This file contains JSON output structs that are logically grouped together
package probe

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"
)

// AllSystemMetrics contains all metrics that can be collected on the current platform.
// This is used for the --probe CLI command to output comprehensive system information.
type AllSystemMetrics struct {
	// Metadata about the collection
	Timestamp   time.Time `json:"timestamp"`
	Platform    string    `json:"platform"`
	Hostname    string    `json:"hostname,omitempty"`
	CollectedAt int64     `json:"collected_at_ns"`

	// Basic system metrics
	CPU    *CPUMetricsJSON    `json:"cpu,omitempty"`
	Memory *MemoryMetricsJSON `json:"memory,omitempty"`
	Load   *LoadMetricsJSON   `json:"load,omitempty"`

	// Disk metrics
	Disk *DiskMetricsJSON `json:"disk,omitempty"`

	// Network metrics
	Network *NetworkMetricsJSON `json:"network,omitempty"`

	// I/O metrics
	IO *IOMetricsJSON `json:"io,omitempty"`

	// Process metrics
	Process *ProcessMetricsJSON `json:"process,omitempty"`

	// Thermal metrics (Linux only)
	Thermal *ThermalMetricsJSON `json:"thermal,omitempty"`

	// Context switches (Linux only)
	ContextSwitches *ContextSwitchMetricsJSON `json:"context_switches,omitempty"`

	// Network connections (Linux only)
	Connections *ConnectionMetricsJSON `json:"connections,omitempty"`

	// Resource quotas and container info
	Quota     *QuotaMetricsJSON     `json:"quota,omitempty"`
	Container *ContainerMetricsJSON `json:"container,omitempty"`
	Runtime   *RuntimeMetricsJSON   `json:"runtime,omitempty"`
}

// CPUMetricsJSON contains CPU-related metrics for JSON output.
// It includes usage percentage, core count, and optional pressure metrics.
type CPUMetricsJSON struct {
	UsagePercent float64          `json:"usage_percent"`
	Cores        uint32           `json:"cores"`
	FrequencyMHz uint64           `json:"frequency_mhz"`
	Pressure     *CPUPressureJSON `json:"pressure,omitempty"`
}

// CPUPressureJSON contains PSI pressure metrics for CPU.
// It tracks CPU contention using Linux Pressure Stall Information.
type CPUPressureJSON struct {
	SomeAvg10   float64 `json:"some_avg10"`
	SomeAvg60   float64 `json:"some_avg60"`
	SomeAvg300  float64 `json:"some_avg300"`
	SomeTotalUs uint64  `json:"some_total_us"`
}

// MemoryMetricsJSON contains memory-related metrics for JSON output.
// It includes total, available, used, cached, and swap memory statistics.
type MemoryMetricsJSON struct {
	TotalBytes     uint64              `json:"total_bytes"`
	AvailableBytes uint64              `json:"available_bytes"`
	UsedBytes      uint64              `json:"used_bytes"`
	CachedBytes    uint64              `json:"cached_bytes"`
	BuffersBytes   uint64              `json:"buffers_bytes"`
	SwapTotalBytes uint64              `json:"swap_total_bytes"`
	SwapUsedBytes  uint64              `json:"swap_used_bytes"`
	UsedPercent    float64             `json:"used_percent"`
	Pressure       *MemoryPressureJSON `json:"pressure,omitempty"`
}

// MemoryPressureJSON contains PSI pressure metrics for memory.
// It tracks memory contention using Linux Pressure Stall Information.
type MemoryPressureJSON struct {
	SomeAvg10   float64 `json:"some_avg10"`
	SomeAvg60   float64 `json:"some_avg60"`
	SomeAvg300  float64 `json:"some_avg300"`
	SomeTotalUs uint64  `json:"some_total_us"`
	FullAvg10   float64 `json:"full_avg10"`
	FullAvg60   float64 `json:"full_avg60"`
	FullAvg300  float64 `json:"full_avg300"`
	FullTotalUs uint64  `json:"full_total_us"`
}

// LoadMetricsJSON contains load average metrics for JSON output.
// It provides 1, 5, and 15 minute system load averages.
type LoadMetricsJSON struct {
	Load1Min  float64 `json:"load_1min"`
	Load5Min  float64 `json:"load_5min"`
	Load15Min float64 `json:"load_15min"`
}

// DiskMetricsJSON contains disk-related metrics for JSON output.
// Uses types from collect_all.go: PartitionInfo, DiskUsageInfo, DiskIOInfo.
type DiskMetricsJSON struct {
	Partitions []PartitionInfo `json:"partitions,omitempty"`
	Usage      []DiskUsageInfo `json:"usage,omitempty"`
	IO         []DiskIOInfo    `json:"io,omitempty"`
}

// NetworkMetricsJSON contains network-related metrics for JSON output.
// Uses types from collect_all.go: NetInterfaceInfo, NetStatsInfo.
type NetworkMetricsJSON struct {
	Interfaces []NetInterfaceJSON `json:"interfaces,omitempty"`
	Stats      []NetStatsJSON     `json:"stats,omitempty"`
}

// NetInterfaceJSON contains information about a network interface.
// It includes name, MAC address, MTU, and status flags.
type NetInterfaceJSON struct {
	Name       string   `json:"name"`
	MACAddress string   `json:"mac_address"`
	MTU        uint32   `json:"mtu"`
	IsUp       bool     `json:"is_up"`
	IsLoopback bool     `json:"is_loopback"`
	Flags      []string `json:"flags,omitempty"`
}

// NetStatsJSON contains network statistics for an interface.
// It tracks bytes, packets, errors, and drops for both directions.
type NetStatsJSON struct {
	Interface   string `json:"interface"`
	BytesRecv   uint64 `json:"bytes_recv"`
	BytesSent   uint64 `json:"bytes_sent"`
	PacketsRecv uint64 `json:"packets_recv"`
	PacketsSent uint64 `json:"packets_sent"`
	ErrorsIn    uint64 `json:"errors_in"`
	ErrorsOut   uint64 `json:"errors_out"`
	DropsIn     uint64 `json:"drops_in"`
	DropsOut    uint64 `json:"drops_out"`
}

// IOMetricsJSON contains I/O-related metrics for JSON output.
// It includes read/write operations, bytes, and optional pressure metrics.
type IOMetricsJSON struct {
	ReadOps    uint64          `json:"read_ops"`
	ReadBytes  uint64          `json:"read_bytes"`
	WriteOps   uint64          `json:"write_ops"`
	WriteBytes uint64          `json:"write_bytes"`
	Pressure   *IOPressureJSON `json:"pressure,omitempty"`
}

// IOPressureJSON contains PSI pressure metrics for I/O.
// It tracks I/O contention using Linux Pressure Stall Information.
type IOPressureJSON struct {
	SomeAvg10   float64 `json:"some_avg10"`
	SomeAvg60   float64 `json:"some_avg60"`
	SomeAvg300  float64 `json:"some_avg300"`
	SomeTotalUs uint64  `json:"some_total_us"`
	FullAvg10   float64 `json:"full_avg10"`
	FullAvg60   float64 `json:"full_avg60"`
	FullAvg300  float64 `json:"full_avg300"`
	FullTotalUs uint64  `json:"full_total_us"`
}

// ProcessMetricsJSON contains information about the current process and system processes.
// It includes PID, process count, and resource usage for top processes.
type ProcessMetricsJSON struct {
	CurrentPID   int32             `json:"current_pid"`
	ProcessCount int               `json:"process_count"`
	TopProcesses []ProcessInfoJSON `json:"top_processes,omitempty"`
}

// ProcessInfoJSON contains information about a process.
// It includes CPU, memory, thread, and file descriptor statistics.
type ProcessInfoJSON struct {
	PID            int32   `json:"pid"`
	CPUPercent     float64 `json:"cpu_percent"`
	MemoryRSSBytes uint64  `json:"memory_rss_bytes"`
	MemoryVMSBytes uint64  `json:"memory_vms_bytes"`
	MemoryPercent  float64 `json:"memory_percent"`
	NumThreads     uint32  `json:"num_threads"`
	NumFDs         uint32  `json:"num_fds"`
	State          string  `json:"state"`
}

// ThermalMetricsJSON contains thermal sensor information.
// It indicates support status and provides zone-specific temperature data.
type ThermalMetricsJSON struct {
	Supported bool              `json:"supported"`
	Zones     []ThermalZoneJSON `json:"zones,omitempty"`
}

// ThermalZoneJSON contains information about a thermal zone.
// It includes name, current temperature, and optional threshold values.
type ThermalZoneJSON struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	TempCelsius float64  `json:"temp_celsius"`
	TempMax     *float64 `json:"temp_max,omitempty"`
	TempCrit    *float64 `json:"temp_crit,omitempty"`
}

// ContextSwitchMetricsJSON contains context switch statistics.
// It tracks system-wide and per-process context switch counts.
type ContextSwitchMetricsJSON struct {
	SystemTotal uint64                 `json:"system_total"`
	Self        *ContextSwitchInfoJSON `json:"self,omitempty"`
}

// ContextSwitchInfoJSON contains context switch counts for a process.
// It separates voluntary and involuntary context switches.
type ContextSwitchInfoJSON struct {
	Voluntary   uint64 `json:"voluntary"`
	Involuntary uint64 `json:"involuntary"`
}

// ConnectionMetricsJSON contains network connection information.
// It includes TCP stats, connections, UDP sockets, and Unix sockets.
type ConnectionMetricsJSON struct {
	TCPStats       *TcpStatsJSON    `json:"tcp_stats,omitempty"`
	TCPConnections []TcpConnJSON    `json:"tcp_connections,omitempty"`
	UDPSockets     []UdpConnJSON    `json:"udp_sockets,omitempty"`
	UnixSockets    []UnixSockJSON   `json:"unix_sockets,omitempty"`
	ListeningPorts []ListenInfoJSON `json:"listening_ports,omitempty"`
}

// TcpStatsJSON contains aggregated TCP statistics.
// It tracks connection counts by state for monitoring and diagnostics.
type TcpStatsJSON struct {
	Established uint32 `json:"established"`
	SynSent     uint32 `json:"syn_sent"`
	SynRecv     uint32 `json:"syn_recv"`
	FinWait1    uint32 `json:"fin_wait1"`
	FinWait2    uint32 `json:"fin_wait2"`
	TimeWait    uint32 `json:"time_wait"`
	Close       uint32 `json:"close"`
	CloseWait   uint32 `json:"close_wait"`
	LastAck     uint32 `json:"last_ack"`
	Listen      uint32 `json:"listen"`
	Closing     uint32 `json:"closing"`
	Total       uint32 `json:"total"`
}

// TcpConnJSON contains information about a TCP connection.
// It includes local/remote endpoints, state, and owning process.
type TcpConnJSON struct {
	Family      string `json:"family"`
	LocalAddr   string `json:"local_addr"`
	LocalPort   uint16 `json:"local_port"`
	RemoteAddr  string `json:"remote_addr"`
	RemotePort  uint16 `json:"remote_port"`
	State       string `json:"state"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}

// UdpConnJSON contains information about a UDP socket.
// It includes local/remote endpoints and owning process.
type UdpConnJSON struct {
	Family      string `json:"family"`
	LocalAddr   string `json:"local_addr"`
	LocalPort   uint16 `json:"local_port"`
	RemoteAddr  string `json:"remote_addr"`
	RemotePort  uint16 `json:"remote_port"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}

// UnixSockJSON contains information about a Unix socket.
// It includes path, type, state, and owning process.
type UnixSockJSON struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	State       string `json:"state"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}

// ListenInfoJSON contains information about a listening port.
// It includes protocol, address, port, and owning process.
type ListenInfoJSON struct {
	Protocol    string `json:"protocol"`
	Address     string `json:"address"`
	Port        uint16 `json:"port"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}

// QuotaMetricsJSON contains resource quota information.
// It indicates support status and provides limit and usage data.
type QuotaMetricsJSON struct {
	Supported bool           `json:"supported"`
	Limits    *QuotaInfoJSON `json:"limits,omitempty"`
	Usage     *UsageInfoJSON `json:"usage,omitempty"`
}

// QuotaInfoJSON contains resource limit information.
// It includes CPU, memory, process, and file descriptor limits.
type QuotaInfoJSON struct {
	CPUQuotaUs       uint64 `json:"cpu_quota_us,omitempty"`
	CPUPeriodUs      uint64 `json:"cpu_period_us,omitempty"`
	MemoryLimitBytes uint64 `json:"memory_limit_bytes,omitempty"`
	PidsLimit        uint64 `json:"pids_limit,omitempty"`
	NofileLimit      uint64 `json:"nofile_limit,omitempty"`
}

// UsageInfoJSON contains current resource usage.
// It includes memory, process count, and CPU utilization.
type UsageInfoJSON struct {
	MemoryBytes      uint64  `json:"memory_bytes"`
	MemoryLimitBytes uint64  `json:"memory_limit_bytes,omitempty"`
	PidsCurrent      uint64  `json:"pids_current"`
	PidsLimit        uint64  `json:"pids_limit,omitempty"`
	CPUPercent       float64 `json:"cpu_percent"`
	CPULimitPercent  float64 `json:"cpu_limit_percent,omitempty"`
}

// ContainerMetricsJSON contains container detection information.
// It indicates containerization status, runtime, and container ID.
type ContainerMetricsJSON struct {
	IsContainerized bool   `json:"is_containerized"`
	Runtime         string `json:"runtime,omitempty"`
	ContainerID     string `json:"container_id,omitempty"`
}

// RuntimeMetricsJSON contains full runtime detection information.
// It includes container runtime, orchestrator, and available runtimes.
type RuntimeMetricsJSON struct {
	IsContainerized   bool                       `json:"is_containerized"`
	ContainerRuntime  string                     `json:"container_runtime,omitempty"`
	Orchestrator      string                     `json:"orchestrator,omitempty"`
	ContainerID       string                     `json:"container_id,omitempty"`
	WorkloadID        string                     `json:"workload_id,omitempty"`
	WorkloadName      string                     `json:"workload_name,omitempty"`
	Namespace         string                     `json:"namespace,omitempty"`
	AvailableRuntimes []AvailableRuntimeInfoJSON `json:"available_runtimes,omitempty"`
}

// AvailableRuntimeInfoJSON contains info about an available runtime on the host.
// It includes runtime name, socket path, version, and running status.
type AvailableRuntimeInfoJSON struct {
	Runtime    string `json:"runtime"`
	SocketPath string `json:"socket_path,omitempty"`
	Version    string `json:"version,omitempty"`
	IsRunning  bool   `json:"is_running"`
}

// CollectAllMetrics collects all available metrics for the current platform.
// Returns a comprehensive snapshot of system metrics as JSON-serializable struct.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - *AllSystemMetrics: collected system metrics
//   - error: nil on success, error if probe not initialized
//
//nolint:cyclop,funlen,ktn-func-maxloc // Comprehensive metrics collection requires many collection calls
func CollectAllMetrics(ctx context.Context) (*AllSystemMetrics, error) {
	// Check if probe is initialized
	if err := checkInitialized(); err != nil {
		// Return early if not initialized
		return nil, err
	}

	// Build initial result with metadata
	now := time.Now()
	hostname, _ := os.Hostname()
	result := &AllSystemMetrics{
		Timestamp:   now,
		Platform:    runtime.GOOS,
		Hostname:    hostname,
		CollectedAt: now.UnixNano(),
	}

	// Create collector instance
	collector := NewCollector()

	// Collect CPU metrics
	if cpu, err := collector.CPU().CollectSystem(ctx); err == nil {
		result.CPU = &CPUMetricsJSON{
			UsagePercent: cpu.UsagePercent,
			Cores:        uint32(runtime.NumCPU()),
		}
		// Try to get pressure metrics (Linux only)
		if pressure, err := collector.CPU().CollectPressure(ctx); err == nil {
			result.CPU.Pressure = &CPUPressureJSON{
				SomeAvg10:   pressure.SomeAvg10,
				SomeAvg60:   pressure.SomeAvg60,
				SomeAvg300:  pressure.SomeAvg300,
				SomeTotalUs: pressure.SomeTotal,
			}
		}
	}

	// Collect Memory metrics
	if mem, err := collector.Memory().CollectSystem(ctx); err == nil {
		result.Memory = &MemoryMetricsJSON{
			TotalBytes:     mem.Total,
			AvailableBytes: mem.Available,
			UsedBytes:      mem.Used,
			CachedBytes:    mem.Cached,
			BuffersBytes:   mem.Buffers,
			SwapTotalBytes: mem.SwapTotal,
			SwapUsedBytes:  mem.SwapUsed,
			UsedPercent:    mem.UsagePercent,
		}
		// Try to get pressure metrics (Linux only)
		if pressure, err := collector.Memory().CollectPressure(ctx); err == nil {
			result.Memory.Pressure = &MemoryPressureJSON{
				SomeAvg10:   pressure.SomeAvg10,
				SomeAvg60:   pressure.SomeAvg60,
				SomeAvg300:  pressure.SomeAvg300,
				SomeTotalUs: pressure.SomeTotal,
				FullAvg10:   pressure.FullAvg10,
				FullAvg60:   pressure.FullAvg60,
				FullAvg300:  pressure.FullAvg300,
				FullTotalUs: pressure.FullTotal,
			}
		}
	}

	// Collect Load metrics
	if load, err := collector.cpu.CollectLoadAverage(ctx); err == nil {
		result.Load = &LoadMetricsJSON{
			Load1Min:  load.Load1,
			Load5Min:  load.Load5,
			Load15Min: load.Load15,
		}
	}

	// Collect Disk metrics
	result.Disk = collectDiskMetricsJSON(ctx, collector)

	// Collect Network metrics
	result.Network = collectNetworkMetricsJSON(ctx, collector)

	// Collect I/O metrics
	result.IO = collectIOMetricsJSON(ctx, collector)

	// Collect Process metrics
	result.Process = collectProcessMetricsJSON(ctx)

	// Collect Thermal metrics (Linux only)
	result.Thermal = collectThermalMetricsJSON()

	// Collect Context switches (Linux only)
	result.ContextSwitches = collectContextSwitchMetricsJSON()

	// Collect Network connections (Linux only)
	result.Connections = collectConnectionMetricsJSON(ctx)

	// Collect Quota metrics
	result.Quota = collectQuotaMetricsJSON()

	// Collect Container/Runtime metrics
	result.Container = collectContainerMetricsJSON()
	result.Runtime = collectRuntimeMetricsJSON()

	// Return the collected metrics
	return result, nil
}

// collectDiskMetricsJSON collects all disk-related metrics.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance to use
//
// Returns:
//   - *DiskMetricsJSON: collected disk metrics
//
//nolint:funlen,ktn-func-maxloc // Disk metrics collection requires gathering partition, usage, and I/O data
func collectDiskMetricsJSON(ctx context.Context, coll *Collector) *DiskMetricsJSON {
	// Initialize disk metrics struct
	disk := &DiskMetricsJSON{}

	// Collect partition information
	if partitions, err := coll.Disk().ListPartitions(ctx); err == nil {
		disk.Partitions = make([]PartitionInfo, 0, len(partitions))
		// Iterate over each partition
		for _, pt := range partitions {
			disk.Partitions = append(disk.Partitions, PartitionInfo{
				Device:     pt.Device,
				MountPoint: pt.Mountpoint,
				FSType:     pt.FSType,
				Options:    joinOptions(pt.Options),
			})
		}
	}

	// Collect disk usage information
	if usage, err := coll.Disk().CollectAllUsage(ctx); err == nil {
		disk.Usage = make([]DiskUsageInfo, 0, len(usage))
		// Iterate over each usage entry
		for _, us := range usage {
			disk.Usage = append(disk.Usage, DiskUsageInfo{
				Path:        us.Path,
				TotalBytes:  us.Total,
				UsedBytes:   us.Used,
				FreeBytes:   us.Free,
				UsedPercent: us.UsagePercent,
				InodesTotal: us.InodesTotal,
				InodesUsed:  us.InodesUsed,
				InodesFree:  us.InodesFree,
			})
		}
	}

	// Collect disk I/O information
	if ioStats, err := coll.Disk().CollectIO(ctx); err == nil {
		disk.IO = make([]DiskIOInfo, 0, len(ioStats))
		// Iterate over each I/O entry
		for _, io := range ioStats {
			disk.IO = append(disk.IO, DiskIOInfo{
				Device:           io.Device,
				ReadsCompleted:   io.ReadCount,
				SectorsRead:      io.ReadBytes / sectorSize,
				ReadTimeMs:       uint64(io.ReadTime.Milliseconds()),
				WritesCompleted:  io.WriteCount,
				SectorsWritten:   io.WriteBytes / sectorSize,
				WriteTimeMs:      uint64(io.WriteTime.Milliseconds()),
				IOInProgress:     io.IOInProgress,
				IOTimeMs:         uint64(io.IOTime.Milliseconds()),
				WeightedIOTimeMs: uint64(io.WeightedIOTime.Milliseconds()),
			})
		}
	}

	// Return the collected disk metrics
	return disk
}

// joinOptions joins slice of options into a comma-separated string.
//
// Params:
//   - opts: slice of option strings to join
//
// Returns:
//   - string: comma-separated options string
func joinOptions(opts []string) string {
	// Return empty string for empty slice
	if len(opts) == 0 {
		// Return empty string
		return ""
	}
	// Use strings.Join for efficient string concatenation
	return strings.Join(opts, ",")
}

// containsFlag checks if a flag is present in the flags slice.
//
// Params:
//   - flags: slice of flags to search
//   - flag: flag to search for
//
// Returns:
//   - bool: true if flag is present, false otherwise
func containsFlag(flags []string, flag string) bool {
	// Use slices.Contains for efficient membership check
	return slices.Contains(flags, flag)
}

// collectNetworkMetricsJSON collects all network-related metrics.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance to use
//
// Returns:
//   - *NetworkMetricsJSON: collected network metrics
func collectNetworkMetricsJSON(ctx context.Context, coll *Collector) *NetworkMetricsJSON {
	// Initialize network metrics struct
	network := &NetworkMetricsJSON{}

	// Collect interface information
	if ifaces, err := coll.Network().ListInterfaces(ctx); err == nil {
		network.Interfaces = make([]NetInterfaceJSON, 0, len(ifaces))
		// Iterate over each interface
		for _, iface := range ifaces {
			// Derive IsUp and IsLoopback from flags
			isUp := containsFlag(iface.Flags, "up")
			isLoopback := containsFlag(iface.Flags, "loopback")
			network.Interfaces = append(network.Interfaces, NetInterfaceJSON{
				Name:       iface.Name,
				MACAddress: iface.HardwareAddr,
				MTU:        uint32(iface.MTU),
				IsUp:       isUp,
				IsLoopback: isLoopback,
				Flags:      iface.Flags,
			})
		}
	}

	// Collect network statistics
	if stats, err := coll.Network().CollectAllStats(ctx); err == nil {
		network.Stats = make([]NetStatsJSON, 0, len(stats))
		// Iterate over each stats entry
		for _, st := range stats {
			network.Stats = append(network.Stats, NetStatsJSON{
				Interface:   st.Interface,
				BytesRecv:   st.BytesRecv,
				BytesSent:   st.BytesSent,
				PacketsRecv: st.PacketsRecv,
				PacketsSent: st.PacketsSent,
				ErrorsIn:    st.ErrorsIn,
				ErrorsOut:   st.ErrorsOut,
				DropsIn:     st.DropsIn,
				DropsOut:    st.DropsOut,
			})
		}
	}

	// Return the collected network metrics
	return network
}

// collectIOMetricsJSON collects all I/O-related metrics.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance to use
//
// Returns:
//   - *IOMetricsJSON: collected I/O metrics
func collectIOMetricsJSON(ctx context.Context, coll *Collector) *IOMetricsJSON {
	// Initialize I/O metrics struct
	ioMetrics := &IOMetricsJSON{}

	// Collect I/O statistics
	if stats, err := coll.IO().CollectStats(ctx); err == nil {
		ioMetrics.ReadOps = stats.ReadOpsTotal
		ioMetrics.ReadBytes = stats.ReadBytesTotal
		ioMetrics.WriteOps = stats.WriteOpsTotal
		ioMetrics.WriteBytes = stats.WriteBytesTotal
	}

	// Collect I/O pressure (Linux only)
	if pressure, err := coll.IO().CollectPressure(ctx); err == nil {
		ioMetrics.Pressure = &IOPressureJSON{
			SomeAvg10:   pressure.SomeAvg10,
			SomeAvg60:   pressure.SomeAvg60,
			SomeAvg300:  pressure.SomeAvg300,
			SomeTotalUs: pressure.SomeTotal,
			FullAvg10:   pressure.FullAvg10,
			FullAvg60:   pressure.FullAvg60,
			FullAvg300:  pressure.FullAvg300,
			FullTotalUs: pressure.FullTotal,
		}
	}

	// Return the collected I/O metrics
	return ioMetrics
}

// collectProcessMetricsJSON collects process-related metrics.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - *ProcessMetricsJSON: collected process metrics
func collectProcessMetricsJSON(ctx context.Context) *ProcessMetricsJSON {
	// Get current process ID
	pid := os.Getpid()
	pm := &ProcessMetricsJSON{
		CurrentPID: int32(pid),
	}

	// Get current process info using ProcessCollector
	pc := NewProcessCollector()
	cpuInfo, cpuErr := pc.CollectCPU(ctx, pid)
	memInfo, memErr := pc.CollectMemory(ctx, pid)

	// Check if both collections succeeded
	if cpuErr == nil && memErr == nil {
		pm.TopProcesses = []ProcessInfoJSON{{
			PID:            int32(cpuInfo.PID),
			CPUPercent:     cpuInfo.UsagePercent,
			MemoryRSSBytes: memInfo.RSS,
			MemoryVMSBytes: memInfo.VMS,
			MemoryPercent:  memInfo.UsagePercent,
		}}
	}

	// Return the collected process metrics
	return pm
}

// collectThermalMetricsJSON collects thermal sensor metrics (Linux only).
//
// Returns:
//   - *ThermalMetricsJSON: collected thermal metrics
func collectThermalMetricsJSON() *ThermalMetricsJSON {
	// Initialize thermal metrics struct
	thermal := &ThermalMetricsJSON{
		Supported: ThermalIsSupported(),
	}

	// Return early if not supported
	if !thermal.Supported {
		// Return unsupported thermal metrics
		return thermal
	}

	// Collect thermal zones
	if zones, err := CollectThermalZones(); err == nil {
		thermal.Zones = make([]ThermalZoneJSON, 0, len(zones))
		// Iterate over each zone
		for _, zn := range zones {
			// ThermalZone and ThermalZoneJSON have identical underlying types.
			thermal.Zones = append(thermal.Zones, ThermalZoneJSON(zn))
		}
	}

	// Return the collected thermal metrics
	return thermal
}

// collectContextSwitchMetricsJSON collects context switch metrics (Linux only).
//
// Returns:
//   - *ContextSwitchMetricsJSON: collected context switch metrics
func collectContextSwitchMetricsJSON() *ContextSwitchMetricsJSON {
	// Initialize context switch metrics struct
	cs := &ContextSwitchMetricsJSON{}

	// Collect system-wide context switches
	if total, err := CollectSystemContextSwitches(); err == nil {
		cs.SystemTotal = total
	}

	// Collect self context switches
	if self, err := CollectSelfContextSwitches(); err == nil {
		cs.Self = &ContextSwitchInfoJSON{
			Voluntary:   self.Voluntary,
			Involuntary: self.Involuntary,
		}
	}

	// Return the collected context switch metrics
	return cs
}

// collectConnectionMetricsJSON collects network connection metrics (Linux only).
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - *ConnectionMetricsJSON: collected connection metrics
//
//nolint:cyclop,funlen,ktn-func-maxloc,ktn-func-cyclo // Connection metrics collection requires gathering data from multiple sources
func collectConnectionMetricsJSON(ctx context.Context) *ConnectionMetricsJSON {
	// Initialize connection metrics struct
	conn := &ConnectionMetricsJSON{}
	connCollector := NewConnectionCollector()

	// Collect TCP Stats
	if stats, err := connCollector.CollectTCPStats(ctx); err == nil {
		conn.TCPStats = &TcpStatsJSON{
			Established: stats.Established,
			SynSent:     stats.SynSent,
			SynRecv:     stats.SynRecv,
			FinWait1:    stats.FinWait1,
			FinWait2:    stats.FinWait2,
			TimeWait:    stats.TimeWait,
			Close:       stats.Close,
			CloseWait:   stats.CloseWait,
			LastAck:     stats.LastAck,
			Listen:      stats.Listen,
			Closing:     stats.Closing,
			Total:       stats.Total(),
		}
	}

	// Collect TCP Connections
	if tcpConns, err := connCollector.CollectTCP(ctx); err == nil {
		conn.TCPConnections = make([]TcpConnJSON, 0, len(tcpConns))
		// Iterate over each TCP connection
		for _, tc := range tcpConns {
			conn.TCPConnections = append(conn.TCPConnections, TcpConnJSON{
				Family:      tc.Family.String(),
				LocalAddr:   tc.LocalAddr,
				LocalPort:   tc.LocalPort,
				RemoteAddr:  tc.RemoteAddr,
				RemotePort:  tc.RemotePort,
				State:       tc.State.String(),
				PID:         tc.PID,
				ProcessName: tc.ProcessName,
			})
		}
	}

	// Collect UDP Sockets
	if udpConns, err := connCollector.CollectUDP(ctx); err == nil {
		conn.UDPSockets = make([]UdpConnJSON, 0, len(udpConns))
		// Iterate over each UDP socket
		for _, uc := range udpConns {
			conn.UDPSockets = append(conn.UDPSockets, UdpConnJSON{
				Family:      uc.Family.String(),
				LocalAddr:   uc.LocalAddr,
				LocalPort:   uc.LocalPort,
				RemoteAddr:  uc.RemoteAddr,
				RemotePort:  uc.RemotePort,
				PID:         uc.PID,
				ProcessName: uc.ProcessName,
			})
		}
	}

	// Collect Unix Sockets
	if unixSocks, err := connCollector.CollectUnix(ctx); err == nil {
		conn.UnixSockets = make([]UnixSockJSON, 0, len(unixSocks))
		// Iterate over each Unix socket
		for _, us := range unixSocks {
			conn.UnixSockets = append(conn.UnixSockets, UnixSockJSON{
				Path:        us.Path,
				Type:        us.SocketType,
				State:       us.State.String(),
				PID:         us.PID,
				ProcessName: us.ProcessName,
			})
		}
	}

	// Collect Listening Ports
	if listening, err := connCollector.CollectListeningPorts(ctx); err == nil {
		conn.ListeningPorts = make([]ListenInfoJSON, 0, len(listening))
		// Iterate over each listening port
		for _, lp := range listening {
			conn.ListeningPorts = append(conn.ListeningPorts, ListenInfoJSON{
				Protocol:    "tcp",
				Address:     lp.LocalAddr,
				Port:        lp.LocalPort,
				PID:         lp.PID,
				ProcessName: lp.ProcessName,
			})
		}
	}

	// Return the collected connection metrics
	return conn
}

// collectQuotaMetricsJSON collects resource quota metrics.
//
// Returns:
//   - *QuotaMetricsJSON: collected quota metrics
func collectQuotaMetricsJSON() *QuotaMetricsJSON {
	// Initialize quota metrics struct
	quota := &QuotaMetricsJSON{
		Supported: true, // Probe is supported if we got this far
	}

	// Get current process ID
	pid := os.Getpid()

	// Collect quota limits
	if limits, err := ReadQuotaLimits(pid); err == nil {
		quota.Limits = &QuotaInfoJSON{
			CPUQuotaUs:       limits.CPUQuotaUS,
			CPUPeriodUs:      limits.CPUPeriodUS,
			MemoryLimitBytes: limits.MemoryLimitBytes,
			PidsLimit:        limits.PIDsLimit,
			NofileLimit:      limits.NofileLimit,
		}
	} else {
		// Quota limits unavailable, mark as unsupported
		quota.Supported = false
	}

	// Collect quota usage
	if usage, err := ReadQuotaUsage(pid); err == nil {
		quota.Usage = &UsageInfoJSON{
			MemoryBytes:      usage.MemoryBytes,
			MemoryLimitBytes: usage.MemoryLimitBytes,
			PidsCurrent:      usage.PIDsCurrent,
			PidsLimit:        usage.PIDsLimit,
			CPUPercent:       usage.CPUPercent,
			CPULimitPercent:  usage.CPULimitPercent,
		}
	}

	// Return the collected quota metrics
	return quota
}

// collectContainerMetricsJSON collects container detection information.
//
// Returns:
//   - *ContainerMetricsJSON: collected container metrics
func collectContainerMetricsJSON() *ContainerMetricsJSON {
	// Detect container environment
	info, err := DetectContainer()
	// Check if detection failed
	if err != nil {
		// Return not containerized on error
		return &ContainerMetricsJSON{IsContainerized: false}
	}

	// Return the collected container metrics
	return &ContainerMetricsJSON{
		IsContainerized: info.IsContainerized,
		Runtime:         info.Runtime.String(),
		ContainerID:     info.ContainerID,
	}
}

// collectRuntimeMetricsJSON collects full runtime detection information.
//
// Returns:
//   - *RuntimeMetricsJSON: collected runtime metrics
func collectRuntimeMetricsJSON() *RuntimeMetricsJSON {
	// Detect runtime environment
	info, err := DetectRuntime()
	// Check if detection failed
	if err != nil {
		// Return not containerized on error
		return &RuntimeMetricsJSON{IsContainerized: false}
	}

	// Build runtime metrics struct
	rm := &RuntimeMetricsJSON{
		IsContainerized:  info.IsContainerized,
		ContainerRuntime: info.ContainerRuntime.String(),
		Orchestrator:     info.Orchestrator.String(),
		ContainerID:      info.ContainerID,
		WorkloadID:       info.WorkloadID,
		WorkloadName:     info.WorkloadName,
		Namespace:        info.Namespace,
	}

	// Check if available runtimes exist
	if len(info.AvailableRuntimes) > 0 {
		rm.AvailableRuntimes = make([]AvailableRuntimeInfoJSON, 0, len(info.AvailableRuntimes))
		// Iterate over each available runtime
		for _, ar := range info.AvailableRuntimes {
			rm.AvailableRuntimes = append(rm.AvailableRuntimes, AvailableRuntimeInfoJSON{
				Runtime:    ar.Runtime.String(),
				SocketPath: ar.SocketPath,
				Version:    ar.Version,
				IsRunning:  ar.IsRunning,
			})
		}
	}

	// Return the collected runtime metrics
	return rm
}

// CollectAllMetricsJSON collects all metrics and returns them as a JSON string.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - string: JSON-encoded metrics
//   - error: nil on success, error if collection or encoding fails
func CollectAllMetricsJSON(ctx context.Context) (string, error) {
	// Collect all metrics
	metrics, err := CollectAllMetrics(ctx)
	// Check if collection failed
	if err != nil {
		// Return empty string with error
		return "", err
	}

	// Encode metrics to JSON
	jsonBytes, err := json.Marshal(metrics)
	// Check if encoding failed
	if err != nil {
		// Return empty string with error
		return "", err
	}

	// Return the JSON string
	return string(jsonBytes), nil
}
