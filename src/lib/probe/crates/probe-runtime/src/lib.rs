//! Universal container and orchestrator runtime detection.
//!
//! This crate provides comprehensive detection of container runtimes and orchestrators:
//! - **Inside detection**: Determine if we're running inside a container
//! - **Available detection**: Find what runtimes are available on the host
//!
//! Supported runtimes:
//! - Docker, Podman, containerd, CRI-O, LXC/LXD
//! - Kubernetes, Nomad, Docker Swarm, OpenShift
//! - AWS ECS/Fargate, Google GKE, Azure AKS
//! - FreeBSD Jail, systemd-nspawn, Firecracker

pub mod available;
pub mod detector;
pub mod inside;
pub mod platform;

use std::collections::HashMap;

/// Container or orchestrator runtime type.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
#[repr(u8)]
pub enum ContainerRuntime {
    /// Not containerized / no runtime.
    None = 0,

    // Container runtimes (1-19)
    /// Docker container runtime.
    Docker = 1,
    /// Podman container runtime.
    Podman = 2,
    /// containerd runtime.
    Containerd = 3,
    /// CRI-O runtime.
    CriO = 4,
    /// LXC container.
    Lxc = 5,
    /// LXD container.
    Lxd = 6,
    /// systemd-nspawn container.
    SystemdNspawn = 7,
    /// Firecracker microVM.
    Firecracker = 8,
    /// FreeBSD Jail.
    FreeBsdJail = 9,

    // Orchestrators (20-39)
    /// Kubernetes orchestrator.
    Kubernetes = 20,
    /// HashiCorp Nomad orchestrator.
    Nomad = 21,
    /// Docker Swarm orchestrator.
    DockerSwarm = 22,
    /// OpenShift (Kubernetes variant).
    OpenShift = 23,

    // Cloud-specific (40-59)
    /// AWS ECS (Elastic Container Service).
    AwsEcs = 40,
    /// AWS Fargate (serverless ECS).
    AwsFargate = 41,
    /// Google Kubernetes Engine.
    GoogleGke = 42,
    /// Azure Kubernetes Service.
    AzureAks = 43,

    /// Unknown runtime detected.
    Unknown = 254,
}

impl ContainerRuntime {
    /// Returns the string name of the runtime.
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::None => "none",
            Self::Docker => "docker",
            Self::Podman => "podman",
            Self::Containerd => "containerd",
            Self::CriO => "cri-o",
            Self::Lxc => "lxc",
            Self::Lxd => "lxd",
            Self::SystemdNspawn => "systemd-nspawn",
            Self::Firecracker => "firecracker",
            Self::FreeBsdJail => "freebsd-jail",
            Self::Kubernetes => "kubernetes",
            Self::Nomad => "nomad",
            Self::DockerSwarm => "docker-swarm",
            Self::OpenShift => "openshift",
            Self::AwsEcs => "aws-ecs",
            Self::AwsFargate => "aws-fargate",
            Self::GoogleGke => "google-gke",
            Self::AzureAks => "azure-aks",
            Self::Unknown => "unknown",
        }
    }

    /// Returns whether this is an orchestrator (vs a container runtime).
    pub fn is_orchestrator(&self) -> bool {
        matches!(
            self,
            Self::Kubernetes
                | Self::Nomad
                | Self::DockerSwarm
                | Self::OpenShift
                | Self::AwsEcs
                | Self::AwsFargate
                | Self::GoogleGke
                | Self::AzureAks
        )
    }
}

impl std::fmt::Display for ContainerRuntime {
    fn fmt(&self, f: &mut std::fmt::Formatter<'_>) -> std::fmt::Result {
        write!(f, "{}", self.as_str())
    }
}

/// Information about running inside a container.
#[derive(Debug, Clone, Default)]
pub struct InsideInfo {
    /// The detected container runtime.
    pub runtime: ContainerRuntime,

    /// The orchestrator (if different from runtime).
    pub orchestrator: Option<ContainerRuntime>,

    /// Container ID (64-char hex for Docker, varies by runtime).
    pub container_id: Option<String>,

    /// Workload ID (allocation ID for Nomad, pod UID for K8s, etc.).
    pub workload_id: Option<String>,

    /// Workload name (job name, pod name, etc.).
    pub workload_name: Option<String>,

    /// Namespace (Kubernetes namespace, Nomad namespace, etc.).
    pub namespace: Option<String>,

    /// Additional runtime-specific metadata.
    pub metadata: HashMap<String, String>,
}

impl Default for ContainerRuntime {
    fn default() -> Self {
        Self::None
    }
}

/// Information about a runtime available on the host.
#[derive(Debug, Clone, Default)]
pub struct AvailableRuntime {
    /// The runtime type.
    pub runtime: ContainerRuntime,

    /// Unix socket path (e.g., /var/run/docker.sock).
    pub socket_path: Option<String>,

    /// API endpoint URL.
    pub api_endpoint: Option<String>,

    /// Version string.
    pub version: Option<String>,

    /// Whether the runtime is currently running/responsive.
    pub is_running: bool,
}

/// Complete runtime environment information.
#[derive(Debug, Clone, Default)]
pub struct RuntimeInfo {
    /// Whether running inside a container.
    pub is_containerized: bool,

    /// Container runtime (if containerized).
    pub container_runtime: Option<ContainerRuntime>,

    /// Orchestrator (may differ from runtime).
    pub orchestrator: Option<ContainerRuntime>,

    /// Container ID.
    pub container_id: Option<String>,

    /// Workload/allocation ID.
    pub workload_id: Option<String>,

    /// Workload/pod name.
    pub workload_name: Option<String>,

    /// Namespace.
    pub namespace: Option<String>,

    /// Available runtimes on the host.
    pub available_runtimes: Vec<AvailableRuntime>,

    /// Additional metadata.
    pub metadata: HashMap<String, String>,
}

/// Trait for detecting if running inside a specific runtime.
pub trait InsideDetector: Send + Sync {
    /// Detect if running inside this runtime.
    fn detect(&self) -> Option<InsideInfo>;

    /// Priority (higher = checked first).
    fn priority(&self) -> u8;

    /// Detector name for debugging.
    fn name(&self) -> &'static str;
}

/// Trait for detecting available runtimes on the host.
pub trait AvailableDetector: Send + Sync {
    /// Detect available instances of this runtime.
    fn detect(&self) -> Vec<AvailableRuntime>;

    /// Detector name for debugging.
    fn name(&self) -> &'static str;
}

/// Error type for runtime detection.
#[derive(Debug, thiserror::Error)]
pub enum RuntimeError {
    /// I/O error during detection.
    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),

    /// Parse error.
    #[error("Parse error: {0}")]
    Parse(String),

    /// Platform not supported.
    #[error("Platform not supported: {0}")]
    NotSupported(String),
}

/// Result type for runtime operations.
pub type Result<T> = std::result::Result<T, RuntimeError>;
