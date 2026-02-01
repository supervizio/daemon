//! FreeBSD-specific utilities.

use std::process::Command;

/// Get the jail ID if running in a jail.
pub fn get_jail_id() -> Option<i32> {
    let output = Command::new("sysctl").args(["-n", "security.jail.jid"]).output().ok()?;

    if output.status.success() {
        let jid_str = String::from_utf8_lossy(&output.stdout);
        let jid: i32 = jid_str.trim().parse().ok()?;
        if jid > 0 {
            return Some(jid);
        }
    }

    None
}

/// Check if running inside a jail.
pub fn is_jailed() -> bool {
    let output = Command::new("sysctl").args(["-n", "security.jail.jailed"]).output();

    if let Ok(output) = output {
        if output.status.success() {
            let value = String::from_utf8_lossy(&output.stdout);
            return value.trim() == "1";
        }
    }

    false
}

/// Get the jail name if running in a jail.
pub fn get_jail_name() -> Option<String> {
    let output = Command::new("jls")
        .args(["-j", &get_jail_id()?.to_string(), "-n", "name"])
        .output()
        .ok()?;

    if output.status.success() {
        let output_str = String::from_utf8_lossy(&output.stdout);
        // Format: name=<jailname>
        if let Some(name) = output_str.trim().strip_prefix("name=") {
            return Some(name.to_string());
        }
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_is_jailed() {
        // This test only makes sense on FreeBSD
        #[cfg(target_os = "freebsd")]
        {
            let _ = is_jailed();
        }
    }
}
