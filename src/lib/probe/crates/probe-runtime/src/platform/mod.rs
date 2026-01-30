//! Platform-specific utilities.

#[cfg(target_os = "linux")]
pub mod linux;

#[cfg(target_os = "freebsd")]
pub mod freebsd;

#[cfg(target_os = "macos")]
pub mod darwin;

/// Get the current platform name.
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
pub fn supports_cgroups() -> bool {
    cfg!(target_os = "linux")
}

/// Check if the platform supports FreeBSD jails.
pub fn supports_jails() -> bool {
    cfg!(target_os = "freebsd")
}
