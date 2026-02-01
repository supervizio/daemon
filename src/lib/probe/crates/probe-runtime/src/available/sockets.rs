//! Unix socket detection for container runtimes.

use crate::{AvailableDetector, AvailableRuntime, ContainerRuntime};
use std::os::unix::net::UnixStream;
use std::path::Path;
use std::time::Duration;

/// Known socket paths for container runtimes.
const SOCKET_PATHS: &[(&str, ContainerRuntime)] = &[
    // Docker
    ("/var/run/docker.sock", ContainerRuntime::Docker),
    ("/run/docker.sock", ContainerRuntime::Docker),
    // Podman (rootful)
    ("/var/run/podman/podman.sock", ContainerRuntime::Podman),
    ("/run/podman/podman.sock", ContainerRuntime::Podman),
    // containerd
    ("/var/run/containerd/containerd.sock", ContainerRuntime::Containerd),
    ("/run/containerd/containerd.sock", ContainerRuntime::Containerd),
    // CRI-O
    ("/var/run/crio/crio.sock", ContainerRuntime::CriO),
    ("/run/crio/crio.sock", ContainerRuntime::CriO),
    // LXD
    ("/var/lib/lxd/unix.socket", ContainerRuntime::Lxd),
    ("/var/snap/lxd/common/lxd/unix.socket", ContainerRuntime::Lxd),
];

/// Detects available runtimes via Unix sockets.
pub struct SocketDetector;

impl AvailableDetector for SocketDetector {
    fn detect(&self) -> Vec<AvailableRuntime> {
        let mut available = Vec::new();

        // Check standard socket paths
        for (path, runtime) in SOCKET_PATHS {
            if Path::new(path).exists() {
                available.push(AvailableRuntime {
                    runtime: *runtime,
                    socket_path: Some((*path).to_string()),
                    is_running: probe_socket(path),
                    ..Default::default()
                });
            }
        }

        // Check rootless Podman socket
        if let Some(runtime) = check_rootless_podman() {
            available.push(runtime);
        }

        // Check macOS Docker Desktop socket
        #[cfg(target_os = "macos")]
        if let Some(runtime) = check_macos_docker() {
            available.push(runtime);
        }

        // Check Colima socket (macOS)
        #[cfg(target_os = "macos")]
        if let Some(runtime) = check_colima() {
            available.push(runtime);
        }

        // Deduplicate by runtime type (keep first found)
        let mut seen = std::collections::HashSet::new();
        available.retain(|r| seen.insert(r.runtime));

        available
    }

    fn name(&self) -> &'static str {
        "sockets"
    }
}

/// Probe a Unix socket to check if it's responsive.
fn probe_socket(path: &str) -> bool {
    UnixStream::connect(path)
        .and_then(|stream| stream.set_read_timeout(Some(Duration::from_millis(100))))
        .is_ok()
}

/// Check for rootless Podman socket in `XDG_RUNTIME_DIR`.
fn check_rootless_podman() -> Option<AvailableRuntime> {
    let xdg_runtime = std::env::var("XDG_RUNTIME_DIR").ok()?;
    let socket_path = format!("{xdg_runtime}/podman/podman.sock");

    if Path::new(&socket_path).exists() {
        return Some(AvailableRuntime {
            runtime: ContainerRuntime::Podman,
            socket_path: Some(socket_path.clone()),
            is_running: probe_socket(&socket_path),
            ..Default::default()
        });
    }

    None
}

/// Check for Docker Desktop socket on macOS.
#[cfg(target_os = "macos")]
fn check_macos_docker() -> Option<AvailableRuntime> {
    let home = std::env::var("HOME").ok()?;

    // Docker Desktop locations
    let paths = [
        format!("{}/.docker/run/docker.sock", home),
        format!("{}/Library/Containers/com.docker.docker/Data/docker.sock", home),
    ];

    for path in paths {
        if Path::new(&path).exists() {
            return Some(AvailableRuntime {
                runtime: ContainerRuntime::Docker,
                socket_path: Some(path.clone()),
                is_running: probe_socket(&path),
                ..Default::default()
            });
        }
    }

    None
}

/// Check for Colima Docker socket on macOS.
#[cfg(target_os = "macos")]
fn check_colima() -> Option<AvailableRuntime> {
    let home = std::env::var("HOME").ok()?;
    let socket_path = format!("{}/.colima/default/docker.sock", home);

    if Path::new(&socket_path).exists() {
        return Some(AvailableRuntime {
            runtime: ContainerRuntime::Docker,
            socket_path: Some(socket_path.clone()),
            is_running: probe_socket(&socket_path),
            ..Default::default()
        });
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_socket_detector() {
        let detector = SocketDetector;
        let available = detector.detect();
        // Returns whatever is available on the system
        for runtime in &available {
            assert!(runtime.socket_path.is_some());
        }
    }
}
