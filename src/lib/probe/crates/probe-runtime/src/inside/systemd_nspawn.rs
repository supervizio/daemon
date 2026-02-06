//! systemd-nspawn container inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::fs;
use std::path::Path;

/// Detects if running inside a systemd-nspawn container.
pub struct SystemdNspawnInsideDetector;

impl InsideDetector for SystemdNspawnInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Method 1: Check /run/host/ which nspawn mounts
        if Path::new("/run/host/container-manager").exists()
            && let Ok(manager) = fs::read_to_string("/run/host/container-manager")
            && manager.trim() == "systemd-nspawn"
        {
            return Some(InsideInfo {
                runtime: ContainerRuntime::SystemdNspawn,
                container_id: get_machine_name(),
                ..Default::default()
            });
        }

        // Method 2: Check container= in /proc/1/environ
        if check_init_environ_nspawn() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::SystemdNspawn,
                container_id: get_machine_name(),
                ..Default::default()
            });
        }

        // Method 3: Check cgroup for machine.slice
        if check_cgroup_machine_slice() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::SystemdNspawn,
                container_id: get_machine_name_from_cgroup(),
                ..Default::default()
            });
        }

        // Method 4: Check for /run/systemd/nspawn marker
        if Path::new("/run/systemd/nspawn").exists() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::SystemdNspawn,
                container_id: get_machine_name(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        75
    }

    fn name(&self) -> &'static str {
        "systemd-nspawn"
    }
}

/// Check /proc/1/environ for container=systemd-nspawn.
fn check_init_environ_nspawn() -> bool {
    if let Ok(environ) = fs::read_to_string("/proc/1/environ") {
        for var in environ.split('\0') {
            if var == "container=systemd-nspawn" {
                return true;
            }
        }
    }
    false
}

/// Check if cgroup contains machine.slice (systemd machine management).
fn check_cgroup_machine_slice() -> bool {
    if let Ok(content) = fs::read_to_string("/proc/self/cgroup") {
        return content.contains("/machine.slice/") || content.contains("machine-");
    }
    false
}

/// Get machine name from /etc/machine-id or hostname.
fn get_machine_name() -> Option<String> {
    // Try /run/host/machine-name first (nspawn specific)
    if let Ok(name) = fs::read_to_string("/run/host/machine-name") {
        return Some(name.trim().to_string());
    }

    // Try hostname
    if let Ok(hostname) = fs::read_to_string("/etc/hostname") {
        return Some(hostname.trim().to_string());
    }

    None
}

/// Get machine name from cgroup path.
fn get_machine_name_from_cgroup() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        // Pattern: /machine.slice/machine-<name>.scope
        if let Some(pos) = line.find("/machine.slice/machine-") {
            let rest = &line[pos + 23..];
            let name = rest.trim_end_matches(".scope");
            // Unescape - systemd escapes special chars
            let name = unescape_systemd_name(name);
            if !name.is_empty() {
                return Some(name);
            }
        }
    }

    None
}

/// Unescape systemd unit name (e.g., \x2d -> -)
fn unescape_systemd_name(name: &str) -> String {
    let mut result = String::with_capacity(name.len());
    let mut chars = name.chars().peekable();

    while let Some(c) = chars.next() {
        if c == '\\' && chars.peek() == Some(&'x') {
            chars.next(); // consume 'x'
            let hex: String = chars.by_ref().take(2).collect();
            if let Ok(byte) = u8::from_str_radix(&hex, 16) {
                result.push(byte as char);
                continue;
            }
        }
        result.push(c);
    }

    result
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_unescape_systemd_name() {
        assert_eq!(unescape_systemd_name("my\\x2dcontainer"), "my-container");
        assert_eq!(unescape_systemd_name("simple"), "simple");
    }
}
