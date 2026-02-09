//go:build cgo

// doc://
// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

import (
	"context"
	"encoding/json"
	"os"
	"runtime"
	"slices"
	"strings"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
)

// CollectAllMetrics collects all available metrics for the current platform.
// Returns a comprehensive snapshot of system metrics as JSON-serializable struct.
//
// Params:
//   - ctx: context for cancellation
//   - cfg: metrics configuration controlling which metrics to collect
//
// Returns:
//   - *AllSystemMetrics: collected system metrics
//   - error: nil on success, error if probe not initialized
func CollectAllMetrics(ctx context.Context, cfg *config.MetricsConfig) (*AllSystemMetrics, error) {
	// check if probe library is initialized
	if err := checkInitialized(); err != nil {
		// return error if not initialized
		return nil, err
	}

	// check global enabled flag
	if cfg == nil || !cfg.Enabled {
		// return minimal result with only metadata
		return &AllSystemMetrics{
			Timestamp:   time.Now(),
			Platform:    runtime.GOOS,
			Hostname:    getHostname(),
			CollectedAt: time.Now().UnixNano(),
		}, nil
	}

	now := time.Now()
	hostname := getHostname()
	result := &AllSystemMetrics{
		Timestamp:   now,
		Platform:    runtime.GOOS,
		Hostname:    hostname,
		CollectedAt: now.UnixNano(),
	}

	collector := NewCollector()
	collectBasicMetrics(ctx, collector, result, cfg)
	collectResourceMetrics(ctx, collector, result, cfg)
	collectSystemMetrics(ctx, result, cfg)

	// return collected metrics
	return result, nil
}

// getHostname returns the system hostname, empty string on error.
//
// Returns:
//   - string: system hostname
func getHostname() string {
	hostname, _ := os.Hostname()
	// return hostname or empty string on error
	return hostname
}

// collectBasicMetrics collects CPU, memory, and load metrics.
//
// Params:
//   - ctx: context for cancellation
//   - collector: collector instance
//   - result: result structure to populate
//   - cfg: metrics configuration
func collectBasicMetrics(ctx context.Context, collector *Collector, result *AllSystemMetrics, cfg *config.MetricsConfig) {
	// collect CPU metrics if enabled
	if cfg.CPU.Enabled {
		result.CPU = collectCPUMetricsWithPressure(ctx, collector, cfg.CPU.Pressure)
	}
	// collect memory metrics if enabled
	if cfg.Memory.Enabled {
		result.Memory = collectMemoryMetricsWithPressure(ctx, collector, cfg.Memory.Pressure)
	}
	// collect load metrics if enabled
	if cfg.Load.Enabled {
		result.Load = collectLoadMetricsJSON(ctx, collector)
	}
}

// collectCPUMetricsWithPressure collects CPU metrics including pressure.
//
// Params:
//   - ctx: context for cancellation.
//   - collector: collector instance.
//   - collectPressure: whether to collect pressure stall information.
//
// Returns:
//   - *CPUMetricsJSON: collected CPU metrics, nil on error.
func collectCPUMetricsWithPressure(ctx context.Context, collector *Collector, collectPressure bool) *CPUMetricsJSON {
	cpu, err := collector.Cpu().CollectSystem(ctx)
	// check for CPU collection error
	if err != nil {
		// return nil on error
		return nil
	}

	cpuMetrics := &CPUMetricsJSON{
		UsagePercent: cpu.UsagePercent,
		Cores:        uint32(runtime.NumCPU()),
	}

	// add CPU pressure if enabled and available
	if collectPressure {
		// attempt to collect CPU pressure information
		if pressure, err := collector.Cpu().CollectPressure(ctx); err == nil {
			cpuMetrics.Pressure = &CPUPressureJSON{
				SomeAvg10:   pressure.SomeAvg10,
				SomeAvg60:   pressure.SomeAvg60,
				SomeAvg300:  pressure.SomeAvg300,
				SomeTotalUs: pressure.SomeTotal,
			}
		}
	}

	// return collected CPU metrics with optional pressure data
	return cpuMetrics
}

// collectMemoryMetricsWithPressure collects memory metrics including pressure.
//
// Params:
//   - ctx: context for cancellation.
//   - collector: collector instance.
//   - collectPressure: whether to collect pressure stall information.
//
// Returns:
//   - *MemoryMetricsJSON: collected memory metrics, nil on error.
func collectMemoryMetricsWithPressure(ctx context.Context, collector *Collector, collectPressure bool) *MemoryMetricsJSON {
	mem, err := collector.Memory().CollectSystem(ctx)
	// check for memory collection error
	if err != nil {
		// return nil on error
		return nil
	}

	memMetrics := &MemoryMetricsJSON{
		TotalBytes:     mem.Total,
		AvailableBytes: mem.Available,
		UsedBytes:      mem.Used,
		CachedBytes:    mem.Cached,
		BuffersBytes:   mem.Buffers,
		SwapTotalBytes: mem.SwapTotal,
		SwapUsedBytes:  mem.SwapUsed,
		UsedPercent:    mem.UsagePercent,
	}

	// add memory pressure if enabled and available
	if collectPressure {
		// attempt to collect memory pressure information
		if pressure, err := collector.Memory().CollectPressure(ctx); err == nil {
			memMetrics.Pressure = &MemoryPressureJSON{
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

	// return collected memory metrics with optional pressure data
	return memMetrics
}

// collectLoadMetricsJSON collects load average metrics.
//
// Params:
//   - ctx: context for cancellation.
//   - collector: collector instance.
//
// Returns:
//   - *LoadMetricsJSON: collected load average metrics, nil on error.
func collectLoadMetricsJSON(ctx context.Context, collector *Collector) *LoadMetricsJSON {
	load, err := collector.cpu.CollectLoadAverage(ctx)
	// check for load average collection error
	if err != nil {
		// return nil on error
		return nil
	}
	// return collected load average metrics
	return &LoadMetricsJSON{
		Load1Min:  load.Load1,
		Load5Min:  load.Load5,
		Load15Min: load.Load15,
	}
}

// collectResourceMetrics collects disk, network, and I/O metrics.
//
// Params:
//   - ctx: context for cancellation
//   - collector: collector instance
//   - result: result structure to populate
//   - cfg: metrics configuration
func collectResourceMetrics(ctx context.Context, collector *Collector, result *AllSystemMetrics, cfg *config.MetricsConfig) {
	// collect disk metrics if enabled
	if cfg.Disk.Enabled {
		result.Disk = collectDiskMetricsJSON(ctx, collector, &cfg.Disk)
	}
	// collect network metrics if enabled
	if cfg.Network.Enabled {
		result.Network = collectNetworkMetricsJSON(ctx, collector, &cfg.Network)
	}
	// collect I/O metrics if enabled
	if cfg.IO.Enabled {
		result.IO = collectIOMetricsJSON(ctx, collector, cfg.IO.Pressure)
	}
}

// collectSystemMetrics collects process, thermal, and connection metrics.
//
// Params:
//   - ctx: context for cancellation
//   - result: result structure to populate
//   - cfg: metrics configuration
func collectSystemMetrics(ctx context.Context, result *AllSystemMetrics, cfg *config.MetricsConfig) {
	// collect process metrics if enabled
	if cfg.Process.Enabled {
		result.Process = collectProcessMetricsJSON(ctx)
	}
	// collect thermal metrics if enabled
	if cfg.Thermal.Enabled {
		result.Thermal = collectThermalMetricsJSON()
	}
	// collect context switch metrics (always enabled, minimal overhead)
	result.ContextSwitches = collectContextSwitchMetricsJSON()
	// collect connection metrics if enabled
	if cfg.Connections.Enabled {
		result.Connections = collectConnectionMetricsJSON(ctx, &cfg.Connections)
	}
	// collect quota metrics if enabled
	if cfg.Quota.Enabled {
		result.Quota = collectQuotaMetricsJSON()
	}
	// collect container metrics if enabled
	if cfg.Container.Enabled {
		result.Container = collectContainerMetricsJSON()
	}
	// collect runtime metrics if enabled
	if cfg.Runtime.Enabled {
		result.Runtime = collectRuntimeMetricsJSON()
	}
}

// collectDiskMetricsJSON collects all disk-related metrics.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance to use
//   - cfg: disk metrics configuration
//
// Returns:
//   - *DiskMetricsJSON: collected disk metrics
func collectDiskMetricsJSON(ctx context.Context, coll *Collector, cfg *config.DiskMetricsConfig) *DiskMetricsJSON {
	disk := &DiskMetricsJSON{}
	// collect partitions if enabled
	if cfg.Partitions {
		disk.Partitions = extractPartitionInfo(ctx, coll)
	}
	// collect usage if enabled
	if cfg.Usage {
		disk.Usage = extractDiskUsageInfo(ctx, coll)
	}
	// collect I/O if enabled
	if cfg.IO {
		disk.IO = extractDiskIOInfo(ctx, coll)
	}
	// return collected disk metrics
	return disk
}

// extractPartitionInfo extracts partition information.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance
//
// Returns:
//   - []PartitionInfo: partition information
func extractPartitionInfo(ctx context.Context, coll *Collector) []PartitionInfo {
	partitions, err := coll.Disk().ListPartitions(ctx)
	// check for partition listing error
	if err != nil {
		// return nil on error
		return nil
	}

	// preallocate slice with capacity
	result := make([]PartitionInfo, 0, len(partitions))
	// iterate through each partition
	for _, pt := range partitions {
		result = append(result, PartitionInfo{
			Device:     pt.Device,
			MountPoint: pt.Mountpoint,
			FSType:     pt.FSType,
			Options:    joinOptions(pt.Options),
		})
	}
	// return extracted partition info
	return result
}

// extractDiskUsageInfo extracts disk usage information.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance
//
// Returns:
//   - []DiskUsageInfo: disk usage information
func extractDiskUsageInfo(ctx context.Context, coll *Collector) []DiskUsageInfo {
	usage, err := coll.Disk().CollectAllUsage(ctx)
	// check for usage collection error
	if err != nil {
		// return nil on error
		return nil
	}

	// preallocate slice with capacity
	result := make([]DiskUsageInfo, 0, len(usage))
	// iterate through each usage entry
	for _, us := range usage {
		result = append(result, DiskUsageInfo{
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
	// return extracted disk usage info
	return result
}

// extractDiskIOInfo extracts disk I/O information.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance
//
// Returns:
//   - []DiskIOInfo: disk I/O information
func extractDiskIOInfo(ctx context.Context, coll *Collector) []DiskIOInfo {
	ioStats, err := coll.Disk().CollectIO(ctx)
	// check for I/O collection error
	if err != nil {
		// return nil on error
		return nil
	}

	// preallocate slice with capacity
	result := make([]DiskIOInfo, 0, len(ioStats))
	// iterate through each I/O stats entry
	for _, io := range ioStats {
		result = append(result, DiskIOInfo{
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
	// return extracted disk I/O info
	return result
}

// joinOptions joins slice of options into a comma-separated string.
//
// Params:
//   - opts: slice of option strings to join
//
// Returns:
//   - string: comma-separated options string
func joinOptions(opts []string) string {
	// return empty string for empty slice
	if len(opts) == 0 {
		// return empty string
		return ""
	}
	// use strings.Join for efficient string concatenation
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
	// use slices.Contains for efficient membership check
	return slices.Contains(flags, flag)
}

// collectNetworkMetricsJSON collects all network-related metrics.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance to use
//   - cfg: network metrics configuration
//
// Returns:
//   - *NetworkMetricsJSON: collected network metrics
func collectNetworkMetricsJSON(ctx context.Context, coll *Collector, cfg *config.NetworkMetricsConfig) *NetworkMetricsJSON {
	// initialize network metrics struct
	network := &NetworkMetricsJSON{}

	// collect interface information if enabled and collection succeeds
	if cfg.Interfaces {
		// attempt to collect network interfaces
		if ifaces, err := coll.Network().ListInterfaces(ctx); err == nil {
			// preallocate slice with capacity
			network.Interfaces = make([]NetInterfaceJSON, 0, len(ifaces))
			// convert each interface to JSON format.
			for _, iface := range ifaces {
				network.Interfaces = append(network.Interfaces, NetInterfaceJSON{
					Name:       iface.Name,
					MACAddress: iface.HardwareAddr,
					MTU:        uint32(iface.MTU),
					IsUp:       containsFlag(iface.Flags, "up"),
					IsLoopback: containsFlag(iface.Flags, "loopback"),
					Flags:      iface.Flags,
				})
			}
		}
	}

	// collect network statistics if enabled and collection succeeds
	if cfg.Stats {
		// attempt to collect network statistics
		if stats, err := coll.Network().CollectAllStats(ctx); err == nil {
			// preallocate slice with capacity
			network.Stats = make([]NetStatsJSON, 0, len(stats))
			// iterate over each stats entry
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
	}

	// return the collected network metrics
	return network
}

// collectIOMetricsJSON collects all I/O-related metrics.
//
// Params:
//   - ctx: context for cancellation
//   - coll: collector instance to use
//   - collectPressure: whether to collect pressure stall information
//
// Returns:
//   - *IOMetricsJSON: collected I/O metrics
func collectIOMetricsJSON(ctx context.Context, coll *Collector, collectPressure bool) *IOMetricsJSON {
	// initialize I/O metrics struct
	ioMetrics := &IOMetricsJSON{}

	// collect I/O statistics if collection succeeds
	if stats, err := coll.Io().CollectStats(ctx); err == nil {
		ioMetrics.ReadOps = stats.ReadOpsTotal
		ioMetrics.ReadBytes = stats.ReadBytesTotal
		ioMetrics.WriteOps = stats.WriteOpsTotal
		ioMetrics.WriteBytes = stats.WriteBytesTotal
	}

	// collect I/O pressure if enabled and available (Linux only)
	if collectPressure {
		// attempt to collect I/O pressure information
		if pressure, err := coll.Io().CollectPressure(ctx); err == nil {
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
	}

	// return the collected I/O metrics
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
	// get current process ID
	pid := os.Getpid()
	pm := &ProcessMetricsJSON{
		CurrentPID: int32(pid),
	}

	// get current process info using ProcessCollector
	pc := NewProcessCollector()
	cpuInfo, cpuErr := pc.CollectCPU(ctx, pid)
	memInfo, memErr := pc.CollectMemory(ctx, pid)
	fdsInfo, fdsErr := pc.CollectFDs(ctx, pid)
	ioInfo, ioErr := pc.CollectIO(ctx, pid)

	// check if CPU and memory collections succeeded
	if cpuErr == nil && memErr == nil {
		processInfo := ProcessInfoJSON{
			PID:            int32(cpuInfo.PID),
			CPUPercent:     cpuInfo.UsagePercent,
			MemoryRSSBytes: memInfo.RSS,
			MemoryVMSBytes: memInfo.VMS,
			MemoryPercent:  memInfo.UsagePercent,
		}

		// add FD count if available
		if fdsErr == nil {
			processInfo.NumFDs = fdsInfo.Count
		}

		// add I/O stats if available
		if ioErr == nil {
			processInfo.ReadBytesPerSec = ioInfo.ReadBytesPerSec
			processInfo.WriteBytesPerSec = ioInfo.WriteBytesPerSec
		}

		pm.TopProcesses = []ProcessInfoJSON{processInfo}
	}

	// return the collected process metrics
	return pm
}

// collectThermalMetricsJSON collects thermal sensor metrics (Linux only).
//
// Returns:
//   - *ThermalMetricsJSON: collected thermal metrics
func collectThermalMetricsJSON() *ThermalMetricsJSON {
	// initialize thermal metrics struct
	thermal := &ThermalMetricsJSON{
		Supported: ThermalIsSupported(),
	}

	// return early if not supported
	if !thermal.Supported {
		// return unsupported thermal metrics
		return thermal
	}

	// collect thermal zones
	if zones, err := CollectThermalZones(); err == nil {
		// preallocate slice with capacity
		thermal.Zones = make([]ThermalZoneJSON, 0, len(zones))
		// iterate over each zone
		for _, zn := range zones {
			// ThermalZone and ThermalZoneJSON have identical underlying types
			thermal.Zones = append(thermal.Zones, ThermalZoneJSON(zn))
		}
	}

	// return the collected thermal metrics
	return thermal
}

// collectContextSwitchMetricsJSON collects context switch metrics (Linux only).
//
// Returns:
//   - *ContextSwitchMetricsJSON: collected context switch metrics
func collectContextSwitchMetricsJSON() *ContextSwitchMetricsJSON {
	// initialize context switch metrics struct
	cs := &ContextSwitchMetricsJSON{}

	// collect system-wide context switches
	if total, err := CollectSystemContextSwitches(); err == nil {
		cs.SystemTotal = total
	}

	// collect self context switches
	if self, err := CollectSelfContextSwitches(); err == nil {
		cs.Self = &ContextSwitchInfoJSON{
			Voluntary:   self.Voluntary,
			Involuntary: self.Involuntary,
		}
	}

	// return the collected context switch metrics
	return cs
}

// collectConnectionMetricsJSON collects network connection metrics (Linux only).
//
// Params:
//   - ctx: context for cancellation
//   - cfg: connection metrics configuration
//
// Returns:
//   - *ConnectionMetricsJSON: collected connection metrics
func collectConnectionMetricsJSON(ctx context.Context, cfg *config.ConnectionMetricsConfig) *ConnectionMetricsJSON {
	conn := &ConnectionMetricsJSON{}
	collector := NewConnectionCollector()

	// collect TCP stats if enabled
	if cfg.TCPStats {
		conn.TCPStats = collectTCPStatsJSON(ctx, collector)
	}
	// collect TCP connections if enabled
	if cfg.TCPConnections {
		conn.TCPConnections = collectTCPConnectionsJSON(ctx, collector)
	}
	// collect UDP sockets if enabled
	if cfg.UDPSockets {
		conn.UDPSockets = collectUDPSocketsJSON(ctx, collector)
	}
	// collect Unix sockets if enabled
	if cfg.UnixSockets {
		conn.UnixSockets = collectUnixSocketsJSON(ctx, collector)
	}
	// collect listening ports if enabled
	if cfg.ListeningPorts {
		conn.ListeningPorts = collectListeningPortsJSON(ctx, collector)
	}

	// return populated connection metrics
	return conn
}

// collectTCPStatsJSON collects TCP statistics.
//
// Params:
//   - ctx: context for cancellation
//   - collector: the connection collector
//
// Returns:
//   - *TcpStatsJSON: collected TCP stats or nil on error
func collectTCPStatsJSON(ctx context.Context, collector *ConnectionCollector) *TcpStatsJSON {
	stats, err := collector.CollectTCPStats(ctx)
	// return nil if collection failed
	if err != nil {
		// skip on error
		return nil
	}
	// return TCP stats with all fields
	return &TcpStatsJSON{
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

// collectTCPConnectionsJSON collects TCP connections using pooled slices.
//
// Params:
//   - ctx: context for cancellation
//   - collector: the connection collector
//
// Returns:
//   - []TcpConnJSON: collected TCP connections
func collectTCPConnectionsJSON(ctx context.Context, collector *ConnectionCollector) []TcpConnJSON {
	tcpConns, err := collector.CollectTCP(ctx)
	// return nil if collection failed
	if err != nil {
		// skip on error
		return nil
	}

	// get pooled slice
	resultPtr := getTCPConnSlice()
	result := *resultPtr

	// convert each TCP connection to JSON format
	for _, tc := range tcpConns {
		result = append(result, TcpConnJSON{
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

	// use slices.Clone() for efficient copying
	resultCopy := slices.Clone(result)

	// return pooled slice
	putTCPConnSlice(resultPtr)

	// return collected TCP connections
	return resultCopy
}

// collectUDPSocketsJSON collects UDP sockets using pooled slices.
//
// Params:
//   - ctx: context for cancellation
//   - collector: the connection collector
//
// Returns:
//   - []UdpConnJSON: collected UDP sockets
func collectUDPSocketsJSON(ctx context.Context, collector *ConnectionCollector) []UdpConnJSON {
	udpConns, err := collector.CollectUDP(ctx)
	// return nil if collection failed
	if err != nil {
		// skip on error
		return nil
	}

	// get pooled slice
	resultPtr := getUDPConnSlice()
	result := *resultPtr

	// convert each UDP socket to JSON format
	for _, uc := range udpConns {
		result = append(result, UdpConnJSON{
			Family:      uc.Family.String(),
			LocalAddr:   uc.LocalAddr,
			LocalPort:   uc.LocalPort,
			RemoteAddr:  uc.RemoteAddr,
			RemotePort:  uc.RemotePort,
			PID:         uc.PID,
			ProcessName: uc.ProcessName,
		})
	}

	// use slices.Clone() for efficient copying
	resultCopy := slices.Clone(result)

	// return pooled slice
	putUDPConnSlice(resultPtr)

	// return collected UDP sockets
	return resultCopy
}

// collectUnixSocketsJSON collects Unix sockets using pooled slices.
//
// Params:
//   - ctx: context for cancellation
//   - collector: the connection collector
//
// Returns:
//   - []UnixSockJSON: collected Unix sockets
func collectUnixSocketsJSON(ctx context.Context, collector *ConnectionCollector) []UnixSockJSON {
	unixSocks, err := collector.CollectUnix(ctx)
	// return nil if collection failed
	if err != nil {
		// skip on error
		return nil
	}

	// get pooled slice
	resultPtr := getUnixSockSlice()
	result := *resultPtr

	// convert each Unix socket to JSON format
	for _, us := range unixSocks {
		result = append(result, UnixSockJSON{
			Path:        us.Path,
			Type:        us.SocketType,
			State:       us.State.String(),
			PID:         us.PID,
			ProcessName: us.ProcessName,
		})
	}

	// use slices.Clone() for efficient copying
	resultCopy := slices.Clone(result)

	// return pooled slice
	putUnixSockSlice(resultPtr)

	// return collected Unix sockets
	return resultCopy
}

// collectListeningPortsJSON collects listening ports using pooled slices.
//
// Params:
//   - ctx: context for cancellation
//   - collector: the connection collector
//
// Returns:
//   - []ListenInfoJSON: collected listening ports
func collectListeningPortsJSON(ctx context.Context, collector *ConnectionCollector) []ListenInfoJSON {
	listening, err := collector.CollectListeningPorts(ctx)
	// return nil if collection failed
	if err != nil {
		// skip on error
		return nil
	}

	// get pooled slice
	resultPtr := getListenInfoSlice()
	result := *resultPtr

	// convert each listening port to JSON format
	for _, lp := range listening {
		result = append(result, ListenInfoJSON{
			Protocol:    "tcp",
			Address:     lp.LocalAddr,
			Port:        lp.LocalPort,
			PID:         lp.PID,
			ProcessName: lp.ProcessName,
		})
	}

	// use slices.Clone() for efficient copying
	resultCopy := slices.Clone(result)

	// return pooled slice
	putListenInfoSlice(resultPtr)

	// return collected listening ports
	return resultCopy
}

// collectQuotaMetricsJSON collects resource quota metrics.
//
// Returns:
//   - *QuotaMetricsJSON: collected quota metrics
func collectQuotaMetricsJSON() *QuotaMetricsJSON {
	// initialize quota metrics struct
	quota := &QuotaMetricsJSON{
		Supported: true, // Probe is supported if we got this far
	}

	// get current process ID
	pid := os.Getpid()

	// collect quota limits
	if limits, err := ReadQuotaLimits(pid); err == nil {
		quota.Limits = &QuotaInfoJSON{
			CPUQuotaUs:       limits.CPUQuotaUS,
			CPUPeriodUs:      limits.CPUPeriodUS,
			MemoryLimitBytes: limits.MemoryLimitBytes,
			PidsLimit:        limits.PIDsLimit,
			NofileLimit:      limits.NofileLimit,
		}
	} else {
		// quota limits unavailable, mark as unsupported
		quota.Supported = false
	}

	// collect quota usage
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

	// return the collected quota metrics
	return quota
}

// collectContainerMetricsJSON collects container detection information.
//
// Returns:
//   - *ContainerMetricsJSON: collected container metrics
func collectContainerMetricsJSON() *ContainerMetricsJSON {
	// detect container environment
	info, err := DetectContainer()
	// check if detection failed
	if err != nil {
		// return not containerized on error
		return &ContainerMetricsJSON{IsContainerized: false}
	}

	// return the collected container metrics
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
	// detect runtime environment
	info, err := DetectRuntime()
	// check if detection failed
	if err != nil {
		// return not containerized on error
		return &RuntimeMetricsJSON{IsContainerized: false}
	}

	// build runtime metrics struct
	rm := &RuntimeMetricsJSON{
		IsContainerized:  info.IsContainerized,
		ContainerRuntime: info.ContainerRuntime.String(),
		Orchestrator:     info.Orchestrator.String(),
		ContainerID:      info.ContainerID,
		WorkloadID:       info.WorkloadID,
		WorkloadName:     info.WorkloadName,
		Namespace:        info.Namespace,
	}

	// check if available runtimes exist
	if len(info.AvailableRuntimes) > 0 {
		// preallocate slice with capacity
		rm.AvailableRuntimes = make([]AvailableRuntimeInfoJSON, 0, len(info.AvailableRuntimes))
		// iterate over each available runtime
		for _, ar := range info.AvailableRuntimes {
			rm.AvailableRuntimes = append(rm.AvailableRuntimes, AvailableRuntimeInfoJSON{
				Runtime:    ar.Runtime.String(),
				SocketPath: ar.SocketPath,
				Version:    ar.Version,
				IsRunning:  ar.IsRunning,
			})
		}
	}

	// return the collected runtime metrics
	return rm
}

// CollectAllMetricsJSON collects all metrics and returns them as a JSON string.
// Uses pooled buffers to reduce allocations.
//
// Params:
//   - ctx: context for cancellation
//   - cfg: metrics configuration controlling which metrics to collect
//
// Returns:
//   - string: JSON-encoded metrics
//   - error: nil on success, error if collection or encoding fails
func CollectAllMetricsJSON(ctx context.Context, cfg *config.MetricsConfig) (string, error) {
	// collect all metrics with configuration
	metrics, err := CollectAllMetrics(ctx, cfg)
	// check if collection failed
	if err != nil {
		// return empty string with error
		return "", err
	}

	// get pooled buffer
	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

	// encode metrics to JSON using pooled buffer
	enc := json.NewEncoder(buf)
	// check if encoding failed
	if err := enc.Encode(metrics); err != nil {
		// return empty string with error
		return "", err
	}

	// return the JSON string (buf.String() makes a copy)
	return buf.String(), nil
}
