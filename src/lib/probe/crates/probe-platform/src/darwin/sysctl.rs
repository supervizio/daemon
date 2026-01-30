//! macOS sysctl and Mach API wrappers

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
    // Use host_processor_info for CPU times
    // This is a simplified implementation
    unsafe {
        let mut count: libc::c_uint = 0;
        let mut info: *mut libc::c_int = ptr::null_mut();
        let mut msg_count: libc::c_uint = 0;

        let host = libc::mach_host_self();
        let result = host_processor_info(
            host,
            PROCESSOR_CPU_LOAD_INFO,
            &mut count,
            &mut info,
            &mut msg_count,
        );

        if result != 0 || info.is_null() {
            return Ok(CpuTimes {
                user_percent: 0.0,
                system_percent: 0.0,
                idle_percent: 100.0,
            });
        }

        // Calculate aggregate CPU times
        let mut total_user: u64 = 0;
        let mut total_system: u64 = 0;
        let mut total_idle: u64 = 0;

        for i in 0..count as usize {
            let base = i * CPU_STATE_MAX;
            total_user += (*info.add(base + CPU_STATE_USER as usize)) as u64;
            total_system += (*info.add(base + CPU_STATE_SYSTEM as usize)) as u64;
            total_idle += (*info.add(base + CPU_STATE_IDLE as usize)) as u64;
        }

        let total = total_user + total_system + total_idle;
        if total == 0 {
            return Ok(CpuTimes {
                user_percent: 0.0,
                system_percent: 0.0,
                idle_percent: 100.0,
            });
        }

        // Deallocate
        libc::vm_deallocate(
            libc::mach_task_self(),
            info as libc::vm_address_t,
            (msg_count as usize) * mem::size_of::<libc::c_int>(),
        );

        Ok(CpuTimes {
            user_percent: (total_user as f64 / total as f64) * 100.0,
            system_percent: (total_system as f64 / total as f64) * 100.0,
            idle_percent: (total_idle as f64 / total as f64) * 100.0,
        })
    }
}

pub fn get_cpu_info() -> Result<CpuInfo> {
    unsafe {
        // Get number of CPUs
        let mut mib = [libc::CTL_HW, libc::HW_NCPU];
        let mut ncpu: libc::c_int = 0;
        let mut len = mem::size_of::<libc::c_int>();

        libc::sysctl(
            mib.as_mut_ptr(),
            2,
            &mut ncpu as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        // Get CPU frequency (hw.cpufrequency)
        let name = CString::new("hw.cpufrequency").unwrap();
        let mut freq: u64 = 0;
        let mut freq_len = mem::size_of::<u64>();

        libc::sysctlbyname(
            name.as_ptr(),
            &mut freq as *mut _ as *mut libc::c_void,
            &mut freq_len,
            ptr::null_mut(),
            0,
        );

        Ok(CpuInfo {
            cores: ncpu as u32,
            frequency_mhz: freq / 1_000_000,
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
    pub swap_total: u64,
    pub swap_used: u64,
}

pub fn get_memory_info() -> Result<MemInfo> {
    unsafe {
        // Get total physical memory
        let mut mib = [libc::CTL_HW, libc::HW_MEMSIZE];
        let mut memsize: u64 = 0;
        let mut len = mem::size_of::<u64>();

        libc::sysctl(
            mib.as_mut_ptr(),
            2,
            &mut memsize as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        // Get page size
        let page_size = libc::sysconf(libc::_SC_PAGESIZE) as u64;

        // Get VM statistics via host_statistics64
        let host = libc::mach_host_self();
        let mut vm_stat: vm_statistics64 = mem::zeroed();
        let mut count = (mem::size_of::<vm_statistics64>() / mem::size_of::<libc::c_int>()) as u32;

        let result = host_statistics64(
            host,
            HOST_VM_INFO64,
            &mut vm_stat as *mut _ as *mut libc::c_int,
            &mut count,
        );

        let (available, cached) = if result == 0 {
            let free = vm_stat.free_count as u64 * page_size;
            let inactive = vm_stat.inactive_count as u64 * page_size;
            let cached = vm_stat.external_page_count as u64 * page_size;
            (free + inactive, cached)
        } else {
            (memsize / 4, 0) // Fallback
        };

        // Get swap info via sysctl
        let name = CString::new("vm.swapusage").unwrap();
        let mut swap: xsw_usage = mem::zeroed();
        let mut swap_len = mem::size_of::<xsw_usage>();

        let swap_result = libc::sysctlbyname(
            name.as_ptr(),
            &mut swap as *mut _ as *mut libc::c_void,
            &mut swap_len,
            ptr::null_mut(),
            0,
        );

        let (swap_total, swap_used) = if swap_result == 0 {
            (swap.xsu_total, swap.xsu_used)
        } else {
            (0, 0)
        };

        Ok(MemInfo {
            total: memsize,
            available,
            cached,
            swap_total,
            swap_used,
        })
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

        // Count file descriptors using proc_pidinfo
        let num_fds = proc_pidinfo_fdcount(pid);

        Ok(ProcessInfo {
            rss: kinfo.kp_eproc.e_xrssize as u64 * 4096,
            vsize: kinfo.kp_eproc.e_vm.vm_map as u64,
            num_threads: 1, // Would need task_info for accurate count
            num_fds,
            state: match kinfo.kp_proc.p_stat as i32 {
                SIDL => 1,
                SRUN => 1,
                SSLEEP => 2,
                SSTOP => 5,
                SZOMB => 4,
                _ => 0,
            },
        })
    }
}

pub fn list_pids() -> Result<Vec<i32>> {
    unsafe {
        // Get number of processes
        let count = libc::proc_listallpids(ptr::null_mut(), 0);
        if count <= 0 {
            return Ok(Vec::new());
        }

        let mut pids: Vec<i32> = vec![0; count as usize];
        let actual = libc::proc_listallpids(
            pids.as_mut_ptr() as *mut libc::c_void,
            (count as usize * mem::size_of::<i32>()) as i32,
        );

        if actual <= 0 {
            return Ok(Vec::new());
        }

        pids.truncate((actual as usize) / mem::size_of::<i32>());
        pids.retain(|&p| p > 0);
        Ok(pids)
    }
}

fn proc_pidinfo_fdcount(pid: i32) -> u32 {
    unsafe {
        let size = libc::proc_pidinfo(pid, PROC_PIDLISTFDS, 0, ptr::null_mut(), 0);
        if size > 0 {
            (size as usize / mem::size_of::<proc_fdinfo>()) as u32
        } else {
            0
        }
    }
}

// ============================================================================
// DISK
// ============================================================================

pub fn get_mounts() -> Result<Vec<Partition>> {
    unsafe {
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
            if device.starts_with("devfs")
                || device.starts_with("map ")
                || mount_point.starts_with("/System/Volumes/")
            {
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
    // IOKit would be needed for detailed stats
    // Return empty for now - this requires significant IOKit bindings
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
                } else if sa_family == libc::AF_LINK {
                    // Get MAC address from AF_LINK
                    let sdl = ifa.ifa_addr as *const sockaddr_dl;
                    if (*sdl).sdl_alen == 6 {
                        let mac_ptr = (sdl as *const u8).add(mem::size_of::<sockaddr_dl>() - 12);
                        let mac = format!(
                            "{:02x}:{:02x}:{:02x}:{:02x}:{:02x}:{:02x}",
                            *mac_ptr,
                            *mac_ptr.add(1),
                            *mac_ptr.add(2),
                            *mac_ptr.add(3),
                            *mac_ptr.add(4),
                            *mac_ptr.add(5)
                        );
                        iface.mac_address = mac;
                    }
                }
            }

            addr = ifa.ifa_next;
        }

        libc::freeifaddrs(addrs);
        Ok(interfaces.into_values().collect())
    }
}

pub fn get_network_stats() -> Result<Vec<NetStats>> {
    // Use sysctl for interface statistics
    unsafe {
        let mut mib = [libc::CTL_NET, libc::PF_ROUTE, 0, 0, libc::NET_RT_IFLIST2, 0];

        let mut len: usize = 0;
        if libc::sysctl(
            mib.as_mut_ptr(),
            6,
            ptr::null_mut(),
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Err(Error::Io(std::io::Error::last_os_error()));
        }

        let mut buf: Vec<u8> = vec![0; len];
        if libc::sysctl(
            mib.as_mut_ptr(),
            6,
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Err(Error::Io(std::io::Error::last_os_error()));
        }

        let mut stats = Vec::new();
        let mut offset = 0;

        while offset < len {
            let ifm = buf.as_ptr().add(offset) as *const if_msghdr;
            let msg_len = (*ifm).ifm_msglen as usize;

            if (*ifm).ifm_type as i32 == RTM_IFINFO2 {
                let ifm2 = ifm as *const if_msghdr2;
                let data = &(*ifm2).ifm_data;

                // Get interface name from index
                let mut ifname = [0i8; libc::IF_NAMESIZE];
                if !libc::if_indextoname((*ifm2).ifm_index as u32, ifname.as_mut_ptr()).is_null() {
                    let name = cstr_to_string(ifname.as_ptr());

                    stats.push(NetStats {
                        interface: name,
                        rx_bytes: data.ifi_ibytes,
                        rx_packets: data.ifi_ipackets,
                        rx_errors: data.ifi_ierrors,
                        rx_drops: data.ifi_iqdrops,
                        tx_bytes: data.ifi_obytes,
                        tx_packets: data.ifi_opackets,
                        tx_errors: data.ifi_oerrors,
                        tx_drops: 0,
                    });
                }
            }

            offset += msg_len;
        }

        Ok(stats)
    }
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

// ============================================================================
// MACH AND SYSTEM TYPES
// ============================================================================

const PROCESSOR_CPU_LOAD_INFO: libc::c_int = 2;
const CPU_STATE_USER: libc::c_int = 0;
const CPU_STATE_SYSTEM: libc::c_int = 1;
const CPU_STATE_IDLE: libc::c_int = 2;
const CPU_STATE_MAX: usize = 4;

const HOST_VM_INFO64: libc::c_int = 4;

const SIDL: i32 = 1;
const SRUN: i32 = 2;
const SSLEEP: i32 = 3;
const SSTOP: i32 = 4;
const SZOMB: i32 = 5;

const PROC_PIDLISTFDS: libc::c_int = 1;

const RTM_IFINFO2: i32 = 0x12;

#[repr(C)]
struct vm_statistics64 {
    free_count: u64,
    active_count: u64,
    inactive_count: u64,
    wire_count: u64,
    zero_fill_count: u64,
    reactivations: u64,
    pageins: u64,
    pageouts: u64,
    faults: u64,
    cow_faults: u64,
    lookups: u64,
    hits: u64,
    purges: u64,
    purgeable_count: u64,
    speculative_count: u64,
    decompressions: u64,
    compressions: u64,
    swapins: u64,
    swapouts: u64,
    compressor_page_count: u64,
    throttled_count: u64,
    external_page_count: u64,
    internal_page_count: u64,
    total_uncompressed_pages_in_compressor: u64,
}

#[repr(C)]
struct xsw_usage {
    xsu_total: u64,
    xsu_avail: u64,
    xsu_used: u64,
    xsu_pagesize: u32,
    xsu_encrypted: bool,
}

#[repr(C)]
struct sockaddr_dl {
    sdl_len: u8,
    sdl_family: u8,
    sdl_index: u16,
    sdl_type: u8,
    sdl_nlen: u8,
    sdl_alen: u8,
    sdl_slen: u8,
    sdl_data: [libc::c_char; 12],
}

#[repr(C)]
struct proc_fdinfo {
    proc_fd: i32,
    proc_fdtype: u32,
}

#[repr(C)]
struct if_data64 {
    ifi_type: u8,
    ifi_typelen: u8,
    ifi_physical: u8,
    ifi_addrlen: u8,
    ifi_hdrlen: u8,
    ifi_recvquota: u8,
    ifi_xmitquota: u8,
    ifi_unused1: u8,
    ifi_mtu: u32,
    ifi_metric: u32,
    ifi_baudrate: u64,
    ifi_ipackets: u64,
    ifi_ierrors: u64,
    ifi_opackets: u64,
    ifi_oerrors: u64,
    ifi_collisions: u64,
    ifi_ibytes: u64,
    ifi_obytes: u64,
    ifi_imcasts: u64,
    ifi_omcasts: u64,
    ifi_iqdrops: u64,
    ifi_noproto: u64,
    ifi_recvtiming: u32,
    ifi_xmittiming: u32,
    ifi_lastchange: libc::timeval32,
}

#[repr(C)]
struct if_msghdr {
    ifm_msglen: u16,
    ifm_version: u8,
    ifm_type: u8,
    ifm_addrs: i32,
    ifm_flags: i32,
    ifm_index: u16,
    ifm_data: if_data64,
}

#[repr(C)]
struct if_msghdr2 {
    ifm_msglen: u16,
    ifm_version: u8,
    ifm_type: u8,
    ifm_addrs: i32,
    ifm_flags: i32,
    ifm_index: u16,
    ifm_snd_len: i32,
    ifm_snd_maxlen: i32,
    ifm_snd_drops: i32,
    ifm_timer: i32,
    ifm_data: if_data64,
}

// External Mach functions
extern "C" {
    fn host_processor_info(
        host: libc::mach_port_t,
        flavor: libc::c_int,
        out_processor_count: *mut libc::c_uint,
        out_processor_info: *mut *mut libc::c_int,
        out_processor_infoCnt: *mut libc::c_uint,
    ) -> libc::c_int;

    fn host_statistics64(
        host: libc::mach_port_t,
        flavor: libc::c_int,
        host_info_out: *mut libc::c_int,
        host_info_outCnt: *mut u32,
    ) -> libc::c_int;
}
