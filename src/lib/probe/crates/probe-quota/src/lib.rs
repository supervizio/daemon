//! probe-quota - Resource Quota Detection (Read-Only)
//!
//! This crate provides cross-platform detection of resource limits:
//! - Linux: read cgroups v1/v2 limits
//! - FreeBSD: read rctl limits
//! - macOS/BSD: read rlimits
//!
//! **Note**: This crate only READS existing limits, it does NOT apply them.
//! Quota enforcement should be handled by containers (Docker, Podman) or
//! orchestrators (Kubernetes, systemd).

use thiserror::Error;

// Platform-specific modules
#[cfg(target_os = "linux")]
mod linux;

#[cfg(target_os = "freebsd")]
mod freebsd;

#[cfg(any(target_os = "macos", target_os = "openbsd", target_os = "netbsd"))]
mod rlimit;

// Re-exports
#[cfg(target_os = "linux")]
pub use linux::LinuxQuotaReader;

#[cfg(target_os = "freebsd")]
pub use freebsd::FreeBSDQuotaReader;

#[cfg(any(target_os = "macos", target_os = "openbsd", target_os = "netbsd"))]
pub use rlimit::RlimitQuotaReader;

/// Error types for quota operations.
#[derive(Error, Debug)]
pub enum Error {
    /// Quota detection not supported on this platform.
    #[error("quota detection not supported")]
    NotSupported,

    /// Permission denied.
    #[error("permission denied: {0}")]
    Permission(String),

    /// Process not found.
    #[error("process not found: {0}")]
    NotFound(i32),

    /// I/O error.
    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),

    /// Parse error.
    #[error("parse error: {0}")]
    Parse(String),
}

/// Result type alias for quota operations.
pub type Result<T> = std::result::Result<T, Error>;

/// Resource limits detected for a process.
///
/// All fields are `Option<u64>` where:
/// - `None` means the limit could not be determined or is not applicable
/// - `Some(u64::MAX)` means unlimited
/// - `Some(value)` is the actual limit
#[derive(Debug, Clone, Default)]
pub struct QuotaLimits {
    /// CPU quota in microseconds per period (cgroups).
    /// `None` if not in a cgroup or no CPU limit set.
    pub cpu_quota_us: Option<u64>,

    /// CPU period in microseconds (cgroups).
    /// Typically 100000 (100ms). Used with cpu_quota_us to calculate %.
    pub cpu_period_us: Option<u64>,

    /// Memory limit in bytes.
    /// `None` if no limit, `Some(u64::MAX)` if unlimited.
    pub memory_limit_bytes: Option<u64>,

    /// Maximum number of processes/threads.
    /// From cgroups pids.max or RLIMIT_NPROC.
    pub pids_limit: Option<u64>,

    /// Maximum file descriptors.
    /// From RLIMIT_NOFILE.
    pub nofile_limit: Option<u64>,

    /// Maximum CPU time in seconds.
    /// From RLIMIT_CPU.
    pub cpu_time_limit_secs: Option<u64>,

    /// Maximum heap/data size in bytes.
    /// From RLIMIT_DATA.
    pub data_limit_bytes: Option<u64>,

    /// I/O read bandwidth limit in bytes/sec.
    /// From cgroups io.max or rctl.
    pub io_read_bps: Option<u64>,

    /// I/O write bandwidth limit in bytes/sec.
    /// From cgroups io.max or rctl.
    pub io_write_bps: Option<u64>,
}

impl QuotaLimits {
    /// Check if any CPU limit is set.
    pub fn has_cpu_limit(&self) -> bool {
        self.cpu_quota_us.is_some() && self.cpu_quota_us != Some(u64::MAX)
    }

    /// Check if memory limit is set.
    pub fn has_memory_limit(&self) -> bool {
        self.memory_limit_bytes.is_some() && self.memory_limit_bytes != Some(u64::MAX)
    }

    /// Calculate CPU limit as percentage (0.0 - 100.0+ for multi-core).
    /// Returns `None` if no CPU limit is set.
    pub fn cpu_limit_percent(&self) -> Option<f64> {
        match (self.cpu_quota_us, self.cpu_period_us) {
            (Some(quota), Some(period)) if quota != u64::MAX && period > 0 => {
                Some((quota as f64 / period as f64) * 100.0)
            }
            _ => None,
        }
    }
}

/// Current resource usage for a process.
#[derive(Debug, Clone, Default)]
pub struct QuotaUsage {
    /// Current memory usage in bytes.
    pub memory_bytes: u64,

    /// Memory limit in bytes (if any).
    pub memory_limit_bytes: Option<u64>,

    /// Current number of processes/threads.
    pub pids_current: u64,

    /// PIDs limit (if any).
    pub pids_limit: Option<u64>,

    /// Current CPU usage percentage.
    pub cpu_percent: f64,

    /// CPU limit percentage (if any).
    pub cpu_limit_percent: Option<f64>,
}

impl QuotaUsage {
    /// Calculate memory usage as percentage of limit.
    /// Returns `None` if no limit is set.
    pub fn memory_usage_percent(&self) -> Option<f64> {
        self.memory_limit_bytes.and_then(|limit| {
            if limit > 0 && limit != u64::MAX {
                Some((self.memory_bytes as f64 / limit as f64) * 100.0)
            } else {
                None
            }
        })
    }

    /// Calculate PIDs usage as percentage of limit.
    /// Returns `None` if no limit is set.
    pub fn pids_usage_percent(&self) -> Option<f64> {
        self.pids_limit.and_then(|limit| {
            if limit > 0 && limit != u64::MAX {
                Some((self.pids_current as f64 / limit as f64) * 100.0)
            } else {
                None
            }
        })
    }
}

/// Trait for reading resource quotas (detection only).
pub trait QuotaReader: Send + Sync {
    /// Read resource limits for a process.
    ///
    /// Returns the detected limits for the given PID.
    /// For the current process, use `std::process::id() as i32`.
    fn read_limits(&self, pid: i32) -> Result<QuotaLimits>;

    /// Read current resource usage for a process.
    ///
    /// Returns usage metrics that can be compared against limits.
    fn read_usage(&self, pid: i32) -> Result<QuotaUsage>;
}

/// Container runtime detection.
#[derive(Debug, Clone, PartialEq, Eq)]
pub enum ContainerRuntime {
    /// Not in a container.
    None,
    /// Docker container.
    Docker,
    /// Podman container.
    Podman,
    /// LXC/LXD container.
    LXC,
    /// Kubernetes pod.
    Kubernetes,
    /// FreeBSD jail.
    FreeBSDJail,
    /// Unknown container type.
    Unknown,
}

/// Container information.
#[derive(Debug, Clone)]
pub struct ContainerInfo {
    /// Whether running in a container.
    pub is_containerized: bool,
    /// Detected container runtime.
    pub runtime: ContainerRuntime,
    /// Container ID if available.
    pub container_id: Option<String>,
}

impl Default for ContainerInfo {
    fn default() -> Self {
        Self {
            is_containerized: false,
            runtime: ContainerRuntime::None,
            container_id: None,
        }
    }
}

/// Check if quota detection is supported on this platform.
pub fn is_supported() -> bool {
    #[cfg(target_os = "linux")]
    {
        // Cgroups are always readable if the fs exists
        std::path::Path::new("/sys/fs/cgroup").exists()
            || std::path::Path::new("/proc/self/cgroup").exists()
    }

    #[cfg(target_os = "freebsd")]
    {
        // rctl is available on modern FreeBSD
        true
    }

    #[cfg(any(target_os = "macos", target_os = "openbsd", target_os = "netbsd"))]
    {
        // getrlimit is always available
        true
    }

    #[cfg(not(any(
        target_os = "linux",
        target_os = "freebsd",
        target_os = "macos",
        target_os = "openbsd",
        target_os = "netbsd"
    )))]
    {
        false
    }
}

/// Create a platform-specific quota reader.
pub fn new_reader() -> Box<dyn QuotaReader> {
    #[cfg(target_os = "linux")]
    {
        Box::new(LinuxQuotaReader::new())
    }

    #[cfg(target_os = "freebsd")]
    {
        Box::new(FreeBSDQuotaReader::new())
    }

    #[cfg(any(target_os = "macos", target_os = "openbsd", target_os = "netbsd"))]
    {
        Box::new(RlimitQuotaReader::new())
    }

    #[cfg(not(any(
        target_os = "linux",
        target_os = "freebsd",
        target_os = "macos",
        target_os = "openbsd",
        target_os = "netbsd"
    )))]
    {
        Box::new(StubQuotaReader)
    }
}

/// Stub reader for unsupported platforms.
#[cfg(not(any(
    target_os = "linux",
    target_os = "freebsd",
    target_os = "macos",
    target_os = "openbsd",
    target_os = "netbsd"
)))]
struct StubQuotaReader;

#[cfg(not(any(
    target_os = "linux",
    target_os = "freebsd",
    target_os = "macos",
    target_os = "openbsd",
    target_os = "netbsd"
)))]
impl QuotaReader for StubQuotaReader {
    fn read_limits(&self, _pid: i32) -> Result<QuotaLimits> {
        Err(Error::NotSupported)
    }

    fn read_usage(&self, _pid: i32) -> Result<QuotaUsage> {
        Err(Error::NotSupported)
    }
}

/// Detect container runtime.
pub fn detect_container() -> ContainerInfo {
    #[cfg(target_os = "linux")]
    {
        linux::detect_container()
    }

    #[cfg(target_os = "freebsd")]
    {
        freebsd::detect_container()
    }

    #[cfg(not(any(target_os = "linux", target_os = "freebsd")))]
    {
        ContainerInfo::default()
    }
}
