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
        let name = CString::new("kern.cp_time").unwrap();
        #[cfg(target_os = "openbsd")]
        let name = CString::new("kern.cp_time").unwrap();
        #[cfg(target_os = "netbsd")]
        let name = CString::new("kern.cp_time").unwrap();

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
            return Ok(CpuTimes {
                user_percent: 0.0,
                system_percent: 0.0,
                idle_percent: 100.0,
            });
        }

        let user = cp_time[0] + cp_time[1]; // user + nice
        let system = cp_time[2] + cp_time[3]; // sys + intr
        let idle = cp_time[4];
        let total = user + system + idle;

        if total == 0 {
            return Ok(CpuTimes {
                user_percent: 0.0,
                system_percent: 0.0,
                idle_percent: 100.0,
            });
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
        let name = CString::new("hw.ncpu").unwrap();
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
        let freq_name = CString::new("hw.clockrate").unwrap();
        #[cfg(target_os = "openbsd")]
        let freq_name = CString::new("hw.cpuspeed").unwrap();
        #[cfg(target_os = "netbsd")]
        let freq_name = CString::new("machdep.tsc_freq").unwrap();

        let mut freq: libc::c_int = 0;
        let mut freq_len = mem::size_of::<libc::c_int>();

        libc::sysctlbyname(
            freq_name.as_ptr(),
            &mut freq as *mut _ as *mut libc::c_void,
            &mut freq_len,
            ptr::null_mut(),
            0,
        );

        Ok(CpuInfo {
            cores: ncpu as u32,
            frequency_mhz: freq as u64,
        })
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
        let name = CString::new("hw.physmem").unwrap();
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
        let free_name = CString::new("vm.stats.vm.v_free_count").unwrap();
        #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
        let free_name = CString::new("vm.uvmexp").unwrap();

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
            let cache_name = CString::new("vm.stats.vm.v_cache_count").unwrap();
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
        let name = CString::new("vm.swap_info").unwrap();
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

        Ok(LoadAvg {
            load_1min: loadavg[0],
            load_5min: loadavg[1],
            load_15min: loadavg[2],
        })
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
            let mut mib = [
                libc::CTL_KERN,
                libc::KERN_PROC,
                libc::KERN_PROC_PID,
                pid as libc::c_int,
            ];

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
            if libc::sysctl(
                mib.as_mut_ptr(),
                3,
                ptr::null_mut(),
                &mut len,
                ptr::null_mut(),
                0,
            ) != 0
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
            let pids: Vec<i32> = kinfos[..actual_count]
                .iter()
                .map(|k| k.ki_pid)
                .filter(|&p| p > 0)
                .collect();

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

                partitions.push(Partition {
                    device,
                    mount_point,
                    fs_type,
                    options: String::new(),
                });
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
            used_percent: if total > 0 {
                (used as f64 / total as f64) * 100.0
            } else {
                0.0
            },
            inodes_total: stat.f_files as u64,
            inodes_used: (stat.f_files as u64).saturating_sub(stat.f_ffree as u64),
            inodes_free: stat.f_ffree as u64,
        })
    }
}

pub fn get_disk_io_stats() -> Result<Vec<DiskIOStats>> {
    // Would need devstat on FreeBSD
    Ok(Vec::new())
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

            let iface = interfaces
                .entry(name.clone())
                .or_insert_with(|| NetInterface {
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
