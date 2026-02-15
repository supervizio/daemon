//! BSD sysctl wrappers

use crate::{DiskIOStats, DiskUsage, Error, NetInterface, NetStats, Partition, Result};
use std::ffi::CString;
use std::mem;
use std::ptr;

// OpenBSD doesn't have sysctlbyname in libc. Provide a Rust implementation
// that maps well-known sysctl names to MIB arrays and calls sysctl().
#[cfg(target_os = "openbsd")]
#[unsafe(no_mangle)]
pub unsafe extern "C" fn sysctlbyname(
    name: *const libc::c_char,
    oldp: *mut libc::c_void,
    oldlenp: *mut usize,
    newp: *mut libc::c_void,
    newlen: usize,
) -> libc::c_int {
    // OpenBSD sysctl MIB constants (from sys/sysctl.h)
    const CTL_KERN: libc::c_int = 1;
    const CTL_HW: libc::c_int = 6;
    const KERN_CPTIME: libc::c_int = 40;
    const HW_NCPU: libc::c_int = 3;
    const HW_PHYSMEM64: libc::c_int = 19;
    const HW_CPUSPEED: libc::c_int = 12;
    const HW_DISKSTATS: libc::c_int = 9;

    let name_cstr = unsafe { std::ffi::CStr::from_ptr(name) };
    let name_bytes = name_cstr.to_bytes();

    // Map sysctl name to MIB array
    let mib: [libc::c_int; 2] = match name_bytes {
        b"kern.cp_time" => [CTL_KERN, KERN_CPTIME],
        b"hw.ncpu" => [CTL_HW, HW_NCPU],
        b"hw.cpuspeed" => [CTL_HW, HW_CPUSPEED],
        b"hw.physmem" => [CTL_HW, HW_PHYSMEM64],
        b"hw.diskstats" => [CTL_HW, HW_DISKSTATS],
        _ => return -1,
    };

    unsafe { libc::sysctl(mib.as_ptr() as *mut libc::c_int, 2, oldp, oldlenp, newp, newlen) }
}

// Cross-BSD wrapper: uses libc on FreeBSD/NetBSD, shim on OpenBSD
#[cfg(not(target_os = "openbsd"))]
#[inline(always)]
unsafe fn do_sysctlbyname(
    name: *const libc::c_char,
    oldp: *mut libc::c_void,
    oldlenp: *mut usize,
    newp: *mut libc::c_void,
    newlen: usize,
) -> libc::c_int {
    unsafe { libc::sysctlbyname(name, oldp, oldlenp, newp, newlen) }
}

#[cfg(target_os = "openbsd")]
#[inline(always)]
unsafe fn do_sysctlbyname(
    name: *const libc::c_char,
    oldp: *mut libc::c_void,
    oldlenp: *mut usize,
    newp: *mut libc::c_void,
    newlen: usize,
) -> libc::c_int {
    unsafe { sysctlbyname(name, oldp, oldlenp, newp, newlen) }
}

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
        // kern.cp_time on FreeBSD/OpenBSD/NetBSD
        let name = CString::new("kern.cp_time")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

        let mut cp_time: [u64; 5] = [0; 5]; // user, nice, sys, intr, idle
        let mut len = mem::size_of_val(&cp_time);

        let result = do_sysctlbyname(
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

        do_sysctlbyname(
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

        do_sysctlbyname(
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
        // Get page size - must check for error (-1)
        let page_size_raw = libc::sysconf(libc::_SC_PAGESIZE);
        if page_size_raw <= 0 {
            return Err(Error::Platform("failed to get page size".to_string()));
        }
        let page_size = page_size_raw as u64;

        // Get total physical memory
        let name = CString::new("hw.physmem")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
        let mut physmem: u64 = 0;
        let mut len = mem::size_of::<u64>();

        do_sysctlbyname(
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

        let mut free_pages: u64 = 0;

        #[cfg(target_os = "freebsd")]
        {
            let mut free_len = mem::size_of::<u64>();
            do_sysctlbyname(
                free_name.as_ptr(),
                &mut free_pages as *mut _ as *mut libc::c_void,
                &mut free_len,
                ptr::null_mut(),
                0,
            );
        }

        // OpenBSD/NetBSD: Use single uvmexp call for all metrics (optimization)
        // This consolidates 3 separate get_uvmexp() calls into 1
        #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
        let (cached, swap_total, swap_used) = {
            let uvm = get_uvmexp()?;
            free_pages = uvm.free as u64;
            let cached = (uvm.vnodepages as u64).saturating_mul(page_size);
            let swap_total = (uvm.swpages as u64).saturating_mul(page_size);
            let swap_used = (uvm.swpginuse as u64).saturating_mul(page_size);
            (cached, swap_total, swap_used)
        };

        // Get cached pages (FreeBSD specific)
        #[cfg(target_os = "freebsd")]
        let cached = {
            let cache_name = CString::new("vm.stats.vm.v_cache_count")
                .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;
            let mut cache_pages: u64 = 0;
            let mut cache_len = mem::size_of::<u64>();
            do_sysctlbyname(
                cache_name.as_ptr(),
                &mut cache_pages as *mut _ as *mut libc::c_void,
                &mut cache_len,
                ptr::null_mut(),
                0,
            );
            cache_pages.saturating_mul(page_size)
        };

        // Get swap info
        #[cfg(target_os = "freebsd")]
        let (swap_total, swap_used) = get_swap_freebsd();

        // Get buffer cache size (platform-specific)
        let buffers = get_buffers_bytes(page_size);

        Ok(MemInfo {
            total: physmem,
            available: free_pages.saturating_mul(page_size),
            cached,
            buffers,
            swap_total,
            swap_used,
        })
    }
}

/// Returns the size of the filesystem buffer cache in bytes.
///
/// - FreeBSD: `vfs.bufspace` sysctl (bytes directly)
/// - OpenBSD: `kern.bcstats` sysctl → `numbufpages × page_size`
/// - NetBSD: `vm.bufmem` sysctl (bytes directly)
fn get_buffers_bytes(page_size: u64) -> u64 {
    #[cfg(target_os = "freebsd")]
    {
        // vfs.bufspace returns bytes directly (u64).
        // On ZFS systems this may return 0 (ZFS uses its own ARC cache).
        let name = match CString::new("vfs.bufspace") {
            Ok(n) => n,
            Err(_) => return 0,
        };
        let mut bufspace: u64 = 0;
        let mut len = mem::size_of::<u64>();
        unsafe {
            do_sysctlbyname(
                name.as_ptr(),
                &mut bufspace as *mut _ as *mut libc::c_void,
                &mut len,
                ptr::null_mut(),
                0,
            );
        }
        bufspace
    }

    #[cfg(target_os = "openbsd")]
    {
        // kern.bcstats via sysctl MIB [CTL_VFS, VFS_GENERIC, VFS_BCACHESTAT]
        // Returns bcachestats struct; we use numbufpages × page_size.
        const CTL_VFS: libc::c_int = 10;
        const VFS_GENERIC: libc::c_int = 0;
        const VFS_BCACHESTAT: libc::c_int = 3;

        #[repr(C)]
        struct BcacheStats {
            numbufs: i64,
            numbufpages: i64,
            numdirtypages: i64,
            numcleanpages: i64,
            pendingwrites: i64,
            pendingreads: i64,
            numwrites: i64,
            numreads: i64,
            cachehits: i64,
            busymapped: i64,
            dmapages: i64,
            highpages: i64,
            delwribufs: i64,
            kvaslots: i64,
            kvaslots_avail: i64,
            highflips: i64,
            highflops: i64,
            dmaflips: i64,
        }

        let mut mib = [CTL_VFS, VFS_GENERIC, VFS_BCACHESTAT];
        let mut bcstats: BcacheStats = unsafe { mem::zeroed() };
        let mut len = mem::size_of::<BcacheStats>();

        let result = unsafe {
            libc::sysctl(
                mib.as_mut_ptr(),
                3,
                &mut bcstats as *mut _ as *mut libc::c_void,
                &mut len,
                ptr::null_mut(),
                0,
            )
        };

        if result != 0 {
            return 0;
        }

        (bcstats.numbufpages as u64).saturating_mul(page_size)
    }

    #[cfg(target_os = "netbsd")]
    {
        // vm.bufmem returns bytes directly (unsigned long).
        let name = match CString::new("vm.bufmem") {
            Ok(n) => n,
            Err(_) => return 0,
        };
        let mut bufmem: libc::c_ulong = 0;
        let mut len = mem::size_of::<libc::c_ulong>();
        unsafe {
            do_sysctlbyname(
                name.as_ptr(),
                &mut bufmem as *mut _ as *mut libc::c_void,
                &mut len,
                ptr::null_mut(),
                0,
            );
        }
        bufmem as u64
    }
}

/// FreeBSD xswdev structure (from sys/swap_pager.h).
///
/// This structure is returned by the vm.swap_info sysctl for each swap device.
#[cfg(target_os = "freebsd")]
#[repr(C)]
struct Xswdev {
    /// Structure version (for compatibility checking).
    xsw_version: u32,
    /// Device identifier.
    xsw_dev: u64, // dev_t is 64-bit on FreeBSD
    /// Swap flags.
    xsw_flags: i32,
    /// Total blocks available.
    xsw_nblks: i32,
    /// Blocks in use.
    xsw_used: i32,
}

#[cfg(target_os = "freebsd")]
const XSWDEV_VERSION: u32 = 2;

#[cfg(target_os = "freebsd")]
fn get_swap_freebsd() -> (u64, u64) {
    unsafe {
        // Get the number of swap devices
        let nswapdev_name = match CString::new("vm.nswapdev") {
            Ok(n) => n,
            Err(_) => return (0, 0),
        };

        let mut nswapdev: i32 = 0;
        let mut len = mem::size_of::<i32>();

        if do_sysctlbyname(
            nswapdev_name.as_ptr(),
            &mut nswapdev as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return (0, 0);
        }

        if nswapdev <= 0 {
            return (0, 0);
        }

        // Query each swap device via vm.swap_info.<index>
        // Check page_size for error (-1 would cause overflow)
        let page_size_raw = libc::sysconf(libc::_SC_PAGESIZE);
        if page_size_raw <= 0 {
            return (0, 0);
        }
        let page_size = page_size_raw as u64;

        let mut total_blocks: u64 = 0;
        let mut used_blocks: u64 = 0;

        for i in 0..nswapdev {
            let name = match CString::new(format!("vm.swap_info.{}", i)) {
                Ok(n) => n,
                Err(_) => continue,
            };

            let mut xsw: Xswdev = mem::zeroed();
            let mut xsw_len = mem::size_of::<Xswdev>();

            if do_sysctlbyname(
                name.as_ptr(),
                &mut xsw as *mut _ as *mut libc::c_void,
                &mut xsw_len,
                ptr::null_mut(),
                0,
            ) != 0
            {
                continue;
            }

            // Verify version compatibility
            if xsw.xsw_version != XSWDEV_VERSION {
                // Try falling back to vm.swap_total / vm.swap_reserved
                return get_swap_freebsd_fallback();
            }

            total_blocks += xsw.xsw_nblks as u64;
            used_blocks += xsw.xsw_used as u64;
        }

        // Convert blocks to bytes (blocks are in pages on FreeBSD)
        // Use saturating_mul to prevent overflow
        let total_bytes = total_blocks.saturating_mul(page_size);
        let used_bytes = used_blocks.saturating_mul(page_size);

        (total_bytes, used_bytes)
    }
}

/// Fallback swap detection using simple vm.swap_total sysctl.
#[cfg(target_os = "freebsd")]
fn get_swap_freebsd_fallback() -> (u64, u64) {
    unsafe {
        // Check page_size for error (-1 would cause overflow)
        let page_size_raw = libc::sysconf(libc::_SC_PAGESIZE);
        if page_size_raw <= 0 {
            return (0, 0);
        }
        let page_size = page_size_raw as u64;

        // Try vm.swap_total (total swap pages)
        let total_name = match CString::new("vm.swap_total") {
            Ok(n) => n,
            Err(_) => return (0, 0),
        };

        let mut total_pages: u64 = 0;
        let mut len = mem::size_of::<u64>();

        if do_sysctlbyname(
            total_name.as_ptr(),
            &mut total_pages as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return (0, 0);
        }

        // Try vm.swap_reserved (reserved/used swap pages)
        let reserved_name = match CString::new("vm.swap_reserved") {
            Ok(n) => n,
            Err(_) => return (total_pages.saturating_mul(page_size), 0),
        };

        let mut reserved_pages: u64 = 0;
        len = mem::size_of::<u64>();

        do_sysctlbyname(
            reserved_name.as_ptr(),
            &mut reserved_pages as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        (total_pages.saturating_mul(page_size), reserved_pages.saturating_mul(page_size))
    }
}

// ============================================================================
// UVMEXP (OpenBSD/NetBSD)
// ============================================================================

/// OpenBSD/NetBSD uvmexp structure for virtual memory statistics.
///
/// This is a partial representation of the full uvmexp structure,
/// containing the fields most commonly needed for memory metrics.
#[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
#[repr(C)]
#[derive(Debug, Clone, Copy)]
struct Uvmexp {
    pagesize: i32,  // Page size in bytes
    pagemask: i32,  // Page mask
    pageshift: i32, // Page shift
    npages: i32,    // Total managed pages
    free: i32,      // Free pages
    active: i32,    // Active pages
    inactive: i32,  // Inactive pages
    paging: i32,    // Pages being paged
    wired: i32,     // Wired pages
    zeropages: i32, // Zero-fill pages
    reserve_pagedaemon: i32,
    reserve_kernel: i32,
    // Pageout params
    anonpages: i32,  // Anonymous pages
    vnodepages: i32, // Vnode pages (file cache)
    vtextpages: i32, // Vnode text pages
    freemin: i32,
    freetarg: i32,
    inactarg: i32,
    wiredmax: i32,
    // Swap
    nswapdev: i32,  // Number of swap devices
    swpages: i32,   // Total swap pages
    swpginuse: i32, // Swap pages in use
    swpgonly: i32,  // Swap pages only in swap
    nswget: i32,
    // Padding to ensure structure is large enough
    _padding: [i32; 64],
}

/// Gets uvmexp statistics via sysctl VM_UVMEXP.
#[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
fn get_uvmexp() -> Result<Uvmexp> {
    unsafe {
        #[cfg(target_os = "openbsd")]
        const VM_UVMEXP: libc::c_int = 4;
        #[cfg(target_os = "netbsd")]
        const VM_UVMEXP: libc::c_int = 2;

        let mut mib = [libc::CTL_VM, VM_UVMEXP];
        let mut uvm: Uvmexp = mem::zeroed();
        let mut len = mem::size_of::<Uvmexp>();

        let result = libc::sysctl(
            mib.as_mut_ptr(),
            2,
            &mut uvm as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        if result != 0 {
            return Err(Error::Platform(format!(
                "sysctl VM_UVMEXP failed: {}",
                std::io::Error::last_os_error()
            )));
        }

        Ok(uvm)
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
                num_fds: 0, // Fallback: libc crate may not expose fd_nfiles field
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

        #[cfg(target_os = "openbsd")]
        {
            get_process_info_openbsd(pid)
        }

        #[cfg(target_os = "netbsd")]
        {
            get_process_info_netbsd(pid)
        }
    }
}

#[cfg(target_os = "openbsd")]
fn get_process_info_openbsd(pid: i32) -> Result<ProcessInfo> {
    unsafe {
        // OpenBSD kinfo_proc structure (simplified, key fields only)
        #[repr(C)]
        struct KinfoProc {
            p_forw: u64,         // 0: forward link
            p_back: u64,         // 8: backward link
            p_paddr: u64,        // 16: address of proc
            p_addr: u64,         // 24: kernel virtual addr
            p_fd: u64,           // 32: ptr to open files
            p_stats: u64,        // 40: accounting info
            p_limit: u64,        // 48: process limits
            p_vmspace: u64,      // 56: address space
            p_sigacts: u64,      // 64: signal actions
            p_sess: u64,         // 72: session
            p_tsess: u64,        // 80: controlling tty session
            p_ru: u64,           // 88: rusage pointer
            p_eflag: i32,        // 96: various flags
            p_exitsig: i32,      // 100: exit signal
            p_flag: i32,         // 104: process flags
            p_pid: i32,          // 108: process id
            p_ppid: i32,         // 112: parent pid
            p_sid: i32,          // 116: session id
            _pgid: i32,          // 120
            p_tpgid: i32,        // 124
            p_uid: u32,          // 128: real uid
            p_ruid: u32,         // 132: real uid
            p_gid: u32,          // 136: real gid
            p_rgid: u32,         // 140: real gid
            p_groups: [u32; 16], // 144: groups
            p_ngroups: i16,      // 208
            p_jobc: i16,         // 210: job control
            p_tdev: u32,         // 212: controlling tty
            p_estcpu: u32,       // 216: time averaged cpu
            p_rtime_sec: i64,    // 220: real time
            p_rtime_usec: i64,   // 228
            p_cpticks: i32,      // 236: cpu ticks
            p_pctcpu: u32,       // 240: cpu usage %
            p_swtime: u32,       // 244
            p_slptime: u32,      // 248
            p_schedflags: i32,   // 252
            p_uticks: u64,       // 256
            p_sticks: u64,       // 264
            p_iticks: u64,       // 272
            p_tracep: u64,       // 280: trace pointer
            p_traceflag: i32,    // 288
            p_holdcnt: i32,      // 292
            p_siglist: i32,      // 296
            p_sigmask: u32,      // 300
            p_sigignore: u32,    // 304
            p_sigcatch: u32,     // 308
            p_stat: i8,          // 312: process state
            p_priority: i8,      // 313
            p_usrpri: i8,        // 314
            p_nice: i8,          // 315
            p_xstat: u16,        // 316
            p_spare: u16,        // 318
            p_comm: [u8; 24],    // 320: command name
            p_wmesg: [u8; 8],    // 344: wait message
            p_wchan: u64,        // 352: wait channel
            p_login: [u8; 32],   // 360: login name
            p_vm_rssize: i32,    // 392: RSS in pages
            p_vm_tsize: i32,     // 396: text size
            p_vm_dsize: i32,     // 400: data size
            p_vm_ssize: i32,     // 404: stack size
            p_uvalid: i64,       // 408
            p_ustart_sec: i64,   // 416
            p_ustart_usec: i64,  // 424
            p_uutime_sec: u32,   // 432
            p_uutime_usec: u32,  // 436
            p_ustime_sec: u32,   // 440
            p_ustime_usec: u32,  // 444
            p_uru_maxrss: u64,   // 448
            p_uru_ixrss: u64,    // 456
            p_uru_idrss: u64,    // 464
            p_uru_isrss: u64,    // 472
            p_uru_minflt: u64,   // 480
            p_uru_majflt: u64,   // 488
            p_uru_nswap: u64,    // 496
            p_uru_inblock: u64,  // 504
            p_uru_oublock: u64,  // 512
            p_uru_msgsnd: u64,   // 520
            p_uru_msgrcv: u64,   // 528
            p_uru_nsignals: u64, // 536
            p_uru_nvcsw: u64,    // 544: voluntary ctx switches
            p_uru_nivcsw: u64,   // 552: involuntary ctx switches
            _rest: [u8; 256],    // padding for future fields
        }

        let mut mib = [
            libc::CTL_KERN,
            libc::KERN_PROC,
            libc::KERN_PROC_PID,
            pid as libc::c_int,
            mem::size_of::<KinfoProc>() as libc::c_int,
            1,
        ];

        let mut kinfo: KinfoProc = mem::zeroed();
        let mut len = mem::size_of::<KinfoProc>();

        let result = libc::sysctl(
            mib.as_mut_ptr(),
            6,
            &mut kinfo as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        if result != 0 || len == 0 {
            return Err(Error::NotFound(format!("process {} not found", pid)));
        }

        // Check page_size for error (-1 would cause overflow)
        let page_size_raw = libc::sysconf(libc::_SC_PAGESIZE);
        if page_size_raw <= 0 {
            return Err(Error::Platform("failed to get page size".to_string()));
        }
        let page_size = page_size_raw as u64;

        Ok(ProcessInfo {
            rss: (kinfo.p_vm_rssize as u64).saturating_mul(page_size),
            vsize: ((kinfo.p_vm_tsize + kinfo.p_vm_dsize + kinfo.p_vm_ssize) as u64)
                .saturating_mul(page_size),
            num_threads: 1, // OpenBSD doesn't expose thread count easily
            num_fds: 0,     // Would need KERN_FILE sysctl
            state: match kinfo.p_stat {
                1 => 1, // SIDL -> Running (idle)
                2 => 1, // SRUN -> Running
                3 => 2, // SSLEEP -> Sleeping
                4 => 5, // SSTOP -> Stopped
                5 => 4, // SZOMB -> Zombie
                6 => 3, // SDEAD -> Waiting
                7 => 6, // SONPROC -> Running (on CPU)
                _ => 0,
            },
        })
    }
}

#[cfg(target_os = "netbsd")]
fn get_process_info_netbsd(pid: i32) -> Result<ProcessInfo> {
    unsafe {
        // NetBSD uses kinfo_proc2 via KERN_PROC2
        #[repr(C)]
        struct KinfoProc2 {
            p_forw: u64,
            p_back: u64,
            p_paddr: u64,
            p_addr: u64,
            p_fd: u64,
            p_cwdi: u64,
            p_stats: u64,
            p_limit: u64,
            p_vmspace: u64,
            p_sigacts: u64,
            p_sess: u64,
            p_tsess: u64,
            p_ru: u64,
            p_eflag: i32,
            p_exitsig: i32,
            p_flag: i32,
            p_pid: i32,
            p_ppid: i32,
            p_sid: i32,
            _pgid: i32,
            p_tpgid: i32,
            p_uid: u32,
            p_ruid: u32,
            p_gid: u32,
            p_rgid: u32,
            p_groups: [u32; 16],
            p_ngroups: i16,
            p_jobc: i16,
            p_tdev: u32,
            p_estcpu: u32,
            p_rtime_sec: i32,
            p_rtime_usec: i32,
            p_cpticks: i32,
            p_pctcpu: u32,
            p_swtime: u32,
            p_slptime: u32,
            p_schedflags: i32,
            p_uticks: u64,
            p_sticks: u64,
            p_iticks: u64,
            p_tracep: u64,
            p_traceflag: i32,
            p_holdcnt: i32,
            p_svuid: u32,
            p_svgid: u32,
            p_ename: [u8; 17],
            p_comm: [u8; 24],
            p_stat: i8,
            p_nice: i8,
            p_xstat: u16,
            p_acflag: u16,
            p_vm_rssize: i32, // RSS in pages
            p_vm_tsize: i64,  // text size
            p_vm_dsize: i64,  // data size
            p_vm_ssize: i64,  // stack size
            p_vm_vsize: i64,  // total virtual size
            p_uru_maxrss: u64,
            p_uru_ixrss: u64,
            p_uru_idrss: u64,
            p_uru_isrss: u64,
            p_uru_minflt: u64,
            p_uru_majflt: u64,
            p_uru_nswap: u64,
            p_uru_inblock: u64,
            p_uru_oublock: u64,
            p_uru_msgsnd: u64,
            p_uru_msgrcv: u64,
            p_uru_nsignals: u64,
            p_uru_nvcsw: u64,  // voluntary ctx switches
            p_uru_nivcsw: u64, // involuntary ctx switches
            p_nlwps: i32,      // number of LWPs (threads)
            p_nrlwps: i32,     // running LWPs
            _rest: [u8; 256],
        }

        const KERN_PROC2: libc::c_int = 47;

        let mut mib = [
            libc::CTL_KERN,
            KERN_PROC2,
            libc::KERN_PROC_PID,
            pid as libc::c_int,
            mem::size_of::<KinfoProc2>() as libc::c_int,
            1,
        ];

        let mut kinfo: KinfoProc2 = mem::zeroed();
        let mut len = mem::size_of::<KinfoProc2>();

        let result = libc::sysctl(
            mib.as_mut_ptr(),
            6,
            &mut kinfo as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        if result != 0 || len == 0 {
            return Err(Error::NotFound(format!("process {} not found", pid)));
        }

        // Check page_size for error (-1 would cause overflow)
        let page_size_raw = libc::sysconf(libc::_SC_PAGESIZE);
        if page_size_raw <= 0 {
            return Err(Error::Platform("failed to get page size".to_string()));
        }
        let page_size = page_size_raw as u64;

        Ok(ProcessInfo {
            rss: (kinfo.p_vm_rssize as u64).saturating_mul(page_size),
            vsize: kinfo.p_vm_vsize as u64,
            num_threads: kinfo.p_nlwps as u32,
            num_fds: 0, // Would need KERN_FILE sysctl
            state: match kinfo.p_stat {
                1 => 1, // SIDL -> Running
                2 => 1, // SRUN -> Running
                3 => 2, // SSLEEP -> Sleeping
                4 => 5, // SSTOP -> Stopped
                5 => 4, // SZOMB -> Zombie
                6 => 3, // SDEAD -> Waiting
                7 => 6, // SONPROC -> Running
                _ => 0,
            },
        })
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

        #[cfg(target_os = "openbsd")]
        {
            list_pids_openbsd()
        }

        #[cfg(target_os = "netbsd")]
        {
            list_pids_netbsd()
        }
    }
}

#[cfg(target_os = "openbsd")]
fn list_pids_openbsd() -> Result<Vec<i32>> {
    unsafe {
        // Minimal struct to get just PIDs
        #[repr(C)]
        #[derive(Clone, Copy)]
        struct KinfoProcMin {
            _padding: [u8; 108],
            p_pid: i32,
            _rest: [u8; 700],
        }

        let mut mib = [
            libc::CTL_KERN,
            libc::KERN_PROC,
            libc::KERN_PROC_ALL,
            0,
            mem::size_of::<KinfoProcMin>() as libc::c_int,
            0,
        ];

        // Get size first
        let mut len: usize = 0;
        if libc::sysctl(mib.as_mut_ptr(), 6, ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        // Get count estimate
        let count = len / mem::size_of::<KinfoProcMin>() + 10;
        mib[5] = count as libc::c_int;
        len = count * mem::size_of::<KinfoProcMin>();

        let mut kinfos: Vec<KinfoProcMin> = vec![mem::zeroed(); count];

        if libc::sysctl(
            mib.as_mut_ptr(),
            6,
            kinfos.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Ok(Vec::new());
        }

        let actual_count = len / mem::size_of::<KinfoProcMin>();
        let pids: Vec<i32> =
            kinfos[..actual_count].iter().map(|k| k.p_pid).filter(|&p| p > 0).collect();

        Ok(pids)
    }
}

#[cfg(target_os = "netbsd")]
fn list_pids_netbsd() -> Result<Vec<i32>> {
    unsafe {
        const KERN_PROC2: libc::c_int = 47;

        // Minimal struct to get just PIDs
        #[repr(C)]
        #[derive(Clone, Copy)]
        struct KinfoProc2Min {
            _padding: [u8; 76],
            p_pid: i32,
            _rest: [u8; 600],
        }

        let mut mib = [
            libc::CTL_KERN,
            KERN_PROC2,
            libc::KERN_PROC_ALL,
            0,
            mem::size_of::<KinfoProc2Min>() as libc::c_int,
            0,
        ];

        // Get size first
        let mut len: usize = 0;
        if libc::sysctl(mib.as_mut_ptr(), 6, ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        // Get count estimate
        let count = len / mem::size_of::<KinfoProc2Min>() + 10;
        mib[5] = count as libc::c_int;
        len = count * mem::size_of::<KinfoProc2Min>();

        let mut kinfos: Vec<KinfoProc2Min> = vec![mem::zeroed(); count];

        if libc::sysctl(
            mib.as_mut_ptr(),
            6,
            kinfos.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Ok(Vec::new());
        }

        let actual_count = len / mem::size_of::<KinfoProc2Min>();
        let pids: Vec<i32> =
            kinfos[..actual_count].iter().map(|k| k.p_pid).filter(|&p| p > 0).collect();

        Ok(pids)
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

        #[cfg(target_os = "openbsd")]
        {
            get_mounts_openbsd()
        }

        #[cfg(target_os = "netbsd")]
        {
            get_mounts_netbsd()
        }
    }
}

#[cfg(target_os = "openbsd")]
fn get_mounts_openbsd() -> Result<Vec<Partition>> {
    unsafe {
        // OpenBSD uses getmntinfo with statfs structure
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
            if fs_type == "mfs" || fs_type == "kernfs" || fs_type == "procfs" {
                continue;
            }

            partitions.push(Partition { device, mount_point, fs_type, options: String::new() });
        }

        Ok(partitions)
    }
}

#[cfg(target_os = "netbsd")]
fn get_mounts_netbsd() -> Result<Vec<Partition>> {
    unsafe {
        // NetBSD uses getvfsstat with statvfs structure
        // First, get the count
        let count = libc::getvfsstat(ptr::null_mut(), 0, libc::MNT_NOWAIT);
        if count <= 0 {
            return Ok(Vec::new());
        }

        // Allocate buffer
        let mut fs_list: Vec<libc::statvfs> = vec![mem::zeroed(); count as usize];
        let buf_size = count as usize * mem::size_of::<libc::statvfs>();

        let actual = libc::getvfsstat(fs_list.as_mut_ptr(), buf_size as usize, libc::MNT_NOWAIT);
        if actual <= 0 {
            return Ok(Vec::new());
        }

        let mut partitions = Vec::new();
        for i in 0..actual as usize {
            let fs = &fs_list[i];

            let device = cstr_to_string(fs.f_mntfromname.as_ptr());
            let mount_point = cstr_to_string(fs.f_mntonname.as_ptr());
            let fs_type = cstr_to_string(fs.f_fstypename.as_ptr());

            // Skip pseudo filesystems
            if fs_type == "kernfs" || fs_type == "procfs" || fs_type == "ptyfs" {
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
        #[cfg(any(target_os = "freebsd", target_os = "netbsd"))]
        let mut stat: libc::statvfs = mem::zeroed();
        #[cfg(not(any(target_os = "freebsd", target_os = "netbsd")))]
        let mut stat: libc::statfs = mem::zeroed();

        #[cfg(any(target_os = "freebsd", target_os = "netbsd"))]
        let result = libc::statvfs(c_path.as_ptr(), &mut stat);
        #[cfg(not(any(target_os = "freebsd", target_os = "netbsd")))]
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
        freebsd::get_disk_io_stats()
    }
}

// ============================================================================
// FREEBSD DISK I/O (devstat)
// ============================================================================

#[cfg(target_os = "freebsd")]
mod freebsd {
    use super::*;

    /// Bintime structure matching FreeBSD sys/time.h.
    #[repr(C)]
    #[derive(Debug, Clone, Copy)]
    struct Bintime {
        sec: i64,
        frac: u64,
    }

    /// FreeBSD devstat structure (from sys/devicestat.h).
    /// Layout verified against FreeBSD 15 with offsetof() checks.
    /// sizeof = 288, DEVSTAT_N_TRANS_FLAGS = 4.
    #[repr(C)]
    struct Devstat {
        sequence0: u32,                  // offset 0
        allocated: i32,                  // offset 4
        start_count: u32,                // offset 8
        end_count: u32,                  // offset 12
        busy_from: Bintime,              // offset 16 (16 bytes)
        dev_links_next: *mut Devstat,    // offset 32 (STAILQ_ENTRY)
        device_number: u32,              // offset 40
        device_name: [libc::c_char; 16], // offset 44
        unit_number: i32,                // offset 60
        bytes: [u64; 4],                 // offset 64 (N_TRANS_FLAGS=4)
        operations: [u64; 4],            // offset 96
        duration: [Bintime; 4],          // offset 128
        busy_time: Bintime,              // offset 192
        creation_time: Bintime,          // offset 208
        block_size: u32,                 // offset 224
        _pad0: u32,                      // padding for alignment
        tag_types: [u64; 3],             // offset 232
        flags: u32,                      // offset 256
        device_type: u32,                // offset 260
        priority: u32,                   // offset 264
        _pad1: u32,                      // padding for pointer alignment
        id: *const libc::c_void,         // offset 272
        sequence1: u32,                  // offset 280
        _pad2: u32,                      // padding to 288
    }

    /// Device info structure (from devstat.h).
    /// sizeof = 32.
    #[repr(C)]
    struct Devinfo {
        devices: *mut Devstat,    // offset 0
        mem_ptr: *mut u8,         // offset 8
        generation: libc::c_long, // offset 16
        numdevs: libc::c_int,     // offset 24
        _pad: u32,                // padding to 32
    }

    /// Statistics info for devstat_getdevs (from devstat.h).
    /// sizeof = 80, CPUSTATES = 5.
    #[repr(C)]
    struct Statinfo {
        cp_time: [libc::c_long; 5], // offset 0 (40 bytes)
        tk_nin: libc::c_long,       // offset 40
        tk_nout: libc::c_long,      // offset 48
        dinfo: *mut Devinfo,        // offset 56
        snap_time: [u8; 16],        // offset 64 (long double = 16 bytes)
    }

    unsafe extern "C" {
        fn devstat_checkversion(kd: *mut libc::c_void) -> libc::c_int;
        fn devstat_getdevs(kd: *mut libc::c_void, stats: *mut Statinfo) -> libc::c_int;
    }

    /// Collects disk I/O statistics on FreeBSD via libdevstat.
    pub fn get_disk_io_stats() -> Result<Vec<DiskIOStats>> {
        unsafe {
            // Check devstat version compatibility
            if devstat_checkversion(ptr::null_mut()) < 0 {
                return Err(Error::Platform("devstat version mismatch".to_string()));
            }

            // Initialize structures (zeroed, then set dinfo pointer)
            let mut dinfo: Devinfo = mem::zeroed();
            let mut stats: Statinfo = mem::zeroed();
            stats.dinfo = &mut dinfo;

            // Get device statistics
            if devstat_getdevs(ptr::null_mut(), &mut stats) < 0 {
                return Err(Error::Platform(format!(
                    "devstat_getdevs failed: {}",
                    std::io::Error::last_os_error()
                )));
            }

            if dinfo.devices.is_null() || dinfo.numdevs <= 0 {
                return Ok(Vec::new());
            }

            let mut results = Vec::with_capacity(dinfo.numdevs as usize);

            for i in 0..dinfo.numdevs as isize {
                let ds = &*dinfo.devices.offset(i);

                // Get device name
                let name = cstr_to_string(ds.device_name.as_ptr());
                if name.is_empty() {
                    continue;
                }

                let device = format!("{}{}", name, ds.unit_number);

                // Skip devices with no activity
                let reads = ds.operations[0]; // DEVSTAT_READ
                let writes = ds.operations[1]; // DEVSTAT_WRITE
                if reads == 0 && writes == 0 {
                    continue;
                }

                let read_bytes = ds.bytes[0];
                let write_bytes = ds.bytes[1];

                // Calculate time in microseconds from bintime (sec + frac/2^64)
                let read_time_us = bintime_to_us(ds.duration[0].sec, ds.duration[0].frac);
                let write_time_us = bintime_to_us(ds.duration[1].sec, ds.duration[1].frac);
                let busy_time_us = bintime_to_us(ds.busy_time.sec, ds.busy_time.frac);

                // IO in progress: difference between start and end counts
                let io_in_progress = ds.start_count.saturating_sub(ds.end_count) as u64;

                results.push(DiskIOStats {
                    device,
                    reads_completed: reads,
                    read_bytes,
                    read_time_us,
                    writes_completed: writes,
                    write_bytes,
                    write_time_us,
                    io_in_progress,
                    io_time_us: busy_time_us,
                    weighted_io_time_us: busy_time_us,
                });
            }

            Ok(results)
        }
    }

    /// Convert bintime (seconds + fraction) to microseconds.
    fn bintime_to_us(sec: i64, frac: u64) -> u64 {
        // bintime fraction is scaled by 2^64
        // frac / 2^64 gives the fractional seconds
        // Clamp negative values (kernel error/uninitialized) to zero
        let sec_clamped = sec.max(0) as u64;
        // Multiply by 1_000_000 to get microseconds
        let frac_us = (frac as u128 * 1_000_000) >> 64;
        sec_clamped.saturating_mul(1_000_000).saturating_add(frac_us as u64)
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
                do_sysctlbyname(name.as_ptr(), ptr::null_mut(), &mut len, ptr::null_mut(), 0);

            if result != 0 || len == 0 {
                return Ok(Vec::new());
            }

            // Allocate buffer
            let count = len / mem::size_of::<DiskStats>();
            let mut stats: Vec<DiskStats> = vec![mem::zeroed(); count];

            let result = do_sysctlbyname(
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

                // Convert time from seconds + microseconds to microseconds
                let time_us = disk
                    .ds_time_sec
                    .max(0)
                    .saturating_mul(1_000_000)
                    .saturating_add(disk.ds_time_usec.max(0));

                let read_bytes = disk.ds_rbytes;
                let write_bytes = disk.ds_wbytes;

                results.push(DiskIOStats {
                    device,
                    reads_completed: disk.ds_rxfer,
                    read_bytes,
                    read_time_us: (time_us / 2) as u64,
                    writes_completed: disk.ds_wxfer,
                    write_bytes,
                    write_time_us: (time_us / 2) as u64,
                    io_in_progress: disk.ds_busy as u64,
                    io_time_us: time_us as u64,
                    weighted_io_time_us: time_us as u64,
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
                do_sysctlbyname(name.as_ptr(), ptr::null_mut(), &mut len, ptr::null_mut(), 0);

            if result != 0 || len == 0 {
                return Ok(Vec::new());
            }

            // Allocate buffer
            let count = len / mem::size_of::<DiskSysctl>();
            let mut stats: Vec<DiskSysctl> = vec![mem::zeroed(); count];

            let result = do_sysctlbyname(
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

                // Convert time from seconds + microseconds to microseconds
                let time_us = disk
                    .dk_time_sec
                    .max(0)
                    .saturating_mul(1_000_000)
                    .saturating_add(disk.dk_time_usec as i64);

                let read_bytes = disk.dk_rbytes;
                let write_bytes = disk.dk_wbytes;

                results.push(DiskIOStats {
                    device,
                    reads_completed: disk.dk_rxfer,
                    read_bytes,
                    read_time_us: (time_us / 2) as u64,
                    writes_completed: disk.dk_wxfer,
                    write_bytes,
                    write_time_us: (time_us / 2) as u64,
                    io_in_progress: disk.dk_busy as u64,
                    io_time_us: time_us as u64,
                    weighted_io_time_us: time_us as u64,
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
            let read_bytes = fields[5].parse::<u64>().unwrap_or(0) * 512;
            let read_time_us = fields[6].parse::<u64>().unwrap_or(0) * 1000;
            let writes_completed = fields[7].parse().unwrap_or(0);
            let write_bytes = fields[9].parse::<u64>().unwrap_or(0) * 512;
            let write_time_us = fields[10].parse::<u64>().unwrap_or(0) * 1000;
            let io_in_progress = fields[11].parse().unwrap_or(0);
            let io_time_us = fields[12].parse::<u64>().unwrap_or(0) * 1000;
            let weighted_io_time_us = fields[13].parse::<u64>().unwrap_or(0) * 1000;

            results.push(DiskIOStats {
                device,
                reads_completed,
                read_bytes,
                read_time_us,
                writes_completed,
                write_bytes,
                write_time_us,
                io_in_progress,
                io_time_us,
                weighted_io_time_us,
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

/// Gets network interface statistics via sysctl NET_RT_IFLIST.
///
/// # Platform Support
///
/// Works on all BSD platforms (FreeBSD, OpenBSD, NetBSD).
///
/// # Examples
///
/// ```no_run
/// use probe_platform::bsd::sysctl::get_network_stats;
///
/// let stats = get_network_stats()?;
/// for iface in stats {
///     println!("{}: rx={} tx={}", iface.interface, iface.rx_bytes, iface.tx_bytes);
/// }
/// # Ok::<(), probe_platform::Error>(())
/// ```
pub fn get_network_stats() -> Result<Vec<NetStats>> {
    unsafe {
        // MIB path: CTL_NET -> PF_ROUTE -> 0 -> AF_UNSPEC -> NET_RT_IFLIST -> 0
        let mut mib = [
            libc::CTL_NET,
            libc::PF_ROUTE,
            0,
            0, // AF_UNSPEC
            libc::NET_RT_IFLIST,
            0,
        ];

        // Get required buffer size
        let mut len: usize = 0;
        let result =
            libc::sysctl(mib.as_mut_ptr(), 6, ptr::null_mut(), &mut len, ptr::null_mut(), 0);

        if result != 0 || len == 0 {
            return Ok(Vec::new());
        }

        // Allocate buffer
        let mut buf: Vec<u8> = vec![0; len];
        let result = libc::sysctl(
            mib.as_mut_ptr(),
            6,
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        if result != 0 {
            return Err(Error::Io(std::io::Error::last_os_error()));
        }

        let mut stats = Vec::new();
        let mut offset = 0;

        while offset < len {
            // Parse if_msghdr structure
            let ifm = buf.as_ptr().add(offset) as *const IfMsghdr;
            let msg_len = (*ifm).ifm_msglen as usize;

            if msg_len == 0 {
                break;
            }

            // RTM_IFINFO indicates interface information
            if (*ifm).ifm_type as i32 == RTM_IFINFO {
                let data = &(*ifm).ifm_data;

                // Get interface name from index
                let mut ifname = [0i8; 16]; // IF_NAMESIZE
                if !libc::if_indextoname((*ifm).ifm_index as u32, ifname.as_mut_ptr()).is_null() {
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
                        tx_drops: 0, // Not all BSDs expose this
                    });
                }
            }

            offset += msg_len;
        }

        Ok(stats)
    }
}

// RTM_IFINFO message type
const RTM_IFINFO: i32 = 0x0e;

/// BSD if_msghdr structure for routing messages.
#[repr(C)]
struct IfMsghdr {
    ifm_msglen: u16,
    ifm_version: u8,
    ifm_type: u8,
    ifm_addrs: i32,
    ifm_flags: i32,
    ifm_index: u16,
    ifm_data: IfData,
}

/// BSD if_data structure containing interface statistics.
#[repr(C)]
struct IfData {
    ifi_type: u8,
    ifi_physical: u8,
    ifi_addrlen: u8,
    ifi_hdrlen: u8,
    ifi_link_state: u8,
    ifi_spare_char1: u8,
    ifi_spare_char2: u8,
    ifi_datalen: u8,
    ifi_mtu: u64,
    ifi_metric: u64,
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
    ifi_hwassist: u64,
    ifi_epoch: i64,
    ifi_lastchange_sec: i64,
    ifi_lastchange_usec: i64,
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

/// Reads context switches for the current process using getrusage().
///
/// # Platform Support
///
/// Works on all BSD platforms (FreeBSD, OpenBSD, NetBSD) via POSIX getrusage().
///
/// # Examples
///
/// ```no_run
/// use probe_platform::bsd::sysctl::read_self_context_switches;
///
/// let ctx = read_self_context_switches()?;
/// println!("Voluntary: {}, Involuntary: {}", ctx.voluntary, ctx.involuntary);
/// # Ok::<(), probe_platform::Error>(())
/// ```
pub fn read_self_context_switches() -> Result<ContextSwitches> {
    unsafe {
        let mut usage: libc::rusage = mem::zeroed();
        let result = libc::getrusage(libc::RUSAGE_SELF, &mut usage);

        if result != 0 {
            return Err(Error::Platform(format!(
                "getrusage failed: {}",
                std::io::Error::last_os_error()
            )));
        }

        Ok(ContextSwitches {
            voluntary: usage.ru_nvcsw as u64,
            involuntary: usage.ru_nivcsw as u64,
        })
    }
}

/// Reads context switches for a specific process.
///
/// # Platform Support
///
/// - **FreeBSD**: Via kinfo_proc.ki_rusage
/// - **OpenBSD**: Via kinfo_proc.p_uru_nvcsw/p_uru_nivcsw
/// - **NetBSD**: Via kinfo_proc2.p_uru_nvcsw/p_uru_nivcsw
///
/// # Arguments
///
/// * `pid` - Process ID to query. Use 0 for the current process.
///
/// # Errors
///
/// Returns [`Error::NotFound`] if the process doesn't exist.
pub fn read_process_context_switches(pid: i32) -> Result<ContextSwitches> {
    // If pid is 0 or current process, use getrusage for efficiency
    if pid == 0 || pid == unsafe { libc::getpid() } {
        return read_self_context_switches();
    }

    #[cfg(target_os = "freebsd")]
    {
        read_process_context_switches_freebsd(pid)
    }

    #[cfg(target_os = "openbsd")]
    {
        read_process_context_switches_openbsd(pid)
    }

    #[cfg(target_os = "netbsd")]
    {
        read_process_context_switches_netbsd(pid)
    }
}

#[cfg(target_os = "freebsd")]
fn read_process_context_switches_freebsd(pid: i32) -> Result<ContextSwitches> {
    unsafe {
        let mut mib = [libc::CTL_KERN, libc::KERN_PROC, libc::KERN_PROC_PID, pid as libc::c_int];

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

        // FreeBSD stores rusage in ki_rusage
        Ok(ContextSwitches {
            voluntary: kinfo.ki_rusage.ru_nvcsw as u64,
            involuntary: kinfo.ki_rusage.ru_nivcsw as u64,
        })
    }
}

#[cfg(target_os = "openbsd")]
fn read_process_context_switches_openbsd(pid: i32) -> Result<ContextSwitches> {
    unsafe {
        // OpenBSD kinfo_proc structure layout
        #[repr(C)]
        struct KinfoProc {
            _padding1: [u8; 232], // Offset to p_uru_nvcsw varies by version
            p_uru_nvcsw: u64,     // Voluntary context switches
            p_uru_nivcsw: u64,    // Involuntary context switches
            _rest: [u8; 512],     // Remaining fields
        }

        let mut mib = [
            libc::CTL_KERN,
            libc::KERN_PROC,
            libc::KERN_PROC_PID,
            pid as libc::c_int,
            mem::size_of::<KinfoProc>() as libc::c_int,
            1,
        ];

        let mut kinfo: KinfoProc = mem::zeroed();
        let mut len = mem::size_of::<KinfoProc>();

        let result = libc::sysctl(
            mib.as_mut_ptr(),
            6,
            &mut kinfo as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        if result != 0 || len == 0 {
            return Err(Error::NotFound(format!("process {} not found", pid)));
        }

        Ok(ContextSwitches { voluntary: kinfo.p_uru_nvcsw, involuntary: kinfo.p_uru_nivcsw })
    }
}

#[cfg(target_os = "netbsd")]
fn read_process_context_switches_netbsd(pid: i32) -> Result<ContextSwitches> {
    unsafe {
        // NetBSD uses kinfo_proc2 via KERN_PROC2
        #[repr(C)]
        struct KinfoProc2 {
            _padding1: [u8; 304], // Offset to context switch fields
            p_uru_nvcsw: u64,     // Voluntary context switches
            p_uru_nivcsw: u64,    // Involuntary context switches
            _rest: [u8; 256],     // Remaining fields
        }

        // NetBSD KERN_PROC2 = 47
        const KERN_PROC2: libc::c_int = 47;

        let mut mib = [
            libc::CTL_KERN,
            KERN_PROC2,
            libc::KERN_PROC_PID,
            pid as libc::c_int,
            mem::size_of::<KinfoProc2>() as libc::c_int,
            1,
        ];

        let mut kinfo: KinfoProc2 = mem::zeroed();
        let mut len = mem::size_of::<KinfoProc2>();

        let result = libc::sysctl(
            mib.as_mut_ptr(),
            6,
            &mut kinfo as *mut _ as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        );

        if result != 0 || len == 0 {
            return Err(Error::NotFound(format!("process {} not found", pid)));
        }

        Ok(ContextSwitches { voluntary: kinfo.p_uru_nvcsw, involuntary: kinfo.p_uru_nivcsw })
    }
}

/// Reads system-wide context switch count.
///
/// # Platform Support
///
/// BSD systems don't expose a direct system-wide context switch counter.
/// This function aggregates from all running processes.
///
/// # Note
///
/// This is an expensive operation as it iterates all processes.
/// For frequent sampling, consider caching the result.
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
    /// Process ID owning this connection (0 if unknown).
    pub pid: i32,
}

/// Lists all network connections on the system.
///
/// # Platform Support
///
/// - **FreeBSD**: Via `net.inet.tcp.pcblist` and `net.inet.udp.pcblist` sysctls
/// - **OpenBSD**: Via `net.inet.tcp.pcblist` and `net.inet.udp.pcblist` sysctls
/// - **NetBSD**: Via `net.inet.tcp.pcblist` sysctl
///
/// # Examples
///
/// ```no_run
/// use probe_platform::bsd::sysctl::list_network_connections;
///
/// let connections = list_network_connections()?;
/// for conn in connections {
///     println!("{} {} -> {} ({})",
///         conn.protocol, conn.local_addr, conn.remote_addr, conn.state);
/// }
/// # Ok::<(), probe_platform::Error>(())
/// ```
pub fn list_network_connections() -> Result<Vec<NetworkConnection>> {
    let mut connections = Vec::new();

    // Get TCP connections
    if let Ok(tcp) = list_tcp_connections() {
        connections.extend(tcp);
    }

    // Get UDP sockets
    if let Ok(udp) = list_udp_connections() {
        connections.extend(udp);
    }

    // Get Unix domain sockets
    if let Ok(unix) = list_unix_connections() {
        connections.extend(unix);
    }

    Ok(connections)
}

/// Lists TCP connections via sysctl.
#[cfg(target_os = "freebsd")]
fn list_tcp_connections() -> Result<Vec<NetworkConnection>> {
    unsafe {
        let name = CString::new("net.inet.tcp.pcblist")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

        // Get required buffer size
        let mut len: usize = 0;
        if do_sysctlbyname(name.as_ptr(), ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        if len == 0 {
            return Ok(Vec::new());
        }

        // Allocate buffer with extra space for potential growth
        len = len * 2;
        let mut buf: Vec<u8> = vec![0; len];

        if do_sysctlbyname(
            name.as_ptr(),
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Ok(Vec::new());
        }

        parse_tcp_pcblist_freebsd(&buf[..len])
    }
}

#[cfg(target_os = "freebsd")]
fn parse_tcp_pcblist_freebsd(buf: &[u8]) -> Result<Vec<NetworkConnection>> {
    // FreeBSD pcblist format uses xinpgen/xtcpcb structures
    // This is a simplified parser that extracts basic connection info

    let mut connections = Vec::new();

    // The buffer starts with a xinpgen header, followed by xtcpcb entries
    // Each entry has an xig_len field indicating its size

    // Skip the header (xinpgen)
    if buf.len() < 32 {
        return Ok(connections);
    }

    let mut offset = 0;

    // Read xinpgen header to get structure size
    #[repr(C)]
    struct Xinpgen {
        xig_len: u32,
        xig_count: u32,
        xig_gen: u64,
        xig_sogen: u64,
    }

    let header = buf.as_ptr() as *const Xinpgen;
    let header_len = unsafe { (*header).xig_len } as usize;

    if header_len == 0 || header_len > buf.len() {
        return Ok(connections);
    }

    offset = header_len;

    // Parse xtcpcb entries
    while offset + 256 < buf.len() {
        // Read length of this entry
        let entry_ptr = unsafe { buf.as_ptr().add(offset) };
        let entry_len = unsafe { *(entry_ptr as *const u32) } as usize;

        if entry_len == 0 || entry_len < 256 {
            break;
        }

        // Extract connection info from xtcpcb
        // Offsets are approximate and may vary by FreeBSD version

        // Local address at offset ~176 (in_addr + port)
        // Remote address at offset ~192 (in_addr + port)
        // State at offset ~216

        if offset + entry_len <= buf.len() {
            // Try to parse addresses (simplified - assumes IPv4)
            let local_ip_offset = offset + 176;
            let local_port_offset = offset + 180;
            let remote_ip_offset = offset + 192;
            let remote_port_offset = offset + 196;
            let state_offset = offset + 216;

            if state_offset + 4 <= buf.len() {
                let local_ip = u32::from_ne_bytes([
                    buf[local_ip_offset],
                    buf[local_ip_offset + 1],
                    buf[local_ip_offset + 2],
                    buf[local_ip_offset + 3],
                ]);
                let local_port =
                    u16::from_be_bytes([buf[local_port_offset], buf[local_port_offset + 1]]);

                let remote_ip = u32::from_ne_bytes([
                    buf[remote_ip_offset],
                    buf[remote_ip_offset + 1],
                    buf[remote_ip_offset + 2],
                    buf[remote_ip_offset + 3],
                ]);
                let remote_port =
                    u16::from_be_bytes([buf[remote_port_offset], buf[remote_port_offset + 1]]);

                let state_val = buf[state_offset] as i32;
                let state = tcp_state_from_int(state_val);

                // Only add if we have a valid port
                if local_port > 0 || state == ConnectionState::Listen {
                    let local_addr = format!(
                        "{}.{}.{}.{}:{}",
                        local_ip & 0xFF,
                        (local_ip >> 8) & 0xFF,
                        (local_ip >> 16) & 0xFF,
                        (local_ip >> 24) & 0xFF,
                        local_port
                    );

                    let remote_addr = if remote_port > 0 {
                        format!(
                            "{}.{}.{}.{}:{}",
                            remote_ip & 0xFF,
                            (remote_ip >> 8) & 0xFF,
                            (remote_ip >> 16) & 0xFF,
                            (remote_ip >> 24) & 0xFF,
                            remote_port
                        )
                    } else {
                        String::new()
                    };

                    connections.push(NetworkConnection {
                        protocol: ConnectionProtocol::Tcp,
                        local_addr,
                        remote_addr,
                        state,
                        pid: 0, // PID not easily available from pcblist
                    });
                }
            }
        }

        offset += entry_len;
    }

    Ok(connections)
}

/// Converts TCP state integer to ConnectionState.
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

/// Lists UDP sockets via sysctl.
#[cfg(target_os = "freebsd")]
fn list_udp_connections() -> Result<Vec<NetworkConnection>> {
    unsafe {
        let name = CString::new("net.inet.udp.pcblist")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

        let mut len: usize = 0;
        if do_sysctlbyname(name.as_ptr(), ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        if len == 0 {
            return Ok(Vec::new());
        }

        len = len * 2;
        let mut buf: Vec<u8> = vec![0; len];

        if do_sysctlbyname(
            name.as_ptr(),
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Ok(Vec::new());
        }

        // UDP parsing is similar to TCP but without state
        let mut connections = Vec::new();

        // Skip xinpgen header
        if len < 32 {
            return Ok(connections);
        }

        let header = buf.as_ptr() as *const u32;
        let header_len = unsafe { *header } as usize;

        if header_len == 0 || header_len > len {
            return Ok(connections);
        }

        let mut offset = header_len;

        while offset + 200 < len {
            let entry_ptr = buf.as_ptr().add(offset);
            let entry_len = unsafe { *(entry_ptr as *const u32) } as usize;

            if entry_len == 0 || entry_len < 100 {
                break;
            }

            // Extract UDP socket info (simplified)
            let local_ip_offset = offset + 176;
            let local_port_offset = offset + 180;

            if local_port_offset + 2 <= len {
                let local_ip = u32::from_ne_bytes([
                    buf[local_ip_offset],
                    buf[local_ip_offset + 1],
                    buf[local_ip_offset + 2],
                    buf[local_ip_offset + 3],
                ]);
                let local_port =
                    u16::from_be_bytes([buf[local_port_offset], buf[local_port_offset + 1]]);

                if local_port > 0 {
                    let local_addr = format!(
                        "{}.{}.{}.{}:{}",
                        local_ip & 0xFF,
                        (local_ip >> 8) & 0xFF,
                        (local_ip >> 16) & 0xFF,
                        (local_ip >> 24) & 0xFF,
                        local_port
                    );

                    connections.push(NetworkConnection {
                        protocol: ConnectionProtocol::Udp,
                        local_addr,
                        remote_addr: String::new(),
                        state: ConnectionState::Established, // UDP is stateless
                        pid: 0,
                    });
                }
            }

            offset += entry_len;
        }

        Ok(connections)
    }
}

/// Lists Unix domain sockets.
#[cfg(target_os = "freebsd")]
fn list_unix_connections() -> Result<Vec<NetworkConnection>> {
    unsafe {
        let name = CString::new("net.local.stream.pcblist")
            .map_err(|e| Error::Platform(format!("invalid sysctl name: {}", e)))?;

        let mut len: usize = 0;
        if do_sysctlbyname(name.as_ptr(), ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        // Unix socket parsing is more complex, returning empty for now
        // A full implementation would parse xunpcb structures
        Ok(Vec::new())
    }
}

// ============================================================================
// OPENBSD NETWORK CONNECTIONS
// ============================================================================

#[cfg(target_os = "openbsd")]
fn list_tcp_connections() -> Result<Vec<NetworkConnection>> {
    unsafe {
        // OpenBSD uses sysctl with CTL_NET, PF_INET, IPPROTO_TCP, TCPCTL_PCBLIST
        const IPPROTO_TCP: libc::c_int = 6;
        const TCPCTL_PCBLIST: libc::c_int = 5;

        let mut mib = [libc::CTL_NET, libc::PF_INET, IPPROTO_TCP, TCPCTL_PCBLIST];

        // Get required buffer size
        let mut len: usize = 0;
        if libc::sysctl(mib.as_mut_ptr(), 4, ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        if len == 0 {
            return Ok(Vec::new());
        }

        // Allocate buffer with extra space
        len = len * 2;
        let mut buf: Vec<u8> = vec![0; len];

        if libc::sysctl(
            mib.as_mut_ptr(),
            4,
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Ok(Vec::new());
        }

        parse_tcp_pcblist_openbsd(&buf[..len])
    }
}

#[cfg(target_os = "openbsd")]
fn parse_tcp_pcblist_openbsd(buf: &[u8]) -> Result<Vec<NetworkConnection>> {
    // OpenBSD pcblist uses struct inpcb and tcpcb
    // Structure: each entry starts with a length field

    let mut connections = Vec::new();
    let mut offset = 0;

    // Skip header
    if buf.len() < 16 {
        return Ok(connections);
    }

    // OpenBSD inpcbtable header
    #[repr(C)]
    struct Inpcbhead {
        total_len: u32,
        count: u32,
        _padding: [u8; 8],
    }

    let header = buf.as_ptr() as *const Inpcbhead;
    let entry_count = unsafe { (*header).count } as usize;
    offset = 16; // Skip header

    // OpenBSD inpcb structure (simplified)
    #[repr(C)]
    struct InpcbEntry {
        inp_len: u32,
        inp_faddr: [u8; 4], // Foreign IPv4 address
        inp_fport: u16,     // Foreign port (network byte order)
        inp_laddr: [u8; 4], // Local IPv4 address
        inp_lport: u16,     // Local port (network byte order)
        t_state: u8,        // TCP state
        _padding: [u8; 3],
    }

    for _ in 0..entry_count.min(4096) {
        if offset + 24 > buf.len() {
            break;
        }

        // Read entry length
        let entry_len =
            u32::from_ne_bytes([buf[offset], buf[offset + 1], buf[offset + 2], buf[offset + 3]])
                as usize;
        if entry_len == 0 || entry_len < 20 {
            break;
        }

        // Parse addresses at known offsets
        let laddr_offset = offset + 8;
        let lport_offset = offset + 12;
        let faddr_offset = offset + 14;
        let fport_offset = offset + 18;
        let state_offset = offset + 20;

        if state_offset + 1 <= buf.len() {
            let local_ip = format!(
                "{}.{}.{}.{}",
                buf[laddr_offset],
                buf[laddr_offset + 1],
                buf[laddr_offset + 2],
                buf[laddr_offset + 3]
            );
            let local_port = u16::from_be_bytes([buf[lport_offset], buf[lport_offset + 1]]);

            let remote_ip = format!(
                "{}.{}.{}.{}",
                buf[faddr_offset],
                buf[faddr_offset + 1],
                buf[faddr_offset + 2],
                buf[faddr_offset + 3]
            );
            let remote_port = u16::from_be_bytes([buf[fport_offset], buf[fport_offset + 1]]);

            let state_val = buf[state_offset] as i32;
            let state = tcp_state_from_int(state_val);

            if local_port > 0 || state == ConnectionState::Listen {
                connections.push(NetworkConnection {
                    protocol: ConnectionProtocol::Tcp,
                    local_addr: format!("{}:{}", local_ip, local_port),
                    remote_addr: if remote_port > 0 {
                        format!("{}:{}", remote_ip, remote_port)
                    } else {
                        String::new()
                    },
                    state,
                    pid: 0,
                });
            }
        }

        offset += entry_len.max(24);
    }

    Ok(connections)
}

#[cfg(target_os = "openbsd")]
fn list_udp_connections() -> Result<Vec<NetworkConnection>> {
    unsafe {
        const IPPROTO_UDP: libc::c_int = 17;
        const UDPCTL_PCBLIST: libc::c_int = 5;

        let mut mib = [libc::CTL_NET, libc::PF_INET, IPPROTO_UDP, UDPCTL_PCBLIST];

        let mut len: usize = 0;
        if libc::sysctl(mib.as_mut_ptr(), 4, ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        if len == 0 {
            return Ok(Vec::new());
        }

        len = len * 2;
        let mut buf: Vec<u8> = vec![0; len];

        if libc::sysctl(
            mib.as_mut_ptr(),
            4,
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Ok(Vec::new());
        }

        // Similar parsing to TCP but simpler (no state)
        let mut connections = Vec::new();
        let mut offset = 16; // Skip header

        while offset + 20 < len {
            let entry_len = u32::from_ne_bytes([
                buf[offset],
                buf[offset + 1],
                buf[offset + 2],
                buf[offset + 3],
            ]) as usize;
            if entry_len == 0 || entry_len < 16 {
                break;
            }

            let laddr_offset = offset + 8;
            let lport_offset = offset + 12;

            if lport_offset + 2 <= len {
                let local_ip = format!(
                    "{}.{}.{}.{}",
                    buf[laddr_offset],
                    buf[laddr_offset + 1],
                    buf[laddr_offset + 2],
                    buf[laddr_offset + 3]
                );
                let local_port = u16::from_be_bytes([buf[lport_offset], buf[lport_offset + 1]]);

                if local_port > 0 {
                    connections.push(NetworkConnection {
                        protocol: ConnectionProtocol::Udp,
                        local_addr: format!("{}:{}", local_ip, local_port),
                        remote_addr: String::new(),
                        state: ConnectionState::Established,
                        pid: 0,
                    });
                }
            }

            offset += entry_len.max(20);
        }

        Ok(connections)
    }
}

#[cfg(target_os = "openbsd")]
fn list_unix_connections() -> Result<Vec<NetworkConnection>> {
    // Unix sockets on OpenBSD require kvm access
    // Return empty for now
    Ok(Vec::new())
}

// ============================================================================
// NETBSD NETWORK CONNECTIONS
// ============================================================================

#[cfg(target_os = "netbsd")]
fn list_tcp_connections() -> Result<Vec<NetworkConnection>> {
    unsafe {
        const IPPROTO_TCP: libc::c_int = 6;
        const TCPCTL_PCBLIST: libc::c_int = 5;

        let mut mib = [libc::CTL_NET, libc::PF_INET, IPPROTO_TCP, TCPCTL_PCBLIST];

        let mut len: usize = 0;
        if libc::sysctl(mib.as_mut_ptr(), 4, ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        if len == 0 {
            return Ok(Vec::new());
        }

        len = len * 2;
        let mut buf: Vec<u8> = vec![0; len];

        if libc::sysctl(
            mib.as_mut_ptr(),
            4,
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Ok(Vec::new());
        }

        parse_tcp_pcblist_netbsd(&buf[..len])
    }
}

#[cfg(target_os = "netbsd")]
fn parse_tcp_pcblist_netbsd(buf: &[u8]) -> Result<Vec<NetworkConnection>> {
    // NetBSD uses struct kinfo_pcb via sysctl
    // Structure is well-documented in sys/socket.h

    let mut connections = Vec::new();

    // NetBSD kinfo_pcb structure
    #[repr(C)]
    struct KinfoPcb {
        ki_pcbaddr: u64,  // PCB address
        ki_ppcbaddr: u64, // Parent PCB address
        ki_sockaddr: u64, // Socket address
        ki_family: u32,   // Address family (AF_INET, AF_INET6)
        ki_type: u32,     // Socket type (SOCK_STREAM, etc)
        ki_protocol: u32, // Protocol (IPPROTO_TCP, etc)
        ki_pflags: u32,   // PCB flags
        ki_sostate: u32,  // Socket state
        ki_prstate: u32,  // Protocol state (TCP state)
        ki_tstate: u32,   // Timer state
        ki_rcvq: u32,     // Receive queue length
        ki_sndq: u32,     // Send queue length
        // Addresses follow in sockaddr_storage format
        ki_s: [u8; 128], // Local address
        ki_d: [u8; 128], // Remote address
    }

    const KINFO_PCB_SIZE: usize = std::mem::size_of::<KinfoPcb>();

    let entry_count = buf.len() / KINFO_PCB_SIZE;

    for i in 0..entry_count {
        let offset = i * KINFO_PCB_SIZE;
        if offset + KINFO_PCB_SIZE > buf.len() {
            break;
        }

        let entry_ptr = unsafe { buf.as_ptr().add(offset) as *const KinfoPcb };
        let entry = unsafe { &*entry_ptr };

        // Only process IPv4 TCP connections
        if entry.ki_family != libc::AF_INET as u32 {
            continue;
        }

        // Parse local address from sockaddr_in at ki_s
        // sockaddr_in: family(2) + port(2) + addr(4) + zero(8)
        let local_port = u16::from_be_bytes([entry.ki_s[2], entry.ki_s[3]]);
        let local_ip =
            format!("{}.{}.{}.{}", entry.ki_s[4], entry.ki_s[5], entry.ki_s[6], entry.ki_s[7]);

        // Parse remote address from sockaddr_in at ki_d
        let remote_port = u16::from_be_bytes([entry.ki_d[2], entry.ki_d[3]]);
        let remote_ip =
            format!("{}.{}.{}.{}", entry.ki_d[4], entry.ki_d[5], entry.ki_d[6], entry.ki_d[7]);

        let state = tcp_state_from_int(entry.ki_prstate as i32);

        if local_port > 0 || state == ConnectionState::Listen {
            connections.push(NetworkConnection {
                protocol: ConnectionProtocol::Tcp,
                local_addr: format!("{}:{}", local_ip, local_port),
                remote_addr: if remote_port > 0 {
                    format!("{}:{}", remote_ip, remote_port)
                } else {
                    String::new()
                },
                state,
                pid: 0,
            });
        }
    }

    Ok(connections)
}

#[cfg(target_os = "netbsd")]
fn list_udp_connections() -> Result<Vec<NetworkConnection>> {
    unsafe {
        const IPPROTO_UDP: libc::c_int = 17;
        const UDPCTL_PCBLIST: libc::c_int = 1;

        let mut mib = [libc::CTL_NET, libc::PF_INET, IPPROTO_UDP, UDPCTL_PCBLIST];

        let mut len: usize = 0;
        if libc::sysctl(mib.as_mut_ptr(), 4, ptr::null_mut(), &mut len, ptr::null_mut(), 0) != 0 {
            return Ok(Vec::new());
        }

        if len == 0 {
            return Ok(Vec::new());
        }

        len = len * 2;
        let mut buf: Vec<u8> = vec![0; len];

        if libc::sysctl(
            mib.as_mut_ptr(),
            4,
            buf.as_mut_ptr() as *mut libc::c_void,
            &mut len,
            ptr::null_mut(),
            0,
        ) != 0
        {
            return Ok(Vec::new());
        }

        // NetBSD kinfo_pcb structure (same as TCP)
        #[repr(C)]
        struct KinfoPcb {
            ki_pcbaddr: u64,
            ki_ppcbaddr: u64,
            ki_sockaddr: u64,
            ki_family: u32,
            ki_type: u32,
            ki_protocol: u32,
            ki_pflags: u32,
            ki_sostate: u32,
            ki_prstate: u32,
            ki_tstate: u32,
            ki_rcvq: u32,
            ki_sndq: u32,
            ki_s: [u8; 128],
            ki_d: [u8; 128],
        }

        const KINFO_PCB_SIZE: usize = std::mem::size_of::<KinfoPcb>();

        let mut connections = Vec::new();
        let entry_count = len / KINFO_PCB_SIZE;

        for i in 0..entry_count {
            let offset = i * KINFO_PCB_SIZE;
            if offset + KINFO_PCB_SIZE > len {
                break;
            }

            let entry_ptr = buf.as_ptr().add(offset) as *const KinfoPcb;
            let entry = &*entry_ptr;

            if entry.ki_family != libc::AF_INET as u32 {
                continue;
            }

            let local_port = u16::from_be_bytes([entry.ki_s[2], entry.ki_s[3]]);
            let local_ip =
                format!("{}.{}.{}.{}", entry.ki_s[4], entry.ki_s[5], entry.ki_s[6], entry.ki_s[7]);

            if local_port > 0 {
                connections.push(NetworkConnection {
                    protocol: ConnectionProtocol::Udp,
                    local_addr: format!("{}:{}", local_ip, local_port),
                    remote_addr: String::new(),
                    state: ConnectionState::Established,
                    pid: 0,
                });
            }
        }

        Ok(connections)
    }
}

#[cfg(target_os = "netbsd")]
fn list_unix_connections() -> Result<Vec<NetworkConnection>> {
    Ok(Vec::new())
}

// ============================================================================
// HELPER FUNCTIONS
// ============================================================================

unsafe fn cstr_to_string(ptr: *const libc::c_char) -> String {
    if ptr.is_null() {
        return String::new();
    }
    unsafe { std::ffi::CStr::from_ptr(ptr).to_string_lossy().into_owned() }
}

// ============================================================================
// TESTS
// ============================================================================

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_get_memory_info_returns_valid_data() {
        let result = get_memory_info();
        assert!(result.is_ok(), "get_memory_info() should succeed on BSD");

        let mem = result.unwrap();
        assert!(mem.total > 0, "total memory should be non-zero");
        assert!(mem.available <= mem.total, "available should not exceed total");
    }

    #[test]
    fn test_get_buffers_bytes_returns_reasonable_value() {
        let page_size_raw = unsafe { libc::sysconf(libc::_SC_PAGESIZE) };
        assert!(page_size_raw > 0, "page size should be positive");
        let page_size = page_size_raw as u64;

        let buffers = get_buffers_bytes(page_size);
        // On real BSD systems, buffers should be >= 0 (may be 0 on ZFS-only FreeBSD).
        // On non-BSD test hosts (Linux CI), the cfg gates make this return 0.
        // We simply verify it doesn't panic and returns a sane value.
        let total_mem = unsafe {
            let name = CString::new("hw.physmem").unwrap();
            let mut physmem: u64 = 0;
            let mut len = mem::size_of::<u64>();
            do_sysctlbyname(
                name.as_ptr(),
                &mut physmem as *mut _ as *mut libc::c_void,
                &mut len,
                ptr::null_mut(),
                0,
            );
            physmem
        };

        if total_mem > 0 {
            assert!(
                buffers <= total_mem,
                "buffers ({buffers}) should not exceed total memory ({total_mem})"
            );
        }
    }

    #[test]
    fn test_buffers_included_in_memory_info() {
        let result = get_memory_info();
        assert!(result.is_ok());

        let mem = result.unwrap();
        // Verify that buffers field is populated (not just hardcoded to 0).
        // On FreeBSD with ZFS, buffers may legitimately be 0.
        // On OpenBSD/NetBSD, it should be > 0 if the system has any buffer cache.
        // We check the value is bounded by total memory.
        assert!(
            mem.buffers <= mem.total,
            "buffers ({}) should not exceed total memory ({})",
            mem.buffers,
            mem.total
        );
    }

    #[test]
    fn test_get_cpu_times_does_not_panic() {
        let result = get_cpu_times();
        assert!(result.is_ok(), "get_cpu_times() should succeed on BSD");

        let times = result.unwrap();
        let total = times.user_percent + times.system_percent + times.idle_percent;
        assert!((total - 100.0).abs() < 1.0, "CPU percentages should sum to ~100%, got {total}");
    }

    #[test]
    fn test_get_cpu_info_returns_valid_data() {
        let result = get_cpu_info();
        assert!(result.is_ok(), "get_cpu_info() should succeed on BSD");

        let info = result.unwrap();
        assert!(info.cores > 0, "should have at least 1 CPU core");
    }

    #[test]
    fn test_get_loadavg_returns_valid_data() {
        let result = get_loadavg();
        assert!(result.is_ok(), "get_loadavg() should succeed on BSD");

        let load = result.unwrap();
        assert!(load.load_1min >= 0.0, "load average should be non-negative");
        assert!(load.load_5min >= 0.0, "load average should be non-negative");
        assert!(load.load_15min >= 0.0, "load average should be non-negative");
    }
}
