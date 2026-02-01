//! Inside detection - determine if running inside a container.
//!
//! Detectors are ordered by priority (highest first):
//! 1. Cloud-specific (AWS Fargate, ECS) - most specific
//! 2. Orchestrators (Kubernetes, Nomad, `OpenShift`)
//! 3. Container runtimes (Docker, Podman, etc.)
//! 4. Platform-specific (FreeBSD Jail, OpenBSD/NetBSD VMs)

mod cloud;
mod containerd;
mod crio;
mod docker;
mod kubernetes;
mod lxc;
mod nomad;
mod podman;
mod swarm;
mod systemd_nspawn;

#[cfg(target_os = "freebsd")]
mod freebsd_jail;

#[cfg(target_os = "openbsd")]
mod openbsd_vm;

#[cfg(target_os = "netbsd")]
mod netbsd_vm;

pub use cloud::{AwsEcsDetector, AwsFargateDetector, AzureAksDetector, GoogleGkeDetector};
pub use containerd::ContainerdInsideDetector;
pub use crio::CriOInsideDetector;
pub use docker::DockerInsideDetector;
pub use kubernetes::KubernetesInsideDetector;
pub use lxc::LxcInsideDetector;
pub use nomad::NomadInsideDetector;
pub use podman::PodmanInsideDetector;
pub use swarm::DockerSwarmInsideDetector;
pub use systemd_nspawn::SystemdNspawnInsideDetector;

#[cfg(target_os = "freebsd")]
pub use freebsd_jail::FreeBsdJailInsideDetector;

#[cfg(target_os = "openbsd")]
pub use openbsd_vm::OpenBsdVmInsideDetector;

#[cfg(target_os = "netbsd")]
pub use netbsd_vm::NetBsdVmInsideDetector;

use crate::InsideDetector;

/// Returns all inside detectors in priority order.
#[must_use]
pub fn all_detectors() -> Vec<Box<dyn InsideDetector>> {
    let mut detectors: Vec<Box<dyn InsideDetector>> = vec![
        // Cloud-specific (highest priority - most specific)
        Box::new(AwsFargateDetector),
        Box::new(AwsEcsDetector),
        Box::new(GoogleGkeDetector),
        Box::new(AzureAksDetector),
        // Orchestrators
        Box::new(KubernetesInsideDetector),
        Box::new(NomadInsideDetector),
        Box::new(DockerSwarmInsideDetector),
        // Container runtimes
        Box::new(DockerInsideDetector),
        Box::new(PodmanInsideDetector),
        Box::new(ContainerdInsideDetector),
        Box::new(CriOInsideDetector),
        Box::new(LxcInsideDetector),
        Box::new(SystemdNspawnInsideDetector),
    ];

    // Platform-specific
    #[cfg(target_os = "freebsd")]
    detectors.push(Box::new(FreeBsdJailInsideDetector));

    #[cfg(target_os = "openbsd")]
    detectors.push(Box::new(OpenBsdVmInsideDetector));

    #[cfg(target_os = "netbsd")]
    detectors.push(Box::new(NetBsdVmInsideDetector));

    // Sort by priority (highest first)
    detectors.sort_by_key(|b| std::cmp::Reverse(b.priority()));

    detectors
}
