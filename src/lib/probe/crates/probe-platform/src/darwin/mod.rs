//! macOS (Darwin) platform implementation
//!
//! Collects system metrics via sysctl and Mach APIs.

mod sysctl;

pub use sysctl::{
    ConnectionProtocol, ConnectionState, ContextSwitches, NetworkConnection, is_thermal_supported,
    list_network_connections, read_process_context_switches, read_self_context_switches,
    read_system_context_switches, read_thermal_zones,
};

use crate::{
    CPUCollector, CPUPressure, DiskCollector, DiskIOStats, DiskUsage, Error, IOCollector,
    IOPressure, IOStats, LoadAverage, LoadCollector, MemoryCollector, MemoryPressure, NetInterface,
    NetStats, NetworkCollector, Partition, ProcessCollector, ProcessMetrics, ProcessState, Result,
    SystemCPU, SystemCollector, SystemMemory,
};

/// macOS system collector implementation.
pub struct DarwinCollector {
    cpu: DarwinCPUCollector,
    memory: DarwinMemoryCollector,
    load: DarwinLoadCollector,
    process: DarwinProcessCollector,
    disk: DarwinDiskCollector,
    network: DarwinNetworkCollector,
    io: DarwinIOCollector,
}

impl DarwinCollector {
    /// Create a new Darwin collector.
    pub fn new() -> Self {
        Self {
            cpu: DarwinCPUCollector,
            memory: DarwinMemoryCollector,
            load: DarwinLoadCollector,
            process: DarwinProcessCollector,
            disk: DarwinDiskCollector,
            network: DarwinNetworkCollector,
            io: DarwinIOCollector,
        }
    }
}

impl Default for DarwinCollector {
    fn default() -> Self {
        Self::new()
    }
}

impl SystemCollector for DarwinCollector {
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

struct DarwinCPUCollector;

impl CPUCollector for DarwinCPUCollector {
    fn collect_system(&self) -> Result<SystemCPU> {
        let cpu_times = sysctl::get_cpu_times()?;
        let cpu_info = sysctl::get_cpu_info()?;

        Ok(SystemCPU {
            user_percent: cpu_times.user_percent,
            system_percent: cpu_times.system_percent,
            idle_percent: cpu_times.idle_percent,
            iowait_percent: 0.0, // Not available on macOS
            steal_percent: 0.0,  // Not available on macOS
            cores: cpu_info.cores,
            frequency_mhz: cpu_info.frequency_mhz,
        })
    }

    fn collect_pressure(&self) -> Result<CPUPressure> {
        // PSI not available on macOS
        Err(Error::NotSupported)
    }
}

// ============================================================================
// MEMORY COLLECTOR
// ============================================================================

struct DarwinMemoryCollector;

impl MemoryCollector for DarwinMemoryCollector {
    fn collect_system(&self) -> Result<SystemMemory> {
        let mem_info = sysctl::get_memory_info()?;

        Ok(SystemMemory {
            total_bytes: mem_info.total,
            available_bytes: mem_info.available,
            used_bytes: mem_info.total.saturating_sub(mem_info.available),
            cached_bytes: mem_info.cached,
            buffers_bytes: 0, // Not available on macOS
            swap_total_bytes: mem_info.swap_total,
            swap_used_bytes: mem_info.swap_used,
        })
    }

    fn collect_pressure(&self) -> Result<MemoryPressure> {
        // PSI not available on macOS
        Err(Error::NotSupported)
    }
}

// ============================================================================
// LOAD COLLECTOR
// ============================================================================

struct DarwinLoadCollector;

impl LoadCollector for DarwinLoadCollector {
    fn collect(&self) -> Result<LoadAverage> {
        let loadavg = sysctl::get_loadavg()?;

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

struct DarwinProcessCollector;

impl ProcessCollector for DarwinProcessCollector {
    fn collect(&self, pid: i32) -> Result<ProcessMetrics> {
        let proc_info = sysctl::get_process_info(pid)?;

        Ok(ProcessMetrics {
            pid,
            cpu_percent: 0.0, // Requires sampling
            memory_rss_bytes: proc_info.rss,
            memory_vms_bytes: proc_info.vsize,
            memory_percent: 0.0, // Requires total memory
            num_threads: proc_info.num_threads,
            num_fds: proc_info.num_fds,
            read_bytes_per_sec: 0,
            write_bytes_per_sec: 0,
            state: match proc_info.state {
                1 => ProcessState::Running,
                2 => ProcessState::Sleeping,
                3 => ProcessState::Waiting,
                4 => ProcessState::Zombie,
                5 => ProcessState::Stopped,
                _ => ProcessState::Unknown,
            },
        })
    }

    fn collect_all(&self) -> Result<Vec<ProcessMetrics>> {
        let pids = sysctl::list_pids()?;
        let results: Vec<ProcessMetrics> =
            pids.into_iter().filter_map(|pid| self.collect(pid).ok()).collect();
        Ok(results)
    }
}

// ============================================================================
// DISK COLLECTOR
// ============================================================================

struct DarwinDiskCollector;

impl DiskCollector for DarwinDiskCollector {
    fn list_partitions(&self) -> Result<Vec<Partition>> {
        sysctl::get_mounts()
    }

    fn collect_usage(&self, path: &str) -> Result<DiskUsage> {
        sysctl::get_disk_usage(path)
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
        sysctl::get_disk_io_stats()
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

struct DarwinNetworkCollector;

impl NetworkCollector for DarwinNetworkCollector {
    fn list_interfaces(&self) -> Result<Vec<NetInterface>> {
        sysctl::get_network_interfaces()
    }

    fn collect_stats(&self, interface: &str) -> Result<NetStats> {
        let all_stats = self.collect_all_stats()?;
        all_stats
            .into_iter()
            .find(|s| s.interface == interface)
            .ok_or_else(|| Error::NotFound(format!("interface {} not found", interface)))
    }

    fn collect_all_stats(&self) -> Result<Vec<NetStats>> {
        sysctl::get_network_stats()
    }
}

// ============================================================================
// I/O COLLECTOR
// ============================================================================

struct DarwinIOCollector;

impl IOCollector for DarwinIOCollector {
    fn collect_stats(&self) -> Result<IOStats> {
        // Aggregate from disk I/O
        let disk_stats = sysctl::get_disk_io_stats()?;

        let mut total = IOStats::default();
        for stat in disk_stats {
            total.read_ops += stat.reads_completed;
            total.read_bytes += stat.sectors_read * 512;
            total.write_ops += stat.writes_completed;
            total.write_bytes += stat.sectors_written * 512;
        }

        Ok(total)
    }

    fn collect_pressure(&self) -> Result<IOPressure> {
        // PSI not available on macOS
        Err(Error::NotSupported)
    }
}
