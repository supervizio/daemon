//! /proc filesystem parsing for Linux
//!
//! Parses various files under /proc to collect system metrics.

use crate::{Error, Result};
use std::fs;

/// CPU statistics from /proc/stat.
#[derive(Debug, Default)]
pub struct ProcStat {
    user: u64,
    nice: u64,
    system: u64,
    idle: u64,
    iowait: u64,
    irq: u64,
    softirq: u64,
    steal: u64,
    total: u64,
}

impl ProcStat {
    /// Read and parse /proc/stat.
    pub fn read() -> Result<Self> {
        let content = fs::read_to_string("/proc/stat")?;
        let line = content
            .lines()
            .next()
            .ok_or_else(|| Error::Platform("empty /proc/stat".into()))?;

        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 9 || parts[0] != "cpu" {
            return Err(Error::Platform("invalid /proc/stat format".into()));
        }

        let user: u64 = parts[1].parse().unwrap_or(0);
        let nice: u64 = parts[2].parse().unwrap_or(0);
        let system: u64 = parts[3].parse().unwrap_or(0);
        let idle: u64 = parts[4].parse().unwrap_or(0);
        let iowait: u64 = parts[5].parse().unwrap_or(0);
        let irq: u64 = parts[6].parse().unwrap_or(0);
        let softirq: u64 = parts[7].parse().unwrap_or(0);
        let steal: u64 = parts[8].parse().unwrap_or(0);

        let total = user + nice + system + idle + iowait + irq + softirq + steal;

        Ok(Self {
            user,
            nice,
            system,
            idle,
            iowait,
            irq,
            softirq,
            steal,
            total,
        })
    }

    /// User CPU percentage.
    pub fn user_percent(&self) -> f64 {
        if self.total == 0 {
            return 0.0;
        }
        (self.user + self.nice) as f64 / self.total as f64 * 100.0
    }

    /// System CPU percentage.
    pub fn system_percent(&self) -> f64 {
        if self.total == 0 {
            return 0.0;
        }
        (self.system + self.irq + self.softirq) as f64 / self.total as f64 * 100.0
    }

    /// Idle CPU percentage.
    pub fn idle_percent(&self) -> f64 {
        if self.total == 0 {
            return 0.0;
        }
        self.idle as f64 / self.total as f64 * 100.0
    }

    /// I/O wait percentage.
    pub fn iowait_percent(&self) -> f64 {
        if self.total == 0 {
            return 0.0;
        }
        self.iowait as f64 / self.total as f64 * 100.0
    }

    /// Steal percentage (VMs).
    pub fn steal_percent(&self) -> f64 {
        if self.total == 0 {
            return 0.0;
        }
        self.steal as f64 / self.total as f64 * 100.0
    }
}

/// CPU information from /proc/cpuinfo.
#[derive(Debug, Default)]
pub struct CpuInfo {
    /// Number of CPU cores.
    pub num_cores: u32,
    /// CPU frequency in MHz.
    pub frequency_mhz: u64,
}

impl CpuInfo {
    /// Read and parse /proc/cpuinfo.
    pub fn read() -> Result<Self> {
        let content = fs::read_to_string("/proc/cpuinfo")?;
        let mut num_cores = 0u32;
        let mut frequency_mhz = 0u64;

        for line in content.lines() {
            if line.starts_with("processor") {
                num_cores += 1;
            } else if line.starts_with("cpu MHz")
                && let Some(value) = line.split(':').nth(1)
                && let Ok(freq) = value.trim().parse::<f64>()
            {
                frequency_mhz = freq as u64;
            }
        }

        Ok(Self {
            num_cores,
            frequency_mhz,
        })
    }
}

/// Memory information from /proc/meminfo.
#[derive(Debug, Default)]
pub struct MemInfo {
    pub mem_total: u64,
    pub mem_free: u64,
    pub mem_available: u64,
    pub buffers: u64,
    pub cached: u64,
    pub swap_total: u64,
    pub swap_free: u64,
}

impl MemInfo {
    /// Read and parse /proc/meminfo.
    pub fn read() -> Result<Self> {
        let content = fs::read_to_string("/proc/meminfo")?;
        let mut info = Self::default();

        for line in content.lines() {
            let parts: Vec<&str> = line.split_whitespace().collect();
            if parts.len() < 2 {
                continue;
            }

            // Values are in kB, convert to bytes
            let value: u64 = parts[1].parse().unwrap_or(0) * 1024;

            match parts[0] {
                "MemTotal:" => info.mem_total = value,
                "MemFree:" => info.mem_free = value,
                "MemAvailable:" => info.mem_available = value,
                "Buffers:" => info.buffers = value,
                "Cached:" => info.cached = value,
                "SwapTotal:" => info.swap_total = value,
                "SwapFree:" => info.swap_free = value,
                _ => {}
            }
        }

        Ok(info)
    }
}

/// Load average from /proc/loadavg.
#[derive(Debug, Default)]
pub struct LoadAvg {
    pub load_1min: f64,
    pub load_5min: f64,
    pub load_15min: f64,
}

impl LoadAvg {
    /// Read and parse /proc/loadavg.
    pub fn read() -> Result<Self> {
        let content = fs::read_to_string("/proc/loadavg")?;
        let parts: Vec<&str> = content.split_whitespace().collect();

        if parts.len() < 3 {
            return Err(Error::Platform("invalid /proc/loadavg format".into()));
        }

        Ok(Self {
            load_1min: parts[0].parse().unwrap_or(0.0),
            load_5min: parts[1].parse().unwrap_or(0.0),
            load_15min: parts[2].parse().unwrap_or(0.0),
        })
    }
}

/// Process statistics from /proc/[pid]/stat.
#[derive(Debug, Default)]
pub struct ProcessStat {
    /// Process ID (used for debugging/logging).
    #[allow(dead_code)]
    pub pid: i32,
    /// Process state character.
    pub state: char,
    /// Number of threads.
    pub num_threads: u32,
    /// User time ticks (used for CPU percentage calculation).
    #[allow(dead_code)]
    pub utime: u64,
    /// System time ticks (used for CPU percentage calculation).
    #[allow(dead_code)]
    pub stime: u64,
}

impl ProcessStat {
    /// Read and parse /proc/[pid]/stat.
    pub fn read(pid: i32) -> Result<Self> {
        let path = format!("/proc/{}/stat", pid);
        let content = fs::read_to_string(&path).map_err(|e| {
            if e.kind() == std::io::ErrorKind::NotFound {
                Error::NotFound(format!("process {} not found", pid))
            } else {
                Error::Io(e)
            }
        })?;

        // Format: pid (comm) state ...
        // Find the closing paren to handle commands with spaces
        let _start = content
            .find('(')
            .ok_or_else(|| Error::Platform(format!("invalid stat format for pid {}", pid)))?;
        let end = content
            .rfind(')')
            .ok_or_else(|| Error::Platform(format!("invalid stat format for pid {}", pid)))?;

        let after_comm = &content[end + 2..]; // Skip ") "
        let fields: Vec<&str> = after_comm.split_whitespace().collect();

        if fields.is_empty() {
            return Err(Error::Platform(format!(
                "insufficient fields in stat for pid {}",
                pid
            )));
        }

        let state = fields[0].chars().next().unwrap_or('?');
        let utime: u64 = fields.get(11).and_then(|s| s.parse().ok()).unwrap_or(0);
        let stime: u64 = fields.get(12).and_then(|s| s.parse().ok()).unwrap_or(0);
        let num_threads: u32 = fields.get(17).and_then(|s| s.parse().ok()).unwrap_or(0);

        Ok(Self {
            pid,
            state,
            num_threads,
            utime,
            stime,
        })
    }
}

/// Process status from /proc/[pid]/status.
#[derive(Debug, Default)]
pub struct ProcessStatus {
    pub vm_size: u64,
    pub vm_rss: u64,
}

impl ProcessStatus {
    /// Read and parse /proc/[pid]/status.
    pub fn read(pid: i32) -> Result<Self> {
        let path = format!("/proc/{}/status", pid);
        let content = fs::read_to_string(&path).map_err(|e| {
            if e.kind() == std::io::ErrorKind::NotFound {
                Error::NotFound(format!("process {} not found", pid))
            } else {
                Error::Io(e)
            }
        })?;

        let mut status = Self::default();

        for line in content.lines() {
            let parts: Vec<&str> = line.split_whitespace().collect();
            if parts.len() < 2 {
                continue;
            }

            // Values are in kB
            let value: u64 = parts[1].parse().unwrap_or(0) * 1024;

            match parts[0] {
                "VmSize:" => status.vm_size = value,
                "VmRSS:" => status.vm_rss = value,
                _ => {}
            }
        }

        Ok(status)
    }
}

/// Count open file descriptors for a process.
pub fn count_fds(pid: i32) -> Result<u32> {
    let path = format!("/proc/{}/fd", pid);
    let entries = fs::read_dir(&path).map_err(|e| {
        if e.kind() == std::io::ErrorKind::NotFound {
            Error::NotFound(format!("process {} not found", pid))
        } else if e.kind() == std::io::ErrorKind::PermissionDenied {
            Error::Permission(format!("cannot read fds for pid {}", pid))
        } else {
            Error::Io(e)
        }
    })?;

    Ok(entries.count() as u32)
}

// ============================================================================
// PRESSURE STALL INFORMATION (PSI)
// ============================================================================

use crate::{
    CPUPressure, DiskIOStats, DiskUsage, IOPressure, IOStats, MemoryPressure, NetInterface,
    NetStats, Partition,
};

/// Parse PSI line: "some avg10=0.00 avg60=0.00 avg300=0.00 total=0"
fn parse_psi_line(line: &str) -> (f64, f64, f64, u64) {
    let mut avg10 = 0.0;
    let mut avg60 = 0.0;
    let mut avg300 = 0.0;
    let mut total = 0u64;

    for part in line.split_whitespace().skip(1) {
        if let Some((key, value)) = part.split_once('=') {
            match key {
                "avg10" => avg10 = value.parse().unwrap_or(0.0),
                "avg60" => avg60 = value.parse().unwrap_or(0.0),
                "avg300" => avg300 = value.parse().unwrap_or(0.0),
                "total" => total = value.parse().unwrap_or(0),
                _ => {}
            }
        }
    }

    (avg10, avg60, avg300, total)
}

/// Read CPU pressure from /proc/pressure/cpu.
pub fn read_cpu_pressure() -> Result<CPUPressure> {
    let content = fs::read_to_string("/proc/pressure/cpu").map_err(|e| {
        if e.kind() == std::io::ErrorKind::NotFound {
            Error::NotSupported
        } else {
            Error::Io(e)
        }
    })?;

    for line in content.lines() {
        if line.starts_with("some") {
            let (avg10, avg60, avg300, total) = parse_psi_line(line);
            return Ok(CPUPressure {
                some_avg10: avg10,
                some_avg60: avg60,
                some_avg300: avg300,
                some_total_us: total,
            });
        }
    }

    Ok(CPUPressure::default())
}

/// Read memory pressure from /proc/pressure/memory.
pub fn read_memory_pressure() -> Result<MemoryPressure> {
    let content = fs::read_to_string("/proc/pressure/memory").map_err(|e| {
        if e.kind() == std::io::ErrorKind::NotFound {
            Error::NotSupported
        } else {
            Error::Io(e)
        }
    })?;

    let mut pressure = MemoryPressure::default();

    for line in content.lines() {
        if line.starts_with("some") {
            let (avg10, avg60, avg300, total) = parse_psi_line(line);
            pressure.some_avg10 = avg10;
            pressure.some_avg60 = avg60;
            pressure.some_avg300 = avg300;
            pressure.some_total_us = total;
        } else if line.starts_with("full") {
            let (avg10, avg60, avg300, total) = parse_psi_line(line);
            pressure.full_avg10 = avg10;
            pressure.full_avg60 = avg60;
            pressure.full_avg300 = avg300;
            pressure.full_total_us = total;
        }
    }

    Ok(pressure)
}

/// Read I/O pressure from /proc/pressure/io.
pub fn read_io_pressure() -> Result<IOPressure> {
    let content = fs::read_to_string("/proc/pressure/io").map_err(|e| {
        if e.kind() == std::io::ErrorKind::NotFound {
            Error::NotSupported
        } else {
            Error::Io(e)
        }
    })?;

    let mut pressure = IOPressure::default();

    for line in content.lines() {
        if line.starts_with("some") {
            let (avg10, avg60, avg300, total) = parse_psi_line(line);
            pressure.some_avg10 = avg10;
            pressure.some_avg60 = avg60;
            pressure.some_avg300 = avg300;
            pressure.some_total_us = total;
        } else if line.starts_with("full") {
            let (avg10, avg60, avg300, total) = parse_psi_line(line);
            pressure.full_avg10 = avg10;
            pressure.full_avg60 = avg60;
            pressure.full_avg300 = avg300;
            pressure.full_total_us = total;
        }
    }

    Ok(pressure)
}

// ============================================================================
// PROCESS ENUMERATION
// ============================================================================

/// List all process IDs from /proc.
pub fn list_processes() -> Result<Vec<i32>> {
    let mut pids = Vec::new();

    for entry in fs::read_dir("/proc")? {
        let entry = entry?;
        if let Some(name) = entry.file_name().to_str()
            && let Ok(pid) = name.parse::<i32>()
        {
            pids.push(pid);
        }
    }

    Ok(pids)
}

// ============================================================================
// DISK METRICS
// ============================================================================

/// Read mounted partitions from /proc/mounts.
pub fn read_mounts() -> Result<Vec<Partition>> {
    let content = fs::read_to_string("/proc/mounts")?;
    let mut partitions = Vec::new();

    for line in content.lines() {
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 4 {
            continue;
        }

        let device = parts[0];
        let mount_point = parts[1];
        let fs_type = parts[2];
        let options = parts[3];

        // Skip pseudo filesystems
        if fs_type == "proc"
            || fs_type == "sysfs"
            || fs_type == "devtmpfs"
            || fs_type == "devpts"
            || fs_type == "cgroup"
            || fs_type == "cgroup2"
            || fs_type == "securityfs"
            || fs_type == "debugfs"
            || fs_type == "tracefs"
            || fs_type == "configfs"
            || fs_type == "fusectl"
            || fs_type == "mqueue"
            || fs_type == "hugetlbfs"
            || fs_type == "pstore"
            || fs_type == "bpf"
            || fs_type == "autofs"
        {
            continue;
        }

        partitions.push(Partition {
            device: device.to_string(),
            mount_point: mount_point.to_string(),
            fs_type: fs_type.to_string(),
            options: options.to_string(),
        });
    }

    Ok(partitions)
}

/// Read disk usage for a path using statvfs.
pub fn read_disk_usage(path: &str) -> Result<DiskUsage> {
    use std::ffi::CString;
    use std::mem::MaybeUninit;

    let c_path = CString::new(path).map_err(|_| Error::Platform("invalid path".into()))?;

    let mut stat: MaybeUninit<libc::statvfs> = MaybeUninit::uninit();

    let ret = unsafe { libc::statvfs(c_path.as_ptr(), stat.as_mut_ptr()) };

    if ret != 0 {
        return Err(Error::Io(std::io::Error::last_os_error()));
    }

    let stat = unsafe { stat.assume_init() };

    let block_size = stat.f_frsize;
    let total_bytes = stat.f_blocks * block_size;
    let free_bytes = stat.f_bfree * block_size;
    let available_bytes = stat.f_bavail * block_size;
    let used_bytes = total_bytes.saturating_sub(free_bytes);

    let used_percent = if total_bytes > 0 {
        (used_bytes as f64 / total_bytes as f64) * 100.0
    } else {
        0.0
    };

    Ok(DiskUsage {
        path: path.to_string(),
        total_bytes,
        used_bytes,
        free_bytes: available_bytes,
        used_percent,
        inodes_total: stat.f_files,
        inodes_used: stat.f_files.saturating_sub(stat.f_ffree),
        inodes_free: stat.f_ffree,
    })
}

/// Read disk I/O statistics from /proc/diskstats.
pub fn read_diskstats() -> Result<Vec<DiskIOStats>> {
    let content = fs::read_to_string("/proc/diskstats")?;
    let mut stats = Vec::new();

    for line in content.lines() {
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 14 {
            continue;
        }

        let device = parts[2];

        // Skip partitions (e.g., sda1, sda2) - only report whole devices
        // Also skip loop devices, ram devices, etc.
        if device.starts_with("loop")
            || device.starts_with("ram")
            || device.starts_with("dm-")
            || (device.len() > 3
                && device.chars().last().is_some_and(|c| c.is_ascii_digit())
                && device
                    .chars()
                    .nth(device.len() - 2)
                    .is_some_and(|c| c.is_ascii_alphabetic()))
        {
            continue;
        }

        stats.push(DiskIOStats {
            device: device.to_string(),
            reads_completed: parts[3].parse().unwrap_or(0),
            sectors_read: parts[5].parse().unwrap_or(0),
            read_time_ms: parts[6].parse().unwrap_or(0),
            writes_completed: parts[7].parse().unwrap_or(0),
            sectors_written: parts[9].parse().unwrap_or(0),
            write_time_ms: parts[10].parse().unwrap_or(0),
            io_in_progress: parts[11].parse().unwrap_or(0),
            io_time_ms: parts[12].parse().unwrap_or(0),
            weighted_io_time_ms: parts[13].parse().unwrap_or(0),
        });
    }

    Ok(stats)
}

// ============================================================================
// NETWORK METRICS
// ============================================================================

/// Read network interfaces from /sys/class/net.
pub fn read_net_interfaces() -> Result<Vec<NetInterface>> {
    let mut interfaces = Vec::new();

    for entry in fs::read_dir("/sys/class/net")? {
        let entry = entry?;
        let name = entry.file_name().to_string_lossy().to_string();
        let iface_path = entry.path();

        // Read MAC address
        let mac_address = fs::read_to_string(iface_path.join("address"))
            .map(|s| s.trim().to_string())
            .unwrap_or_default();

        // Read MTU
        let mtu: u32 = fs::read_to_string(iface_path.join("mtu"))
            .ok()
            .and_then(|s| s.trim().parse().ok())
            .unwrap_or(0);

        // Read flags to check if up
        let flags: u32 = fs::read_to_string(iface_path.join("flags"))
            .ok()
            .and_then(|s| {
                let s = s.trim().trim_start_matches("0x");
                u32::from_str_radix(s, 16).ok()
            })
            .unwrap_or(0);

        let is_up = (flags & 0x1) != 0; // IFF_UP
        let is_loopback = (flags & 0x8) != 0; // IFF_LOOPBACK

        interfaces.push(NetInterface {
            name,
            mac_address,
            ipv4_addresses: Vec::new(), // Would need to use netlink/ioctl
            ipv6_addresses: Vec::new(),
            mtu,
            is_up,
            is_loopback,
        });
    }

    Ok(interfaces)
}

/// Read network statistics from /proc/net/dev.
pub fn read_net_dev() -> Result<Vec<NetStats>> {
    let content = fs::read_to_string("/proc/net/dev")?;
    let mut stats = Vec::new();

    for line in content.lines().skip(2) {
        // Skip header lines
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 17 {
            continue;
        }

        let interface = parts[0].trim_end_matches(':').to_string();

        stats.push(NetStats {
            interface,
            rx_bytes: parts[1].parse().unwrap_or(0),
            rx_packets: parts[2].parse().unwrap_or(0),
            rx_errors: parts[3].parse().unwrap_or(0),
            rx_drops: parts[4].parse().unwrap_or(0),
            tx_bytes: parts[9].parse().unwrap_or(0),
            tx_packets: parts[10].parse().unwrap_or(0),
            tx_errors: parts[11].parse().unwrap_or(0),
            tx_drops: parts[12].parse().unwrap_or(0),
        });
    }

    Ok(stats)
}

// ============================================================================
// I/O METRICS
// ============================================================================

// ============================================================================
// CONTEXT SWITCHES
// ============================================================================

use crate::ContextSwitches;

/// Read system-wide context switch count from /proc/stat.
pub fn read_system_context_switches() -> Result<u64> {
    let content = fs::read_to_string("/proc/stat")?;

    for line in content.lines() {
        if line.starts_with("ctxt ") {
            let parts: Vec<&str> = line.split_whitespace().collect();
            if parts.len() >= 2 {
                return Ok(parts[1].parse().unwrap_or(0));
            }
        }
    }

    Ok(0)
}

/// Read per-process context switches from /proc/[pid]/status.
pub fn read_process_context_switches(pid: i32) -> Result<ContextSwitches> {
    let path = format!("/proc/{}/status", pid);
    let content = fs::read_to_string(&path).map_err(|e| {
        if e.kind() == std::io::ErrorKind::NotFound {
            Error::NotFound(format!("process {} not found", pid))
        } else {
            Error::Io(e)
        }
    })?;

    let mut switches = ContextSwitches::default();

    for line in content.lines() {
        if line.starts_with("voluntary_ctxt_switches:") {
            let parts: Vec<&str> = line.split_whitespace().collect();
            if parts.len() >= 2 {
                switches.voluntary = parts[1].parse().unwrap_or(0);
            }
        } else if line.starts_with("nonvoluntary_ctxt_switches:") {
            let parts: Vec<&str> = line.split_whitespace().collect();
            if parts.len() >= 2 {
                switches.involuntary = parts[1].parse().unwrap_or(0);
            }
        }
    }

    // Also read system-wide total
    switches.system_total = read_system_context_switches().unwrap_or(0);

    Ok(switches)
}

/// Read context switches for the current process.
pub fn read_self_context_switches() -> Result<ContextSwitches> {
    let content = fs::read_to_string("/proc/self/status")?;

    let mut switches = ContextSwitches::default();

    for line in content.lines() {
        if line.starts_with("voluntary_ctxt_switches:") {
            let parts: Vec<&str> = line.split_whitespace().collect();
            if parts.len() >= 2 {
                switches.voluntary = parts[1].parse().unwrap_or(0);
            }
        } else if line.starts_with("nonvoluntary_ctxt_switches:") {
            let parts: Vec<&str> = line.split_whitespace().collect();
            if parts.len() >= 2 {
                switches.involuntary = parts[1].parse().unwrap_or(0);
            }
        }
    }

    switches.system_total = read_system_context_switches().unwrap_or(0);

    Ok(switches)
}

/// Read system-wide I/O statistics (aggregated from diskstats).
pub fn read_io_stats() -> Result<IOStats> {
    let diskstats = read_diskstats()?;

    let mut stats = IOStats::default();

    for disk in diskstats {
        stats.read_ops += disk.reads_completed;
        stats.read_bytes += disk.sectors_read * 512; // 512 bytes per sector
        stats.write_ops += disk.writes_completed;
        stats.write_bytes += disk.sectors_written * 512;
    }

    Ok(stats)
}

#[cfg(test)]
mod context_switch_tests {
    use super::*;

    #[test]
    fn test_read_system_context_switches() {
        let result = read_system_context_switches();
        assert!(result.is_ok());
        // System should have had at least some context switches
        assert!(result.unwrap() > 0);
    }

    #[test]
    fn test_read_self_context_switches() {
        let result = read_self_context_switches();
        assert!(result.is_ok());
        let switches = result.unwrap();
        // Current process should have had at least one context switch
        assert!(switches.voluntary > 0 || switches.involuntary > 0 || switches.system_total > 0);
    }

    #[test]
    fn test_read_process_context_switches() {
        // Read context switches for pid 1 (init/systemd)
        let result = read_process_context_switches(1);
        // This might fail if we don't have permission, which is OK
        if let Ok(switches) = result {
            assert!(switches.system_total > 0);
        }
    }
}
