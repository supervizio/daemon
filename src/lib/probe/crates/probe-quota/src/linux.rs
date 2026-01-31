//! Linux cgroups v1/v2 quota reader.
//!
//! Reads resource limits from cgroups filesystem without applying them.

use crate::{ContainerInfo, ContainerRuntime, Error, QuotaLimits, QuotaReader, QuotaUsage, Result};
use std::fs;
use std::path::{Path, PathBuf};

/// Linux quota reader using cgroups.
pub struct LinuxQuotaReader {
    /// Detected cgroups version (1 or 2).
    cgroup_version: CgroupVersion,
}

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
enum CgroupVersion {
    V1,
    V2,
    Unknown,
}

impl LinuxQuotaReader {
    /// Create a new Linux quota reader.
    pub fn new() -> Self {
        let cgroup_version = detect_cgroup_version();
        Self { cgroup_version }
    }

    /// Get the cgroup path for a process.
    fn get_cgroup_path(&self, pid: i32) -> Result<PathBuf> {
        let cgroup_file = format!("/proc/{}/cgroup", pid);
        let content = fs::read_to_string(&cgroup_file).map_err(|e| {
            if e.kind() == std::io::ErrorKind::NotFound {
                Error::NotFound(pid)
            } else {
                Error::Io(e)
            }
        })?;

        match self.cgroup_version {
            CgroupVersion::V2 => parse_cgroup_v2_path(&content),
            CgroupVersion::V1 => parse_cgroup_v1_path(&content),
            CgroupVersion::Unknown => Err(Error::NotSupported),
        }
    }

    fn read_cgroup_v2_limits(&self, cgroup_path: &Path) -> QuotaLimits {
        let mut limits = QuotaLimits::default();

        // CPU limits from cpu.max: "quota period" or "max period"
        if let Ok(content) = fs::read_to_string(cgroup_path.join("cpu.max"))
            && let Some((quota, period)) = parse_cpu_max(&content)
        {
            limits.cpu_quota_us = Some(quota);
            limits.cpu_period_us = Some(period);
        }

        // Memory limit from memory.max
        if let Ok(content) = fs::read_to_string(cgroup_path.join("memory.max")) {
            limits.memory_limit_bytes = parse_cgroup_value(&content);
        }

        // PIDs limit from pids.max
        if let Ok(content) = fs::read_to_string(cgroup_path.join("pids.max")) {
            limits.pids_limit = parse_cgroup_value(&content);
        }

        // I/O limits from io.max (format: "MAJ:MIN rbps=X wbps=X riops=Y wiops=Z")
        if let Ok(content) = fs::read_to_string(cgroup_path.join("io.max")) {
            let (rbps, wbps) = parse_io_max(&content);
            limits.io_read_bps = rbps;
            limits.io_write_bps = wbps;
        }

        // Also read rlimits for nofile, cpu time, data
        read_rlimits_into(&mut limits);

        limits
    }

    fn read_cgroup_v1_limits(&self, _cgroup_path: &Path) -> QuotaLimits {
        let mut limits = QuotaLimits::default();

        // In cgroups v1, different controllers are in different paths
        // cpu controller: /sys/fs/cgroup/cpu/...
        // memory controller: /sys/fs/cgroup/memory/...

        // CPU quota from cpu.cfs_quota_us and cpu.cfs_period_us
        if let Ok(quota) = fs::read_to_string("/sys/fs/cgroup/cpu/cpu.cfs_quota_us")
            && let Ok(val) = quota.trim().parse::<i64>()
            && val > 0
        {
            limits.cpu_quota_us = Some(val as u64);
        }
        if let Ok(period) = fs::read_to_string("/sys/fs/cgroup/cpu/cpu.cfs_period_us")
            && let Ok(val) = period.trim().parse::<u64>()
        {
            limits.cpu_period_us = Some(val);
        }

        // Memory limit
        if let Ok(content) = fs::read_to_string("/sys/fs/cgroup/memory/memory.limit_in_bytes") {
            limits.memory_limit_bytes = parse_cgroup_value(&content);
        }

        // PIDs limit
        if let Ok(content) = fs::read_to_string("/sys/fs/cgroup/pids/pids.max") {
            limits.pids_limit = parse_cgroup_value(&content);
        }

        // Also read rlimits
        read_rlimits_into(&mut limits);

        limits
    }

    fn read_cgroup_v2_usage(&self, cgroup_path: &Path, limits: &QuotaLimits) -> QuotaUsage {
        let mut usage = QuotaUsage::default();

        // Memory usage from memory.current
        if let Ok(content) = fs::read_to_string(cgroup_path.join("memory.current"))
            && let Ok(val) = content.trim().parse::<u64>()
        {
            usage.memory_bytes = val;
        }
        usage.memory_limit_bytes = limits.memory_limit_bytes;

        // PIDs current from pids.current
        if let Ok(content) = fs::read_to_string(cgroup_path.join("pids.current"))
            && let Ok(val) = content.trim().parse::<u64>()
        {
            usage.pids_current = val;
        }
        usage.pids_limit = limits.pids_limit;

        // CPU limit percentage
        usage.cpu_limit_percent = limits.cpu_limit_percent();

        usage
    }
}

impl Default for LinuxQuotaReader {
    fn default() -> Self {
        Self::new()
    }
}

impl QuotaReader for LinuxQuotaReader {
    fn read_limits(&self, pid: i32) -> Result<QuotaLimits> {
        let cgroup_path = self.get_cgroup_path(pid)?;

        let limits = match self.cgroup_version {
            CgroupVersion::V2 => self.read_cgroup_v2_limits(&cgroup_path),
            CgroupVersion::V1 => self.read_cgroup_v1_limits(&cgroup_path),
            CgroupVersion::Unknown => {
                // Fall back to rlimits only
                let mut limits = QuotaLimits::default();
                read_rlimits_into(&mut limits);
                limits
            }
        };

        Ok(limits)
    }

    fn read_usage(&self, pid: i32) -> Result<QuotaUsage> {
        let limits = self.read_limits(pid)?;
        let cgroup_path = self.get_cgroup_path(pid)?;

        let usage = match self.cgroup_version {
            CgroupVersion::V2 => self.read_cgroup_v2_usage(&cgroup_path, &limits),
            CgroupVersion::V1 => {
                // V1 usage reading - simplified
                QuotaUsage {
                    memory_limit_bytes: limits.memory_limit_bytes,
                    pids_limit: limits.pids_limit,
                    cpu_limit_percent: limits.cpu_limit_percent(),
                    ..Default::default()
                }
            }
            CgroupVersion::Unknown => QuotaUsage::default(),
        };

        Ok(usage)
    }
}

/// Detect cgroups version.
fn detect_cgroup_version() -> CgroupVersion {
    // Check for cgroups v2 unified hierarchy
    if Path::new("/sys/fs/cgroup/cgroup.controllers").exists() {
        return CgroupVersion::V2;
    }

    // Check for cgroups v1
    if Path::new("/sys/fs/cgroup/cpu").exists() || Path::new("/sys/fs/cgroup/memory").exists() {
        return CgroupVersion::V1;
    }

    CgroupVersion::Unknown
}

/// Parse cgroup v2 path from /proc/PID/cgroup.
/// Format: "0::/path/to/cgroup"
fn parse_cgroup_v2_path(content: &str) -> Result<PathBuf> {
    for line in content.lines() {
        let parts: Vec<&str> = line.splitn(3, ':').collect();
        if parts.len() == 3 && parts[0] == "0" {
            let cgroup_relative = parts[2].trim();
            let path =
                PathBuf::from("/sys/fs/cgroup").join(cgroup_relative.trim_start_matches('/'));
            if path.exists() {
                return Ok(path);
            }
        }
    }

    // Default to root cgroup
    Ok(PathBuf::from("/sys/fs/cgroup"))
}

/// Parse cgroup v1 path from /proc/PID/cgroup.
/// Format: "hierarchy-id:controller-list:path"
fn parse_cgroup_v1_path(content: &str) -> Result<PathBuf> {
    for line in content.lines() {
        let parts: Vec<&str> = line.splitn(3, ':').collect();
        if parts.len() == 3 {
            // Look for memory or cpu controller
            let controllers = parts[1];
            if controllers.contains("memory") || controllers.contains("cpu") {
                let path = parts[2].trim();
                return Ok(PathBuf::from(format!("/sys/fs/cgroup/memory{}", path)));
            }
        }
    }

    Ok(PathBuf::from("/sys/fs/cgroup"))
}

/// Parse cpu.max format: "quota period" or "max period".
fn parse_cpu_max(content: &str) -> Option<(u64, u64)> {
    let parts: Vec<&str> = content.split_whitespace().collect();
    if parts.len() >= 2 {
        let quota = if parts[0] == "max" {
            u64::MAX
        } else {
            parts[0].parse().ok()?
        };
        let period = parts[1].parse().ok()?;
        return Some((quota, period));
    }
    None
}

/// Parse cgroup value that can be "max" or a number.
fn parse_cgroup_value(content: &str) -> Option<u64> {
    let trimmed = content.trim();
    if trimmed == "max" {
        Some(u64::MAX)
    } else {
        trimmed.parse().ok()
    }
}

/// Parse io.max format: "MAJ:MIN rbps=X wbps=X riops=Y wiops=Z".
fn parse_io_max(content: &str) -> (Option<u64>, Option<u64>) {
    let mut rbps = None;
    let mut wbps = None;

    for line in content.lines() {
        for part in line.split_whitespace() {
            if let Some(val) = part.strip_prefix("rbps=") {
                rbps = if val == "max" {
                    Some(u64::MAX)
                } else {
                    val.parse().ok()
                };
            } else if let Some(val) = part.strip_prefix("wbps=") {
                wbps = if val == "max" {
                    Some(u64::MAX)
                } else {
                    val.parse().ok()
                };
            }
        }
    }

    (rbps, wbps)
}

/// Read rlimits into QuotaLimits.
fn read_rlimits_into(limits: &mut QuotaLimits) {
    use libc::{RLIMIT_CPU, RLIMIT_DATA, RLIMIT_NOFILE, RLIMIT_NPROC, getrlimit, rlimit};

    unsafe {
        let mut rl = rlimit {
            rlim_cur: 0,
            rlim_max: 0,
        };

        // RLIMIT_NOFILE
        if getrlimit(RLIMIT_NOFILE, &mut rl) == 0 {
            limits.nofile_limit = Some(if rl.rlim_cur == libc::RLIM_INFINITY {
                u64::MAX
            } else {
                rl.rlim_cur
            });
        }

        // RLIMIT_CPU
        if getrlimit(RLIMIT_CPU, &mut rl) == 0 {
            limits.cpu_time_limit_secs = Some(if rl.rlim_cur == libc::RLIM_INFINITY {
                u64::MAX
            } else {
                rl.rlim_cur
            });
        }

        // RLIMIT_DATA
        if getrlimit(RLIMIT_DATA, &mut rl) == 0 {
            limits.data_limit_bytes = Some(if rl.rlim_cur == libc::RLIM_INFINITY {
                u64::MAX
            } else {
                rl.rlim_cur
            });
        }

        // RLIMIT_NPROC (if not already set from cgroups)
        if limits.pids_limit.is_none() && getrlimit(RLIMIT_NPROC, &mut rl) == 0 {
            limits.pids_limit = Some(if rl.rlim_cur == libc::RLIM_INFINITY {
                u64::MAX
            } else {
                rl.rlim_cur
            });
        }
    }
}

/// Detect container runtime on Linux.
pub fn detect_container() -> ContainerInfo {
    // Check for Kubernetes first (most specific)
    if std::env::var("KUBERNETES_SERVICE_HOST").is_ok() {
        return ContainerInfo {
            is_containerized: true,
            runtime: ContainerRuntime::Kubernetes,
            container_id: get_container_id_from_cgroup(),
        };
    }

    // Check for Docker
    if Path::new("/.dockerenv").exists() {
        return ContainerInfo {
            is_containerized: true,
            runtime: ContainerRuntime::Docker,
            container_id: get_container_id_from_cgroup(),
        };
    }

    // Check for Podman
    if Path::new("/run/.containerenv").exists() || Path::new("/.containerenv").exists() {
        return ContainerInfo {
            is_containerized: true,
            runtime: ContainerRuntime::Podman,
            container_id: get_container_id_from_cgroup(),
        };
    }

    // Check cgroup for container hints
    if let Ok(content) = fs::read_to_string("/proc/1/cgroup") {
        if content.contains("/docker/") || content.contains("/docker-") {
            return ContainerInfo {
                is_containerized: true,
                runtime: ContainerRuntime::Docker,
                container_id: extract_container_id(&content, "docker"),
            };
        }
        if content.contains("/kubepods/") || content.contains("/kubepods.slice/") {
            return ContainerInfo {
                is_containerized: true,
                runtime: ContainerRuntime::Kubernetes,
                container_id: extract_container_id(&content, "kubepods"),
            };
        }
        if content.contains("/lxc/") {
            return ContainerInfo {
                is_containerized: true,
                runtime: ContainerRuntime::LXC,
                container_id: extract_container_id(&content, "lxc"),
            };
        }
    }

    // Check if we're in a non-root cgroup (likely containerized)
    if let Ok(content) = fs::read_to_string("/proc/1/cgroup") {
        let lines: Vec<&str> = content.lines().collect();
        if !lines.is_empty() {
            let parts: Vec<&str> = lines[0].splitn(3, ':').collect();
            if parts.len() == 3 && parts[2] != "/" {
                return ContainerInfo {
                    is_containerized: true,
                    runtime: ContainerRuntime::Unknown,
                    container_id: None,
                };
            }
        }
    }

    ContainerInfo::default()
}

fn get_container_id_from_cgroup() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;
    extract_container_id(&content, "")
}

fn extract_container_id(content: &str, hint: &str) -> Option<String> {
    for line in content.lines() {
        let parts: Vec<&str> = line.splitn(3, ':').collect();
        if parts.len() == 3 {
            let path = parts[2];
            // Look for 64-char hex ID (Docker/Podman container ID)
            for segment in path.split('/') {
                if segment.len() == 64 && segment.chars().all(|c| c.is_ascii_hexdigit()) {
                    return Some(segment.to_string());
                }
                // Also check if segment contains the hint and has an ID after it
                if !hint.is_empty()
                    && segment.contains(hint)
                    && let Some(id_part) = segment.split('-').next_back()
                    && id_part.len() >= 12
                    && id_part.chars().all(|c| c.is_ascii_hexdigit())
                {
                    return Some(id_part.to_string());
                }
            }
        }
    }
    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_cpu_max() {
        assert_eq!(parse_cpu_max("100000 100000\n"), Some((100000, 100000)));
        assert_eq!(parse_cpu_max("max 100000\n"), Some((u64::MAX, 100000)));
        assert_eq!(parse_cpu_max("50000 100000"), Some((50000, 100000)));
    }

    #[test]
    fn test_parse_cgroup_value() {
        assert_eq!(parse_cgroup_value("max\n"), Some(u64::MAX));
        assert_eq!(parse_cgroup_value("1073741824\n"), Some(1073741824));
        assert_eq!(parse_cgroup_value("100"), Some(100));
    }

    #[test]
    fn test_parse_io_max() {
        let content = "8:0 rbps=104857600 wbps=52428800\n";
        let (rbps, wbps) = parse_io_max(content);
        assert_eq!(rbps, Some(104857600));
        assert_eq!(wbps, Some(52428800));

        let content2 = "8:0 rbps=max wbps=max\n";
        let (rbps2, wbps2) = parse_io_max(content2);
        assert_eq!(rbps2, Some(u64::MAX));
        assert_eq!(wbps2, Some(u64::MAX));
    }
}
