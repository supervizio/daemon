//! Available detection - find runtimes available on the host.
//!
//! Detects runtimes through:
//! - Unix sockets (Docker, Podman, containerd, CRI-O, LXD)
//! - CLI tools (docker, podman, kubectl, nomad, etc.)
//! - Configuration files (kubeconfig, nomad config)

mod cli;
mod kubernetes;
mod nomad;
mod sockets;

pub use cli::CliDetector;
pub use kubernetes::KubernetesAvailableDetector;
pub use nomad::NomadAvailableDetector;
pub use sockets::SocketDetector;

use crate::AvailableDetector;

/// Returns all available detectors.
pub fn all_detectors() -> Vec<Box<dyn AvailableDetector>> {
    vec![
        Box::new(SocketDetector),
        Box::new(CliDetector),
        Box::new(KubernetesAvailableDetector),
        Box::new(NomadAvailableDetector),
    ]
}
