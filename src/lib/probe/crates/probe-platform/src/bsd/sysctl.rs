//! BSD sysctl wrappers

use crate::{DiskIOStats, DiskUsage, Error, NetInterface, NetStats, Partition, Result};
use std::ffi::CString;
use std::mem;
use std::ptr;

// ============================================================================
// CPU
// ============================================================================

pub struct CpuTimes {
    pub user_percent: f64,
    pub system_percent: f64,
    pub idle_percent: f64,
}

pub struct CpuInfo {
    pub cores: u32,
    pub frequency_mhz: u64,
}

pub fn get_cpu_times() -> Result<CpuTimes> {
    unsafe {
        // kern.cp_time on FreeBSD/NetBSD, kern.cp_time2 on OpenBSD
        #[cfg(target_os = "freebsd")]
        let name = CString::new("kern.cp_time")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
        #[cfg(target_os = "openbsd")]
        let name = CString::new("kern.cp_time")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
        #[cfg(target_os = "netbsd")]
        let name = CString::new("kern.cp_time")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

        let mut cp_time: [u64; 5] = [0; 5]; // user, nice, sys, intr, idle
        let mut len = mem::size_of_val(&cp_time);

        let result = libc::sysctlbyname(
            name.as_ptr(),
            cp_time.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        if result != 0 {
            return Ok(CpuTimes { user_percent: 0.0, system_percent: 0.0, idle_percent: 100.0 });
        }

        let user = cp_time[0] + cp_time[1]; // user + nice
        let system = cp_time[2] + cp_time[3]; // sys + intr
        let idle = cp_time[4];
        let total = user + system + idle;

        if total == 0 {
            return Ok(CpuTimes { user_percent: 0.0, system_percent: 0.0, idle_percent: 100.0 });
        }

        Ok(CpuTimes {
            user_percent: (user as f64 / total as f64) * 100.0,
            system_percent: (system as f64 / total as f64) * 100.0,
            idle_percent: (idle as f64 / total as f64) * 100.0,
        })
    }
}

pub fn get_cpu_info() -> Result<CpuInfo> {
    unsafe {
        // Get number of CPUs
        let name = CString::new("hw.ncpu")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
        let mut ncpu: libc::c_int = 0;
        let mut len = mem::size_of::<libc::c_int>();

        libc::sysctlbyname(
            name.as_ptr(),
            &mut ncpu as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        // Get CPU frequency
        #[cfg(target_os = "freebsd")]
        let freq_name = CString::new("hw.clockrate")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
        #[cfg(target_os = "openbsd")]
        let freq_name = CString::new("hw.cpuspeed")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
        #[cfg(target_os = "netbsd")]
        let freq_name = CString::new("machdep.tsc_freq")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

        let mut freq: libc::c_int = 0;
        let mut freq_len = mem::size_of::<libc::c_int>();

        libc::sysctlbyname(
            freq_name.as_ptr(),
            &mut freq as *mut _ as *mut libc::c_void,
            &mut freq_len,
            ptr::null_mut(),
            0,
        );

        Ok(CpuInfo { cores: ncpu as u32, frequency_mhz: freq as u64 })
    }
}

// ============================================================================
// MEMORY
// ============================================================================

pub struct MemInfo {
    pub total: u64,
    pub available: u64,
    pub cached: u64,
    pub buffers: u64,
    pub swap_total: u64,
    pub swap_used: u64,
}

pub fn get_memory_info() -> Result<MemInfo> {
    unsafe {
        // Get page size
        let page_size = libc::sysconf(libc::_SC_PAGESIZE) as u64;

        // Get total physical memory
        let name = CString::new("hw.physmem")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
        let mut physmem: u64 = 0;
        let mut len = mem::size_of::<u64>();

        libc::sysctlbyname(
            name.as_ptr(),
            &mut physmem as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        // Get free pages
        #[cfg(target_os = "freebsd")]
        let free_name = CString::new("vm.stats.vm.v_free_count")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
        #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
        let free_name = CString::new("vm.uvmexp")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

        let mut free_pages: u64 = 0;
        let mut free_len = mem::size_of::<u64>();

        #[cfg(target_os = "freebsd")]
        {
            libc::sysctlbyname(
                free_name.as_ptr(),
                &mut free_pages as *mut _ as *mut libc::c_void,
                &mut free_len,
                ptr::null_mut(),
                0,
            );
        }

        #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
        {
            // OpenBSD/NetBSD use uvmexp structure
            free_pages = physmem / page_size / 4; // Rough estimate
        }

        // Get cached pages (FreeBSD specific)
        #[cfg(target_os = "freebsd")]
        let cached = {
            let cache_name = CString::new("vm.stats.vm.v_cache_count")
                .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
            let mut cache_pages: u64 = 0;
            let mut cache_len = mem::size_of::<u64>();
            libc::sysctlbyname(
                cache_name.as_ptr(),
                &mut cache_pages as *mut _ as *mut libc::c_void,
                &mut cache_len,
                ptr::null_mut(),
                0,
            );
            cache_pages * page_size
        };

        #[cfg(not(target_os = "freebsd"))]
        let cached = 0u64;

        // Get swap info
        #[cfg(target_os = "freebsd")]
        let (swap_total, swap_used) = get_swap_freebsd();
        #[cfg(not(target_os = "freebsd"))]
        let (swap_total, swap_used) = (0u64, 0u64);

        Ok(MemInfo {
            total: physmem,
            available: free_pages * page_size,
            cached,
            buffers: 0,
            swap_total,
            swap_used,
        })
    }
}

#[cfg(target_os = "freebsd")]
fn get_swap_freebsd() -> (u64, u64) {
    unsafe {
        let name = CString::new("vm.swap_info").ok()?;
        let mut len: usize = 0;

        // First call to get size
        if libc::sysctlbyname(name.as_ptr(), ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return (0, 0);
        }

        if len == 0 {
            return (0, 0);
        }

        let mut buf: Vec<u8> = vec![0; len];
        if libc::sysctlbyname(
            name.as_ptr(),
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return (0, 0);
        }

        // Parse swap info (simplified)
        (0, 0)
    }
}

// ============================================================================
// LOAD AVERAGE
// ============================================================================

pub struct LoadAvg {
    pub load_1min: f64,
    pub load_5min: f64,
    pub load_15min: f64,
}

pub fn get_loadavg() -> Result<LoadAvg> {
    unsafe {
        let mut loadavg: [libc::c_double; 3] = [0.0; 3];
        let result = libc::getloadavg(loadavg.as_mut_ptr(), 3);

        if result < 0 {
            return Err(Error::Platform("getloadavg failed".to_string()));
        }

        Ok(LoadAvg { load_1min: loadavg[0], load_5min: loadavg[1], load_15min: loadavg[2] })
    }
}

// ============================================================================
// PROCESS
// ============================================================================

pub struct ProcessInfo {
    pub rss: u64,
    pub vsize: u64,
    pub num_threads: u32,
    pub num_fds: u32,
    pub state: u8,
}

pub fn get_process_info(pid: i32) -> Result<ProcessInfo> {
    unsafe {
        #[cfg(target_os = "freebsd")]
        {
            let mut mib =
                [libc::CTL_KERN, libc::KERN_PROC, libc::KERN_PROC_PID, pid as libc::c_int];

            let mut kinfo: libc::kinfo_proc = mem::zeroed();
            let mut len = mem::size_of::<libc::kinfo_proc>();

            let result = libc::sysctl(
                mib.as_mut_ptr(),
                4,
                &mut kinfo as *mut _ as *mut libc::c_void,
                &mut len,
                ptr::null_mut(),
                0,
            );

            if result != 0 || len == 0 {
                return Err(Error::NotFound(format!("process {} not found", pid)));
            }

            Ok(ProcessInfo {
                rss: kinfo.ki_rssize as u64 * 4096,
                vsize: kinfo.ki_size as u64,
                num_threads: kinfo.ki_numthreads as u32,
                num_fds: kinfo.ki_fd.fd_nfiles as u32,
                state: match kinfo.ki_stat as i32 {
                    SRUN => 1,
                    SSLEEP => 2,
                    SWAIT => 3,
                    SZOMB => 4,
                    SSTOP => 5,
                    _ => 0,
                },
            })
        }

        #[cfg(not(target_os = "freebsd"))]
        {
            // OpenBSD/NetBSD have different kinfo_proc structures
            Err(Error::NotSupported)
        }
    }
}

pub fn list_pids() -> Result<Vec<i32>> {
    unsafe {
        #[cfg(target_os = "freebsd")]
        {
            let mut mib = [libc::CTL_KERN, libc::KERN_PROC, libc::KERN_PROC_ALL, 0];

            let mut len: usize = 0;

            // Get size first
            if libc::sysctl(mib.as_mut_ptr(), 3, ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0
            {
                return Ok(Vec::new());
            }

            let count = len / mem::size_of::<libc::kinfo_proc>();
            let mut kinfos: Vec<libc::kinfo_proc> = vec![mem::zeroed(); count];

            if libc::sysctl(
                mib.as_mut_ptr(),
                3,
                kinfos.as_mut_ptr() as *mut libc::c_void,
                &mut len,
                ptr::null_mut(),
                0,
            ) != 0
            {
                return Ok(Vec::new());
            }

            let actual_count = len / mem::size_of::<libc::kinfo_proc>();
            let pids: Vec<i32> =
                kinfos[..actual_count].iter().map(|k| k.ki_pid).filter(|&p| p > 0).collect();

            Ok(pids)
        }

        #[cfg(not(target_os = "freebsd"))]
        {
            Ok(Vec::new())
        }
    }
}

#[cfg(target_os = "freebsd")]
const SRUN: i32 = 2;
#[cfg(target_os = "freebsd")]
const SSLEEP: i32 = 3;
#[cfg(target_os = "freebsd")]
const SWAIT: i32 = 4;
#[cfg(target_os = "freebsd")]
const SZOMB: i32 = 5;
#[cfg(target_os = "freebsd")]
const SSTOP: i32 = 6;

// ============================================================================
// DISK
// ============================================================================

pub fn get_mounts() -> Result<Vec<Partition>> {
    unsafe {
        #[cfg(target_os = "freebsd")]
        {
            let count = libc::getmntinfo(ptr::null_mut(), libc::MNT_NOWAIT);
            if count <= 0 {
                return Ok(Vec::new());
            }

            let mut fs_list: *mut libc::statfs = ptr::null_mut();
            let actual = libc::getmntinfo(&mut fs_list, libc::MNT_NOWAIT);
            if actual <= 0 || fs_list.is_null() {
                return Ok(Vec::new());
            }

            let mut partitions = Vec::new();
            for i in 0..actual {
                let fs = &*fs_list.add(i as usize);

                let device = cstr_to_string(fs.f_mntfromname.as_ptr());
                let mount_point = cstr_to_string(fs.f_mntonname.as_ptr());
                let fs_type = cstr_to_string(fs.f_fstypename.as_ptr());

                // Skip pseudo filesystems
                if fs_type == "devfs" || fs_type == "tmpfs" || fs_type == "fdescfs" {
                    continue;
                }

                partitions.push(Partition { device, mount_point, fs_type, options: String::new() });
            }

            Ok(partitions)
        }

        #[cfg(not(target_os = "freebsd"))]
        {
            Ok(Vec::new())
        }
    }
}

pub fn get_disk_usage(path: &str) -> Result<DiskUsage> {
    unsafe {
        let c_path = CString::new(path).map_err(|_| Error::Platform("invalid path".to_string()))?;
        let mut stat: libc::statfs = mem::zeroed();

        let result = libc::statfs(c_path.as_ptr(), &mut stat);
        if result != 0 {
            return Err(Error::NotFound(format!("path {} not found", path)));
        }

        let block_size = stat.f_bsize as u64;
        let total = stat.f_blocks as u64 * block_size;
        let free = stat.f_bfree as u64 * block_size;
        let available = stat.f_bavail as u64 * block_size;
        let used = total.saturating_sub(free);

        Ok(DiskUsage {
            path: path.to_string(),
            total_bytes: total,
            used_bytes: used,
            free_bytes: available,
            used_percent: if total > 0 { (used as f64 / total as f64) * 100.0 } else { 0.0 },
            inodes_total: stat.f_files as u64,
            inodes_used: (stat.f_files as u64).saturating_sub(stat.f_ffree as u64),
            inodes_free: stat.f_ffree as u64,
        })
    }
}

// ============================================================================
// DISK I/O STATISTICS
// ============================================================================

/// Gets disk I/O statistics for the current BSD platform.
///
/// # Platform Support
///
/// - **OpenBSD**: Via `sysctl hw.diskstats`
/// - **NetBSD**: Via `sysctl hw.diskstats` or `/proc/diskstats`
/// - **FreeBSD**: Returns empty vector (requires devstat library)
///
/// # Examples
///
/// ```no_run
/// use probe_platform::bsd::sysctl;
///
/// let stats = sysctl::get_disk_io_stats()?;
/// for disk in stats {
///     println!("{}: {} reads, {} writes",
///         disk.device,
///         disk.reads_completed,
///         disk.writes_completed
///     );
/// }
/// # Ok::<(), probe_platform::Error>(())
/// ```
///
/// # Errors
///
/// Returns an error if the platform-specific syscalls fail.
/// Returns an empty vector if no disks are found or stats are unavailable.
pub fn get_disk_io_stats() -> Result<Vec<DiskIOStats>> {
    #[cfg(target_os = "openbsd")]
    {
        openbsd::get_disk_io_stats()
    }

    #[cfg(target_os = "netbsd")]
    {
        netbsd::get_disk_io_stats()
    }

    #[cfg(target_os = "freebsd")]
    {
        // FreeBSD requires libdevstat which needs C bindings
        // TODO: Implement via devstat_getdevs() and devstat_compute_statistics()
        Ok(Vec::new())
    }
}

// ============================================================================
// OPENBSD DISK I/O
// ============================================================================

#[cfg(target_os = "openbsd")]
mod openbsd {
    use super::*;

    /// OpenBSD disk statistics structure (from sys/disk.h).
    ///
    /// This structure matches the kernel's `struct diskstats` definition.
    #[repr(C)]
    #[derive(Debug, Clone, Copy)]
    struct DiskStats {
        /// Device name (e.g., "sd0", "wd0").
        ds_name: [libc::c_char; 16],
        /// Number of busy units.
        ds_busy: i32,
        /// Read operations completed.
        ds_rxfer: u64,
        /// Write operations completed.
        ds_wxfer: u64,
        /// Bytes read.
        ds_rbytes: u64,
        /// Bytes written.
        ds_wbytes: u64,
        /// Attachment timestamp.
        ds_attachtime_sec: i64,
        ds_attachtime_usec: i64,
        /// Total time in operation.
        ds_timestamp_sec: i64,
        ds_timestamp_usec: i64,
        /// Time spent processing requests.
        ds_time_sec: i64,
        ds_time_usec: i64,
    }

    /// Collects disk I/O statistics on OpenBSD via sysctl hw.diskstats.
    ///
    /// # Errors
    ///
    /// Returns an error if the sysctl call fails.
    /// Returns an empty vector if no disk devices are found.
    pub fn get_disk_io_stats() -> Result<Vec<DiskIOStats>> {
        unsafe {
            let name = CString::new("hw.diskstats")
                .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

            // Get required buffer size
            let mut len: usize = 0;
            let result =
                libc::sysctlbyname(name.as_ptr(), ptr::null_mut(), &mut len, ptr::null_mut(), 0);

            if result != 0 || len == 0 {
                return Ok(Vec::new());
            }

            // Allocate buffer
            let count = len / mem::size_of::<DiskStats>();
            let mut stats: Vec<DiskStats> = vec![mem::zeroed(); count];

            let result = libc::sysctlbyname(
                name.as_ptr(),
                stats.as_mut_ptr() as *mut libc::c_void,
                &mut len,
                ptr::null_mut(),
                0,
            );

            if result != 0 {
                return Err(Error::Platform(format!(
                    "sysctl hw.diskstats failed: {}",
                    std::io::Error::last_os_error()
                )));
            }

            let actual_count = len / mem::size_of::<DiskStats>();
            let mut results = Vec::with_capacity(actual_count);

            for disk in &stats[..actual_count] {
                let device = cstr_to_string(disk.ds_name.as_ptr());

                // Skip devices with no activity
                if disk.ds_rxfer == 0 && disk.ds_wxfer == 0 {
                    continue;
                }

                // Convert time from seconds + microseconds to milliseconds
                let time_ms = (disk.ds_time_sec * 1000).saturating_add(disk.ds_time_usec / 1000);

                // OpenBSD provides bytes, convert to sectors (512 bytes)
                let sectors_read = disk.ds_rbytes / 512;
                let sectors_written = disk.ds_wbytes / 512;

                results.push(DiskIOStats {
                    device,
                    reads_completed: disk.ds_rxfer,
                    sectors_read,
                    read_time_ms: (time_ms / 2).max(1) as u64, // Estimate split
                    writes_completed: disk.ds_wxfer,
                    sectors_written,
                    write_time_ms: (time_ms / 2).max(1) as u64, // Estimate split
                    io_in_progress: disk.ds_busy as u64,
                    io_time_ms: time_ms as u64,
                    weighted_io_time_ms: time_ms as u64,
                });
            }

            Ok(results)
        }
    }
}

// ============================================================================
// NETBSD DISK I/O
// ============================================================================

#[cfg(target_os = "netbsd")]
mod netbsd {
    use super::*;

    /// NetBSD disk statistics structure (from sys/disk.h).
    ///
    /// This structure matches the kernel's `struct disk_sysctl` definition.
    #[repr(C)]
    #[derive(Debug, Clone, Copy)]
    struct DiskSysctl {
        /// Device name (e.g., "wd0", "sd0").
        dk_name: [libc::c_char; 16],
        /// Transfer count.
        dk_xfer: u64,
        /// Number of seeks (not used on modern drives).
        dk_seek: u64,
        /// Bytes transferred.
        dk_bytes: u64,
        /// Read operations.
        dk_rxfer: u64,
        /// Write operations.
        dk_wxfer: u64,
        /// Bytes read.
        dk_rbytes: u64,
        /// Bytes written.
        dk_wbytes: u64,
        /// Attachment time.
        dk_attachtime_sec: i64,
        dk_attachtime_usec: u64,
        /// Total operation time.
        dk_timestamp_sec: i64,
        dk_timestamp_usec: u64,
        /// Time spent in operations.
        dk_time_sec: i64,
        dk_time_usec: u64,
        /// Number of busy units.
        dk_busy: i32,
    }

    /// Collects disk I/O statistics on NetBSD.
    ///
    /// Tries multiple methods in order of preference:
    /// 1. sysctl hw.diskstats (preferred)
    /// 2. /proc/diskstats if procfs is mounted
    ///
    /// # Errors
    ///
    /// Returns an empty vector if all methods fail or no disk devices are found.
    pub fn get_disk_io_stats() -> Result<Vec<DiskIOStats>> {
        // Try sysctl first
        match get_disk_io_stats_sysctl() {
            Ok(stats) if !stats.is_empty() => return Ok(stats),
            _ => {}
        }

        // Try procfs as fallback
        match get_disk_io_stats_procfs() {
            Ok(stats) if !stats.is_empty() => return Ok(stats),
            _ => {}
        }

        // Return empty if no method succeeds
        Ok(Vec::new())
    }

    /// Gets disk I/O stats via sysctl hw.diskstats.
    fn get_disk_io_stats_sysctl() -> Result<Vec<DiskIOStats>> {
        unsafe {
            let name = CString::new("hw.diskstats")
                .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

            // Get required buffer size
            let mut len: usize = 0;
            let result =
                libc::sysctlbyname(name.as_ptr(), ptr::null_mut(), &mut len, ptr::null_mut(), 0);

            if result != 0 || len == 0 {
                return Ok(Vec::new());
            }

            // Allocate buffer
            let count = len / mem::size_of::<DiskSysctl>();
            let mut stats: Vec<DiskSysctl> = vec![mem::zeroed(); count];

            let result = libc::sysctlbyname(
                name.as_ptr(),
                stats.as_mut_ptr() as *mut libc::c_void,
                &mut len,
                ptr::null_mut(),
                0,
            );

            if result != 0 {
                return Err(Error::Platform(format!(
                    "sysctl hw.diskstats failed: {}",
                    std::io::Error::last_os_error()
                )));
            }

            let actual_count = len / mem::size_of::<DiskSysctl>();
            let mut results = Vec::with_capacity(actual_count);

            for disk in &stats[..actual_count] {
                let device = cstr_to_string(disk.dk_name.as_ptr());

                // Skip devices with no activity
                if disk.dk_rxfer == 0 && disk.dk_wxfer == 0 {
                    continue;
                }

                // Convert time from seconds + microseconds to milliseconds
                let time_ms =
                    (disk.dk_time_sec * 1000).saturating_add((disk.dk_time_usec / 1000) as i64);

                // NetBSD provides bytes, convert to sectors (512 bytes)
                let sectors_read = disk.dk_rbytes / 512;
                let sectors_written = disk.dk_wbytes / 512;

                results.push(DiskIOStats {
                    device,
                    reads_completed: disk.dk_rxfer,
                    sectors_read,
                    read_time_ms: (time_ms / 2).max(1) as u64,
                    writes_completed: disk.dk_wxfer,
                    sectors_written,
                    write_time_ms: (time_ms / 2).max(1) as u64,
                    io_in_progress: disk.dk_busy as u64,
                    io_time_ms: time_ms as u64,
                    weighted_io_time_ms: time_ms as u64,
                });
            }

            Ok(results)
        }
    }

    /// Gets disk I/O stats from /proc/diskstats if available.
    fn get_disk_io_stats_procfs() -> Result<Vec<DiskIOStats>> {
        use std::fs;
        use std::io::{BufRead, BufReader};

        let file = fs::File::open("/proc/diskstats").map_err(|_| Error::NotSupported)?;

        let reader = BufReader::new(file);
        let mut results = Vec::new();

        for line in reader.lines() {
            let line = line.map_err(Error::Io)?;
            let fields: Vec<&str> = line.split_whitespace().collect();

            // /proc/diskstats format: major minor name reads ... writes ...
            if fields.len() < 14 {
                continue;
            }

            // Parse fields (same format as Linux)
            let device = fields[2].to_string();
            let reads_completed = fields[3].parse().unwrap_or(0);
            let sectors_read = fields[5].parse().unwrap_or(0);
            let read_time_ms = fields[6].parse().unwrap_or(0);
            let writes_completed = fields[7].parse().unwrap_or(0);
            let sectors_written = fields[9].parse().unwrap_or(0);
            let write_time_ms = fields[10].parse().unwrap_or(0);
            let io_in_progress = fields[11].parse().unwrap_or(0);
            let io_time_ms = fields[12].parse().unwrap_or(0);
            let weighted_io_time_ms = fields[13].parse().unwrap_or(0);

            results.push(DiskIOStats {
                device,
                reads_completed,
                sectors_read,
                read_time_ms,
                writes_completed,
                sectors_written,
                write_time_ms,
                io_in_progress,
                io_time_ms,
                weighted_io_time_ms,
            });
        }

        Ok(results)
    }
}

// ============================================================================
// NETWORK
// ============================================================================

pub fn get_network_interfaces() -> Result<Vec<NetInterface>> {
    unsafe {
        let mut addrs: *mut libc::ifaddrs = ptr::null_mut();
        if libc::getifaddrs(&mut addrs) != 0 {
            return Err(Error::Io(std::io::Error::last_os_error()));
        }

        let mut interfaces: std::collections::HashMap<String, NetInterface> =
            std::collections::HashMap::new();

        let mut addr = addrs;
        while !addr.is_null() {
            let ifa = &*addr;
            let name = cstr_to_string(ifa.ifa_name);

            let iface = interfaces.entry(name.clone()).or_insert_with(|| NetInterface {
                name: name.clone(),
                mac_address: String::new(),
                ipv4_addresses: Vec::new(),
                ipv6_addresses: Vec::new(),
                mtu: 0,
                is_up: (ifa.ifa_flags as i32 & libc::IFF_UP) != 0,
                is_loopback: (ifa.ifa_flags as i32 & libc::IFF_LOOPBACK) != 0,
            });

            if !ifa.ifa_addr.is_null() {
                let sa_family = (*ifa.ifa_addr).sa_family as i32;

                if sa_family == libc::AF_INET {
                    let sin = ifa.ifa_addr as *const libc::sockaddr_in;
                    let ip = std::net::Ipv4Addr::from(u32::from_be((*sin).sin_addr.s_addr));
                    iface.ipv4_addresses.push(ip.to_string());
                } else if sa_family == libc::AF_INET6 {
                    let sin6 = ifa.ifa_addr as *const libc::sockaddr_in6;
                    let ip = std::net::Ipv6Addr::from((*sin6).sin6_addr.s6_addr);
                    iface.ipv6_addresses.push(ip.to_string());
                }
            }

            addr = ifa.ifa_next;
        }

        libc::freeifaddrs(addrs);
        Ok(interfaces.into_values().collect())
    }
}

pub fn get_network_stats() -> Result<Vec<NetStats>> {
    // Would need sysctl NET_RT_IFLIST
    Ok(Vec::new())
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

unsafe fn cstr_to_string(ptr: *const libc::c_char) -> String {
    if ptr.is_null() {
        return String::new();
    }
    std::ffi::CStr::from_ptr(ptr).to_string_lossy().into_owned()
}
