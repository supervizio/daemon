//! Platform-specific utilities.

#[cfg(target_os = "linux")]
pub mod linux;

#[cfg(target_os = "freebsd")]
pub mod freebsd;

#[cfg(target_os = "macos")]
pub mod darwin;

#[cfg(target_os = "openbsd")]
pub mod openbsd;

#[cfg(target_os = "netbsd")]
pub mod netbsd;

/// Get the current platform name.
#[must_use]
pub fn current_platform() -> &'static str {
    #[cfg(target_os = "linux")]
    {
        "linux"
    }
    #[cfg(target_os = "macos")]
    {
        "darwin"
    }
    #[cfg(target_os = "freebsd")]
    {
        "freebsd"
    }
    #[cfg(target_os = "openbsd")]
    {
        "openbsd"
    }
    #[cfg(target_os = "netbsd")]
    {
        "netbsd"
    }
    #[cfg(not(any(
        target_os = "linux",
        target_os = "macos",
        target_os = "freebsd",
        target_os = "openbsd",
        target_os = "netbsd"
    )))]
    {
        "unknown"
    }
}

/// Check if the platform supports cgroups.
#[must_use]
pub fn supports_cgroups() -> bool {
    cfg!(target_os = "linux")
}

/// Check if the platform supports FreeBSD jails.
#[must_use]
pub fn supports_jails() -> bool {
    cfg!(target_os = "freebsd")
}

/// Check if the platform supports hypervisor detection.
#[must_use]
pub fn supports_hypervisor_detection() -> bool {
    cfg!(any(target_os = "openbsd", target_os = "netbsd"))
}
