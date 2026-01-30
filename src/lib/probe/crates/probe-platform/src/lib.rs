//! probe-platform - Platform-specific implementations
//!
//! This crate provides system metrics collection for each supported platform:
//! - Linux: via /proc filesystem
//! - macOS: via Mach APIs and sysctl
//! - BSD: via sysctl and kvm

pub use probe_metrics::{
    AddressFamily, CPUCollector, CPUPressure, ConnectionCollector, ContextSwitches, DiskCollector,
    DiskIOStats, DiskUsage, Error, IOCollector, IOPressure, IOStats, LoadAverage, LoadCollector,
    MemoryCollector, MemoryPressure, NetInterface, NetStats, NetworkCollector, Partition,
    ProcessCollector, ProcessMetrics, ProcessState, Result, SocketState, SystemCPU, SystemCollector,
    SystemMemory, TcpConnection, TcpStats, ThermalCollector, ThermalZone, UdpConnection, UnixSocket,
};

// Platform-specific modules
#[cfg(target_os = "linux")]
pub mod linux;

#[cfg(target_os = "macos")]
pub mod darwin;

#[cfg(any(target_os = "freebsd", target_os = "openbsd", target_os = "netbsd"))]
pub mod bsd;

// Re-export the platform-specific collector
#[cfg(target_os = "linux")]
pub use linux::LinuxCollector as PlatformCollector;

#[cfg(target_os = "macos")]
pub use darwin::DarwinCollector as PlatformCollector;

#[cfg(any(target_os = "freebsd", target_os = "openbsd", target_os = "netbsd"))]
pub use bsd::BsdCollector as PlatformCollector;

// Fallback for unsupported platforms
#[cfg(not(any(
    target_os = "linux",
    target_os = "macos",
    target_os = "freebsd",
    target_os = "openbsd",
    target_os = "netbsd"
)))]
pub mod stub;

#[cfg(not(any(
    target_os = "linux",
    target_os = "macos",
    target_os = "freebsd",
    target_os = "openbsd",
    target_os = "netbsd"
)))]
pub use stub::StubCollector as PlatformCollector;

/// Create a new platform-specific collector.
pub fn new_collector() -> PlatformCollector {
    PlatformCollector::new()
}
