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
            return Ok(CpuTimes { user_percent: 0.0, system_percent: 0.0, idle_percent: 100.0 });
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
            return Ok(CpuTimes { user_percent: 0.0, system_percent: 0.0, idle_percent: 100.0 });
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

        Ok(CpuInfo { cores: ncpu as u32, frequency_mhz: freq / 1_000_000 })
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

        let result = libc::sysctl(
            mib.as_mut_ptr(),
            2,
            &mut memsize as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        if result != 0 || memsize == 0 {
            return Err(Error::Platform("failed to get total memory".to_string()));
        }

        // Get page size - must check for error (-1)
        let page_size_raw = libc::sysconf(libc::_SC_PAGESIZE);
        if page_size_raw <= 0 {
            return Err(Error::Platform("failed to get page size".to_string()));
        }
        let page_size = page_size_raw as u64;

        // Get memory stats via host_statistics (canonical macOS API)
        let host = libc::mach_host_self();
        let mut vm_stat: vm_statistics = mem::zeroed();
        let mut count = (mem::size_of::<vm_statistics>() / mem::size_of::<u32>()) as u32;

        let result = host_statistics(
            host,
            HOST_VM_INFO,
            &mut vm_stat as *mut _ as *mut libc::c_int,
            &mut count,
        );

        let available = if result == 0 {
            // Available = free + inactive + speculative + purgeable (all can be reclaimed)
            let free = u64::from(vm_stat.free_count);
            let inactive = u64::from(vm_stat.inactive_count);
            let speculative = u64::from(vm_stat.speculative_count);
            let purgeable = u64::from(vm_stat.purgeable_count);

            // Use checked arithmetic to prevent overflow
            let total_pages =
                free.saturating_add(inactive).saturating_add(speculative).saturating_add(purgeable);

            total_pages.saturating_mul(page_size)
        } else {
            // Fallback: estimate available as 10% of total (conservative)
            memsize / 10
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

        let (swap_total, swap_used) =
            if swap_result == 0 { (swap.xsu_total, swap.xsu_used) } else { (0, 0) };

        Ok(MemInfo { total: memsize, available, cached: 0, swap_total, swap_used })
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
        // Use proc_pidinfo with PROC_PIDTBSDINFO for modern macOS API
        let mut bsd_info: ProcBsdInfo = mem::zeroed();
        let size = libc::proc_pidinfo(
            pid,
            PROC_PIDTBSDINFO,
            0,
            &mut bsd_info as *mut _ as *mut libc::c_void,
            mem::size_of::<ProcBsdInfo>() as i32,
        );

        if size <= 0 {
            return Err(Error::NotFound(format!("process {} not found", pid)));
        }

        // Get task info for memory stats
        let mut task_info: ProcTaskInfo = mem::zeroed();
        let task_size = libc::proc_pidinfo(
            pid,
            PROC_PIDTASKINFO,
            0,
            &mut task_info as *mut _ as *mut libc::c_void,
            mem::size_of::<ProcTaskInfo>() as i32,
        );

        let (rss, vsize, num_threads) = if task_size > 0 {
            (
                task_info.pti_resident_size,
                task_info.pti_virtual_size,
                task_info.pti_threadnum as u32,
            )
        } else {
            (0, 0, 1)
        };

        // Count file descriptors using proc_pidinfo
        let num_fds = proc_pidinfo_fdcount(pid);

        Ok(ProcessInfo {
            rss,
            vsize,
            num_threads,
            num_fds,
            state: match bsd_info.pbi_status {
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
        if size > 0 { (size as usize / mem::size_of::<proc_fdinfo>()) as u32 } else { 0 }
    }
}

// ============================================================================
// DISK
// ============================================================================

pub fn get_mounts() -> Result<Vec<Partition>> {
    unsafe {
        let mut fs_list: *mut libc::statfs = ptr::null_mut();
        let count = libc::getmntinfo(&mut fs_list, libc::MNT_NOWAIT);
        if count <= 0 || fs_list.is_null() {
            return Ok(Vec::new());
        }

        let mut partitions = Vec::new();
        for i in 0..count {
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

            partitions.push(Partition { device, mount_point, fs_type, options: String::new() });
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
            used_percent: if total > 0 { (used as f64 / total as f64) * 100.0 } else { 0.0 },
            inodes_total: stat.f_files as u64,
            inodes_used: (stat.f_files as u64).saturating_sub(stat.f_ffree as u64),
            inodes_free: stat.f_ffree as u64,
        })
    }
}

/// Gets disk I/O statistics on macOS via IOKit.
///
/// Uses IOKit's IOServiceGetMatchingServices to enumerate disk devices
/// and IORegistryEntryCreateCFProperties to read statistics.
///
/// # Examples
///
/// ```no_run
/// use probe_platform::darwin::sysctl::get_disk_io_stats;
///
/// let stats = get_disk_io_stats()?;
/// for disk in stats {
///     println!("{}: {} reads, {} writes", disk.device, disk.reads_completed, disk.writes_completed);
/// }
/// # Ok::<(), probe_platform::Error>(())
/// ```
pub fn get_disk_io_stats() -> Result<Vec<DiskIOStats>> {
    unsafe {
        // Get the IOKit master port
        let mut master_port: libc::mach_port_t = 0;
        if IOMainPort(libc::mach_host_self(), &mut master_port) != 0 {
            // Fallback to deprecated IOMasterPort for older macOS versions
            if IOMasterPort(libc::mach_host_self(), &mut master_port) != 0 {
                return Ok(Vec::new());
            }
        }

        // Create matching dictionary for IOBlockStorageDriver
        let matching = IOServiceMatching(b"IOBlockStorageDriver\0".as_ptr() as *const libc::c_char);
        if matching.is_null() {
            return Ok(Vec::new());
        }

        // Get iterator for matching services
        let mut iterator: u32 = 0;
        if IOServiceGetMatchingServices(master_port, matching, &mut iterator) != 0 {
            return Ok(Vec::new());
        }

        let mut results = Vec::new();

        // Iterate over all disk drivers
        loop {
            let service = IOIteratorNext(iterator);
            if service == 0 {
                break;
            }

            // Get the disk's parent to find the device name
            let mut parent: u32 = 0;
            if IORegistryEntryGetParentEntry(
                service,
                b"IOService\0".as_ptr() as *const libc::c_char,
                &mut parent,
            ) == 0
            {
                // Get device name from parent
                let mut name_buf = [0i8; 128];
                if IORegistryEntryGetName(parent, name_buf.as_mut_ptr()) == 0 {
                    let device_name = cstr_to_string(name_buf.as_ptr());

                    // Get statistics dictionary from the driver
                    let mut props: *mut libc::c_void = ptr::null_mut();
                    if IORegistryEntryCreateCFProperties(
                        service,
                        &mut props as *mut *mut libc::c_void as *mut _,
                        std::ptr::null(),
                        0,
                    ) == 0
                        && !props.is_null()
                    {
                        // Extract statistics from the dictionary
                        let stats_key = CFStringCreateWithCString(
                            std::ptr::null(),
                            b"Statistics\0".as_ptr() as *const libc::c_char,
                            0x08000100, // kCFStringEncodingUTF8
                        );

                        if !stats_key.is_null() {
                            let stats_dict =
                                CFDictionaryGetValue(props as *const _, stats_key as *const _);
                            if !stats_dict.is_null() {
                                let stats = parse_iokit_disk_stats(stats_dict, &device_name);
                                if stats.reads_completed > 0 || stats.writes_completed > 0 {
                                    results.push(stats);
                                }
                            }
                            CFRelease(stats_key as *const _);
                        }

                        CFRelease(props);
                    }
                }
                IOObjectRelease(parent);
            }

            IOObjectRelease(service);
        }

        IOObjectRelease(iterator);

        Ok(results)
    }
}

/// Parse IOKit disk statistics dictionary.
unsafe fn parse_iokit_disk_stats(stats_dict: *const libc::c_void, device: &str) -> DiskIOStats {
    let mut stats = DiskIOStats {
        device: device.to_string(),
        reads_completed: 0,
        read_bytes: 0,
        read_time_us: 0,
        writes_completed: 0,
        write_bytes: 0,
        write_time_us: 0,
        io_in_progress: 0,
        io_time_us: 0,
        weighted_io_time_us: 0,
    };

    // Key strings for statistics
    let keys = [
        (b"Operations (Read)\0".as_ptr(), &mut stats.reads_completed as *mut u64),
        (b"Operations (Write)\0".as_ptr(), &mut stats.writes_completed as *mut u64),
        (b"Bytes (Read)\0".as_ptr(), &mut stats.read_bytes as *mut u64),
        (b"Bytes (Write)\0".as_ptr(), &mut stats.write_bytes as *mut u64),
        (b"Total Time (Read)\0".as_ptr(), &mut stats.read_time_us as *mut u64),
        (b"Total Time (Write)\0".as_ptr(), &mut stats.write_time_us as *mut u64),
    ];

    for (key_bytes, value_ptr) in keys {
        // SAFETY: Calling CoreFoundation functions with valid parameters
        let key = unsafe {
            CFStringCreateWithCString(
                std::ptr::null(),
                key_bytes as *const libc::c_char,
                0x08000100,
            )
        };
        if key.is_null() {
            continue;
        }

        // SAFETY: stats_dict is a valid CFDictionary, key is a valid CFString
        let value = unsafe { CFDictionaryGetValue(stats_dict, key as *const _) };
        if !value.is_null() {
            let mut num: i64 = 0;
            // SAFETY: value is a valid CFNumber, num is a valid pointer
            if unsafe { CFNumberGetValue(value, 4, &mut num as *mut _ as *mut _) } {
                // 4 = kCFNumberSInt64Type
                // SAFETY: value_ptr points to a field in stats which is valid
                unsafe { *value_ptr = num as u64 };
            }
        }
        // SAFETY: key is a valid CF object that we created above
        unsafe { CFRelease(key as *const _) };
    }

    // Convert nanoseconds to microseconds
    stats.read_time_us /= 1_000;
    stats.write_time_us /= 1_000;

    // Calculate total IO time
    stats.io_time_us = stats.read_time_us + stats.write_time_us;
    stats.weighted_io_time_us = stats.io_time_us;

    stats
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
        if libc::sysctl(mib.as_mut_ptr(), 6, ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
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
// NETWORK CONNECTIONS
// ============================================================================

/// Network connection state.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ConnectionState {
    /// Connection is closed.
    Closed,
    /// Listening for connections.
    Listen,
    /// SYN sent, awaiting SYN-ACK.
    SynSent,
    /// SYN received, awaiting ACK.
    SynReceived,
    /// Connection established.
    Established,
    /// Received FIN, waiting for close.
    CloseWait,
    /// FIN sent, awaiting ACK.
    FinWait1,
    /// FIN sent and ACKed, awaiting peer FIN.
    FinWait2,
    /// Both sides sent FIN.
    Closing,
    /// Awaiting final ACK of our FIN.
    LastAck,
    /// Waiting for old packets to expire.
    TimeWait,
    /// Unknown state.
    Unknown,
}

impl std::fmt::Display for ConnectionState {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ConnectionState::Closed => write!(f, "CLOSED"),
            ConnectionState::Listen => write!(f, "LISTEN"),
            ConnectionState::SynSent => write!(f, "SYN_SENT"),
            ConnectionState::SynReceived => write!(f, "SYN_RCVD"),
            ConnectionState::Established => write!(f, "ESTABLISHED"),
            ConnectionState::CloseWait => write!(f, "CLOSE_WAIT"),
            ConnectionState::FinWait1 => write!(f, "FIN_WAIT_1"),
            ConnectionState::FinWait2 => write!(f, "FIN_WAIT_2"),
            ConnectionState::Closing => write!(f, "CLOSING"),
            ConnectionState::LastAck => write!(f, "LAST_ACK"),
            ConnectionState::TimeWait => write!(f, "TIME_WAIT"),
            ConnectionState::Unknown => write!(f, "UNKNOWN"),
        }
    }
}

/// Network connection protocol.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum ConnectionProtocol {
    /// TCP connection.
    Tcp,
    /// UDP socket.
    Udp,
    /// Unix domain socket (stream).
    UnixStream,
    /// Unix domain socket (datagram).
    UnixDgram,
}

impl std::fmt::Display for ConnectionProtocol {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        match self {
            ConnectionProtocol::Tcp => write!(f, "tcp"),
            ConnectionProtocol::Udp => write!(f, "udp"),
            ConnectionProtocol::UnixStream => write!(f, "unix-stream"),
            ConnectionProtocol::UnixDgram => write!(f, "unix-dgram"),
        }
    }
}

/// A network connection or socket.
#[derive(Debug, Clone)]
pub struct NetworkConnection {
    /// Protocol (TCP, UDP, Unix).
    pub protocol: ConnectionProtocol,
    /// Local address (IP:port or path).
    pub local_addr: String,
    /// Remote address (IP:port or path), empty for listening sockets.
    pub remote_addr: String,
    /// Connection state (TCP only).
    pub state: ConnectionState,
    /// Process ID owning this connection.
    pub pid: i32,
}

/// Lists all network connections on the system.
///
/// Uses libproc to enumerate sockets for all processes.
///
/// # Examples
///
/// ```no_run
/// use probe_platform::darwin::sysctl::list_network_connections;
///
/// let connections = list_network_connections()?;
/// for conn in connections {
///     println!("{} {} -> {} ({})", conn.protocol, conn.local_addr, conn.remote_addr, conn.state);
/// }
/// # Ok::<(), probe_platform::Error>(())
/// ```
pub fn list_network_connections() -> Result<Vec<NetworkConnection>> {
    let mut connections = Vec::new();

    // Get list of all PIDs
    let pids = list_pids()?;

    for pid in pids {
        if let Ok(mut conns) = list_process_connections(pid) {
            connections.append(&mut conns);
        }
    }

    Ok(connections)
}

/// List network connections for a specific process.
fn list_process_connections(pid: i32) -> Result<Vec<NetworkConnection>> {
    unsafe {
        // First, get the buffer size needed
        let buffer_size = libc::proc_pidinfo(pid, PROC_PIDLISTFDS, 0, ptr::null_mut(), 0);

        if buffer_size <= 0 {
            return Ok(Vec::new());
        }

        // Allocate buffer for file descriptors
        let fd_count = buffer_size as usize / mem::size_of::<proc_fdinfo>();
        let mut fds: Vec<proc_fdinfo> = vec![mem::zeroed(); fd_count];

        let actual_size = libc::proc_pidinfo(
            pid,
            PROC_PIDLISTFDS,
            0,
            fds.as_mut_ptr() as *mut libc::c_void,
            buffer_size,
        );

        if actual_size <= 0 {
            return Ok(Vec::new());
        }

        let actual_count = actual_size as usize / mem::size_of::<proc_fdinfo>();
        let mut connections = Vec::new();

        for i in 0..actual_count {
            let fd_info = &fds[i];

            // Only process socket file descriptors
            if fd_info.proc_fdtype != PROX_FDTYPE_SOCKET {
                continue;
            }

            // Get socket info
            let mut socket_info: socket_fdinfo = mem::zeroed();
            let info_size = libc::proc_pidfdinfo(
                pid,
                fd_info.proc_fd,
                PROC_PIDFDSOCKETINFO,
                &mut socket_info as *mut _ as *mut libc::c_void,
                mem::size_of::<socket_fdinfo>() as i32,
            );

            if info_size <= 0 {
                continue;
            }

            // Parse socket info based on family
            let family = socket_info.psi.soi_family as i32;
            let socket_type = socket_info.psi.soi_type as i32;
            let protocol = socket_info.psi.soi_protocol as i32;

            if family == libc::AF_INET || family == libc::AF_INET6 {
                if protocol == libc::IPPROTO_TCP {
                    // TCP connection
                    let tcp_info = &socket_info.psi.soi_proto.pri_tcp;
                    let state = tcp_state_from_int(tcp_info.tcpsi_state as i32);

                    let (local_addr, remote_addr) = if family == libc::AF_INET {
                        parse_inet4_addrs(&socket_info.psi.soi_proto.pri_in)
                    } else {
                        parse_inet6_addrs(&socket_info.psi.soi_proto.pri_in)
                    };

                    connections.push(NetworkConnection {
                        protocol: ConnectionProtocol::Tcp,
                        local_addr,
                        remote_addr,
                        state,
                        pid,
                    });
                } else if protocol == libc::IPPROTO_UDP {
                    // UDP socket
                    let (local_addr, remote_addr) = if family == libc::AF_INET {
                        parse_inet4_addrs(&socket_info.psi.soi_proto.pri_in)
                    } else {
                        parse_inet6_addrs(&socket_info.psi.soi_proto.pri_in)
                    };

                    connections.push(NetworkConnection {
                        protocol: ConnectionProtocol::Udp,
                        local_addr,
                        remote_addr,
                        state: ConnectionState::Established,
                        pid,
                    });
                }
            } else if family == libc::AF_UNIX {
                // Unix socket
                let protocol = if socket_type == libc::SOCK_STREAM {
                    ConnectionProtocol::UnixStream
                } else {
                    ConnectionProtocol::UnixDgram
                };

                connections.push(NetworkConnection {
                    protocol,
                    local_addr: String::new(), // Unix socket paths require additional parsing
                    remote_addr: String::new(),
                    state: ConnectionState::Established,
                    pid,
                });
            }
        }

        Ok(connections)
    }
}

/// Parse IPv4 addresses from socket info.
fn parse_inet4_addrs(in_info: &in_sockinfo) -> (String, String) {
    // SAFETY: We know the socket family is AF_INET, so accessing ina_46 is valid
    let (local_ip, remote_ip) = unsafe {
        let local =
            std::net::Ipv4Addr::from(u32::from_be(in_info.insi_laddr.ina_46.i46a_addr4.s_addr));
        let remote =
            std::net::Ipv4Addr::from(u32::from_be(in_info.insi_faddr.ina_46.i46a_addr4.s_addr));
        (local, remote)
    };
    let local_port = u16::from_be(in_info.insi_lport as u16);
    let remote_port = u16::from_be(in_info.insi_fport as u16);

    let local_addr = format!("{}:{}", local_ip, local_port);
    let remote_addr =
        if remote_port > 0 { format!("{}:{}", remote_ip, remote_port) } else { String::new() };

    (local_addr, remote_addr)
}

/// Parse IPv6 addresses from socket info.
fn parse_inet6_addrs(in_info: &in_sockinfo) -> (String, String) {
    // SAFETY: We know the socket family is AF_INET6, so accessing ina_6 is valid
    let (local_ip, remote_ip) = unsafe {
        let local = std::net::Ipv6Addr::from(in_info.insi_laddr.ina_6.s6_addr);
        let remote = std::net::Ipv6Addr::from(in_info.insi_faddr.ina_6.s6_addr);
        (local, remote)
    };
    let local_port = u16::from_be(in_info.insi_lport as u16);
    let remote_port = u16::from_be(in_info.insi_fport as u16);

    let local_addr = format!("[{}]:{}", local_ip, local_port);
    let remote_addr =
        if remote_port > 0 { format!("[{}]:{}", remote_ip, remote_port) } else { String::new() };

    (local_addr, remote_addr)
}

/// Convert TCP state to ConnectionState.
fn tcp_state_from_int(state: i32) -> ConnectionState {
    match state {
        0 => ConnectionState::Closed,
        1 => ConnectionState::Listen,
        2 => ConnectionState::SynSent,
        3 => ConnectionState::SynReceived,
        4 => ConnectionState::Established,
        5 => ConnectionState::CloseWait,
        6 => ConnectionState::FinWait1,
        7 => ConnectionState::Closing,
        8 => ConnectionState::LastAck,
        9 => ConnectionState::FinWait2,
        10 => ConnectionState::TimeWait,
        _ => ConnectionState::Unknown,
    }
}

// libproc socket structures
const PROX_FDTYPE_SOCKET: u32 = 2;
const PROC_PIDFDSOCKETINFO: i32 = 3;

#[repr(C)]
struct socket_fdinfo {
    pfi: proc_fileinfo,
    psi: socket_info,
}

#[repr(C)]
struct proc_fileinfo {
    fi_openflags: u32,
    fi_status: u32,
    fi_offset: i64,
    fi_type: i32,
    fi_guardflags: u32,
}

#[repr(C)]
#[derive(Clone, Copy)]
struct socket_info {
    soi_stat: vinfo_stat,
    soi_so: u64,
    soi_pcb: u64,
    soi_type: i32,
    soi_protocol: i32,
    soi_family: i32,
    soi_options: i16,
    soi_linger: i16,
    soi_state: i16,
    soi_qlen: i16,
    soi_incqlen: i16,
    soi_qlimit: i16,
    soi_timeo: i16,
    soi_error: u16,
    soi_oobmark: u32,
    soi_rcv: sockbuf_info,
    soi_snd: sockbuf_info,
    soi_kind: i32,
    _padding: u32,
    soi_proto: proto_info,
}

#[repr(C)]
#[derive(Clone, Copy)]
struct vinfo_stat {
    vst_dev: u32,
    vst_mode: u16,
    vst_nlink: u16,
    vst_ino: u64,
    vst_uid: u32,
    vst_gid: u32,
    vst_atime: i64,
    vst_atimensec: i64,
    vst_mtime: i64,
    vst_mtimensec: i64,
    vst_ctime: i64,
    vst_ctimensec: i64,
    vst_birthtime: i64,
    vst_birthtimensec: i64,
    vst_size: i64,
    vst_blocks: i64,
    vst_blksize: i32,
    vst_flags: u32,
    vst_gen: u32,
    vst_rdev: u32,
    vst_qspare: [i64; 2],
}

#[repr(C)]
#[derive(Clone, Copy)]
struct sockbuf_info {
    sbi_cc: u32,
    sbi_hiwat: u32,
    sbi_mbcnt: u32,
    sbi_mbmax: u32,
    sbi_lowat: u32,
    sbi_flags: i16,
    sbi_timeo: i16,
}

#[repr(C)]
#[derive(Clone, Copy)]
union proto_info {
    pri_in: in_sockinfo,
    pri_tcp: tcp_sockinfo,
    pri_un: un_sockinfo,
    pri_ndrv: ndrv_info,
    pri_kern_event: kern_event_info,
    pri_kern_ctl: kern_ctl_info,
}

#[repr(C)]
#[derive(Clone, Copy)]
struct in_sockinfo {
    insi_fport: i32,
    insi_lport: i32,
    insi_gencnt: u64,
    insi_flags: u32,
    insi_flow: u32,
    insi_vflag: u8,
    insi_ip_ttl: u8,
    _padding: [u8; 2],
    insi_faddr: in_addr_union,
    insi_laddr: in_addr_union,
    insi_v4: in4in6_addr,
    insi_v6: in4in6_addr,
}

#[repr(C)]
#[derive(Clone, Copy)]
union in_addr_union {
    ina_46: in4in6_addr,
    ina_6: libc::in6_addr,
}

#[repr(C)]
#[derive(Clone, Copy)]
struct in4in6_addr {
    i46a_pad32: [u32; 3],
    i46a_addr4: libc::in_addr,
}

#[repr(C)]
#[derive(Clone, Copy)]
struct tcp_sockinfo {
    tcpsi_ini: in_sockinfo,
    tcpsi_state: i32,
    tcpsi_timer: [i32; 4],
    tcpsi_mss: i32,
    tcpsi_flags: u32,
    _padding1: u32,
    tcpsi_tp: u64,
}

#[repr(C)]
#[derive(Clone, Copy)]
struct un_sockinfo {
    unsi_conn_so: u64,
    unsi_conn_pcb: u64,
    unsi_addr: un_addr,
    unsi_caddr: un_addr,
}

#[repr(C)]
#[derive(Clone, Copy)]
struct un_addr {
    ua_sun: libc::sockaddr_un,
    ua_dummy: [i8; 16],
}

#[repr(C)]
#[derive(Clone, Copy)]
struct ndrv_info {
    ndrvsi_if_family: u32,
    ndrvsi_if_unit: u32,
    ndrvsi_if_name: [libc::c_char; 16],
}

#[repr(C)]
#[derive(Clone, Copy)]
struct kern_event_info {
    kesi_vendor_code_filter: u32,
    kesi_class_filter: u32,
    kesi_subclass_filter: u32,
}

#[repr(C)]
#[derive(Clone, Copy)]
struct kern_ctl_info {
    kcsi_id: u32,
    kcsi_reg_unit: u32,
    kcsi_flags: u32,
    kcsi_recvbufsize: u32,
    kcsi_sendbufsize: u32,
    kcsi_unit: u32,
    kcsi_name: [libc::c_char; 96],
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

unsafe fn cstr_to_string(ptr: *const libc::c_char) -> String {
    if ptr.is_null() {
        return String::new();
    }
    // SAFETY: Caller guarantees ptr is valid and null-terminated
    unsafe { std::ffi::CStr::from_ptr(ptr).to_string_lossy().into_owned() }
}

// ============================================================================
// MACH AND SYSTEM TYPES
// ============================================================================

const PROCESSOR_CPU_LOAD_INFO: libc::c_int = 2;
const CPU_STATE_USER: libc::c_int = 0;
const CPU_STATE_SYSTEM: libc::c_int = 1;
const CPU_STATE_IDLE: libc::c_int = 2;
const CPU_STATE_MAX: usize = 4;

const HOST_VM_INFO: libc::c_int = 2;

const SIDL: u32 = 1;
const SRUN: u32 = 2;
const SSLEEP: u32 = 3;
const SSTOP: u32 = 4;
const SZOMB: u32 = 5;

const PROC_PIDLISTFDS: libc::c_int = 1;
const PROC_PIDTBSDINFO: libc::c_int = 3;
const PROC_PIDTASKINFO: libc::c_int = 4;
const RUSAGE_INFO_V4: libc::c_int = 4;

/// BSD process info structure from proc_pidinfo(PROC_PIDTBSDINFO)
#[repr(C)]
struct ProcBsdInfo {
    pbi_flags: u32,
    pbi_status: u32,
    pbi_xstatus: u32,
    pbi_pid: u32,
    pbi_ppid: u32,
    pbi_uid: libc::uid_t,
    pbi_gid: libc::gid_t,
    pbi_ruid: libc::uid_t,
    pbi_rgid: libc::gid_t,
    pbi_svuid: libc::uid_t,
    pbi_svgid: libc::gid_t,
    _reserved: u32,
    pbi_comm: [libc::c_char; 16],
    pbi_name: [libc::c_char; 32],
    pbi_nfiles: u32,
    pbi_pgid: u32,
    pbi_pjobc: u32,
    e_tdev: u32,
    e_tpgid: u32,
    pbi_nice: i32,
    pbi_start_tvsec: u64,
    pbi_start_tvusec: u64,
}

/// Task info structure from proc_pidinfo(PROC_PIDTASKINFO)
#[repr(C)]
struct ProcTaskInfo {
    pti_virtual_size: u64,
    pti_resident_size: u64,
    pti_total_user: u64,
    pti_total_system: u64,
    pti_threads_user: u64,
    pti_threads_system: u64,
    pti_policy: i32,
    pti_faults: i32,
    pti_pageins: i32,
    pti_cow_faults: i32,
    pti_messages_sent: i32,
    pti_messages_received: i32,
    pti_syscalls_mach: i32,
    pti_syscalls_unix: i32,
    pti_csw: i32,
    pti_threadnum: i32,
    pti_numrunning: i32,
    pti_priority: i32,
}

const RTM_IFINFO2: i32 = 0x12;

/// vm_statistics structure (32-bit counters) - more reliable for memory info.
#[repr(C)]
struct vm_statistics {
    free_count: u32,
    active_count: u32,
    inactive_count: u32,
    wire_count: u32,
    zero_fill_count: u32,
    reactivations: u32,
    pageins: u32,
    pageouts: u32,
    faults: u32,
    cow_faults: u32,
    lookups: u32,
    hits: u32,
    purgeable_count: u32,
    purges: u32,
    speculative_count: u32,
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
#[derive(Clone, Copy)]
struct proc_fdinfo {
    proc_fd: i32,
    proc_fdtype: u32,
}

/// 32-bit timeval structure for interface statistics.
/// Defined locally since `libc::timeval32` may not be available on all platforms.
#[repr(C)]
#[derive(Clone, Copy)]
struct Timeval32 {
    tv_sec: i32,
    tv_usec: i32,
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
    ifi_lastchange: Timeval32,
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

// ============================================================================
// CONTEXT SWITCHES
// ============================================================================

/// Context switch counts (voluntary and involuntary).
#[derive(Debug, Clone, Copy, Default)]
pub struct ContextSwitches {
    /// Voluntary context switches (process yielded CPU).
    pub voluntary: u64,
    /// Involuntary context switches (preempted by scheduler).
    pub involuntary: u64,
}

/// Reads context switches for the current process using task_info().
///
/// Uses TASK_EVENTS_INFO which provides the csw (context switches) field.
/// macOS doesn't distinguish between voluntary/involuntary at the Mach level,
/// so we report all as voluntary.
///
/// # Examples
///
/// ```no_run
/// use probe_platform::darwin::sysctl::read_self_context_switches;
///
/// let ctx = read_self_context_switches()?;
/// println!("Context switches: {}", ctx.voluntary + ctx.involuntary);
/// # Ok::<(), probe_platform::Error>(())
/// ```
pub fn read_self_context_switches() -> Result<ContextSwitches> {
    unsafe {
        let mut events: TaskEventsInfo = mem::zeroed();
        let mut count = (mem::size_of::<TaskEventsInfo>() / mem::size_of::<libc::c_int>()) as u32;

        let result = task_info(
            libc::mach_task_self(),
            TASK_EVENTS_INFO,
            &mut events as *mut _ as *mut libc::c_int,
            &mut count,
        );

        if result != 0 {
            return Err(Error::Platform(format!("task_info failed: {}", result)));
        }

        // macOS provides total context switches, not split by voluntary/involuntary
        // We report all as voluntary for consistency
        Ok(ContextSwitches { voluntary: events.csw as u64, involuntary: 0 })
    }
}

/// Reads context switches for a specific process.
///
/// # Arguments
///
/// * `pid` - Process ID to query. Use 0 for the current process.
///
/// # Note
///
/// For non-self processes, this requires the process to be owned by the
/// same user or have appropriate privileges.
pub fn read_process_context_switches(pid: i32) -> Result<ContextSwitches> {
    if pid == 0 || pid == unsafe { libc::getpid() } {
        return read_self_context_switches();
    }

    unsafe {
        // Get task port for the target process
        let mut task: libc::mach_port_t = 0;
        let result = task_for_pid(libc::mach_task_self(), pid, &mut task);

        if result != 0 {
            // Fall back to getrusage via proc_pidinfo if task_for_pid fails
            // (which it will for processes not owned by us)
            return read_process_context_switches_via_rusage(pid);
        }

        let mut events: TaskEventsInfo = mem::zeroed();
        let mut count = (mem::size_of::<TaskEventsInfo>() / mem::size_of::<libc::c_int>()) as u32;

        let result = task_info(
            task,
            TASK_EVENTS_INFO,
            &mut events as *mut _ as *mut libc::c_int,
            &mut count,
        );

        // Deallocate the task port
        mach_port_deallocate(libc::mach_task_self(), task);

        if result != 0 {
            return Err(Error::Platform(format!("task_info failed: {}", result)));
        }

        Ok(ContextSwitches { voluntary: events.csw as u64, involuntary: 0 })
    }
}

fn read_process_context_switches_via_rusage(pid: i32) -> Result<ContextSwitches> {
    unsafe {
        // Use proc_pid_rusage for processes we can't get task ports for
        let mut usage: RusageInfoV4 = mem::zeroed();
        let result = libc::proc_pid_rusage(
            pid,
            RUSAGE_INFO_V4,
            &mut usage as *mut _ as *mut libc::rusage_info_t,
        );

        if result != 0 {
            return Err(Error::NotFound(format!("process {} not found or not accessible", pid)));
        }

        // rusage_info_v4 contains context switches
        Ok(ContextSwitches {
            voluntary: usage.ri_voluntary_context_switches,
            involuntary: usage.ri_involuntary_context_switches,
        })
    }
}

/// Reads system-wide context switch count.
///
/// Aggregates context switches from all running processes.
///
/// # Note
///
/// This is an expensive operation. Consider caching for frequent sampling.
pub fn read_system_context_switches() -> Result<ContextSwitches> {
    let pids = list_pids()?;
    let mut total = ContextSwitches::default();

    for pid in pids {
        if let Ok(ctx) = read_process_context_switches(pid) {
            total.voluntary = total.voluntary.saturating_add(ctx.voluntary);
            total.involuntary = total.involuntary.saturating_add(ctx.involuntary);
        }
    }

    Ok(total)
}

// ============================================================================
// MACH AND SYSTEM TYPES
// ============================================================================

const TASK_EVENTS_INFO: libc::c_int = 2;

/// Task events info structure from Mach.
#[repr(C)]
struct TaskEventsInfo {
    faults: i32,            // page faults
    pageins: i32,           // actual pageins
    cow_faults: i32,        // copy-on-write faults
    messages_sent: i32,     // messages sent
    messages_received: i32, // messages received
    syscalls_mach: i32,     // Mach system calls
    syscalls_unix: i32,     // Unix system calls
    csw: i32,               // context switches
}

/// Rusage info v4 structure with context switches.
#[repr(C)]
struct RusageInfoV4 {
    ri_uuid: [u8; 16],
    ri_user_time: u64,
    ri_system_time: u64,
    ri_pkg_idle_wkups: u64,
    ri_interrupt_wkups: u64,
    ri_pageins: u64,
    ri_wired_size: u64,
    ri_resident_size: u64,
    ri_phys_footprint: u64,
    ri_proc_start_abstime: u64,
    ri_proc_exit_abstime: u64,
    ri_child_user_time: u64,
    ri_child_system_time: u64,
    ri_child_pkg_idle_wkups: u64,
    ri_child_interrupt_wkups: u64,
    ri_child_pageins: u64,
    ri_child_elapsed_abstime: u64,
    ri_diskio_bytesread: u64,
    ri_diskio_byteswritten: u64,
    ri_cpu_time_qos_default: u64,
    ri_cpu_time_qos_maintenance: u64,
    ri_cpu_time_qos_background: u64,
    ri_cpu_time_qos_utility: u64,
    ri_cpu_time_qos_legacy: u64,
    ri_cpu_time_qos_user_initiated: u64,
    ri_cpu_time_qos_user_interactive: u64,
    ri_billed_system_time: u64,
    ri_serviced_system_time: u64,
    ri_logical_writes: u64,
    ri_lifetime_max_phys_footprint: u64,
    ri_instructions: u64,
    ri_cycles: u64,
    ri_billed_energy: u64,
    ri_serviced_energy: u64,
    ri_interval_max_phys_footprint: u64,
    ri_runnable_time: u64,
    ri_voluntary_context_switches: u64,
    ri_involuntary_context_switches: u64,
}

// External Mach functions
unsafe extern "C" {
    fn host_processor_info(
        host: libc::mach_port_t,
        flavor: libc::c_int,
        out_processor_count: *mut libc::c_uint,
        out_processor_info: *mut *mut libc::c_int,
        out_processor_infoCnt: *mut libc::c_uint,
    ) -> libc::c_int;

    fn host_statistics(
        host: libc::mach_port_t,
        flavor: libc::c_int,
        host_info_out: *mut libc::c_int,
        host_info_outCnt: *mut u32,
    ) -> libc::c_int;

    fn task_info(
        target_task: libc::mach_port_t,
        flavor: libc::c_int,
        task_info_out: *mut libc::c_int,
        task_info_outCnt: *mut u32,
    ) -> libc::c_int;

    fn task_for_pid(
        target_tport: libc::mach_port_t,
        pid: libc::c_int,
        t: *mut libc::mach_port_t,
    ) -> libc::c_int;

    fn mach_port_deallocate(task: libc::mach_port_t, name: libc::mach_port_t) -> libc::c_int;
}

// ============================================================================
// IOKIT FFI BINDINGS
// ============================================================================

// IOKit functions for disk I/O statistics
#[link(name = "IOKit", kind = "framework")]
unsafe extern "C" {
    fn IOMainPort(
        bootstrap_port: libc::mach_port_t,
        master_port: *mut libc::mach_port_t,
    ) -> libc::c_int;

    fn IOMasterPort(
        bootstrap_port: libc::mach_port_t,
        master_port: *mut libc::mach_port_t,
    ) -> libc::c_int;

    fn IOServiceMatching(name: *const libc::c_char) -> *mut libc::c_void;

    fn IOServiceGetMatchingServices(
        master_port: libc::mach_port_t,
        matching: *mut libc::c_void,
        existing: *mut u32,
    ) -> libc::c_int;

    fn IOIteratorNext(iterator: u32) -> u32;

    fn IOObjectRelease(object: u32) -> libc::c_int;

    fn IORegistryEntryGetParentEntry(
        entry: u32,
        plane: *const libc::c_char,
        parent: *mut u32,
    ) -> libc::c_int;

    fn IORegistryEntryGetName(entry: u32, name: *mut libc::c_char) -> libc::c_int;

    fn IORegistryEntryCreateCFProperties(
        entry: u32,
        properties: *mut *mut libc::c_void,
        allocator: *const libc::c_void,
        options: u32,
    ) -> libc::c_int;
}

// CoreFoundation functions for dictionary access
#[link(name = "CoreFoundation", kind = "framework")]
unsafe extern "C" {
    fn CFRelease(cf: *const libc::c_void);

    fn CFStringCreateWithCString(
        allocator: *const libc::c_void,
        c_str: *const libc::c_char,
        encoding: u32,
    ) -> *const libc::c_void;

    fn CFDictionaryGetValue(
        dict: *const libc::c_void,
        key: *const libc::c_void,
    ) -> *const libc::c_void;

    fn CFNumberGetValue(
        number: *const libc::c_void,
        number_type: libc::c_int,
        value_ptr: *mut libc::c_void,
    ) -> bool;
}

// ============================================================================
// THERMAL MONITORING (SMC)
// ============================================================================

/// Read thermal zone information from macOS SMC.
///
/// macOS uses the System Management Controller (SMC) for thermal monitoring.
/// This function reads CPU and GPU temperatures via IOKit SMC access.
///
/// # Examples
///
/// ```no_run
/// use probe_platform::darwin::sysctl::read_thermal_zones;
///
/// let zones = read_thermal_zones()?;
/// for zone in zones {
///     println!("{}: {:.1}°C", zone.name, zone.temp_celsius);
/// }
/// # Ok::<(), probe_platform::Error>(())
/// ```
pub fn read_thermal_zones() -> Result<Vec<crate::ThermalZone>> {
    unsafe {
        // Get the IOKit master port
        let mut master_port: libc::mach_port_t = 0;
        if IOMainPort(libc::mach_host_self(), &mut master_port) != 0 {
            if IOMasterPort(libc::mach_host_self(), &mut master_port) != 0 {
                return Err(Error::Platform("failed to get IOKit master port".to_string()));
            }
        }

        // Create matching dictionary for AppleSMC
        let matching = IOServiceMatching(b"AppleSMC\0".as_ptr() as *const libc::c_char);
        if matching.is_null() {
            return Err(Error::Platform("failed to create SMC matching dict".to_string()));
        }

        // Get SMC service
        let mut iterator: u32 = 0;
        if IOServiceGetMatchingServices(master_port, matching, &mut iterator) != 0 {
            return Err(Error::Platform("SMC service not found".to_string()));
        }

        let smc_service = IOIteratorNext(iterator);
        IOObjectRelease(iterator);

        if smc_service == 0 {
            return Err(Error::NotSupported);
        }

        // Open connection to SMC
        let mut conn: u32 = 0;
        if IOServiceOpen(smc_service, libc::mach_task_self(), 0, &mut conn) != 0 {
            IOObjectRelease(smc_service);
            return Err(Error::Platform("failed to open SMC connection".to_string()));
        }

        IOObjectRelease(smc_service);

        let mut zones = Vec::new();

        // Common SMC temperature keys
        let temp_keys = [
            ("TC0P", "CPU Proximity"),
            ("TC0D", "CPU Die"),
            ("TC0H", "CPU Heatsink"),
            ("TCXC", "CPU PECI"),
            ("TG0P", "GPU Proximity"),
            ("TG0D", "GPU Die"),
            ("TG0H", "GPU Heatsink"),
            ("Tm0P", "Memory Proximity"),
            ("TN0P", "North Bridge"),
            ("TA0P", "Ambient"),
        ];

        for (key, label) in temp_keys {
            if let Some(temp) = read_smc_temperature(conn, key) {
                zones.push(crate::ThermalZone {
                    name: "smc".to_string(),
                    label: label.to_string(),
                    temp_celsius: temp,
                    temp_max: Some(100.0),  // Default max temp
                    temp_crit: Some(105.0), // Default critical temp
                });
            }
        }

        IOServiceClose(conn);

        if zones.is_empty() {
            return Err(Error::NotSupported);
        }

        Ok(zones)
    }
}

/// Read a temperature value from SMC.
unsafe fn read_smc_temperature(conn: u32, key: &str) -> Option<f64> {
    // SMC input/output structure
    #[repr(C)]
    struct SmcKeyData {
        key: u32,
        vers: SmcVers,
        p_limit_data: SmcPLimitData,
        key_info: SmcKeyInfo,
        result: u8,
        status: u8,
        data8: u8,
        data32: u32,
        bytes: [u8; 32],
    }

    #[repr(C)]
    struct SmcVers {
        major: u8,
        minor: u8,
        build: u8,
        reserved: u8,
        release: u16,
    }

    #[repr(C)]
    struct SmcPLimitData {
        version: u16,
        length: u16,
        cpu_p_limit: u32,
        gpu_p_limit: u32,
        mem_p_limit: u32,
    }

    #[repr(C)]
    struct SmcKeyInfo {
        data_size: u32,
        data_type: u32,
        data_attributes: u8,
    }

    const SMC_CMD_READ_KEYINFO: u8 = 9;
    const SMC_CMD_READ_BYTES: u8 = 5;

    // Convert key to u32 (big-endian 4-char code)
    let key_bytes = key.as_bytes();
    if key_bytes.len() != 4 {
        return None;
    }
    let key_code = u32::from_be_bytes([key_bytes[0], key_bytes[1], key_bytes[2], key_bytes[3]]);

    // SAFETY: SmcKeyData is a C-compatible struct that can be safely zero-initialized
    let mut input: SmcKeyData = unsafe { mem::zeroed() };
    input.key = key_code;
    input.data8 = SMC_CMD_READ_KEYINFO;

    let mut output: SmcKeyData = unsafe { mem::zeroed() };
    let input_size = mem::size_of::<SmcKeyData>();
    let mut output_size = mem::size_of::<SmcKeyData>();

    // SAFETY: Calling IOKit function with valid parameters
    if unsafe {
        IOConnectCallStructMethod(
            conn,
            2, // kSMCHandleYPCEvent
            &input as *const _ as *const libc::c_void,
            input_size,
            &mut output as *mut _ as *mut libc::c_void,
            &mut output_size,
        )
    } != 0
    {
        return None;
    }

    // Check if key exists
    if output.result != 0 || output.key_info.data_size == 0 {
        return None;
    }

    // Read the actual value
    input.key_info.data_size = output.key_info.data_size;
    input.data8 = SMC_CMD_READ_BYTES;

    output = unsafe { mem::zeroed() };

    // SAFETY: Calling IOKit function with valid parameters
    if unsafe {
        IOConnectCallStructMethod(
            conn,
            2,
            &input as *const _ as *const libc::c_void,
            input_size,
            &mut output as *mut _ as *mut libc::c_void,
            &mut output_size,
        )
    } != 0
    {
        return None;
    }

    if output.result != 0 {
        return None;
    }

    // Convert SMC data to temperature
    // SMC stores temperatures in various formats, most common is sp78 (signed 7.8 fixed point)
    let data_type = output.key_info.data_type;

    // 'sp78' = 0x73703738
    if data_type == 0x73703738 || data_type == 0x66707838 {
        // Fixed point 7.8 format: high byte is integer part, low byte is fractional (scaled by 256)
        let int_part = output.bytes[0] as i8;
        let frac_part = output.bytes[1] as u8;
        let temp = f64::from(int_part) + (f64::from(frac_part) / 256.0);
        if temp > -40.0 && temp < 150.0 {
            return Some(temp);
        }
    }

    // 'flt ' = floating point
    if data_type == 0x666c7420 && output.key_info.data_size >= 4 {
        let temp = f32::from_be_bytes([
            output.bytes[0],
            output.bytes[1],
            output.bytes[2],
            output.bytes[3],
        ]);
        if temp > -40.0 && temp < 150.0 {
            return Some(f64::from(temp));
        }
    }

    None
}

/// Check if thermal monitoring is supported.
#[must_use]
pub fn is_thermal_supported() -> bool {
    read_thermal_zones().is_ok()
}

// Additional IOKit functions for SMC access
#[link(name = "IOKit", kind = "framework")]
unsafe extern "C" {
    fn IOServiceOpen(
        service: u32,
        owning_task: libc::mach_port_t,
        connection_type: u32,
        connection: *mut u32,
    ) -> libc::c_int;

    fn IOServiceClose(connection: u32) -> libc::c_int;

    fn IOConnectCallStructMethod(
        connection: u32,
        selector: u32,
        input_struct: *const libc::c_void,
        input_struct_cnt: usize,
        output_struct: *mut libc::c_void,
        output_struct_cnt: *mut usize,
    ) -> libc::c_int;
}
