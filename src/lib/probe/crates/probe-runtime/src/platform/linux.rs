//! Linux-specific utilities.

use std::fs;

/// Cgroup version detected on the system.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum CgroupVersion {
    /// cgroup v1 (legacy).
    V1,
    /// cgroup v2 (unified).
    V2,
    /// Hybrid (both v1 and v2).
    Hybrid,
    /// Unknown or not available.
    Unknown,
}

/// Detect the cgroup version in use.
pub fn detect_cgroup_version() -> CgroupVersion {
    // Check for cgroup v2 unified hierarchy
    let v2_exists = fs::metadata("/sys/fs/cgroup/cgroup.controllers").is_ok();

    // Check for cgroup v1
    let v1_exists =
        fs::metadata("/sys/fs/cgroup/memory").is_ok() || fs::metadata("/sys/fs/cgroup/cpu").is_ok();

    match (v2_exists, v1_exists) {
        (true, false) => CgroupVersion::V2,
        (false, true) => CgroupVersion::V1,
        (true, true) => CgroupVersion::Hybrid,
        (false, false) => CgroupVersion::Unknown,
    }
}

/// Read cgroup file for the current process.
pub fn read_self_cgroup() -> Option<String> {
    fs::read_to_string("/proc/self/cgroup").ok()
}

/// Get the cgroup path for the current process.
pub fn get_cgroup_path() -> Option<String> {
    let content = read_self_cgroup()?;

    // For cgroup v2, look for the unified entry (0::)
    for line in content.lines() {
        if let Some(stripped) = line.strip_prefix("0::") {
            return Some(stripped.to_string());
        }
    }

    // For cgroup v1, use any path (they should be similar)
    for line in content.lines() {
        if let Some((_, path)) = line.rsplit_once(':')
            && !path.is_empty()
            && path != "/"
        {
            return Some(path.to_string());
        }
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_detect_cgroup_version() {
        let version = detect_cgroup_version();
        // Should return something on Linux
        let _ = version;
    }

    #[test]
    fn test_get_cgroup_path() {
        let path = get_cgroup_path();
        // May or may not return something depending on environment
        let _ = path;
    }
}
