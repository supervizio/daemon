//! FreeBSD Jail inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};

/// Detects if running inside a FreeBSD jail.
#[cfg(target_os = "freebsd")]
pub struct FreeBsdJailInsideDetector;

#[cfg(target_os = "freebsd")]
impl InsideDetector for FreeBsdJailInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Check sysctl security.jail.jailed
        if is_jailed() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::FreeBsdJail,
                container_id: get_jail_id(),
                workload_name: get_jail_name(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        85
    }

    fn name(&self) -> &'static str {
        "freebsd-jail"
    }
}

/// Check if we're inside a jail using sysctl.
#[cfg(target_os = "freebsd")]
fn is_jailed() -> bool {
    use std::process::Command;

    // sysctl -n security.jail.jailed returns 1 if jailed, 0 otherwise
    if let Ok(output) = Command::new("sysctl")
        .args(["-n", "security.jail.jailed"])
        .output()
    {
        if output.status.success() {
            let value = String::from_utf8_lossy(&output.stdout);
            return value.trim() == "1";
        }
    }

    // Alternative: check if jail_get syscall returns valid data
    false
}

/// Get jail ID.
#[cfg(target_os = "freebsd")]
fn get_jail_id() -> Option<String> {
    use std::process::Command;

    // sysctl -n security.jail.jid returns the jail ID
    if let Ok(output) = Command::new("sysctl")
        .args(["-n", "security.jail.jid"])
        .output()
    {
        if output.status.success() {
            let jid = String::from_utf8_lossy(&output.stdout).trim().to_string();
            if !jid.is_empty() && jid != "0" {
                return Some(jid);
            }
        }
    }

    None
}

/// Get jail name.
#[cfg(target_os = "freebsd")]
fn get_jail_name() -> Option<String> {
    use std::process::Command;

    // jls -n name returns the jail name
    if let Ok(output) = Command::new("jls").args(["-n", "name"]).output() {
        if output.status.success() {
            let name = String::from_utf8_lossy(&output.stdout).trim().to_string();
            // Format is "name=<jailname>"
            if let Some(stripped) = name.strip_prefix("name=") {
                return Some(stripped.to_string());
            }
        }
    }

    None
}

#[cfg(target_os = "freebsd")]
#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_detector() {
        let detector = FreeBsdJailInsideDetector;
        // Will only detect if actually in a jail
        let _ = detector.detect();
    }
}
