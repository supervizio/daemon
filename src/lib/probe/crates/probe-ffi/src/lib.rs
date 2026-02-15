//! probe-ffi - C ABI interface for the probe library
//!
//! This crate exposes the FFI functions that Go calls via CGO.
//! All types are repr(C) for C ABI compatibility.

use libc::{c_char, c_int};
use std::ffi::CString;
use std::ptr;
use std::sync::OnceLock;

use probe_metrics::{ProcessState as MetricsProcessState, SystemCollector};
use probe_platform::{PlatformCollector, new_collector};

// Global collector instance
static COLLECTOR: OnceLock<PlatformCollector> = OnceLock::new();

// ============================================================================
// ERROR CODES
// ============================================================================

/// Operation successful.
pub const PROBE_OK: c_int = 0;
/// Operation not supported on this platform.
pub const PROBE_ERR_NOT_SUPPORTED: c_int = 1;
/// Permission denied.
pub const PROBE_ERR_PERMISSION: c_int = 2;
/// Resource not found.
pub const PROBE_ERR_NOT_FOUND: c_int = 3;
/// Invalid parameter.
pub const PROBE_ERR_INVALID_PARAM: c_int = 4;
/// I/O error.
pub const PROBE_ERR_IO: c_int = 5;
/// Internal error.
pub const PROBE_ERR_INTERNAL: c_int = 99;

// ============================================================================
// C-COMPATIBLE TYPES
// ============================================================================

/// Result type for FFI calls.
#[repr(C)]
pub struct ProbeResult {
    /// Whether the operation succeeded.
    pub success: bool,
    /// Error code (PROBE_OK if success).
    pub error_code: c_int,
    /// Error message (NULL if success). Caller must NOT free this.
    pub error_message: *const c_char,
}

impl ProbeResult {
    fn ok() -> Self {
        Self { success: true, error_code: PROBE_OK, error_message: ptr::null() }
    }

    fn err(code: c_int, message: *const c_char) -> Self {
        Self { success: false, error_code: code, error_message: message }
    }

    fn from_metrics_error(e: probe_metrics::Error) -> Self {
        match e {
            probe_metrics::Error::NotSupported => {
                Self::err(PROBE_ERR_NOT_SUPPORTED, c"operation not supported".as_ptr())
            }
            probe_metrics::Error::Permission(_) => {
                Self::err(PROBE_ERR_PERMISSION, c"permission denied".as_ptr())
            }
            probe_metrics::Error::NotFound(_) => {
                Self::err(PROBE_ERR_NOT_FOUND, c"resource not found".as_ptr())
            }
            probe_metrics::Error::Io(_) => Self::err(PROBE_ERR_IO, c"I/O error".as_ptr()),
            probe_metrics::Error::Platform(_) => {
                Self::err(PROBE_ERR_INTERNAL, c"platform error".as_ptr())
            }
        }
    }
}

/// System CPU metrics.
#[repr(C)]
pub struct SystemCPU {
    pub user_percent: f64,
    pub system_percent: f64,
    pub idle_percent: f64,
    pub iowait_percent: f64,
    pub steal_percent: f64,
    pub cores: u32,
    pub frequency_mhz: u64,
}

impl From<probe_metrics::SystemCPU> for SystemCPU {
    fn from(cpu: probe_metrics::SystemCPU) -> Self {
        Self {
            user_percent: cpu.user_percent,
            system_percent: cpu.system_percent,
            idle_percent: cpu.idle_percent,
            iowait_percent: cpu.iowait_percent,
            steal_percent: cpu.steal_percent,
            cores: cpu.cores,
            frequency_mhz: cpu.frequency_mhz,
        }
    }
}

/// System memory metrics.
#[repr(C)]
pub struct SystemMemory {
    pub total_bytes: u64,
    pub available_bytes: u64,
    pub used_bytes: u64,
    pub cached_bytes: u64,
    pub buffers_bytes: u64,
    pub swap_total_bytes: u64,
    pub swap_used_bytes: u64,
}

impl From<probe_metrics::SystemMemory> for SystemMemory {
    fn from(mem: probe_metrics::SystemMemory) -> Self {
        Self {
            total_bytes: mem.total_bytes,
            available_bytes: mem.available_bytes,
            used_bytes: mem.used_bytes,
            cached_bytes: mem.cached_bytes,
            buffers_bytes: mem.buffers_bytes,
            swap_total_bytes: mem.swap_total_bytes,
            swap_used_bytes: mem.swap_used_bytes,
        }
    }
}

/// Load average.
#[repr(C)]
pub struct LoadAverage {
    pub load_1min: f64,
    pub load_5min: f64,
    pub load_15min: f64,
}

impl From<probe_metrics::LoadAverage> for LoadAverage {
    fn from(load: probe_metrics::LoadAverage) -> Self {
        Self { load_1min: load.load_1min, load_5min: load.load_5min, load_15min: load.load_15min }
    }
}

/// Process state.
#[repr(C)]
pub enum ProcessState {
    Running = 0,
    Sleeping = 1,
    Waiting = 2,
    Zombie = 3,
    Stopped = 4,
    Unknown = 255,
}

impl From<MetricsProcessState> for ProcessState {
    fn from(state: MetricsProcessState) -> Self {
        match state {
            MetricsProcessState::Running => ProcessState::Running,
            MetricsProcessState::Sleeping => ProcessState::Sleeping,
            MetricsProcessState::Waiting => ProcessState::Waiting,
            MetricsProcessState::Zombie => ProcessState::Zombie,
            MetricsProcessState::Stopped => ProcessState::Stopped,
            MetricsProcessState::Unknown => ProcessState::Unknown,
        }
    }
}

/// Process metrics.
#[repr(C)]
pub struct ProcessMetrics {
    pub pid: i32,
    pub cpu_percent: f64,
    pub memory_rss_bytes: u64,
    pub memory_vms_bytes: u64,
    pub memory_percent: f64,
    pub num_threads: u32,
    pub num_fds: u32,
    pub read_bytes_per_sec: u64,
    pub write_bytes_per_sec: u64,
    pub state: ProcessState,
}

impl From<probe_metrics::ProcessMetrics> for ProcessMetrics {
    fn from(p: probe_metrics::ProcessMetrics) -> Self {
        Self {
            pid: p.pid,
            cpu_percent: p.cpu_percent,
            memory_rss_bytes: p.memory_rss_bytes,
            memory_vms_bytes: p.memory_vms_bytes,
            memory_percent: p.memory_percent,
            num_threads: p.num_threads,
            num_fds: p.num_fds,
            read_bytes_per_sec: p.read_bytes_per_sec,
            write_bytes_per_sec: p.write_bytes_per_sec,
            state: p.state.into(),
        }
    }
}

/// Resource quota limits (read-only detection).
#[repr(C)]
#[derive(Default)]
pub struct QuotaLimits {
    /// CPU quota in microseconds per period (0 = not set, u64::MAX = unlimited).
    pub cpu_quota_us: u64,
    /// CPU period in microseconds (typically 100000).
    pub cpu_period_us: u64,
    /// Memory limit in bytes (0 = not set, u64::MAX = unlimited).
    pub memory_limit_bytes: u64,
    /// Maximum PIDs (0 = not set, u64::MAX = unlimited).
    pub pids_limit: u64,
    /// Maximum file descriptors (0 = not set, u64::MAX = unlimited).
    pub nofile_limit: u64,
    /// Maximum CPU time in seconds (0 = not set, u64::MAX = unlimited).
    pub cpu_time_limit_secs: u64,
    /// Maximum data/heap size in bytes (0 = not set, u64::MAX = unlimited).
    pub data_limit_bytes: u64,
    /// I/O read bandwidth limit in bytes/sec (0 = not set).
    pub io_read_bps: u64,
    /// I/O write bandwidth limit in bytes/sec (0 = not set).
    pub io_write_bps: u64,
    /// Flags indicating which fields are valid.
    pub flags: u32,
}

// QuotaLimits flags
const QUOTA_FLAG_CPU: u32 = 1 << 0;
const QUOTA_FLAG_MEMORY: u32 = 1 << 1;
const QUOTA_FLAG_PIDS: u32 = 1 << 2;
const QUOTA_FLAG_NOFILE: u32 = 1 << 3;
const QUOTA_FLAG_CPU_TIME: u32 = 1 << 4;
const QUOTA_FLAG_DATA: u32 = 1 << 5;
const QUOTA_FLAG_IO_READ: u32 = 1 << 6;
const QUOTA_FLAG_IO_WRITE: u32 = 1 << 7;

impl From<probe_quota::QuotaLimits> for QuotaLimits {
    fn from(l: probe_quota::QuotaLimits) -> Self {
        let mut flags = 0u32;

        if l.cpu_quota_us.is_some() {
            flags |= QUOTA_FLAG_CPU;
        }
        if l.memory_limit_bytes.is_some() {
            flags |= QUOTA_FLAG_MEMORY;
        }
        if l.pids_limit.is_some() {
            flags |= QUOTA_FLAG_PIDS;
        }
        if l.nofile_limit.is_some() {
            flags |= QUOTA_FLAG_NOFILE;
        }
        if l.cpu_time_limit_secs.is_some() {
            flags |= QUOTA_FLAG_CPU_TIME;
        }
        if l.data_limit_bytes.is_some() {
            flags |= QUOTA_FLAG_DATA;
        }
        if l.io_read_bps.is_some() {
            flags |= QUOTA_FLAG_IO_READ;
        }
        if l.io_write_bps.is_some() {
            flags |= QUOTA_FLAG_IO_WRITE;
        }

        Self {
            cpu_quota_us: l.cpu_quota_us.unwrap_or(0),
            cpu_period_us: l.cpu_period_us.unwrap_or(0),
            memory_limit_bytes: l.memory_limit_bytes.unwrap_or(0),
            pids_limit: l.pids_limit.unwrap_or(0),
            nofile_limit: l.nofile_limit.unwrap_or(0),
            cpu_time_limit_secs: l.cpu_time_limit_secs.unwrap_or(0),
            data_limit_bytes: l.data_limit_bytes.unwrap_or(0),
            io_read_bps: l.io_read_bps.unwrap_or(0),
            io_write_bps: l.io_write_bps.unwrap_or(0),
            flags,
        }
    }
}

/// Current resource usage.
#[repr(C)]
pub struct QuotaUsage {
    /// Current memory usage in bytes.
    pub memory_bytes: u64,
    /// Memory limit in bytes (0 = no limit).
    pub memory_limit_bytes: u64,
    /// Current number of processes/threads.
    pub pids_current: u64,
    /// PIDs limit (0 = no limit).
    pub pids_limit: u64,
    /// Current CPU usage percentage.
    pub cpu_percent: f64,
    /// CPU limit percentage (0 = no limit).
    pub cpu_limit_percent: f64,
}

impl Default for QuotaUsage {
    fn default() -> Self {
        Self {
            memory_bytes: 0,
            memory_limit_bytes: 0,
            pids_current: 0,
            pids_limit: 0,
            cpu_percent: 0.0,
            cpu_limit_percent: 0.0,
        }
    }
}

impl From<probe_quota::QuotaUsage> for QuotaUsage {
    fn from(u: probe_quota::QuotaUsage) -> Self {
        Self {
            memory_bytes: u.memory_bytes,
            memory_limit_bytes: u.memory_limit_bytes.unwrap_or(0),
            pids_current: u.pids_current,
            pids_limit: u.pids_limit.unwrap_or(0),
            cpu_percent: u.cpu_percent,
            cpu_limit_percent: u.cpu_limit_percent.unwrap_or(0.0),
        }
    }
}

/// Container runtime type.
#[repr(C)]
pub enum ContainerRuntime {
    None = 0,
    Docker = 1,
    Podman = 2,
    LXC = 3,
    Kubernetes = 4,
    FreeBSDJail = 5,
    Unknown = 255,
}

impl From<probe_quota::ContainerRuntime> for ContainerRuntime {
    fn from(r: probe_quota::ContainerRuntime) -> Self {
        match r {
            probe_quota::ContainerRuntime::None => ContainerRuntime::None,
            probe_quota::ContainerRuntime::Docker => ContainerRuntime::Docker,
            probe_quota::ContainerRuntime::Podman => ContainerRuntime::Podman,
            probe_quota::ContainerRuntime::LXC => ContainerRuntime::LXC,
            probe_quota::ContainerRuntime::Kubernetes => ContainerRuntime::Kubernetes,
            probe_quota::ContainerRuntime::FreeBSDJail => ContainerRuntime::FreeBSDJail,
            probe_quota::ContainerRuntime::Unknown => ContainerRuntime::Unknown,
        }
    }
}

/// Container information.
#[repr(C)]
pub struct ContainerInfo {
    /// Whether running in a container.
    pub is_containerized: bool,
    /// Container runtime type.
    pub runtime: ContainerRuntime,
    /// Container ID (null-terminated, empty if not available).
    pub container_id: [c_char; 65],
}

impl Default for ContainerInfo {
    fn default() -> Self {
        Self { is_containerized: false, runtime: ContainerRuntime::None, container_id: [0; 65] }
    }
}

impl From<probe_quota::ContainerInfo> for ContainerInfo {
    fn from(c: probe_quota::ContainerInfo) -> Self {
        let mut result = Self {
            is_containerized: c.is_containerized,
            runtime: c.runtime.into(),
            container_id: [0; 65],
        };
        if let Some(id) = c.container_id {
            copy_str_to_carray(&id, &mut result.container_id);
        }
        result
    }
}

// ============================================================================
// LIFECYCLE FUNCTIONS
// ============================================================================

/// Initialize the probe library.
/// Must be called once at startup.
#[unsafe(no_mangle)]
pub extern "C" fn probe_init() -> ProbeResult {
    match COLLECTOR.set(new_collector()) {
        Ok(()) => ProbeResult::ok(),
        Err(_) => ProbeResult::ok(), // Already initialized, that's fine
    }
}

/// Shutdown the probe library.
/// Should be called at program exit.
#[unsafe(no_mangle)]
pub extern "C" fn probe_shutdown() {
    // Nothing to clean up currently
}

// ============================================================================
// SYSTEM METRICS FUNCTIONS
// ============================================================================

/// Collect system CPU metrics.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_cpu(out: *mut SystemCPU) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.cpu().collect_system() {
        Ok(cpu) => {
            unsafe { *out = SystemCPU::from(cpu) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Collect system memory metrics.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_memory(out: *mut SystemMemory) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.memory().collect_system() {
        Ok(mem) => {
            unsafe { *out = SystemMemory::from(mem) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Collect system load average.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_load(out: *mut LoadAverage) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.load().collect() {
        Ok(load) => {
            unsafe { *out = LoadAverage::from(load) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

// ============================================================================
// PROCESS METRICS FUNCTIONS
// ============================================================================

/// Collect metrics for a specific process.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_process(pid: i32, out: *mut ProcessMetrics) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.process().collect(pid) {
        Ok(proc) => {
            unsafe { *out = ProcessMetrics::from(proc) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

// ============================================================================
// RESOURCE QUOTA FUNCTIONS (READ-ONLY DETECTION)
// ============================================================================

// Global quota reader instance
static QUOTA_READER: OnceLock<Box<dyn probe_quota::QuotaReader>> = OnceLock::new();

fn get_quota_reader() -> &'static dyn probe_quota::QuotaReader {
    QUOTA_READER.get_or_init(probe_quota::new_reader).as_ref()
}

/// Check if quota detection is supported on this platform.
#[unsafe(no_mangle)]
pub extern "C" fn probe_quota_is_supported() -> bool {
    probe_quota::is_supported()
}

/// Read resource limits for a process.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_quota_read_limits(pid: i32, out: *mut QuotaLimits) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let reader = get_quota_reader();
    match reader.read_limits(pid) {
        Ok(limits) => {
            unsafe { *out = QuotaLimits::from(limits) };
            ProbeResult::ok()
        }
        Err(e) => match e {
            probe_quota::Error::NotFound(_) => {
                ProbeResult::err(PROBE_ERR_NOT_FOUND, c"process not found".as_ptr())
            }
            probe_quota::Error::Permission(_) => {
                ProbeResult::err(PROBE_ERR_PERMISSION, c"permission denied".as_ptr())
            }
            probe_quota::Error::NotSupported => {
                ProbeResult::err(PROBE_ERR_NOT_SUPPORTED, c"not supported".as_ptr())
            }
            _ => ProbeResult::err(PROBE_ERR_INTERNAL, c"internal error".as_ptr()),
        },
    }
}

/// Read current resource usage for a process.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_quota_read_usage(pid: i32, out: *mut QuotaUsage) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let reader = get_quota_reader();
    match reader.read_usage(pid) {
        Ok(usage) => {
            unsafe { *out = QuotaUsage::from(usage) };
            ProbeResult::ok()
        }
        Err(e) => match e {
            probe_quota::Error::NotFound(_) => {
                ProbeResult::err(PROBE_ERR_NOT_FOUND, c"process not found".as_ptr())
            }
            probe_quota::Error::Permission(_) => {
                ProbeResult::err(PROBE_ERR_PERMISSION, c"permission denied".as_ptr())
            }
            probe_quota::Error::NotSupported => {
                ProbeResult::err(PROBE_ERR_NOT_SUPPORTED, c"not supported".as_ptr())
            }
            _ => ProbeResult::err(PROBE_ERR_INTERNAL, c"internal error".as_ptr()),
        },
    }
}

/// Detect container runtime.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_detect_container(out: *mut ContainerInfo) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let info = probe_quota::detect_container();
    unsafe { *out = ContainerInfo::from(info) };
    ProbeResult::ok()
}

// ============================================================================
// PLATFORM INFO FUNCTIONS
// ============================================================================

/// Get the platform name.
#[unsafe(no_mangle)]
pub extern "C" fn probe_get_platform() -> *const c_char {
    #[cfg(target_os = "linux")]
    return c"linux".as_ptr();

    #[cfg(target_os = "macos")]
    return c"darwin".as_ptr();

    #[cfg(target_os = "freebsd")]
    return c"freebsd".as_ptr();

    #[cfg(target_os = "openbsd")]
    return c"openbsd".as_ptr();

    #[cfg(target_os = "netbsd")]
    return c"netbsd".as_ptr();

    #[cfg(not(any(
        target_os = "linux",
        target_os = "macos",
        target_os = "freebsd",
        target_os = "openbsd",
        target_os = "netbsd"
    )))]
    return c"unknown".as_ptr();
}

/// Helper to call libc::uname and return the result.
fn get_uname_info() -> Option<libc::utsname> {
    unsafe {
        let mut info: libc::utsname = std::mem::zeroed();
        if libc::uname(&mut info) == 0 {
            Some(info)
        } else {
            None
        }
    }
}

/// Helper to convert a C char array to a Rust string.
#[allow(clippy::unnecessary_cast)] // c_char is i8 on x86_64, u8 on aarch64
fn carray_to_string(arr: &[c_char]) -> String {
    let len = arr.iter().position(|&c| c == 0).unwrap_or(arr.len());
    let bytes: Vec<u8> = arr[..len].iter().map(|&c| c as u8).collect();
    String::from_utf8_lossy(&bytes).into_owned()
}

/// Get the OS version string (e.g. "Linux 6.12.69", "Darwin 24.6.0").
#[unsafe(no_mangle)]
pub extern "C" fn probe_get_os_version() -> *const c_char {
    static VALUE: OnceLock<CString> = OnceLock::new();
    VALUE
        .get_or_init(|| {
            let s = get_uname_info()
                .map(|u| {
                    let sysname = carray_to_string(&u.sysname);
                    let release = carray_to_string(&u.release);
                    format!("{sysname} {release}")
                })
                .unwrap_or_else(|| "unknown".to_string());
            CString::new(s).unwrap_or_else(|_| CString::new("unknown").unwrap())
        })
        .as_ptr()
}

/// Get the kernel version string (full build string from uname.version).
#[unsafe(no_mangle)]
pub extern "C" fn probe_get_kernel_version() -> *const c_char {
    static VALUE: OnceLock<CString> = OnceLock::new();
    VALUE
        .get_or_init(|| {
            let s = get_uname_info()
                .map(|u| carray_to_string(&u.version))
                .unwrap_or_else(|| "unknown".to_string());
            CString::new(s).unwrap_or_else(|_| CString::new("unknown").unwrap())
        })
        .as_ptr()
}

/// Get the machine architecture (e.g. "x86_64", "aarch64", "arm64").
#[unsafe(no_mangle)]
pub extern "C" fn probe_get_arch() -> *const c_char {
    static VALUE: OnceLock<CString> = OnceLock::new();
    VALUE
        .get_or_init(|| {
            let s = get_uname_info()
                .map(|u| carray_to_string(&u.machine))
                .unwrap_or_else(|| "unknown".to_string());
            CString::new(s).unwrap_or_else(|_| CString::new("unknown").unwrap())
        })
        .as_ptr()
}

// ============================================================================
// PRESSURE METRICS (PSI - Linux only)
// ============================================================================

/// CPU pressure metrics.
#[repr(C)]
pub struct CPUPressure {
    pub some_avg10: f64,
    pub some_avg60: f64,
    pub some_avg300: f64,
    pub some_total_us: u64,
}

impl From<probe_metrics::CPUPressure> for CPUPressure {
    fn from(p: probe_metrics::CPUPressure) -> Self {
        Self {
            some_avg10: p.some_avg10,
            some_avg60: p.some_avg60,
            some_avg300: p.some_avg300,
            some_total_us: p.some_total_us,
        }
    }
}

/// Memory pressure metrics.
#[repr(C)]
pub struct MemoryPressure {
    pub some_avg10: f64,
    pub some_avg60: f64,
    pub some_avg300: f64,
    pub some_total_us: u64,
    pub full_avg10: f64,
    pub full_avg60: f64,
    pub full_avg300: f64,
    pub full_total_us: u64,
}

impl From<probe_metrics::MemoryPressure> for MemoryPressure {
    fn from(p: probe_metrics::MemoryPressure) -> Self {
        Self {
            some_avg10: p.some_avg10,
            some_avg60: p.some_avg60,
            some_avg300: p.some_avg300,
            some_total_us: p.some_total_us,
            full_avg10: p.full_avg10,
            full_avg60: p.full_avg60,
            full_avg300: p.full_avg300,
            full_total_us: p.full_total_us,
        }
    }
}

/// I/O pressure metrics.
#[repr(C)]
pub struct IOPressure {
    pub some_avg10: f64,
    pub some_avg60: f64,
    pub some_avg300: f64,
    pub some_total_us: u64,
    pub full_avg10: f64,
    pub full_avg60: f64,
    pub full_avg300: f64,
    pub full_total_us: u64,
}

impl From<probe_metrics::IOPressure> for IOPressure {
    fn from(p: probe_metrics::IOPressure) -> Self {
        Self {
            some_avg10: p.some_avg10,
            some_avg60: p.some_avg60,
            some_avg300: p.some_avg300,
            some_total_us: p.some_total_us,
            full_avg10: p.full_avg10,
            full_avg60: p.full_avg60,
            full_avg300: p.full_avg300,
            full_total_us: p.full_total_us,
        }
    }
}

/// Collect CPU pressure metrics.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_cpu_pressure(out: *mut CPUPressure) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.cpu().collect_pressure() {
        Ok(pressure) => {
            unsafe { *out = CPUPressure::from(pressure) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Collect memory pressure metrics.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_memory_pressure(out: *mut MemoryPressure) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.memory().collect_pressure() {
        Ok(pressure) => {
            unsafe { *out = MemoryPressure::from(pressure) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Collect I/O pressure metrics.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_io_pressure(out: *mut IOPressure) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.io().collect_pressure() {
        Ok(pressure) => {
            unsafe { *out = IOPressure::from(pressure) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

// ============================================================================
// DISK METRICS
// ============================================================================

/// Maximum path length for disk-related strings.
pub const PROBE_MAX_PATH_LEN: usize = 256;

/// Partition information.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct Partition {
    pub device: [c_char; PROBE_MAX_PATH_LEN],
    pub mount_point: [c_char; PROBE_MAX_PATH_LEN],
    pub fs_type: [c_char; 64],
    pub options: [c_char; PROBE_MAX_PATH_LEN],
}

impl Default for Partition {
    fn default() -> Self {
        Self {
            device: [0; PROBE_MAX_PATH_LEN],
            mount_point: [0; PROBE_MAX_PATH_LEN],
            fs_type: [0; 64],
            options: [0; PROBE_MAX_PATH_LEN],
        }
    }
}

fn copy_str_to_carray<const N: usize>(s: &str, dest: &mut [c_char; N]) {
    let bytes = s.as_bytes();
    let len = bytes.len().min(N - 1);
    for (i, &b) in bytes[..len].iter().enumerate() {
        dest[i] = b as c_char;
    }
    dest[len] = 0;
}

impl From<probe_metrics::Partition> for Partition {
    fn from(p: probe_metrics::Partition) -> Self {
        let mut result = Self::default();
        copy_str_to_carray(&p.device, &mut result.device);
        copy_str_to_carray(&p.mount_point, &mut result.mount_point);
        copy_str_to_carray(&p.fs_type, &mut result.fs_type);
        copy_str_to_carray(&p.options, &mut result.options);
        result
    }
}

/// Disk usage information.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct DiskUsage {
    pub path: [c_char; PROBE_MAX_PATH_LEN],
    pub total_bytes: u64,
    pub used_bytes: u64,
    pub free_bytes: u64,
    pub used_percent: f64,
    pub inodes_total: u64,
    pub inodes_used: u64,
    pub inodes_free: u64,
}

impl Default for DiskUsage {
    fn default() -> Self {
        Self {
            path: [0; PROBE_MAX_PATH_LEN],
            total_bytes: 0,
            used_bytes: 0,
            free_bytes: 0,
            used_percent: 0.0,
            inodes_total: 0,
            inodes_used: 0,
            inodes_free: 0,
        }
    }
}

impl From<probe_metrics::DiskUsage> for DiskUsage {
    fn from(d: probe_metrics::DiskUsage) -> Self {
        let mut result = Self::default();
        copy_str_to_carray(&d.path, &mut result.path);
        result.total_bytes = d.total_bytes;
        result.used_bytes = d.used_bytes;
        result.free_bytes = d.free_bytes;
        result.used_percent = d.used_percent;
        result.inodes_total = d.inodes_total;
        result.inodes_used = d.inodes_used;
        result.inodes_free = d.inodes_free;
        result
    }
}

/// Disk I/O statistics.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct DiskIOStats {
    pub device: [c_char; 64],
    pub reads_completed: u64,
    pub read_bytes: u64,
    pub read_time_us: u64,
    pub writes_completed: u64,
    pub write_bytes: u64,
    pub write_time_us: u64,
    pub io_in_progress: u64,
    pub io_time_us: u64,
    pub weighted_io_time_us: u64,
}

impl Default for DiskIOStats {
    fn default() -> Self {
        Self {
            device: [0; 64],
            reads_completed: 0,
            read_bytes: 0,
            read_time_us: 0,
            writes_completed: 0,
            write_bytes: 0,
            write_time_us: 0,
            io_in_progress: 0,
            io_time_us: 0,
            weighted_io_time_us: 0,
        }
    }
}

impl From<probe_metrics::DiskIOStats> for DiskIOStats {
    fn from(d: probe_metrics::DiskIOStats) -> Self {
        let mut result = Self::default();
        copy_str_to_carray(&d.device, &mut result.device);
        result.reads_completed = d.reads_completed;
        result.read_bytes = d.read_bytes;
        result.read_time_us = d.read_time_us;
        result.writes_completed = d.writes_completed;
        result.write_bytes = d.write_bytes;
        result.write_time_us = d.write_time_us;
        result.io_in_progress = d.io_in_progress;
        result.io_time_us = d.io_time_us;
        result.weighted_io_time_us = d.weighted_io_time_us;
        result
    }
}

/// List result for partitions.
#[repr(C)]
pub struct PartitionList {
    pub items: *mut Partition,
    pub count: usize,
    pub capacity: usize,
}

/// List result for disk I/O stats.
#[repr(C)]
pub struct DiskIOStatsList {
    pub items: *mut DiskIOStats,
    pub count: usize,
    pub capacity: usize,
}

/// List disk partitions.
///
/// # Safety
/// The `out` pointer must be valid. Caller must call `probe_free_partition_list` when done.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_list_partitions(out: *mut PartitionList) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.disk().list_partitions() {
        Ok(partitions) => {
            let mut items: Vec<Partition> = partitions.into_iter().map(|p| p.into()).collect();
            let count = items.len();
            let capacity = items.capacity();
            let ptr = items.as_mut_ptr();
            std::mem::forget(items);

            unsafe {
                (*out).items = ptr;
                (*out).count = count;
                (*out).capacity = capacity;
            }
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Free a partition list returned by `probe_list_partitions`.
///
/// # Safety
/// The list must have been allocated by `probe_list_partitions`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_free_partition_list(list: *mut PartitionList) {
    if list.is_null() {
        return;
    }
    unsafe {
        let list = &mut *list;
        if !list.items.is_null() {
            drop(Vec::from_raw_parts(list.items, list.count, list.capacity));
            list.items = ptr::null_mut();
            list.count = 0;
            list.capacity = 0;
        }
    }
}

/// Collect disk usage for a specific path.
///
/// # Safety
/// The `path` must be a null-terminated C string. The `out` pointer must be valid.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_disk_usage(
    path: *const c_char,
    out: *mut DiskUsage,
) -> ProbeResult {
    if path.is_null() || out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    let path_str = unsafe { std::ffi::CStr::from_ptr(path).to_string_lossy() };

    match collector.disk().collect_usage(&path_str) {
        Ok(usage) => {
            unsafe { *out = DiskUsage::from(usage) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Collect disk I/O statistics for all devices.
///
/// # Safety
/// The `out` pointer must be valid. Caller must call `probe_free_disk_io_list` when done.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_disk_io(out: *mut DiskIOStatsList) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.disk().collect_io() {
        Ok(stats) => {
            let mut items: Vec<DiskIOStats> = stats.into_iter().map(|s| s.into()).collect();
            let count = items.len();
            let capacity = items.capacity();
            let ptr = items.as_mut_ptr();
            std::mem::forget(items);

            unsafe {
                (*out).items = ptr;
                (*out).count = count;
                (*out).capacity = capacity;
            }
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Free a disk I/O stats list.
///
/// # Safety
/// The list must have been allocated by `probe_collect_disk_io`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_free_disk_io_list(list: *mut DiskIOStatsList) {
    if list.is_null() {
        return;
    }
    unsafe {
        let list = &mut *list;
        if !list.items.is_null() {
            drop(Vec::from_raw_parts(list.items, list.count, list.capacity));
            list.items = ptr::null_mut();
            list.count = 0;
            list.capacity = 0;
        }
    }
}

// ============================================================================
// NETWORK METRICS
// ============================================================================

/// Network interface information.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct NetInterface {
    pub name: [c_char; 64],
    pub mac_address: [c_char; 18],
    pub mtu: u32,
    pub is_up: bool,
    pub is_loopback: bool,
}

impl Default for NetInterface {
    fn default() -> Self {
        Self { name: [0; 64], mac_address: [0; 18], mtu: 0, is_up: false, is_loopback: false }
    }
}

impl From<probe_metrics::NetInterface> for NetInterface {
    fn from(n: probe_metrics::NetInterface) -> Self {
        let mut result = Self::default();
        copy_str_to_carray(&n.name, &mut result.name);
        copy_str_to_carray(&n.mac_address, &mut result.mac_address);
        result.mtu = n.mtu;
        result.is_up = n.is_up;
        result.is_loopback = n.is_loopback;
        result
    }
}

/// Network interface statistics.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct NetStats {
    pub interface: [c_char; 64],
    pub rx_bytes: u64,
    pub rx_packets: u64,
    pub rx_errors: u64,
    pub rx_drops: u64,
    pub tx_bytes: u64,
    pub tx_packets: u64,
    pub tx_errors: u64,
    pub tx_drops: u64,
}

impl Default for NetStats {
    fn default() -> Self {
        Self {
            interface: [0; 64],
            rx_bytes: 0,
            rx_packets: 0,
            rx_errors: 0,
            rx_drops: 0,
            tx_bytes: 0,
            tx_packets: 0,
            tx_errors: 0,
            tx_drops: 0,
        }
    }
}

impl From<probe_metrics::NetStats> for NetStats {
    fn from(n: probe_metrics::NetStats) -> Self {
        let mut result = Self::default();
        copy_str_to_carray(&n.interface, &mut result.interface);
        result.rx_bytes = n.rx_bytes;
        result.rx_packets = n.rx_packets;
        result.rx_errors = n.rx_errors;
        result.rx_drops = n.rx_drops;
        result.tx_bytes = n.tx_bytes;
        result.tx_packets = n.tx_packets;
        result.tx_errors = n.tx_errors;
        result.tx_drops = n.tx_drops;
        result
    }
}

/// List result for network interfaces.
#[repr(C)]
pub struct NetInterfaceList {
    pub items: *mut NetInterface,
    pub count: usize,
    pub capacity: usize,
}

/// List result for network stats.
#[repr(C)]
pub struct NetStatsList {
    pub items: *mut NetStats,
    pub count: usize,
    pub capacity: usize,
}

/// List network interfaces.
///
/// # Safety
/// The `out` pointer must be valid. Caller must call `probe_free_net_interface_list` when done.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_list_net_interfaces(out: *mut NetInterfaceList) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.network().list_interfaces() {
        Ok(interfaces) => {
            let mut items: Vec<NetInterface> = interfaces.into_iter().map(|i| i.into()).collect();
            let count = items.len();
            let capacity = items.capacity();
            let ptr = items.as_mut_ptr();
            std::mem::forget(items);

            unsafe {
                (*out).items = ptr;
                (*out).count = count;
                (*out).capacity = capacity;
            }
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Free a network interface list.
///
/// # Safety
/// The list must have been allocated by `probe_list_net_interfaces`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_free_net_interface_list(list: *mut NetInterfaceList) {
    if list.is_null() {
        return;
    }
    unsafe {
        let list = &mut *list;
        if !list.items.is_null() {
            drop(Vec::from_raw_parts(list.items, list.count, list.capacity));
            list.items = ptr::null_mut();
            list.count = 0;
            list.capacity = 0;
        }
    }
}

/// Collect network statistics for all interfaces.
///
/// # Safety
/// The `out` pointer must be valid. Caller must call `probe_free_net_stats_list` when done.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_net_stats(out: *mut NetStatsList) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.network().collect_all_stats() {
        Ok(stats) => {
            let mut items: Vec<NetStats> = stats.into_iter().map(|s| s.into()).collect();
            let count = items.len();
            let capacity = items.capacity();
            let ptr = items.as_mut_ptr();
            std::mem::forget(items);

            unsafe {
                (*out).items = ptr;
                (*out).count = count;
                (*out).capacity = capacity;
            }
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

/// Free a network stats list.
///
/// # Safety
/// The list must have been allocated by `probe_collect_net_stats`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_free_net_stats_list(list: *mut NetStatsList) {
    if list.is_null() {
        return;
    }
    unsafe {
        let list = &mut *list;
        if !list.items.is_null() {
            drop(Vec::from_raw_parts(list.items, list.count, list.capacity));
            list.items = ptr::null_mut();
            list.count = 0;
            list.capacity = 0;
        }
    }
}

// ============================================================================
// I/O METRICS
// ============================================================================

/// System-wide I/O statistics.
#[repr(C)]
pub struct IOStats {
    pub read_ops: u64,
    pub read_bytes: u64,
    pub write_ops: u64,
    pub write_bytes: u64,
}

impl From<probe_metrics::IOStats> for IOStats {
    fn from(io: probe_metrics::IOStats) -> Self {
        Self {
            read_ops: io.read_ops,
            read_bytes: io.read_bytes,
            write_ops: io.write_ops,
            write_bytes: io.write_bytes,
        }
    }
}

/// Collect system-wide I/O statistics.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_io_stats(out: *mut IOStats) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.io().collect_stats() {
        Ok(stats) => {
            unsafe { *out = IOStats::from(stats) };
            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

// ============================================================================
// CONTEXT SWITCHES
// ============================================================================

/// Context switch statistics.
#[repr(C)]
pub struct ContextSwitches {
    /// Voluntary context switches (process yielded CPU).
    pub voluntary: u64,
    /// Involuntary context switches (preempted by scheduler).
    pub involuntary: u64,
    /// System-wide total context switches.
    pub system_total: u64,
}

impl From<probe_metrics::ContextSwitches> for ContextSwitches {
    fn from(cs: probe_metrics::ContextSwitches) -> Self {
        Self { voluntary: cs.voluntary, involuntary: cs.involuntary, system_total: cs.system_total }
    }
}

/// Collect system-wide context switch count.
///
/// # Safety
/// The `out` pointer must be valid.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_system_context_switches(out: *mut u64) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::read_system_context_switches() {
            Ok(count) => {
                unsafe { *out = count };
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(any(target_os = "freebsd", target_os = "openbsd", target_os = "netbsd"))]
    {
        match probe_platform::bsd::read_system_context_switches() {
            Ok(cs) => {
                unsafe { *out = cs.voluntary.saturating_add(cs.involuntary) };
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(any(
        target_os = "linux",
        target_os = "freebsd",
        target_os = "openbsd",
        target_os = "netbsd"
    )))]
    {
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"context switches not supported on this platform".as_ptr(),
        )
    }
}

/// Collect context switches for a specific process.
///
/// # Safety
/// The `out` pointer must be valid.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_process_context_switches(
    pid: i32,
    out: *mut ContextSwitches,
) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::read_process_context_switches(pid) {
            Ok(switches) => {
                unsafe { *out = ContextSwitches::from(switches) };
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(any(target_os = "freebsd", target_os = "openbsd", target_os = "netbsd"))]
    {
        match probe_platform::bsd::read_process_context_switches(pid) {
            Ok(cs) => {
                unsafe {
                    *out = ContextSwitches {
                        voluntary: cs.voluntary,
                        involuntary: cs.involuntary,
                        system_total: 0,
                    }
                };
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(any(
        target_os = "linux",
        target_os = "freebsd",
        target_os = "openbsd",
        target_os = "netbsd"
    )))]
    {
        let _ = pid;
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"context switches not supported on this platform".as_ptr(),
        )
    }
}

/// Collect context switches for the current process.
///
/// # Safety
/// The `out` pointer must be valid.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_self_context_switches(
    out: *mut ContextSwitches,
) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::read_self_context_switches() {
            Ok(switches) => {
                unsafe { *out = ContextSwitches::from(switches) };
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(any(target_os = "freebsd", target_os = "openbsd", target_os = "netbsd"))]
    {
        match probe_platform::bsd::read_self_context_switches() {
            Ok(cs) => {
                unsafe {
                    *out = ContextSwitches {
                        voluntary: cs.voluntary,
                        involuntary: cs.involuntary,
                        system_total: 0,
                    }
                };
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(any(
        target_os = "linux",
        target_os = "freebsd",
        target_os = "openbsd",
        target_os = "netbsd"
    )))]
    {
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"context switches not supported on this platform".as_ptr(),
        )
    }
}

// ============================================================================
// THERMAL METRICS
// ============================================================================

/// Maximum thermal zones to return.
pub const MAX_THERMAL_ZONES: usize = 32;

/// Thermal zone information.
#[repr(C)]
pub struct ThermalZone {
    pub name: [c_char; 64],
    pub label: [c_char; 64],
    pub temp_celsius: f64,
    pub temp_max: f64,
    pub temp_crit: f64,
    pub has_max: bool,
    pub has_crit: bool,
}

impl Default for ThermalZone {
    fn default() -> Self {
        Self {
            name: [0; 64],
            label: [0; 64],
            temp_celsius: 0.0,
            temp_max: 0.0,
            temp_crit: 0.0,
            has_max: false,
            has_crit: false,
        }
    }
}

impl From<probe_metrics::ThermalZone> for ThermalZone {
    fn from(zone: probe_metrics::ThermalZone) -> Self {
        let mut result = Self::default();
        copy_str_to_carray(&zone.name, &mut result.name);
        copy_str_to_carray(&zone.label, &mut result.label);
        result.temp_celsius = zone.temp_celsius;
        if let Some(max) = zone.temp_max {
            result.temp_max = max;
            result.has_max = true;
        }
        if let Some(crit) = zone.temp_crit {
            result.temp_crit = crit;
            result.has_crit = true;
        }
        result
    }
}

/// List of thermal zones.
#[repr(C)]
pub struct ThermalZoneList {
    pub items: *mut ThermalZone,
    pub count: usize,
    pub capacity: usize,
}

/// Check if thermal monitoring is supported.
#[unsafe(no_mangle)]
pub extern "C" fn probe_thermal_is_supported() -> bool {
    #[cfg(target_os = "linux")]
    {
        probe_platform::linux::is_thermal_supported()
    }

    #[cfg(not(target_os = "linux"))]
    {
        false
    }
}

/// Collect thermal zones.
///
/// # Safety
/// The `out` pointer must be valid. Caller must call `probe_free_thermal_list` when done.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_thermal_zones(out: *mut ThermalZoneList) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::read_thermal_zones() {
            Ok(zones) => {
                let mut items: Vec<ThermalZone> = zones.into_iter().map(|z| z.into()).collect();
                let count = items.len();
                let capacity = items.capacity();
                let ptr = items.as_mut_ptr();
                std::mem::forget(items);

                unsafe {
                    (*out).items = ptr;
                    (*out).count = count;
                    (*out).capacity = capacity;
                }
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(target_os = "linux"))]
    {
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"thermal monitoring not supported on this platform".as_ptr(),
        )
    }
}

/// Free a thermal zone list.
///
/// # Safety
/// The list must have been allocated by `probe_collect_thermal_zones`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_free_thermal_list(list: *mut ThermalZoneList) {
    if list.is_null() {
        return;
    }
    unsafe {
        let list = &mut *list;
        if !list.items.is_null() && list.capacity > 0 {
            drop(Vec::from_raw_parts(list.items, list.count, list.capacity));
            list.items = ptr::null_mut();
            list.count = 0;
            list.capacity = 0;
        }
    }
}

// ============================================================================
// AGGREGATED METRICS COLLECTION
// ============================================================================

/// All pressure metrics combined.
#[repr(C)]
pub struct AllPressure {
    pub cpu: CPUPressure,
    pub memory: MemoryPressure,
    pub io: IOPressure,
    /// Whether pressure metrics are available (Linux only).
    pub available: bool,
}

impl Default for AllPressure {
    fn default() -> Self {
        Self {
            cpu: CPUPressure {
                some_avg10: 0.0,
                some_avg60: 0.0,
                some_avg300: 0.0,
                some_total_us: 0,
            },
            memory: MemoryPressure {
                some_avg10: 0.0,
                some_avg60: 0.0,
                some_avg300: 0.0,
                some_total_us: 0,
                full_avg10: 0.0,
                full_avg60: 0.0,
                full_avg300: 0.0,
                full_total_us: 0,
            },
            io: IOPressure {
                some_avg10: 0.0,
                some_avg60: 0.0,
                some_avg300: 0.0,
                some_total_us: 0,
                full_avg10: 0.0,
                full_avg60: 0.0,
                full_avg300: 0.0,
                full_total_us: 0,
            },
            available: false,
        }
    }
}

/// Maximum partitions, disk I/O stats, interfaces, and net stats in AllMetrics.
pub const MAX_ALL_METRICS_ITEMS: usize = 64;

/// All system metrics collected in one call.
#[repr(C)]
pub struct AllMetrics {
    /// System CPU metrics.
    pub cpu: SystemCPU,
    /// System memory metrics.
    pub memory: SystemMemory,
    /// System load average.
    pub load: LoadAverage,
    /// System I/O statistics.
    pub io_stats: IOStats,
    /// Pressure metrics.
    pub pressure: AllPressure,
    /// Timestamp when metrics were collected (microseconds since epoch).
    pub timestamp_us: u64,

    /// Partition count.
    pub partition_count: u32,
    /// Disk usage count.
    pub disk_usage_count: u32,
    /// Disk I/O stats count.
    pub disk_io_count: u32,
    /// Network interface count.
    pub net_interface_count: u32,
    /// Network stats count.
    pub net_stats_count: u32,

    /// Partitions (up to MAX_ALL_METRICS_ITEMS).
    pub partitions: [Partition; MAX_ALL_METRICS_ITEMS],
    /// Disk usage (up to MAX_ALL_METRICS_ITEMS).
    pub disk_usage: [DiskUsage; MAX_ALL_METRICS_ITEMS],
    /// Disk I/O statistics (up to MAX_ALL_METRICS_ITEMS).
    pub disk_io: [DiskIOStats; MAX_ALL_METRICS_ITEMS],
    /// Network interfaces (up to MAX_ALL_METRICS_ITEMS).
    pub net_interfaces: [NetInterface; MAX_ALL_METRICS_ITEMS],
    /// Network statistics (up to MAX_ALL_METRICS_ITEMS).
    pub net_stats: [NetStats; MAX_ALL_METRICS_ITEMS],
}

impl Default for AllMetrics {
    fn default() -> Self {
        Self {
            cpu: SystemCPU {
                user_percent: 0.0,
                system_percent: 0.0,
                idle_percent: 0.0,
                iowait_percent: 0.0,
                steal_percent: 0.0,
                cores: 0,
                frequency_mhz: 0,
            },
            memory: SystemMemory {
                total_bytes: 0,
                available_bytes: 0,
                used_bytes: 0,
                cached_bytes: 0,
                buffers_bytes: 0,
                swap_total_bytes: 0,
                swap_used_bytes: 0,
            },
            load: LoadAverage { load_1min: 0.0, load_5min: 0.0, load_15min: 0.0 },
            io_stats: IOStats { read_ops: 0, read_bytes: 0, write_ops: 0, write_bytes: 0 },
            pressure: AllPressure::default(),
            timestamp_us: 0,
            partition_count: 0,
            disk_usage_count: 0,
            disk_io_count: 0,
            net_interface_count: 0,
            net_stats_count: 0,
            partitions: [Partition::default(); MAX_ALL_METRICS_ITEMS],
            disk_usage: [DiskUsage::default(); MAX_ALL_METRICS_ITEMS],
            disk_io: [DiskIOStats::default(); MAX_ALL_METRICS_ITEMS],
            net_interfaces: [NetInterface::default(); MAX_ALL_METRICS_ITEMS],
            net_stats: [NetStats::default(); MAX_ALL_METRICS_ITEMS],
        }
    }
}

/// Collect all system metrics in one call.
///
/// This is more efficient than calling each collector individually
/// and provides a consistent snapshot of all metrics.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_all(out: *mut AllMetrics) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let collector = match COLLECTOR.get() {
        Some(c) => c,
        None => return ProbeResult::err(PROBE_ERR_INTERNAL, c"not initialized".as_ptr()),
    };

    match collector.collect_all() {
        Ok(metrics) => {
            let result = unsafe { &mut *out };

            // Copy basic metrics
            result.cpu = SystemCPU::from(metrics.cpu);
            result.memory = SystemMemory::from(metrics.memory);
            result.load = LoadAverage::from(metrics.load);
            result.io_stats = IOStats::from(metrics.io_stats);
            result.timestamp_us = metrics.timestamp_us;

            // Copy pressure if available
            if let Some(pressure) = metrics.pressure {
                result.pressure = AllPressure {
                    cpu: CPUPressure::from(pressure.cpu),
                    memory: MemoryPressure::from(pressure.memory),
                    io: IOPressure::from(pressure.io),
                    available: true,
                };
            } else {
                result.pressure = AllPressure::default();
            }

            // Copy partitions
            let part_count = metrics.partitions.len().min(MAX_ALL_METRICS_ITEMS);
            result.partition_count = part_count as u32;
            for (i, p) in metrics.partitions.into_iter().take(part_count).enumerate() {
                result.partitions[i] = Partition::from(p);
            }

            // Copy disk usage
            let usage_count = metrics.disk_usage.len().min(MAX_ALL_METRICS_ITEMS);
            result.disk_usage_count = usage_count as u32;
            for (i, u) in metrics.disk_usage.into_iter().take(usage_count).enumerate() {
                result.disk_usage[i] = DiskUsage::from(u);
            }

            // Copy disk I/O
            let io_count = metrics.disk_io.len().min(MAX_ALL_METRICS_ITEMS);
            result.disk_io_count = io_count as u32;
            for (i, io) in metrics.disk_io.into_iter().take(io_count).enumerate() {
                result.disk_io[i] = DiskIOStats::from(io);
            }

            // Copy network interfaces
            let iface_count = metrics.net_interfaces.len().min(MAX_ALL_METRICS_ITEMS);
            result.net_interface_count = iface_count as u32;
            for (i, iface) in metrics.net_interfaces.into_iter().take(iface_count).enumerate() {
                result.net_interfaces[i] = NetInterface::from(iface);
            }

            // Copy network stats
            let stats_count = metrics.net_stats.len().min(MAX_ALL_METRICS_ITEMS);
            result.net_stats_count = stats_count as u32;
            for (i, stats) in metrics.net_stats.into_iter().take(stats_count).enumerate() {
                result.net_stats[i] = NetStats::from(stats);
            }

            ProbeResult::ok()
        }
        Err(e) => ProbeResult::from_metrics_error(e),
    }
}

// ============================================================================
// UNIVERSAL RUNTIME DETECTION
// ============================================================================

/// Maximum number of available runtimes to return.
pub const MAX_AVAILABLE_RUNTIMES: usize = 16;

/// Extended container runtime type (covers all runtimes).
#[repr(C)]
#[derive(Debug, Clone, Copy)]
pub enum RuntimeType {
    /// Not containerized / no runtime.
    None = 0,
    // Container runtimes (1-19)
    Docker = 1,
    Podman = 2,
    Containerd = 3,
    CriO = 4,
    Lxc = 5,
    Lxd = 6,
    SystemdNspawn = 7,
    Firecracker = 8,
    FreeBsdJail = 9,
    // Orchestrators (20-39)
    Kubernetes = 20,
    Nomad = 21,
    DockerSwarm = 22,
    OpenShift = 23,
    // Cloud-specific (40-59)
    AwsEcs = 40,
    AwsFargate = 41,
    GoogleGke = 42,
    AzureAks = 43,
    // Virtualization (60-79)
    VMware = 60,
    Qemu = 61,
    VirtualBox = 62,
    HyperV = 63,
    Bhyve = 64,
    Xen = 65,
    Parallels = 66,
    /// Unknown runtime.
    Unknown = 254,
}

impl From<probe_runtime::ContainerRuntime> for RuntimeType {
    fn from(r: probe_runtime::ContainerRuntime) -> Self {
        match r {
            probe_runtime::ContainerRuntime::None => Self::None,
            probe_runtime::ContainerRuntime::Docker => Self::Docker,
            probe_runtime::ContainerRuntime::Podman => Self::Podman,
            probe_runtime::ContainerRuntime::Containerd => Self::Containerd,
            probe_runtime::ContainerRuntime::CriO => Self::CriO,
            probe_runtime::ContainerRuntime::Lxc => Self::Lxc,
            probe_runtime::ContainerRuntime::Lxd => Self::Lxd,
            probe_runtime::ContainerRuntime::SystemdNspawn => Self::SystemdNspawn,
            probe_runtime::ContainerRuntime::Firecracker => Self::Firecracker,
            probe_runtime::ContainerRuntime::FreeBsdJail => Self::FreeBsdJail,
            probe_runtime::ContainerRuntime::Kubernetes => Self::Kubernetes,
            probe_runtime::ContainerRuntime::Nomad => Self::Nomad,
            probe_runtime::ContainerRuntime::DockerSwarm => Self::DockerSwarm,
            probe_runtime::ContainerRuntime::OpenShift => Self::OpenShift,
            probe_runtime::ContainerRuntime::AwsEcs => Self::AwsEcs,
            probe_runtime::ContainerRuntime::AwsFargate => Self::AwsFargate,
            probe_runtime::ContainerRuntime::GoogleGke => Self::GoogleGke,
            probe_runtime::ContainerRuntime::AzureAks => Self::AzureAks,
            probe_runtime::ContainerRuntime::VMware => Self::VMware,
            probe_runtime::ContainerRuntime::Qemu => Self::Qemu,
            probe_runtime::ContainerRuntime::VirtualBox => Self::VirtualBox,
            probe_runtime::ContainerRuntime::HyperV => Self::HyperV,
            probe_runtime::ContainerRuntime::Bhyve => Self::Bhyve,
            probe_runtime::ContainerRuntime::Xen => Self::Xen,
            probe_runtime::ContainerRuntime::Parallels => Self::Parallels,
            probe_runtime::ContainerRuntime::Unknown => Self::Unknown,
        }
    }
}

/// Information about a runtime available on the host.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct AvailableRuntimeInfo {
    /// Runtime type.
    pub runtime: RuntimeType,
    /// Unix socket path (null-terminated, empty if not available).
    pub socket_path: [c_char; PROBE_MAX_PATH_LEN],
    /// Version string (null-terminated, empty if not available).
    pub version: [c_char; 64],
    /// Whether the runtime is currently running/responsive.
    pub is_running: bool,
}

impl Default for AvailableRuntimeInfo {
    fn default() -> Self {
        Self {
            runtime: RuntimeType::None,
            socket_path: [0; PROBE_MAX_PATH_LEN],
            version: [0; 64],
            is_running: false,
        }
    }
}

#[allow(clippy::field_reassign_with_default)]
impl From<probe_runtime::AvailableRuntime> for AvailableRuntimeInfo {
    fn from(r: probe_runtime::AvailableRuntime) -> Self {
        let mut result = Self::default();
        result.runtime = r.runtime.into();
        if let Some(path) = r.socket_path {
            copy_str_to_carray(&path, &mut result.socket_path);
        }
        if let Some(version) = r.version {
            copy_str_to_carray(&version, &mut result.version);
        }
        result.is_running = r.is_running;
        result
    }
}

/// Full runtime environment information.
#[repr(C)]
pub struct RuntimeInfo {
    /// Whether running inside a container.
    pub is_containerized: bool,
    /// Container runtime type (if containerized).
    pub container_runtime: RuntimeType,
    /// Orchestrator type (may differ from runtime).
    pub orchestrator: RuntimeType,
    /// Container ID (null-terminated, 64 chars max).
    pub container_id: [c_char; 65],
    /// Workload/allocation ID (null-terminated).
    pub workload_id: [c_char; 65],
    /// Workload/pod name (null-terminated).
    pub workload_name: [c_char; 128],
    /// Namespace (null-terminated).
    pub namespace: [c_char; 64],
    /// Number of available runtimes.
    pub available_count: u32,
    /// Available runtimes on the host.
    pub available_runtimes: [AvailableRuntimeInfo; MAX_AVAILABLE_RUNTIMES],
}

impl Default for RuntimeInfo {
    fn default() -> Self {
        Self {
            is_containerized: false,
            container_runtime: RuntimeType::None,
            orchestrator: RuntimeType::None,
            container_id: [0; 65],
            workload_id: [0; 65],
            workload_name: [0; 128],
            namespace: [0; 64],
            available_count: 0,
            available_runtimes: [AvailableRuntimeInfo::default(); MAX_AVAILABLE_RUNTIMES],
        }
    }
}

#[allow(clippy::field_reassign_with_default)]
impl From<probe_runtime::RuntimeInfo> for RuntimeInfo {
    fn from(info: probe_runtime::RuntimeInfo) -> Self {
        let mut result = Self::default();

        result.is_containerized = info.is_containerized;

        if let Some(runtime) = info.container_runtime {
            result.container_runtime = runtime.into();
        }

        if let Some(orchestrator) = info.orchestrator {
            result.orchestrator = orchestrator.into();
        }

        if let Some(id) = info.container_id {
            copy_str_to_carray(&id, &mut result.container_id);
        }

        if let Some(id) = info.workload_id {
            copy_str_to_carray(&id, &mut result.workload_id);
        }

        if let Some(name) = info.workload_name {
            copy_str_to_carray(&name, &mut result.workload_name);
        }

        if let Some(ns) = info.namespace {
            copy_str_to_carray(&ns, &mut result.namespace);
        }

        let count = info.available_runtimes.len().min(MAX_AVAILABLE_RUNTIMES);
        result.available_count = count as u32;
        for (i, runtime) in info.available_runtimes.into_iter().take(count).enumerate() {
            result.available_runtimes[i] = runtime.into();
        }

        result
    }
}

/// Detect full runtime environment.
///
/// This detects:
/// - Whether running inside a container
/// - The container runtime and orchestrator
/// - Container/workload IDs and names
/// - Available runtimes on the host
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_detect_runtime(out: *mut RuntimeInfo) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    let detector = probe_runtime::detector::UniversalRuntimeDetector::new();
    let info = detector.detect();

    unsafe { *out = RuntimeInfo::from(info) };
    ProbeResult::ok()
}

/// Check if running inside a container (fast check).
///
/// This only checks for containerization, not available runtimes.
#[unsafe(no_mangle)]
pub extern "C" fn probe_is_containerized() -> bool {
    probe_runtime::detector::is_containerized()
}

/// Get container runtime name as string.
///
/// Returns a static string like "docker", "kubernetes", etc.
/// Returns "none" if not containerized.
#[unsafe(no_mangle)]
pub extern "C" fn probe_get_runtime_name() -> *const c_char {
    match probe_runtime::detector::get_container_runtime() {
        Some(runtime) => match runtime {
            probe_runtime::ContainerRuntime::None => c"none".as_ptr(),
            probe_runtime::ContainerRuntime::Docker => c"docker".as_ptr(),
            probe_runtime::ContainerRuntime::Podman => c"podman".as_ptr(),
            probe_runtime::ContainerRuntime::Containerd => c"containerd".as_ptr(),
            probe_runtime::ContainerRuntime::CriO => c"cri-o".as_ptr(),
            probe_runtime::ContainerRuntime::Lxc => c"lxc".as_ptr(),
            probe_runtime::ContainerRuntime::Lxd => c"lxd".as_ptr(),
            probe_runtime::ContainerRuntime::SystemdNspawn => c"systemd-nspawn".as_ptr(),
            probe_runtime::ContainerRuntime::Firecracker => c"firecracker".as_ptr(),
            probe_runtime::ContainerRuntime::FreeBsdJail => c"freebsd-jail".as_ptr(),
            probe_runtime::ContainerRuntime::Kubernetes => c"kubernetes".as_ptr(),
            probe_runtime::ContainerRuntime::Nomad => c"nomad".as_ptr(),
            probe_runtime::ContainerRuntime::DockerSwarm => c"docker-swarm".as_ptr(),
            probe_runtime::ContainerRuntime::OpenShift => c"openshift".as_ptr(),
            probe_runtime::ContainerRuntime::AwsEcs => c"aws-ecs".as_ptr(),
            probe_runtime::ContainerRuntime::AwsFargate => c"aws-fargate".as_ptr(),
            probe_runtime::ContainerRuntime::GoogleGke => c"google-gke".as_ptr(),
            probe_runtime::ContainerRuntime::AzureAks => c"azure-aks".as_ptr(),
            probe_runtime::ContainerRuntime::VMware => c"vmware".as_ptr(),
            probe_runtime::ContainerRuntime::Qemu => c"qemu".as_ptr(),
            probe_runtime::ContainerRuntime::VirtualBox => c"virtualbox".as_ptr(),
            probe_runtime::ContainerRuntime::HyperV => c"hyper-v".as_ptr(),
            probe_runtime::ContainerRuntime::Bhyve => c"bhyve".as_ptr(),
            probe_runtime::ContainerRuntime::Xen => c"xen".as_ptr(),
            probe_runtime::ContainerRuntime::Parallels => c"parallels".as_ptr(),
            probe_runtime::ContainerRuntime::Unknown => c"unknown".as_ptr(),
        },
        None => c"none".as_ptr(),
    }
}

// ============================================================================
// CACHE MANAGEMENT FUNCTIONS
// ============================================================================

use parking_lot::RwLock;
use probe_cache::{CachePolicies, CachedCollector, MetricType};
use std::time::Duration;

/// Global cached collector instance.
static CACHED_COLLECTOR: OnceLock<RwLock<Option<CachedCollector<PlatformCollector>>>> =
    OnceLock::new();

fn get_cached_collector() -> &'static RwLock<Option<CachedCollector<PlatformCollector>>> {
    CACHED_COLLECTOR.get_or_init(|| RwLock::new(None))
}

/// Enable caching for the global collector with default policies.
///
/// After calling this, all metric collection calls will use caching.
/// Call `probe_cache_disable` to disable caching.
#[unsafe(no_mangle)]
pub extern "C" fn probe_cache_enable() -> ProbeResult {
    let mut guard = get_cached_collector().write();
    if guard.is_some() {
        return ProbeResult::ok(); // Already enabled
    }

    *guard = Some(CachedCollector::new(new_collector(), CachePolicies::default()));
    ProbeResult::ok()
}

/// Enable caching with custom TTL policy preset.
///
/// Policy values:
/// - 0: Default (balanced TTLs)
/// - 1: High frequency (shorter TTLs)
/// - 2: Low frequency (longer TTLs)
/// - 3: No cache (TTL=0, for testing)
#[unsafe(no_mangle)]
pub extern "C" fn probe_cache_enable_with_policy(policy: u32) -> ProbeResult {
    let policies = match policy {
        0 => CachePolicies::default(),
        1 => CachePolicies::high_frequency(),
        2 => CachePolicies::low_frequency(),
        3 => CachePolicies::no_cache(),
        _ => return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"invalid policy".as_ptr()),
    };

    let mut guard = get_cached_collector().write();
    *guard = Some(CachedCollector::new(new_collector(), policies));
    ProbeResult::ok()
}

/// Disable caching and revert to direct collection.
#[unsafe(no_mangle)]
pub extern "C" fn probe_cache_disable() -> ProbeResult {
    let mut guard = get_cached_collector().write();
    *guard = None;
    ProbeResult::ok()
}

/// Check if caching is currently enabled.
#[unsafe(no_mangle)]
pub extern "C" fn probe_cache_is_enabled() -> bool {
    get_cached_collector().read().is_some()
}

/// Set the TTL for a specific metric type.
///
/// Metric types:
/// - 0: CPU system
/// - 1: CPU pressure
/// - 2: Memory system
/// - 3: Memory pressure
/// - 4: Load average
/// - 5: Disk partitions
/// - 6: Disk usage
/// - 7: Disk I/O
/// - 8: Network interfaces
/// - 9: Network stats
/// - 10: I/O stats
/// - 11: I/O pressure
///
/// TTL is specified in milliseconds.
#[unsafe(no_mangle)]
pub extern "C" fn probe_cache_set_ttl(metric_type: u8, ttl_ms: u64) -> ProbeResult {
    let metric = match MetricType::from_u8(metric_type) {
        Some(m) => m,
        None => return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"invalid metric type".as_ptr()),
    };

    let mut guard = get_cached_collector().write();
    match guard.as_mut() {
        Some(collector) => {
            collector.set_ttl(metric, Duration::from_millis(ttl_ms));
            ProbeResult::ok()
        }
        None => ProbeResult::err(PROBE_ERR_INTERNAL, c"caching not enabled".as_ptr()),
    }
}

/// Invalidate all cached metrics.
#[unsafe(no_mangle)]
pub extern "C" fn probe_cache_invalidate_all() -> ProbeResult {
    let guard = get_cached_collector().read();
    match guard.as_ref() {
        Some(collector) => {
            collector.invalidate_all();
            ProbeResult::ok()
        }
        None => ProbeResult::err(PROBE_ERR_INTERNAL, c"caching not enabled".as_ptr()),
    }
}

/// Invalidate a specific metric type from the cache.
#[unsafe(no_mangle)]
pub extern "C" fn probe_cache_invalidate(metric_type: u8) -> ProbeResult {
    let metric = match MetricType::from_u8(metric_type) {
        Some(m) => m,
        None => return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"invalid metric type".as_ptr()),
    };

    let guard = get_cached_collector().read();
    match guard.as_ref() {
        Some(collector) => {
            collector.invalidate(metric);
            ProbeResult::ok()
        }
        None => ProbeResult::err(PROBE_ERR_INTERNAL, c"caching not enabled".as_ptr()),
    }
}

// ============================================================================
// CACHED COLLECTION FUNCTIONS
// ============================================================================

/// Collect system CPU metrics with caching (if enabled).
///
/// If caching is disabled, this is equivalent to `probe_collect_cpu`.
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_cpu_cached(out: *mut SystemCPU) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    // Try cached collector first
    {
        let guard = get_cached_collector().read();
        if let Some(collector) = guard.as_ref() {
            return match collector.cpu().collect_system() {
                Ok(cpu) => {
                    unsafe { *out = SystemCPU::from(cpu) };
                    ProbeResult::ok()
                }
                Err(e) => ProbeResult::from_metrics_error(e),
            };
        }
    }

    // Fall back to direct collection
    unsafe { probe_collect_cpu(out) }
}

/// Collect system memory metrics with caching (if enabled).
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_memory_cached(out: *mut SystemMemory) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    {
        let guard = get_cached_collector().read();
        if let Some(collector) = guard.as_ref() {
            return match collector.memory().collect_system() {
                Ok(mem) => {
                    unsafe { *out = SystemMemory::from(mem) };
                    ProbeResult::ok()
                }
                Err(e) => ProbeResult::from_metrics_error(e),
            };
        }
    }

    unsafe { probe_collect_memory(out) }
}

/// Collect system load average with caching (if enabled).
///
/// # Safety
/// The `out` pointer must be valid and properly aligned.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_load_cached(out: *mut LoadAverage) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    {
        let guard = get_cached_collector().read();
        if let Some(collector) = guard.as_ref() {
            return match collector.load().collect() {
                Ok(load) => {
                    unsafe { *out = LoadAverage::from(load) };
                    ProbeResult::ok()
                }
                Err(e) => ProbeResult::from_metrics_error(e),
            };
        }
    }

    unsafe { probe_collect_load(out) }
}

// ============================================================================
// NETWORK CONNECTIONS
// ============================================================================

/// Socket state (matching Linux TCP states).
#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum SocketState {
    /// Unknown state.
    #[default]
    Unknown = 0,
    /// Established connection.
    Established = 1,
    /// Sent SYN.
    SynSent = 2,
    /// Received SYN.
    SynRecv = 3,
    /// FIN-WAIT-1.
    FinWait1 = 4,
    /// FIN-WAIT-2.
    FinWait2 = 5,
    /// TIME-WAIT.
    TimeWait = 6,
    /// Closed.
    Close = 7,
    /// CLOSE-WAIT.
    CloseWait = 8,
    /// LAST-ACK.
    LastAck = 9,
    /// Listening for connections.
    Listen = 10,
    /// CLOSING.
    Closing = 11,
}

impl From<probe_metrics::SocketState> for SocketState {
    fn from(s: probe_metrics::SocketState) -> Self {
        match s {
            probe_metrics::SocketState::Unknown => Self::Unknown,
            probe_metrics::SocketState::Established => Self::Established,
            probe_metrics::SocketState::SynSent => Self::SynSent,
            probe_metrics::SocketState::SynRecv => Self::SynRecv,
            probe_metrics::SocketState::FinWait1 => Self::FinWait1,
            probe_metrics::SocketState::FinWait2 => Self::FinWait2,
            probe_metrics::SocketState::TimeWait => Self::TimeWait,
            probe_metrics::SocketState::Close => Self::Close,
            probe_metrics::SocketState::CloseWait => Self::CloseWait,
            probe_metrics::SocketState::LastAck => Self::LastAck,
            probe_metrics::SocketState::Listen => Self::Listen,
            probe_metrics::SocketState::Closing => Self::Closing,
        }
    }
}

/// Address family.
#[repr(C)]
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
pub enum AddressFamily {
    /// IPv4 address.
    #[default]
    IPv4 = 4,
    /// IPv6 address.
    IPv6 = 6,
}

impl From<probe_metrics::AddressFamily> for AddressFamily {
    fn from(f: probe_metrics::AddressFamily) -> Self {
        match f {
            probe_metrics::AddressFamily::IPv4 => Self::IPv4,
            probe_metrics::AddressFamily::IPv6 => Self::IPv6,
        }
    }
}

/// Maximum address length for IPv6.
pub const MAX_ADDR_LEN: usize = 46;

/// TCP connection information.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct TcpConnection {
    /// Address family (IPv4 or IPv6).
    pub family: AddressFamily,
    /// Local IP address (null-terminated).
    pub local_addr: [c_char; MAX_ADDR_LEN],
    /// Local port.
    pub local_port: u16,
    /// Remote IP address (null-terminated).
    pub remote_addr: [c_char; MAX_ADDR_LEN],
    /// Remote port.
    pub remote_port: u16,
    /// Connection state.
    pub state: SocketState,
    /// Process ID owning this connection (-1 if unknown).
    pub pid: i32,
    /// Process name (null-terminated, empty if unknown).
    pub process_name: [c_char; 64],
    /// Socket inode number.
    pub inode: u64,
    /// Receive queue size.
    pub rx_queue: u32,
    /// Transmit queue size.
    pub tx_queue: u32,
}

impl Default for TcpConnection {
    fn default() -> Self {
        Self {
            family: AddressFamily::IPv4,
            local_addr: [0; MAX_ADDR_LEN],
            local_port: 0,
            remote_addr: [0; MAX_ADDR_LEN],
            remote_port: 0,
            state: SocketState::Unknown,
            pid: -1,
            process_name: [0; 64],
            inode: 0,
            rx_queue: 0,
            tx_queue: 0,
        }
    }
}

#[allow(clippy::field_reassign_with_default)]
impl From<probe_metrics::TcpConnection> for TcpConnection {
    fn from(c: probe_metrics::TcpConnection) -> Self {
        let mut result = Self::default();
        result.family = c.family.into();
        copy_str_to_carray(&c.local_addr, &mut result.local_addr);
        result.local_port = c.local_port;
        copy_str_to_carray(&c.remote_addr, &mut result.remote_addr);
        result.remote_port = c.remote_port;
        result.state = c.state.into();
        result.pid = c.pid;
        copy_str_to_carray(&c.process_name, &mut result.process_name);
        result.inode = c.inode;
        result.rx_queue = c.rx_queue;
        result.tx_queue = c.tx_queue;
        result
    }
}

/// UDP socket information.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct UdpConnection {
    /// Address family (IPv4 or IPv6).
    pub family: AddressFamily,
    /// Local IP address (null-terminated).
    pub local_addr: [c_char; MAX_ADDR_LEN],
    /// Local port.
    pub local_port: u16,
    /// Remote IP address (null-terminated, may be 0.0.0.0).
    pub remote_addr: [c_char; MAX_ADDR_LEN],
    /// Remote port (may be 0 for unconnected).
    pub remote_port: u16,
    /// Connection state.
    pub state: SocketState,
    /// Process ID owning this socket (-1 if unknown).
    pub pid: i32,
    /// Process name (null-terminated, empty if unknown).
    pub process_name: [c_char; 64],
    /// Socket inode number.
    pub inode: u64,
    /// Receive queue size.
    pub rx_queue: u32,
    /// Transmit queue size.
    pub tx_queue: u32,
}

impl Default for UdpConnection {
    fn default() -> Self {
        Self {
            family: AddressFamily::IPv4,
            local_addr: [0; MAX_ADDR_LEN],
            local_port: 0,
            remote_addr: [0; MAX_ADDR_LEN],
            remote_port: 0,
            state: SocketState::Unknown,
            pid: -1,
            process_name: [0; 64],
            inode: 0,
            rx_queue: 0,
            tx_queue: 0,
        }
    }
}

#[allow(clippy::field_reassign_with_default)]
impl From<probe_metrics::UdpConnection> for UdpConnection {
    fn from(c: probe_metrics::UdpConnection) -> Self {
        let mut result = Self::default();
        result.family = c.family.into();
        copy_str_to_carray(&c.local_addr, &mut result.local_addr);
        result.local_port = c.local_port;
        copy_str_to_carray(&c.remote_addr, &mut result.remote_addr);
        result.remote_port = c.remote_port;
        result.state = c.state.into();
        result.pid = c.pid;
        copy_str_to_carray(&c.process_name, &mut result.process_name);
        result.inode = c.inode;
        result.rx_queue = c.rx_queue;
        result.tx_queue = c.tx_queue;
        result
    }
}

/// Unix socket information.
#[repr(C)]
#[derive(Clone, Copy)]
pub struct UnixSocket {
    /// Socket path (null-terminated, may be empty for abstract sockets).
    pub path: [c_char; PROBE_MAX_PATH_LEN],
    /// Socket type (stream, dgram, seqpacket).
    pub socket_type: [c_char; 16],
    /// Connection state.
    pub state: SocketState,
    /// Process ID owning this socket (-1 if unknown).
    pub pid: i32,
    /// Process name (null-terminated, empty if unknown).
    pub process_name: [c_char; 64],
    /// Socket inode number.
    pub inode: u64,
}

impl Default for UnixSocket {
    fn default() -> Self {
        Self {
            path: [0; PROBE_MAX_PATH_LEN],
            socket_type: [0; 16],
            state: SocketState::Unknown,
            pid: -1,
            process_name: [0; 64],
            inode: 0,
        }
    }
}

impl From<probe_metrics::UnixSocket> for UnixSocket {
    fn from(s: probe_metrics::UnixSocket) -> Self {
        let mut result = Self::default();
        copy_str_to_carray(&s.path, &mut result.path);
        copy_str_to_carray(&s.socket_type, &mut result.socket_type);
        result.state = s.state.into();
        result.pid = s.pid;
        copy_str_to_carray(&s.process_name, &mut result.process_name);
        result.inode = s.inode;
        result
    }
}

/// Aggregated TCP connection statistics.
#[repr(C)]
#[derive(Clone, Copy, Default)]
pub struct TcpStats {
    /// Number of established connections.
    pub established: u32,
    /// Number of SYN_SENT connections.
    pub syn_sent: u32,
    /// Number of SYN_RECV connections.
    pub syn_recv: u32,
    /// Number of FIN_WAIT1 connections.
    pub fin_wait1: u32,
    /// Number of FIN_WAIT2 connections.
    pub fin_wait2: u32,
    /// Number of TIME_WAIT connections.
    pub time_wait: u32,
    /// Number of CLOSE connections.
    pub close: u32,
    /// Number of CLOSE_WAIT connections.
    pub close_wait: u32,
    /// Number of LAST_ACK connections.
    pub last_ack: u32,
    /// Number of LISTEN sockets.
    pub listen: u32,
    /// Number of CLOSING connections.
    pub closing: u32,
}

impl From<probe_metrics::TcpStats> for TcpStats {
    fn from(s: probe_metrics::TcpStats) -> Self {
        Self {
            established: s.established,
            syn_sent: s.syn_sent,
            syn_recv: s.syn_recv,
            fin_wait1: s.fin_wait1,
            fin_wait2: s.fin_wait2,
            time_wait: s.time_wait,
            close: s.close,
            close_wait: s.close_wait,
            last_ack: s.last_ack,
            listen: s.listen,
            closing: s.closing,
        }
    }
}

/// List of TCP connections.
#[repr(C)]
pub struct TcpConnectionList {
    pub items: *mut TcpConnection,
    pub count: usize,
    pub capacity: usize,
}

/// List of UDP connections.
#[repr(C)]
pub struct UdpConnectionList {
    pub items: *mut UdpConnection,
    pub count: usize,
    pub capacity: usize,
}

/// List of Unix sockets.
#[repr(C)]
pub struct UnixSocketList {
    pub items: *mut UnixSocket,
    pub count: usize,
    pub capacity: usize,
}

/// Collect all TCP connections.
///
/// # Safety
/// The `out` pointer must be valid. Caller must call `probe_free_tcp_connection_list` when done.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_tcp_connections(out: *mut TcpConnectionList) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::collect_tcp_connections() {
            Ok(connections) => {
                let mut items: Vec<TcpConnection> =
                    connections.into_iter().map(|c| c.into()).collect();
                let count = items.len();
                let capacity = items.capacity();
                let ptr = items.as_mut_ptr();
                std::mem::forget(items);

                unsafe {
                    (*out).items = ptr;
                    (*out).count = count;
                    (*out).capacity = capacity;
                }
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(target_os = "linux"))]
    {
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"TCP connections not supported on this platform".as_ptr(),
        )
    }
}

/// Free a TCP connection list.
///
/// # Safety
/// The list must have been allocated by `probe_collect_tcp_connections`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_free_tcp_connection_list(list: *mut TcpConnectionList) {
    if list.is_null() {
        return;
    }
    unsafe {
        let list = &mut *list;
        if !list.items.is_null() && list.capacity > 0 {
            drop(Vec::from_raw_parts(list.items, list.count, list.capacity));
            list.items = ptr::null_mut();
            list.count = 0;
            list.capacity = 0;
        }
    }
}

/// Collect all UDP sockets.
///
/// # Safety
/// The `out` pointer must be valid. Caller must call `probe_free_udp_connection_list` when done.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_udp_connections(out: *mut UdpConnectionList) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::collect_udp_connections() {
            Ok(connections) => {
                let mut items: Vec<UdpConnection> =
                    connections.into_iter().map(|c| c.into()).collect();
                let count = items.len();
                let capacity = items.capacity();
                let ptr = items.as_mut_ptr();
                std::mem::forget(items);

                unsafe {
                    (*out).items = ptr;
                    (*out).count = count;
                    (*out).capacity = capacity;
                }
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(target_os = "linux"))]
    {
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"UDP connections not supported on this platform".as_ptr(),
        )
    }
}

/// Free a UDP connection list.
///
/// # Safety
/// The list must have been allocated by `probe_collect_udp_connections`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_free_udp_connection_list(list: *mut UdpConnectionList) {
    if list.is_null() {
        return;
    }
    unsafe {
        let list = &mut *list;
        if !list.items.is_null() && list.capacity > 0 {
            drop(Vec::from_raw_parts(list.items, list.count, list.capacity));
            list.items = ptr::null_mut();
            list.count = 0;
            list.capacity = 0;
        }
    }
}

/// Collect all Unix sockets.
///
/// # Safety
/// The `out` pointer must be valid. Caller must call `probe_free_unix_socket_list` when done.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_unix_sockets(out: *mut UnixSocketList) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::collect_unix_sockets() {
            Ok(sockets) => {
                let mut items: Vec<UnixSocket> = sockets.into_iter().map(|s| s.into()).collect();
                let count = items.len();
                let capacity = items.capacity();
                let ptr = items.as_mut_ptr();
                std::mem::forget(items);

                unsafe {
                    (*out).items = ptr;
                    (*out).count = count;
                    (*out).capacity = capacity;
                }
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(target_os = "linux"))]
    {
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"Unix sockets not supported on this platform".as_ptr(),
        )
    }
}

/// Free a Unix socket list.
///
/// # Safety
/// The list must have been allocated by `probe_collect_unix_sockets`.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_free_unix_socket_list(list: *mut UnixSocketList) {
    if list.is_null() {
        return;
    }
    unsafe {
        let list = &mut *list;
        if !list.items.is_null() && list.capacity > 0 {
            drop(Vec::from_raw_parts(list.items, list.count, list.capacity));
            list.items = ptr::null_mut();
            list.count = 0;
            list.capacity = 0;
        }
    }
}

/// Collect TCP connection statistics.
///
/// # Safety
/// The `out` pointer must be valid.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_collect_tcp_stats(out: *mut TcpStats) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::collect_tcp_stats() {
            Ok(stats) => {
                unsafe { *out = TcpStats::from(stats) };
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(target_os = "linux"))]
    {
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"TCP stats not supported on this platform".as_ptr(),
        )
    }
}

/// Find which process owns a specific port.
///
/// # Safety
/// The `out` pointer must be valid. If no process is found, *out will be -1.
#[unsafe(no_mangle)]
pub unsafe extern "C" fn probe_find_process_by_port(
    port: u16,
    tcp: bool,
    out: *mut i32,
) -> ProbeResult {
    if out.is_null() {
        return ProbeResult::err(PROBE_ERR_INVALID_PARAM, c"null pointer".as_ptr());
    }

    #[cfg(target_os = "linux")]
    {
        match probe_platform::linux::find_process_by_port(port, tcp) {
            Ok(Some(pid)) => {
                unsafe { *out = pid };
                ProbeResult::ok()
            }
            Ok(None) => {
                unsafe { *out = -1 };
                ProbeResult::ok()
            }
            Err(e) => ProbeResult::from_metrics_error(e),
        }
    }

    #[cfg(not(target_os = "linux"))]
    {
        let _ = (port, tcp);
        ProbeResult::err(
            PROBE_ERR_NOT_SUPPORTED,
            c"port lookup not supported on this platform".as_ptr(),
        )
    }
}
