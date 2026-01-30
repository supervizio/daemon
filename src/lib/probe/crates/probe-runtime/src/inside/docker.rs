//! Docker container inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::fs;
use std::path::Path;

/// Detects if running inside a Docker container.
pub struct DockerInsideDetector;

impl InsideDetector for DockerInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Method 1: Check /.dockerenv marker file (fastest)
        if Path::new("/.dockerenv").exists() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Docker,
                container_id: get_container_id_from_cgroup(),
                ..Default::default()
            });
        }

        // Method 2: Check cgroup for docker patterns
        if let Some(id) = check_cgroup_docker() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Docker,
                container_id: Some(id),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        90
    }

    fn name(&self) -> &'static str {
        "docker"
    }
}

/// Get container ID from cgroup.
fn get_container_id_from_cgroup() -> Option<String> {
    // Try cgroup v2 first (unified hierarchy)
    if let Ok(content) = fs::read_to_string("/proc/self/cgroup") {
        for line in content.lines() {
            // cgroup v1: 12:memory:/docker/<id>
            // cgroup v2: 0::/docker/<id>
            if let Some(id) = extract_docker_id(line) {
                return Some(id);
            }
        }
    }

    // Try hostname as fallback (Docker sets it to short container ID)
    if let Ok(hostname) = fs::read_to_string("/etc/hostname") {
        let hostname = hostname.trim();
        // Docker container hostnames are 12-char hex strings
        if hostname.len() == 12 && hostname.chars().all(|c| c.is_ascii_hexdigit()) {
            return Some(hostname.to_string());
        }
    }

    None
}

/// Check cgroup for docker patterns and extract container ID.
fn check_cgroup_docker() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        if let Some(id) = extract_docker_id(line) {
            return Some(id);
        }
    }

    None
}

/// Extract Docker container ID from a cgroup line.
fn extract_docker_id(line: &str) -> Option<String> {
    // Patterns to look for:
    // /docker/<64-char-id>
    // /docker-<64-char-id>.scope
    // /system.slice/docker-<64-char-id>.scope

    let patterns = ["/docker/", "/docker-", "docker-"];

    for pattern in patterns {
        if let Some(pos) = line.find(pattern) {
            let start = pos + pattern.len();
            let rest = &line[start..];

            // Extract 64-char hex ID
            let id: String = rest
                .chars()
                .take_while(|c| c.is_ascii_hexdigit())
                .collect();

            if id.len() == 64 {
                return Some(id);
            }
        }
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_docker_id() {
        let id = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";

        // cgroup v1 format
        let line = format!("12:memory:/docker/{}", id);
        assert_eq!(extract_docker_id(&line), Some(id.to_string()));

        // cgroup v2 format
        let line = format!("0::/docker/{}", id);
        assert_eq!(extract_docker_id(&line), Some(id.to_string()));

        // systemd scope format
        let line = format!("0::/system.slice/docker-{}.scope", id);
        assert_eq!(extract_docker_id(&line), Some(id.to_string()));

        // Not docker
        let line = "12:memory:/user.slice";
        assert_eq!(extract_docker_id(line), None);
    }
}
