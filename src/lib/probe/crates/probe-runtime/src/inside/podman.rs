//! Podman container inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::collections::HashMap;
use std::fs;
use std::path::Path;

/// Detects if running inside a Podman container.
pub struct PodmanInsideDetector;

impl InsideDetector for PodmanInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Method 1: Check /run/.containerenv (Podman-specific marker)
        if Path::new("/run/.containerenv").exists() {
            let mut info = InsideInfo {
                runtime: ContainerRuntime::Podman,
                ..Default::default()
            };

            // Parse containerenv for additional info
            if let Ok(content) = fs::read_to_string("/run/.containerenv") {
                info.metadata = parse_containerenv(&content);
                info.container_id = info.metadata.get("id").cloned();
                info.workload_name = info.metadata.get("name").cloned();
            }

            return Some(info);
        }

        // Method 2: Check cgroup for podman patterns
        if let Some((id, metadata)) = check_cgroup_podman() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Podman,
                container_id: Some(id),
                metadata,
                ..Default::default()
            });
        }

        // Method 3: Check for PODMAN_* environment variables
        if std::env::var("container").ok().as_deref() == Some("podman") {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Podman,
                container_id: get_container_id_from_cgroup(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        // Higher than Docker (podman containers might have /.dockerenv too)
        91
    }

    fn name(&self) -> &'static str {
        "podman"
    }
}

/// Parse /run/.containerenv file.
fn parse_containerenv(content: &str) -> HashMap<String, String> {
    let mut meta = HashMap::new();

    for line in content.lines() {
        // Format: key="value" or key=value
        if let Some((key, value)) = line.split_once('=') {
            let value = value.trim_matches('"');
            meta.insert(key.to_string(), value.to_string());
        }
    }

    meta
}

/// Check cgroup for podman patterns.
fn check_cgroup_podman() -> Option<(String, HashMap<String, String>)> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        // Look for libpod- prefix (Podman's internal name)
        if let Some(pos) = line.find("libpod-") {
            let start = pos + "libpod-".len();
            let rest = &line[start..];

            // Extract 64-char hex ID
            let id: String = rest
                .chars()
                .take_while(|c| c.is_ascii_hexdigit())
                .collect();

            if id.len() == 64 {
                return Some((id, HashMap::new()));
            }
        }
    }

    None
}

/// Get container ID from cgroup (fallback).
fn get_container_id_from_cgroup() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        // Look for any container-like path
        for pattern in ["/libpod-", "/crio-", "/docker/"] {
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
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_containerenv() {
        let content = r#"engine="podman-4.0.0"
name="mycontainer"
id="abc123def456"
image="nginx:latest"
"#;

        let meta = parse_containerenv(content);
        assert_eq!(meta.get("engine"), Some(&"podman-4.0.0".to_string()));
        assert_eq!(meta.get("name"), Some(&"mycontainer".to_string()));
        assert_eq!(meta.get("id"), Some(&"abc123def456".to_string()));
    }
}
