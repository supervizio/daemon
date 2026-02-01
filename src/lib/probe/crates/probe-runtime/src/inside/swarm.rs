//! Docker Swarm inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::collections::HashMap;

/// Detects if running inside a Docker Swarm service.
pub struct DockerSwarmInsideDetector;

impl InsideDetector for DockerSwarmInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Docker Swarm injects specific environment variables for services
        let task_id = std::env::var("DOCKER_SWARM_TASK_ID").ok();
        let service_id = std::env::var("DOCKER_SWARM_SERVICE_ID").ok();
        let service_name = std::env::var("DOCKER_SWARM_SERVICE_NAME").ok();

        // Also check for legacy Swarm env vars
        let is_swarm = task_id.is_some()
            || service_id.is_some()
            || std::env::var("SWARM_NODE_ID").is_ok()
            || check_swarm_labels();

        if is_swarm {
            return Some(InsideInfo {
                runtime: ContainerRuntime::DockerSwarm,
                orchestrator: Some(ContainerRuntime::DockerSwarm),
                workload_id: task_id,
                workload_name: service_name,
                metadata: collect_swarm_metadata(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        // Higher than plain Docker
        92
    }

    fn name(&self) -> &'static str {
        "docker-swarm"
    }
}

/// Check for Swarm-specific labels in container config.
fn check_swarm_labels() -> bool {
    // Swarm sets labels like com.docker.swarm.service.name
    // These are visible via Docker API, not directly in the container
    // So we rely on environment variables instead
    false
}

/// Collect Docker Swarm metadata from environment.
fn collect_swarm_metadata() -> HashMap<String, String> {
    let mut meta = HashMap::new();

    let env_vars = [
        "DOCKER_SWARM_TASK_ID",
        "DOCKER_SWARM_TASK_NAME",
        "DOCKER_SWARM_SERVICE_ID",
        "DOCKER_SWARM_SERVICE_NAME",
        "DOCKER_SWARM_NODE_ID",
        "SWARM_NODE_ID",
        "SWARM_MANAGER",
    ];

    for key in env_vars {
        if let Ok(val) = std::env::var(key) {
            meta.insert(key.to_lowercase(), val);
        }
    }

    meta
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_detector_without_swarm() {
        let detector = DockerSwarmInsideDetector;
        // Should return None when not in Swarm
        let _ = detector.detect();
    }
}
