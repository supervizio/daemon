//! POSIX rlimit quota reader.
//!
//! Reads resource limits using getrlimit() syscall.
//! Used on macOS, OpenBSD, and NetBSD where cgroups/rctl are not available.

use crate::{Error, QuotaLimits, QuotaReader, QuotaUsage, Result};
use libc::{getrlimit, rlimit, RLIMIT_CPU, RLIMIT_DATA, RLIMIT_NOFILE, RLIMIT_NPROC};

/// POSIX rlimit quota reader.
pub struct RlimitQuotaReader;

impl RlimitQuotaReader {
    /// Create a new rlimit quota reader.
    pub fn new() -> Self {
        Self
    }
}

impl Default for RlimitQuotaReader {
    fn default() -> Self {
        Self::new()
    }
}

impl QuotaReader for RlimitQuotaReader {
    fn read_limits(&self, _pid: i32) -> Result<QuotaLimits> {
        let mut limits = QuotaLimits::default();

        unsafe {
            let mut rl = rlimit {
                rlim_cur: 0,
                rlim_max: 0,
            };

            // RLIMIT_NOFILE - Maximum file descriptors
            if getrlimit(RLIMIT_NOFILE, &mut rl) == 0 {
                limits.nofile_limit = Some(rlimit_to_u64(rl.rlim_cur));
            }

            // RLIMIT_CPU - Maximum CPU time in seconds
            if getrlimit(RLIMIT_CPU, &mut rl) == 0 {
                limits.cpu_time_limit_secs = Some(rlimit_to_u64(rl.rlim_cur));
            }

            // RLIMIT_DATA - Maximum heap/data size
            if getrlimit(RLIMIT_DATA, &mut rl) == 0 {
                limits.data_limit_bytes = Some(rlimit_to_u64(rl.rlim_cur));
            }

            // RLIMIT_NPROC - Maximum processes
            if getrlimit(RLIMIT_NPROC, &mut rl) == 0 {
                limits.pids_limit = Some(rlimit_to_u64(rl.rlim_cur));
            }

            // Platform-specific limits
            #[cfg(target_os = "macos")]
            {
                // macOS has RLIMIT_RSS but it's not enforced (soft hint only)
                // We still read it for informational purposes
                const RLIMIT_RSS: libc::c_int = 5;
                if getrlimit(RLIMIT_RSS, &mut rl) == 0 {
                    let val = rlimit_to_u64(rl.rlim_cur);
                    // Only set if it's not unlimited (macOS typically sets this to unlimited)
                    if val != u64::MAX {
                        limits.memory_limit_bytes = Some(val);
                    }
                }
            }

            #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
            {
                // OpenBSD/NetBSD: RLIMIT_RSS exists but is "no longer enforced"
                // per the man page, so we read it but note it may not be active
                const RLIMIT_RSS: libc::c_int = 5;
                if getrlimit(RLIMIT_RSS, &mut rl) == 0 {
                    let val = rlimit_to_u64(rl.rlim_cur);
                    if val != u64::MAX {
                        limits.memory_limit_bytes = Some(val);
                    }
                }
            }
        }

        Ok(limits)
    }

    fn read_usage(&self, pid: i32) -> Result<QuotaUsage> {
        let limits = self.read_limits(pid)?;
        let mut usage = QuotaUsage::default();

        usage.memory_limit_bytes = limits.memory_limit_bytes;
        usage.pids_limit = limits.pids_limit;
        usage.cpu_limit_percent = limits.cpu_limit_percent();

        // Note: Getting actual usage requires platform-specific APIs
        // (getrusage, proc_pidinfo on macOS, etc.)
        // For now, we only report the limits

        Ok(usage)
    }
}

/// Convert rlimit value to u64, handling RLIM_INFINITY.
fn rlimit_to_u64(val: libc::rlim_t) -> u64 {
    if val == libc::RLIM_INFINITY {
        u64::MAX
    } else {
        val as u64
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_read_limits() {
        let reader = RlimitQuotaReader::new();
        let limits = reader.read_limits(std::process::id() as i32).unwrap();

        // We should always be able to read nofile limit
        assert!(limits.nofile_limit.is_some());
    }

    #[test]
    fn test_read_usage() {
        let reader = RlimitQuotaReader::new();
        let usage = reader.read_usage(std::process::id() as i32).unwrap();

        // Usage should match limits for applicable fields
        let limits = reader.read_limits(std::process::id() as i32).unwrap();
        assert_eq!(usage.pids_limit, limits.pids_limit);
    }
}
