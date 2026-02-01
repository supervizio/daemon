//! LXC/LXD container inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::fs;
use std::path::Path;

/// Detects if running inside an LXC container.
pub struct LxcInsideDetector;

impl InsideDetector for LxcInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Method 1: Check container= environment variable in /proc/1/environ
        if check_init_environ_lxc() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Lxc,
                container_id: get_container_name_from_cgroup(),
                ..Default::default()
            });
        }

        // Method 2: Check cgroup for /lxc/ pattern
        if let Some(name) = check_cgroup_lxc() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Lxc,
                container_id: Some(name),
                ..Default::default()
            });
        }

        // Method 3: Check for LXC-specific paths
        if Path::new("/dev/lxc").exists() || Path::new("/.lxcfs").exists() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Lxc,
                container_id: get_container_name_from_cgroup(),
                ..Default::default()
            });
        }

        // Method 4: Check for LXD (which uses LXC)
        if std::env::var("LXD_DIR").is_ok() || Path::new("/dev/.lxd-mounts").exists() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Lxd,
                container_id: get_container_name_from_cgroup(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        80
    }

    fn name(&self) -> &'static str {
        "lxc"
    }
}

/// Check /proc/1/environ for container=lxc.
fn check_init_environ_lxc() -> bool {
    if let Ok(environ) = fs::read_to_string("/proc/1/environ") {
        // environ is null-separated
        for var in environ.split('\0') {
            if var == "container=lxc" || var == "container=lxd" {
                return true;
            }
        }
    }
    false
}

/// Check cgroup for LXC patterns.
fn check_cgroup_lxc() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        // Pattern: /lxc/<container-name> or /lxc.payload/<name>
        if let Some(pos) = line.find("/lxc/") {
            let name = &line[pos + 5..];
            let name = name.split('/').next().unwrap_or(name);
            if !name.is_empty() {
                return Some(name.to_string());
            }
        }

        if let Some(pos) = line.find("/lxc.payload/") {
            let name = &line[pos + 13..];
            let name = name.split('/').next().unwrap_or(name);
            if !name.is_empty() {
                return Some(name.to_string());
            }
        }
    }

    None
}

/// Get container name from cgroup.
fn get_container_name_from_cgroup() -> Option<String> {
    check_cgroup_lxc()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_check_cgroup_lxc() {
        // Test with typical LXC cgroup line
        // This just verifies the function works
        let _ = check_cgroup_lxc();
    }
}
