//! containerd runtime inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::fs;

/// Detects if running inside a containerd container (non-K8s).
pub struct ContainerdInsideDetector;

impl InsideDetector for ContainerdInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Check cgroup for containerd patterns
        // Note: K8s uses containerd but will be caught by K8s detector first
        if let Some(id) = check_cgroup_containerd() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Containerd,
                container_id: Some(id),
                ..Default::default()
            });
        }

        // Check for containerd-specific environment
        if std::env::var("CONTAINERD_NAMESPACE").is_ok()
            || std::env::var("CONTAINERD_ADDRESS").is_ok()
        {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Containerd,
                container_id: get_container_id_from_cgroup(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        // Lower than K8s/Docker/Podman which use containerd underneath
        70
    }

    fn name(&self) -> &'static str {
        "containerd"
    }
}

/// Check cgroup for containerd patterns.
fn check_cgroup_containerd() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        // Patterns:
        // /system.slice/containerd.service/cri-containerd-<id>
        // /containerd/<namespace>/<id>

        if line.contains("cri-containerd-") || line.contains("/containerd/") {
            // Don't match if it's clearly kubernetes
            if line.contains("/kubepods") {
                continue;
            }

            // Extract container ID
            if let Some(id) = extract_containerd_id(line) {
                return Some(id);
            }
        }
    }

    None
}

/// Extract containerd container ID from cgroup line.
fn extract_containerd_id(line: &str) -> Option<String> {
    // Try cri-containerd- prefix
    if let Some(pos) = line.find("cri-containerd-") {
        let start = pos + "cri-containerd-".len();
        let rest = &line[start..];

        let id: String = rest.chars().take_while(char::is_ascii_hexdigit).collect();

        if id.len() == 64 {
            return Some(id);
        }
    }

    // Try /containerd/<namespace>/<id> pattern
    if let Some(pos) = line.find("/containerd/") {
        let rest = &line[pos + 12..];
        let parts: Vec<&str> = rest.split('/').collect();
        if parts.len() >= 2 {
            let potential_id = parts[1].trim_end_matches(".scope");
            let id: String = potential_id.chars().take_while(char::is_ascii_hexdigit).collect();

            if id.len() == 64 {
                return Some(id);
            }
        }
    }

    None
}

/// Get container ID from cgroup (fallback).
fn get_container_id_from_cgroup() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        if let Some(id) = extract_containerd_id(line) {
            return Some(id);
        }
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_containerd_id() {
        let id = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";

        let line = format!("0::/cri-containerd-{}.scope", id);
        assert_eq!(extract_containerd_id(&line), Some(id.to_string()));

        // Non-matching
        let line = "0::/user.slice";
        assert_eq!(extract_containerd_id(line), None);
    }
}
