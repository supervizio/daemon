//! FreeBSD rctl quota reader.
//!
//! Reads resource limits from rctl without applying them.
//! Requires `kern.racct.enable=1` in `/boot/loader.conf` for full functionality.

use crate::{ContainerInfo, ContainerRuntime, Error, QuotaLimits, QuotaReader, QuotaUsage, Result};
use std::ffi::CString;
use std::process::Command;

/// FreeBSD quota reader using rctl.
pub struct FreeBSDQuotaReader {
    /// Whether rctl is available.
    rctl_available: bool,
}

impl FreeBSDQuotaReader {
    /// Create a new FreeBSD quota reader.
    pub fn new() -> Self {
        Self {
            rctl_available: check_rctl_available(),
        }
    }

    /// Read rctl rules for a process.
    fn read_rctl_rules(&self, pid: i32) -> Vec<RctlRule> {
        if !self.rctl_available {
            return Vec::new();
        }

        // Use rctl -l to list rules for this process
        let output = Command::new("rctl")
            .args(["-l", &format!("process:{}", pid)])
            .output();

        match output {
            Ok(out) if out.status.success() => {
                let stdout = String::from_utf8_lossy(&out.stdout);
                parse_rctl_output(&stdout)
            }
            _ => Vec::new(),
        }
    }
}

impl Default for FreeBSDQuotaReader {
    fn default() -> Self {
        Self::new()
    }
}

impl QuotaReader for FreeBSDQuotaReader {
    fn read_limits(&self, pid: i32) -> Result<QuotaLimits> {
        let mut limits = QuotaLimits::default();

        // Read rctl rules
        let rules = self.read_rctl_rules(pid);
        for rule in rules {
            match rule.resource.as_str() {
                "pcpu" => {
                    // pcpu is percentage of one CPU core
                    // Convert to microseconds quota with 100ms period
                    let percent = rule.amount;
                    limits.cpu_quota_us = Some(percent * 1000); // percent * 1000us
                    limits.cpu_period_us = Some(100_000); // 100ms standard period
                }
                "memoryuse" => {
                    limits.memory_limit_bytes = Some(rule.amount);
                }
                "maxproc" => {
                    limits.pids_limit = Some(rule.amount);
                }
                "openfiles" => {
                    limits.nofile_limit = Some(rule.amount);
                }
                "cputime" => {
                    limits.cpu_time_limit_secs = Some(rule.amount);
                }
                "datasize" => {
                    limits.data_limit_bytes = Some(rule.amount);
                }
                "readbps" => {
                    limits.io_read_bps = Some(rule.amount);
                }
                "writebps" => {
                    limits.io_write_bps = Some(rule.amount);
                }
                _ => {}
            }
        }

        // Also read rlimits as fallback
        read_rlimits_into(&mut limits);

        Ok(limits)
    }

    fn read_usage(&self, pid: i32) -> Result<QuotaUsage> {
        let limits = self.read_limits(pid)?;
        let mut usage = QuotaUsage::default();

        usage.memory_limit_bytes = limits.memory_limit_bytes;
        usage.pids_limit = limits.pids_limit;
        usage.cpu_limit_percent = limits.cpu_limit_percent();

        // Read current usage from rctl -u
        if self.rctl_available {
            if let Ok(output) = Command::new("rctl")
                .args(["-u", &format!("process:{}", pid)])
                .output()
            {
                if output.status.success() {
                    let stdout = String::from_utf8_lossy(&output.stdout);
                    for line in stdout.lines() {
                        let parts: Vec<&str> = line.split('=').collect();
                        if parts.len() == 2 {
                            let resource = parts[0].split(':').last().unwrap_or("");
                            if let Ok(val) = parts[1].parse::<u64>() {
                                match resource {
                                    "memoryuse" => usage.memory_bytes = val,
                                    "maxproc" => usage.pids_current = val,
                                    _ => {}
                                }
                            }
                        }
                    }
                }
            }
        }

        Ok(usage)
    }
}

/// Parsed rctl rule.
struct RctlRule {
    resource: String,
    amount: u64,
}

/// Check if rctl is available.
fn check_rctl_available() -> bool {
    // Check sysctl kern.racct.enable
    if let Ok(output) = Command::new("sysctl")
        .args(["-n", "kern.racct.enable"])
        .output()
    {
        if output.status.success() {
            let val = String::from_utf8_lossy(&output.stdout);
            return val.trim() == "1";
        }
    }
    false
}

/// Parse rctl output.
/// Format: "subject:id:resource:action=amount"
fn parse_rctl_output(output: &str) -> Vec<RctlRule> {
    let mut rules = Vec::new();

    for line in output.lines() {
        let line = line.trim();
        if line.is_empty() {
            continue;
        }

        // Format: process:1234:pcpu:deny=50
        let parts: Vec<&str> = line.split(':').collect();
        if parts.len() >= 4 {
            let resource = parts[2].to_string();
            // Parse "action=amount"
            if let Some((_, amount_str)) = parts[3].split_once('=') {
                if let Ok(amount) = parse_rctl_amount(amount_str) {
                    rules.push(RctlRule { resource, amount });
                }
            }
        }
    }

    rules
}

/// Parse rctl amount which can have suffixes like k, m, g.
fn parse_rctl_amount(s: &str) -> std::result::Result<u64, std::num::ParseIntError> {
    let s = s.trim().to_lowercase();

    if let Some(num) = s.strip_suffix('k') {
        return num.parse::<u64>().map(|v| v * 1024);
    }
    if let Some(num) = s.strip_suffix('m') {
        return num.parse::<u64>().map(|v| v * 1024 * 1024);
    }
    if let Some(num) = s.strip_suffix('g') {
        return num.parse::<u64>().map(|v| v * 1024 * 1024 * 1024);
    }

    s.parse()
}

/// Read rlimits into QuotaLimits.
fn read_rlimits_into(limits: &mut QuotaLimits) {
    use libc::{getrlimit, rlimit, RLIMIT_CPU, RLIMIT_DATA, RLIMIT_NOFILE, RLIMIT_NPROC};

    unsafe {
        let mut rl = rlimit {
            rlim_cur: 0,
            rlim_max: 0,
        };

        // RLIMIT_NOFILE (if not already set from rctl)
        if limits.nofile_limit.is_none() {
            if getrlimit(RLIMIT_NOFILE, &mut rl) == 0 {
                limits.nofile_limit = Some(if rl.rlim_cur == libc::RLIM_INFINITY {
                    u64::MAX
                } else {
                    rl.rlim_cur
                });
            }
        }

        // RLIMIT_CPU (if not already set)
        if limits.cpu_time_limit_secs.is_none() {
            if getrlimit(RLIMIT_CPU, &mut rl) == 0 {
                limits.cpu_time_limit_secs = Some(if rl.rlim_cur == libc::RLIM_INFINITY {
                    u64::MAX
                } else {
                    rl.rlim_cur
                });
            }
        }

        // RLIMIT_DATA (if not already set)
        if limits.data_limit_bytes.is_none() {
            if getrlimit(RLIMIT_DATA, &mut rl) == 0 {
                limits.data_limit_bytes = Some(if rl.rlim_cur == libc::RLIM_INFINITY {
                    u64::MAX
                } else {
                    rl.rlim_cur
                });
            }
        }

        // RLIMIT_NPROC (if not already set)
        if limits.pids_limit.is_none() {
            if getrlimit(RLIMIT_NPROC, &mut rl) == 0 {
                limits.pids_limit = Some(if rl.rlim_cur == libc::RLIM_INFINITY {
                    u64::MAX
                } else {
                    rl.rlim_cur
                });
            }
        }
    }
}

/// Detect if running in a FreeBSD jail.
pub fn detect_container() -> ContainerInfo {
    // Check sysctl security.jail.jailed
    if let Ok(output) = Command::new("sysctl")
        .args(["-n", "security.jail.jailed"])
        .output()
    {
        if output.status.success() {
            let val = String::from_utf8_lossy(&output.stdout);
            if val.trim() == "1" {
                // Get jail ID
                let jail_id = Command::new("sysctl")
                    .args(["-n", "security.jail.jid"])
                    .output()
                    .ok()
                    .map(|o| String::from_utf8_lossy(&o.stdout).trim().to_string());

                return ContainerInfo {
                    is_containerized: true,
                    runtime: ContainerRuntime::FreeBSDJail,
                    container_id: jail_id,
                };
            }
        }
    }

    ContainerInfo::default()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_rctl_amount() {
        assert_eq!(parse_rctl_amount("1024"), Ok(1024));
        assert_eq!(parse_rctl_amount("1k"), Ok(1024));
        assert_eq!(parse_rctl_amount("1m"), Ok(1024 * 1024));
        assert_eq!(parse_rctl_amount("1g"), Ok(1024 * 1024 * 1024));
        assert_eq!(parse_rctl_amount("100M"), Ok(100 * 1024 * 1024));
    }

    #[test]
    fn test_parse_rctl_output() {
        let output = "process:1234:pcpu:deny=50\nprocess:1234:memoryuse:deny=1g\n";
        let rules = parse_rctl_output(output);
        assert_eq!(rules.len(), 2);
        assert_eq!(rules[0].resource, "pcpu");
        assert_eq!(rules[0].amount, 50);
        assert_eq!(rules[1].resource, "memoryuse");
        assert_eq!(rules[1].amount, 1024 * 1024 * 1024);
    }
}
