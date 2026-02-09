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
	// Check if probe library is initialized.
	if err := checkInitialized(); err != nil {
		// Return error if not initialized.
		return nil, err
	}

	// Check global enabled flag.
	if cfg == nil || !cfg.Enabled {
		// Return minimal result with only metadata.
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

	// Return collected metrics.
	return result, nil
}

// getHostname returns the system hostname, empty string on error.
//
// Returns:
//   - string: system hostname
func getHostname() string {
	hostname, _ := os.Hostname()
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
	// Collect CPU metrics if enabled.
	if cfg.CPU.Enabled {
		result.CPU = collectCPUMetricsWithPressure(ctx, collector, cfg.CPU.Pressure)
	}
	// Collect memory metrics if enabled.
	if cfg.Memory.Enabled {
		result.Memory = collectMemoryMetricsWithPressure(ctx, collector, cfg.Memory.Pressure)
	}
	// Collect load metrics if enabled.
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
	// Check for CPU collection error.
	if err != nil {
		// Return nil on error.
		return nil
	}

	cpuMetrics := &CPUMetricsJSON{
		UsagePercent: cpu.UsagePercent,
		Cores:        uint32(runtime.NumCPU()),
	}

	// Add CPU pressure if enabled and available.
	if collectPressure {
		if pressure, err := collector.Cpu().CollectPressure(ctx); err == nil {
			cpuMetrics.Pressure = &CPUPressureJSON{
				SomeAvg10:   pressure.SomeAvg10,
				SomeAvg60:   pressure.SomeAvg60,
				SomeAvg300:  pressure.SomeAvg300,
				SomeTotalUs: pressure.SomeTotal,
			}
		}
	}

	// Return collected CPU metrics with optional pressure data.
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
	// Check for memory collection error.
	if err != nil {
		// Return nil on error.
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

	// Add memory pressure if enabled and available.
	if collectPressure {
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

	// Return collected memory metrics with optional pressure data.
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
	// Check for load average collection error.
	if err != nil {
		// Return nil on error.
		return nil
	}
	// Return collected load average metrics.
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
	// Collect disk metrics if enabled.
	if cfg.Disk.Enabled {
		result.Disk = collectDiskMetricsJSON(ctx, collector, &cfg.Disk)
	}
	// Collect network metrics if enabled.
	if cfg.Network.Enabled {
		result.Network = collectNetworkMetricsJSON(ctx, collector, &cfg.Network)
	}
	// Collect I/O metrics if enabled.
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
	// Collect process metrics if enabled.
	if cfg.Process.Enabled {
		result.Process = collectProcessMetricsJSON(ctx)
	}
	// Collect thermal metrics if enabled.
	if cfg.Thermal.Enabled {
		result.Thermal = collectThermalMetricsJSON()
	}
	// Collect context switch metrics (always enabled, minimal overhead).
	result.ContextSwitches = collectContextSwitchMetricsJSON()
	// Collect connection metrics if enabled.
	if cfg.Connections.Enabled {
		result.Connections = collectConnectionMetricsJSON(ctx, &cfg.Connections)
	}
	// Collect quota metrics if enabled.
	if cfg.Quota.Enabled {
		result.Quota = collectQuotaMetricsJSON()
	}
	// Collect container metrics if enabled.
	if cfg.Container.Enabled {
		result.Container = collectContainerMetricsJSON()
	}
	// Collect runtime metrics if enabled.
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
	// Collect partitions if enabled.
	if cfg.Partitions {
		disk.Partitions = extractPartitionInfo(ctx, coll)
	}
	// Collect usage if enabled.
	if cfg.Usage {
		disk.Usage = extractDiskUsageInfo(ctx, coll)
	}
	// Collect I/O if enabled.
	if cfg.IO {
		disk.IO = extractDiskIOInfo(ctx, coll)
	}
	// Return collected disk metrics.
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
	// Check for partition listing error.
	if err != nil {
		// Return nil on error.
		return nil
	}

	result := make([]PartitionInfo, 0, len(partitions))
	// Iterate through each partition.
	for _, pt := range partitions {
		result = append(result, PartitionInfo{
			Device:     pt.Device,
			MountPoint: pt.Mountpoint,
			FSType:     pt.FSType,
			Options:    joinOptions(pt.Options),
		})
	}
	// Return extracted partition info.
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
	// Check for usage collection error.
	if err != nil {
		// Return nil on error.
		return nil
	}

	result := make([]DiskUsageInfo, 0, len(usage))
	// Iterate through each usage entry.
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
	// Return extracted disk usage info.
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
	// Check for I/O collection error.
	if err != nil {
		// Return nil on error.
		return nil
	}

	result := make([]DiskIOInfo, 0, len(ioStats))
	// Iterate through each I/O stats entry.
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
	// Return extracted disk I/O info.
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
//   - cfg: network metrics configuration
//
// Returns:
//   - *NetworkMetricsJSON: collected network metrics
func collectNetworkMetricsJSON(ctx context.Context, coll *Collector, cfg *config.NetworkMetricsConfig) *NetworkMetricsJSON {
	// Initialize network metrics struct
	network := &NetworkMetricsJSON{}

	// Collect interface information if enabled and collection succeeds.
	if cfg.Interfaces {
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
	}

	// Collect network statistics if enabled and collection succeeds.
	if cfg.Stats {
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
	}

	// Return the collected network metrics
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
	// Initialize I/O metrics struct
	ioMetrics := &IOMetricsJSON{}

	// Collect I/O statistics if collection succeeds.
	if stats, err := coll.Io().CollectStats(ctx); err == nil {
		ioMetrics.ReadOps = stats.ReadOpsTotal
		ioMetrics.ReadBytes = stats.ReadBytesTotal
		ioMetrics.WriteOps = stats.WriteOpsTotal
		ioMetrics.WriteBytes = stats.WriteBytesTotal
	}

	// Collect I/O pressure if enabled and available (Linux only).
	if collectPressure {
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
	fdsInfo, fdsErr := pc.CollectFDs(ctx, pid)
	ioInfo, ioErr := pc.CollectIO(ctx, pid)

	// Check if CPU and memory collections succeeded.
	if cpuErr == nil && memErr == nil {
		processInfo := ProcessInfoJSON{
			PID:            int32(cpuInfo.PID),
			CPUPercent:     cpuInfo.UsagePercent,
			MemoryRSSBytes: memInfo.RSS,
			MemoryVMSBytes: memInfo.VMS,
			MemoryPercent:  memInfo.UsagePercent,
		}

		// Add FD count if available
		if fdsErr == nil {
			processInfo.NumFDs = fdsInfo.Count
		}

		// Add I/O stats if available
		if ioErr == nil {
			processInfo.ReadBytesPerSec = ioInfo.ReadBytesPerSec
			processInfo.WriteBytesPerSec = ioInfo.WriteBytesPerSec
		}

		pm.TopProcesses = []ProcessInfoJSON{processInfo}
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
//   - cfg: connection metrics configuration
//
// Returns:
//   - *ConnectionMetricsJSON: collected connection metrics
func collectConnectionMetricsJSON(ctx context.Context, cfg *config.ConnectionMetricsConfig) *ConnectionMetricsJSON {
	conn := &ConnectionMetricsJSON{}
	collector := NewConnectionCollector()

	// Collect TCP stats if enabled.
	if cfg.TCPStats {
		conn.TCPStats = collectTCPStatsJSON(ctx, collector)
	}
	// Collect TCP connections if enabled.
	if cfg.TCPConnections {
		conn.TCPConnections = collectTCPConnectionsJSON(ctx, collector)
	}
	// Collect UDP sockets if enabled.
	if cfg.UDPSockets {
		conn.UDPSockets = collectUDPSocketsJSON(ctx, collector)
	}
	// Collect Unix sockets if enabled.
	if cfg.UnixSockets {
		conn.UnixSockets = collectUnixSocketsJSON(ctx, collector)
	}
	// Collect listening ports if enabled.
	if cfg.ListeningPorts {
		conn.ListeningPorts = collectListeningPortsJSON(ctx, collector)
	}

	// Return populated connection metrics.
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
	// Return nil if collection failed.
	if err != nil {
		// Skip on error.
		return nil
	}
	// Return TCP stats with all fields.
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
	// Return nil if collection failed.
	if err != nil {
		// Skip on error.
		return nil
	}

	// Get pooled slice
	resultPtr := getTCPConnSlice()
	result := *resultPtr

	// Convert each TCP connection to JSON format.
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

	// Make a copy for return (caller owns this)
	resultCopy := make([]TcpConnJSON, len(result))
	copy(resultCopy, result)

	// Return pooled slice
	putTCPConnSlice(resultPtr)

	// Return collected TCP connections.
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
	// Return nil if collection failed.
	if err != nil {
		// Skip on error.
		return nil
	}

	// Get pooled slice
	resultPtr := getUDPConnSlice()
	result := *resultPtr

	// Convert each UDP socket to JSON format.
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

	// Make a copy for return (caller owns this)
	resultCopy := make([]UdpConnJSON, len(result))
	copy(resultCopy, result)

	// Return pooled slice
	putUDPConnSlice(resultPtr)

	// Return collected UDP sockets.
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
	// Return nil if collection failed.
	if err != nil {
		// Skip on error.
		return nil
	}

	// Get pooled slice
	resultPtr := getUnixSockSlice()
	result := *resultPtr

	// Convert each Unix socket to JSON format.
	for _, us := range unixSocks {
		result = append(result, UnixSockJSON{
			Path:        us.Path,
			Type:        us.SocketType,
			State:       us.State.String(),
			PID:         us.PID,
			ProcessName: us.ProcessName,
		})
	}

	// Make a copy for return (caller owns this)
	resultCopy := make([]UnixSockJSON, len(result))
	copy(resultCopy, result)

	// Return pooled slice
	putUnixSockSlice(resultPtr)

	// Return collected Unix sockets.
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
	// Return nil if collection failed.
	if err != nil {
		// Skip on error.
		return nil
	}

	// Get pooled slice
	resultPtr := getListenInfoSlice()
	result := *resultPtr

	// Convert each listening port to JSON format.
	for _, lp := range listening {
		result = append(result, ListenInfoJSON{
			Protocol:    "tcp",
			Address:     lp.LocalAddr,
			Port:        lp.LocalPort,
			PID:         lp.PID,
			ProcessName: lp.ProcessName,
		})
	}

	// Make a copy for return (caller owns this)
	resultCopy := make([]ListenInfoJSON, len(result))
	copy(resultCopy, result)

	// Return pooled slice
	putListenInfoSlice(resultPtr)

	// Return collected listening ports.
	return resultCopy
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
	// Collect all metrics with configuration
	metrics, err := CollectAllMetrics(ctx, cfg)
	// Check if collection failed
	if err != nil {
		// Return empty string with error
		return "", err
	}

	// Get pooled buffer
	buf := getJSONBuffer()
	defer putJSONBuffer(buf)

	// Encode metrics to JSON using pooled buffer
	enc := json.NewEncoder(buf)
	if err := enc.Encode(metrics); err != nil {
		// Return empty string with error
		return "", err
	}

	// Return the JSON string (buf.String() makes a copy)
	return buf.String(), nil
}
