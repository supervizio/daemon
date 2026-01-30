//! CRI-O runtime inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::fs;

/// Detects if running inside a CRI-O container.
pub struct CriOInsideDetector;

impl InsideDetector for CriOInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Check cgroup for crio patterns
        if let Some(id) = check_cgroup_crio() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::CriO,
                container_id: Some(id),
                ..Default::default()
            });
        }

        // Check for CRI-O specific files
        if std::path::Path::new("/run/crio").exists() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::CriO,
                container_id: get_container_id_from_cgroup(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        // Lower than K8s which typically uses CRI-O
        72
    }

    fn name(&self) -> &'static str {
        "cri-o"
    }
}

/// Check cgroup for CRI-O patterns.
fn check_cgroup_crio() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        // Patterns:
        // /crio-<id>
        // /crio/<id>
        // /system.slice/crio-<id>.scope

        // Skip if this is clearly kubernetes-managed
        // (K8s detector should have higher priority anyway)

        if let Some(id) = extract_crio_id(line) {
            return Some(id);
        }
    }

    None
}

/// Extract CRI-O container ID from cgroup line.
fn extract_crio_id(line: &str) -> Option<String> {
    for pattern in ["crio-", "/crio/"] {
        if let Some(pos) = line.find(pattern) {
            let start = pos + pattern.len();
            let rest = &line[start..];

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

/// Get container ID from cgroup (fallback).
fn get_container_id_from_cgroup() -> Option<String> {
    check_cgroup_crio()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_crio_id() {
        let id = "a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2c3d4e5f6a1b2";

        let line = format!("0::/crio-{}.scope", id);
        assert_eq!(extract_crio_id(&line), Some(id.to_string()));

        let line = format!("12:memory:/crio/{}", id);
        assert_eq!(extract_crio_id(&line), Some(id.to_string()));
    }
}
