//go:build cgo

package probe

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
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
type CPUMetricsJSON struct {
	UsagePercent float64          `json:"usage_percent"`
	Cores        uint32           `json:"cores"`
	FrequencyMHz uint64           `json:"frequency_mhz"`
	Pressure     *CPUPressureJSON `json:"pressure,omitempty"`
}

// CPUPressureJSON contains PSI pressure metrics for CPU.
type CPUPressureJSON struct {
	SomeAvg10   float64 `json:"some_avg10"`
	SomeAvg60   float64 `json:"some_avg60"`
	SomeAvg300  float64 `json:"some_avg300"`
	SomeTotalUs uint64  `json:"some_total_us"`
}

// MemoryMetricsJSON contains memory-related metrics for JSON output.
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
type LoadMetricsJSON struct {
	Load1Min  float64 `json:"load_1min"`
	Load5Min  float64 `json:"load_5min"`
	Load15Min float64 `json:"load_15min"`
}

// DiskMetricsJSON contains disk-related metrics for JSON output.
// Uses types from collect_all.go: PartitionInfo, DiskUsageInfo, DiskIOInfo
type DiskMetricsJSON struct {
	Partitions []PartitionInfo `json:"partitions,omitempty"`
	Usage      []DiskUsageInfo `json:"usage,omitempty"`
	IO         []DiskIOInfo    `json:"io,omitempty"`
}

// NetworkMetricsJSON contains network-related metrics for JSON output.
// Uses types from collect_all.go: NetInterfaceInfo, NetStatsInfo
type NetworkMetricsJSON struct {
	Interfaces []NetInterfaceJSON `json:"interfaces,omitempty"`
	Stats      []NetStatsJSON     `json:"stats,omitempty"`
}

// NetInterfaceJSON contains information about a network interface.
type NetInterfaceJSON struct {
	Name       string   `json:"name"`
	MACAddress string   `json:"mac_address"`
	MTU        uint32   `json:"mtu"`
	IsUp       bool     `json:"is_up"`
	IsLoopback bool     `json:"is_loopback"`
	Flags      []string `json:"flags,omitempty"`
}

// NetStatsJSON contains network statistics for an interface.
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
type IOMetricsJSON struct {
	ReadOps    uint64          `json:"read_ops"`
	ReadBytes  uint64          `json:"read_bytes"`
	WriteOps   uint64          `json:"write_ops"`
	WriteBytes uint64          `json:"write_bytes"`
	Pressure   *IOPressureJSON `json:"pressure,omitempty"`
}

// IOPressureJSON contains PSI pressure metrics for I/O.
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
type ProcessMetricsJSON struct {
	CurrentPID   int32             `json:"current_pid"`
	ProcessCount int               `json:"process_count"`
	TopProcesses []ProcessInfoJSON `json:"top_processes,omitempty"`
}

// ProcessInfoJSON contains information about a process.
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
type ThermalMetricsJSON struct {
	Supported bool              `json:"supported"`
	Zones     []ThermalZoneJSON `json:"zones,omitempty"`
}

// ThermalZoneJSON contains information about a thermal zone.
type ThermalZoneJSON struct {
	Name        string   `json:"name"`
	Label       string   `json:"label"`
	TempCelsius float64  `json:"temp_celsius"`
	TempMax     *float64 `json:"temp_max,omitempty"`
	TempCrit    *float64 `json:"temp_crit,omitempty"`
}

// ContextSwitchMetricsJSON contains context switch statistics.
type ContextSwitchMetricsJSON struct {
	SystemTotal uint64                 `json:"system_total"`
	Self        *ContextSwitchInfoJSON `json:"self,omitempty"`
}

// ContextSwitchInfoJSON contains context switch counts for a process.
type ContextSwitchInfoJSON struct {
	Voluntary   uint64 `json:"voluntary"`
	Involuntary uint64 `json:"involuntary"`
}

// ConnectionMetricsJSON contains network connection information.
type ConnectionMetricsJSON struct {
	TCPStats       *TcpStatsJSON    `json:"tcp_stats,omitempty"`
	TCPConnections []TcpConnJSON    `json:"tcp_connections,omitempty"`
	UDPSockets     []UdpConnJSON    `json:"udp_sockets,omitempty"`
	UnixSockets    []UnixSockJSON   `json:"unix_sockets,omitempty"`
	ListeningPorts []ListenInfoJSON `json:"listening_ports,omitempty"`
}

// TcpStatsJSON contains aggregated TCP statistics.
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
type UnixSockJSON struct {
	Path        string `json:"path"`
	Type        string `json:"type"`
	State       string `json:"state"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}

// ListenInfoJSON contains information about a listening port.
type ListenInfoJSON struct {
	Protocol    string `json:"protocol"`
	Address     string `json:"address"`
	Port        uint16 `json:"port"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}

// QuotaMetricsJSON contains resource quota information.
type QuotaMetricsJSON struct {
	Supported bool           `json:"supported"`
	Limits    *QuotaInfoJSON `json:"limits,omitempty"`
	Usage     *UsageInfoJSON `json:"usage,omitempty"`
}

// QuotaInfoJSON contains resource limit information.
type QuotaInfoJSON struct {
	CPUQuotaUs       uint64 `json:"cpu_quota_us,omitempty"`
	CPUPeriodUs      uint64 `json:"cpu_period_us,omitempty"`
	MemoryLimitBytes uint64 `json:"memory_limit_bytes,omitempty"`
	PidsLimit        uint64 `json:"pids_limit,omitempty"`
	NofileLimit      uint64 `json:"nofile_limit,omitempty"`
}

// UsageInfoJSON contains current resource usage.
type UsageInfoJSON struct {
	MemoryBytes      uint64  `json:"memory_bytes"`
	MemoryLimitBytes uint64  `json:"memory_limit_bytes,omitempty"`
	PidsCurrent      uint64  `json:"pids_current"`
	PidsLimit        uint64  `json:"pids_limit,omitempty"`
	CPUPercent       float64 `json:"cpu_percent"`
	CPULimitPercent  float64 `json:"cpu_limit_percent,omitempty"`
}

// ContainerMetricsJSON contains container detection information.
type ContainerMetricsJSON struct {
	IsContainerized bool   `json:"is_containerized"`
	Runtime         string `json:"runtime,omitempty"`
	ContainerID     string `json:"container_id,omitempty"`
}

// RuntimeMetricsJSON contains full runtime detection information.
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
type AvailableRuntimeInfoJSON struct {
	Runtime    string `json:"runtime"`
	SocketPath string `json:"socket_path,omitempty"`
	Version    string `json:"version,omitempty"`
	IsRunning  bool   `json:"is_running"`
}

// CollectAllMetrics collects all available metrics for the current platform.
// Returns a comprehensive snapshot of system metrics as JSON-serializable struct.
func CollectAllMetrics(ctx context.Context) (*AllSystemMetrics, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	now := time.Now()
	hostname, _ := os.Hostname()
	result := &AllSystemMetrics{
		Timestamp:   now,
		Platform:    runtime.GOOS,
		Hostname:    hostname,
		CollectedAt: now.UnixNano(),
	}

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

	return result, nil
}

// collectDiskMetricsJSON collects all disk-related metrics.
func collectDiskMetricsJSON(ctx context.Context, c *Collector) *DiskMetricsJSON {
	disk := &DiskMetricsJSON{}

	if partitions, err := c.Disk().ListPartitions(ctx); err == nil {
		disk.Partitions = make([]PartitionInfo, len(partitions))
		for i, p := range partitions {
			disk.Partitions[i] = PartitionInfo{
				Device:     p.Device,
				MountPoint: p.Mountpoint,
				FSType:     p.FSType,
				Options:    joinOptions(p.Options),
			}
		}
	}

	if usage, err := c.Disk().CollectAllUsage(ctx); err == nil {
		disk.Usage = make([]DiskUsageInfo, len(usage))
		for i, u := range usage {
			disk.Usage[i] = DiskUsageInfo{
				Path:        u.Path,
				TotalBytes:  u.Total,
				UsedBytes:   u.Used,
				FreeBytes:   u.Free,
				UsedPercent: u.UsagePercent,
				InodesTotal: u.InodesTotal,
				InodesUsed:  u.InodesUsed,
				InodesFree:  u.InodesFree,
			}
		}
	}

	if io, err := c.Disk().CollectIO(ctx); err == nil {
		disk.IO = make([]DiskIOInfo, len(io))
		for i, d := range io {
			disk.IO[i] = DiskIOInfo{
				Device:           d.Device,
				ReadsCompleted:   d.ReadCount,
				SectorsRead:      d.ReadBytes / 512, // Convert bytes to sectors
				ReadTimeMs:       uint64(d.ReadTime.Milliseconds()),
				WritesCompleted:  d.WriteCount,
				SectorsWritten:   d.WriteBytes / 512, // Convert bytes to sectors
				WriteTimeMs:      uint64(d.WriteTime.Milliseconds()),
				IOInProgress:     d.IOInProgress,
				IOTimeMs:         uint64(d.IOTime.Milliseconds()),
				WeightedIOTimeMs: uint64(d.WeightedIOTime.Milliseconds()),
			}
		}
	}

	return disk
}

// joinOptions joins slice of options into a comma-separated string.
func joinOptions(opts []string) string {
	if len(opts) == 0 {
		return ""
	}
	result := opts[0]
	for i := 1; i < len(opts); i++ {
		result += "," + opts[i]
	}
	return result
}

// containsFlag checks if a flag is present in the flags slice.
func containsFlag(flags []string, flag string) bool {
	for _, f := range flags {
		if f == flag {
			return true
		}
	}
	return false
}

// collectNetworkMetricsJSON collects all network-related metrics.
func collectNetworkMetricsJSON(ctx context.Context, c *Collector) *NetworkMetricsJSON {
	network := &NetworkMetricsJSON{}

	if ifaces, err := c.Network().ListInterfaces(ctx); err == nil {
		network.Interfaces = make([]NetInterfaceJSON, len(ifaces))
		for i, iface := range ifaces {
			// Derive IsUp and IsLoopback from flags
			isUp := containsFlag(iface.Flags, "up")
			isLoopback := containsFlag(iface.Flags, "loopback")
			network.Interfaces[i] = NetInterfaceJSON{
				Name:       iface.Name,
				MACAddress: iface.HardwareAddr,
				MTU:        uint32(iface.MTU),
				IsUp:       isUp,
				IsLoopback: isLoopback,
				Flags:      iface.Flags,
			}
		}
	}

	if stats, err := c.Network().CollectAllStats(ctx); err == nil {
		network.Stats = make([]NetStatsJSON, len(stats))
		for i, s := range stats {
			network.Stats[i] = NetStatsJSON{
				Interface:   s.Interface,
				BytesRecv:   s.BytesRecv,
				BytesSent:   s.BytesSent,
				PacketsRecv: s.PacketsRecv,
				PacketsSent: s.PacketsSent,
				ErrorsIn:    s.ErrorsIn,
				ErrorsOut:   s.ErrorsOut,
				DropsIn:     s.DropsIn,
				DropsOut:    s.DropsOut,
			}
		}
	}

	return network
}

// collectIOMetricsJSON collects all I/O-related metrics.
func collectIOMetricsJSON(ctx context.Context, c *Collector) *IOMetricsJSON {
	io := &IOMetricsJSON{}

	if stats, err := c.IO().CollectStats(ctx); err == nil {
		io.ReadOps = stats.ReadOpsTotal
		io.ReadBytes = stats.ReadBytesTotal
		io.WriteOps = stats.WriteOpsTotal
		io.WriteBytes = stats.WriteBytesTotal
	}

	if pressure, err := c.IO().CollectPressure(ctx); err == nil {
		io.Pressure = &IOPressureJSON{
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

	return io
}

// collectProcessMetricsJSON collects process-related metrics.
func collectProcessMetricsJSON(ctx context.Context) *ProcessMetricsJSON {
	pid := os.Getpid()
	pm := &ProcessMetricsJSON{
		CurrentPID: int32(pid),
	}

	// Get current process info using ProcessCollector
	pc := NewProcessCollector()
	cpuInfo, cpuErr := pc.CollectCPU(ctx, pid)
	memInfo, memErr := pc.CollectMemory(ctx, pid)

	if cpuErr == nil && memErr == nil {
		pm.TopProcesses = []ProcessInfoJSON{{
			PID:            int32(cpuInfo.PID),
			CPUPercent:     cpuInfo.UsagePercent,
			MemoryRSSBytes: memInfo.RSS,
			MemoryVMSBytes: memInfo.VMS,
			MemoryPercent:  memInfo.UsagePercent,
		}}
	}

	return pm
}

// collectThermalMetricsJSON collects thermal sensor metrics (Linux only).
func collectThermalMetricsJSON() *ThermalMetricsJSON {
	thermal := &ThermalMetricsJSON{
		Supported: ThermalIsSupported(),
	}

	if !thermal.Supported {
		return thermal
	}

	if zones, err := CollectThermalZones(); err == nil {
		thermal.Zones = make([]ThermalZoneJSON, len(zones))
		for i, z := range zones {
			info := ThermalZoneJSON{
				Name:        z.Name,
				Label:       z.Label,
				TempCelsius: z.TempCelsius,
				TempMax:     z.TempMax,
				TempCrit:    z.TempCrit,
			}
			thermal.Zones[i] = info
		}
	}

	return thermal
}

// collectContextSwitchMetricsJSON collects context switch metrics (Linux only).
func collectContextSwitchMetricsJSON() *ContextSwitchMetricsJSON {
	cs := &ContextSwitchMetricsJSON{}

	if total, err := CollectSystemContextSwitches(); err == nil {
		cs.SystemTotal = total
	}

	if self, err := CollectSelfContextSwitches(); err == nil {
		cs.Self = &ContextSwitchInfoJSON{
			Voluntary:   self.Voluntary,
			Involuntary: self.Involuntary,
		}
	}

	return cs
}

// collectConnectionMetricsJSON collects network connection metrics (Linux only).
func collectConnectionMetricsJSON(ctx context.Context) *ConnectionMetricsJSON {
	conn := &ConnectionMetricsJSON{}
	connCollector := NewConnectionCollector()

	// TCP Stats
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

	// TCP Connections
	if tcpConns, err := connCollector.CollectTCP(ctx); err == nil {
		conn.TCPConnections = make([]TcpConnJSON, len(tcpConns))
		for i, c := range tcpConns {
			conn.TCPConnections[i] = TcpConnJSON{
				Family:      c.Family.String(),
				LocalAddr:   c.LocalAddr,
				LocalPort:   c.LocalPort,
				RemoteAddr:  c.RemoteAddr,
				RemotePort:  c.RemotePort,
				State:       c.State.String(),
				PID:         c.PID,
				ProcessName: c.ProcessName,
			}
		}
	}

	// UDP Sockets
	if udpConns, err := connCollector.CollectUDP(ctx); err == nil {
		conn.UDPSockets = make([]UdpConnJSON, len(udpConns))
		for i, c := range udpConns {
			conn.UDPSockets[i] = UdpConnJSON{
				Family:      c.Family.String(),
				LocalAddr:   c.LocalAddr,
				LocalPort:   c.LocalPort,
				RemoteAddr:  c.RemoteAddr,
				RemotePort:  c.RemotePort,
				PID:         c.PID,
				ProcessName: c.ProcessName,
			}
		}
	}

	// Unix Sockets
	if unixSocks, err := connCollector.CollectUnix(ctx); err == nil {
		conn.UnixSockets = make([]UnixSockJSON, len(unixSocks))
		for i, s := range unixSocks {
			conn.UnixSockets[i] = UnixSockJSON{
				Path:        s.Path,
				Type:        s.SocketType,
				State:       s.State.String(),
				PID:         s.PID,
				ProcessName: s.ProcessName,
			}
		}
	}

	// Listening Ports
	if listening, err := connCollector.CollectListeningPorts(ctx); err == nil {
		conn.ListeningPorts = make([]ListenInfoJSON, len(listening))
		for i, l := range listening {
			conn.ListeningPorts[i] = ListenInfoJSON{
				Protocol:    "tcp",
				Address:     l.LocalAddr,
				Port:        l.LocalPort,
				PID:         l.PID,
				ProcessName: l.ProcessName,
			}
		}
	}

	return conn
}

// collectQuotaMetricsJSON collects resource quota metrics.
func collectQuotaMetricsJSON() *QuotaMetricsJSON {
	quota := &QuotaMetricsJSON{
		Supported: true, // Probe is supported if we got this far
	}

	pid := os.Getpid()

	if limits, err := ReadQuotaLimits(pid); err == nil {
		quota.Limits = &QuotaInfoJSON{
			CPUQuotaUs:       limits.CPUQuotaUS,
			CPUPeriodUs:      limits.CPUPeriodUS,
			MemoryLimitBytes: limits.MemoryLimitBytes,
			PidsLimit:        limits.PIDsLimit,
			NofileLimit:      limits.NofileLimit,
		}
	} else {
		quota.Supported = false
	}

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

	return quota
}

// collectContainerMetricsJSON collects container detection information.
func collectContainerMetricsJSON() *ContainerMetricsJSON {
	info, err := DetectContainer()
	if err != nil {
		return &ContainerMetricsJSON{IsContainerized: false}
	}

	return &ContainerMetricsJSON{
		IsContainerized: info.IsContainerized,
		Runtime:         info.Runtime.String(),
		ContainerID:     info.ContainerID,
	}
}

// collectRuntimeMetricsJSON collects full runtime detection information.
func collectRuntimeMetricsJSON() *RuntimeMetricsJSON {
	info, err := DetectRuntime()
	if err != nil {
		return &RuntimeMetricsJSON{IsContainerized: false}
	}

	rm := &RuntimeMetricsJSON{
		IsContainerized:  info.IsContainerized,
		ContainerRuntime: info.ContainerRuntime.String(),
		Orchestrator:     info.Orchestrator.String(),
		ContainerID:      info.ContainerID,
		WorkloadID:       info.WorkloadID,
		WorkloadName:     info.WorkloadName,
		Namespace:        info.Namespace,
	}

	if len(info.AvailableRuntimes) > 0 {
		rm.AvailableRuntimes = make([]AvailableRuntimeInfoJSON, len(info.AvailableRuntimes))
		for i, ar := range info.AvailableRuntimes {
			rm.AvailableRuntimes[i] = AvailableRuntimeInfoJSON{
				Runtime:    ar.Runtime.String(),
				SocketPath: ar.SocketPath,
				Version:    ar.Version,
				IsRunning:  ar.IsRunning,
			}
		}
	}

	return rm
}

// CollectAllMetricsJSON collects all metrics and returns them as a JSON string.
func CollectAllMetricsJSON(ctx context.Context) (string, error) {
	metrics, err := CollectAllMetrics(ctx)
	if err != nil {
		return "", err
	}

	jsonBytes, err := json.Marshal(metrics)
	if err != nil {
		return "", err
	}

	return string(jsonBytes), nil
}
