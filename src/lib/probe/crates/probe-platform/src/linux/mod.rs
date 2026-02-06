//! Linux platform implementation
//!
//! Collects system metrics via the /proc and /sys filesystems.

mod connections;
mod procfs;
mod thermal;

pub use connections::{
    build_socket_pid_map, collect_process_connections, collect_tcp_connections, collect_tcp_stats,
    collect_udp_connections, collect_unix_sockets, find_process_by_port,
};
pub use procfs::{
    read_process_context_switches, read_self_context_switches, read_system_context_switches,
};
pub use thermal::{is_thermal_supported, read_thermal_zones};

use crate::{
    CPUCollector, CPUPressure, ConnectionCollector, DiskCollector, DiskIOStats, DiskUsage, Error,
    IOCollector, IOPressure, IOStats, LoadAverage, LoadCollector, MemoryCollector, MemoryPressure,
    NetInterface, NetStats, NetworkCollector, Partition, ProcessCollector, ProcessMetrics,
    ProcessState, Result, SystemCPU, SystemCollector, SystemMemory, TcpConnection, TcpStats,
    ThermalCollector, ThermalZone, UdpConnection, UnixSocket,
};

/// Linux system collector implementation.
pub struct LinuxCollector {
    cpu: LinuxCPUCollector,
    memory: LinuxMemoryCollector,
    load: LinuxLoadCollector,
    process: LinuxProcessCollector,
    disk: LinuxDiskCollector,
    network: LinuxNetworkCollector,
    io: LinuxIOCollector,
}

impl LinuxCollector {
    /// Create a new Linux collector.
    pub fn new() -> Self {
        Self {
            cpu: LinuxCPUCollector,
            memory: LinuxMemoryCollector,
            load: LinuxLoadCollector,
            process: LinuxProcessCollector,
            disk: LinuxDiskCollector,
            network: LinuxNetworkCollector,
            io: LinuxIOCollector,
        }
    }
}

impl Default for LinuxCollector {
    fn default() -> Self {
        Self::new()
    }
}

impl SystemCollector for LinuxCollector {
    fn cpu(&self) -> &dyn CPUCollector {
        &self.cpu
    }

    fn memory(&self) -> &dyn MemoryCollector {
        &self.memory
    }

    fn load(&self) -> &dyn LoadCollector {
        &self.load
    }

    fn process(&self) -> &dyn ProcessCollector {
        &self.process
    }

    fn disk(&self) -> &dyn DiskCollector {
        &self.disk
    }

    fn network(&self) -> &dyn NetworkCollector {
        &self.network
    }

    fn io(&self) -> &dyn IOCollector {
        &self.io
    }
}

// ============================================================================
// CPU COLLECTOR
// ============================================================================

struct LinuxCPUCollector;

impl CPUCollector for LinuxCPUCollector {
    fn collect_system(&self) -> Result<SystemCPU> {
        let stat = procfs::ProcStat::read()?;
        let cpuinfo = procfs::CpuInfo::read()?;

        Ok(SystemCPU {
            user_percent: stat.user_percent(),
            system_percent: stat.system_percent(),
            idle_percent: stat.idle_percent(),
            iowait_percent: stat.iowait_percent(),
            steal_percent: stat.steal_percent(),
            cores: cpuinfo.num_cores,
            frequency_mhz: cpuinfo.frequency_mhz,
        })
    }

    fn collect_pressure(&self) -> Result<CPUPressure> {
        procfs::read_cpu_pressure()
    }
}

// ============================================================================
// MEMORY COLLECTOR
// ============================================================================

struct LinuxMemoryCollector;

impl MemoryCollector for LinuxMemoryCollector {
    fn collect_system(&self) -> Result<SystemMemory> {
        let meminfo = procfs::MemInfo::read()?;

        Ok(SystemMemory {
            total_bytes: meminfo.mem_total,
            available_bytes: meminfo.mem_available,
            used_bytes: meminfo.mem_total.saturating_sub(meminfo.mem_available),
            cached_bytes: meminfo.cached,
            buffers_bytes: meminfo.buffers,
            swap_total_bytes: meminfo.swap_total,
            swap_used_bytes: meminfo.swap_total.saturating_sub(meminfo.swap_free),
        })
    }

    fn collect_pressure(&self) -> Result<MemoryPressure> {
        procfs::read_memory_pressure()
    }
}

// ============================================================================
// LOAD COLLECTOR
// ============================================================================

struct LinuxLoadCollector;

impl LoadCollector for LinuxLoadCollector {
    fn collect(&self) -> Result<LoadAverage> {
        let loadavg = procfs::LoadAvg::read()?;

        Ok(LoadAverage {
            load_1min: loadavg.load_1min,
            load_5min: loadavg.load_5min,
            load_15min: loadavg.load_15min,
        })
    }
}

// ============================================================================
// PROCESS COLLECTOR
// ============================================================================

struct LinuxProcessCollector;

impl ProcessCollector for LinuxProcessCollector {
    fn collect(&self, pid: i32) -> Result<ProcessMetrics> {
        let stat = procfs::ProcessStat::read(pid)?;
        let status = procfs::ProcessStatus::read(pid)?;

        Ok(ProcessMetrics {
            pid,
            cpu_percent: 0.0,
            memory_rss_bytes: status.vm_rss,
            memory_vms_bytes: status.vm_size,
            memory_percent: 0.0,
            num_threads: stat.num_threads,
            num_fds: procfs::count_fds(pid).unwrap_or(0),
            read_bytes_per_sec: 0,
            write_bytes_per_sec: 0,
            state: match stat.state {
                'R' => ProcessState::Running,
                'S' => ProcessState::Sleeping,
                'D' => ProcessState::Waiting,
                'Z' => ProcessState::Zombie,
                'T' => ProcessState::Stopped,
                _ => ProcessState::Unknown,
            },
        })
    }

    fn collect_all(&self) -> Result<Vec<ProcessMetrics>> {
        procfs::list_processes()?
            .into_iter()
            .filter_map(|pid| self.collect(pid).ok())
            .collect::<Vec<_>>()
            .pipe(Ok)
    }
}

// Helper trait for functional style
trait Pipe: Sized {
    fn pipe<F, R>(self, f: F) -> R
    where
        F: FnOnce(Self) -> R,
    {
        f(self)
    }
}

impl<T> Pipe for T {}

// ============================================================================
// DISK COLLECTOR
// ============================================================================

struct LinuxDiskCollector;

impl DiskCollector for LinuxDiskCollector {
    fn list_partitions(&self) -> Result<Vec<Partition>> {
        procfs::read_mounts()
    }

    fn collect_usage(&self, path: &str) -> Result<DiskUsage> {
        procfs::read_disk_usage(path)
    }

    fn collect_all_usage(&self) -> Result<Vec<DiskUsage>> {
        let partitions = self.list_partitions()?;
        let mut usages = Vec::new();

        for partition in partitions {
            if let Ok(usage) = self.collect_usage(&partition.mount_point) {
                usages.push(usage);
            }
        }

        Ok(usages)
    }

    fn collect_io(&self) -> Result<Vec<DiskIOStats>> {
        procfs::read_diskstats()
    }

    fn collect_device_io(&self, device: &str) -> Result<DiskIOStats> {
        let stats = self.collect_io()?;
        stats
            .into_iter()
            .find(|s| s.device == device)
            .ok_or_else(|| Error::NotFound(format!("device {} not found", device)))
    }
}

// ============================================================================
// NETWORK COLLECTOR
// ============================================================================

struct LinuxNetworkCollector;

impl NetworkCollector for LinuxNetworkCollector {
    fn list_interfaces(&self) -> Result<Vec<NetInterface>> {
        procfs::read_net_interfaces()
    }

    fn collect_stats(&self, interface: &str) -> Result<NetStats> {
        let all_stats = self.collect_all_stats()?;
        all_stats
            .into_iter()
            .find(|s| s.interface == interface)
            .ok_or_else(|| Error::NotFound(format!("interface {} not found", interface)))
    }

    fn collect_all_stats(&self) -> Result<Vec<NetStats>> {
        procfs::read_net_dev()
    }
}

// ============================================================================
// I/O COLLECTOR
// ============================================================================

struct LinuxIOCollector;

impl IOCollector for LinuxIOCollector {
    fn collect_stats(&self) -> Result<IOStats> {
        procfs::read_io_stats()
    }

    fn collect_pressure(&self) -> Result<IOPressure> {
        procfs::read_io_pressure()
    }
}

// ============================================================================
// THERMAL COLLECTOR
// ============================================================================

/// Linux thermal collector using /sys/class/hwmon.
pub struct LinuxThermalCollector;

impl ThermalCollector for LinuxThermalCollector {
    fn is_supported(&self) -> bool {
        thermal::is_thermal_supported()
    }

    fn list_zones(&self) -> Result<Vec<ThermalZone>> {
        thermal::read_thermal_zones()
    }

    fn collect_temperatures(&self) -> Result<Vec<ThermalZone>> {
        thermal::read_thermal_zones()
    }
}

// ============================================================================
// CONNECTION COLLECTOR
// ============================================================================

/// Linux connection collector using /proc/net.
pub struct LinuxConnectionCollector;

impl ConnectionCollector for LinuxConnectionCollector {
    fn collect_tcp(&self) -> Result<Vec<TcpConnection>> {
        connections::collect_tcp_connections()
    }

    fn collect_udp(&self) -> Result<Vec<UdpConnection>> {
        connections::collect_udp_connections()
    }

    fn collect_unix(&self) -> Result<Vec<UnixSocket>> {
        connections::collect_unix_sockets()
    }

    fn collect_tcp_stats(&self) -> Result<TcpStats> {
        connections::collect_tcp_stats()
    }

    fn collect_process_connections(
        &self,
        pid: i32,
    ) -> Result<(Vec<TcpConnection>, Vec<UdpConnection>)> {
        connections::collect_process_connections(pid)
    }

    fn find_process_by_port(&self, port: u16, tcp: bool) -> Result<Option<i32>> {
        connections::find_process_by_port(port, tcp)
    }
}
