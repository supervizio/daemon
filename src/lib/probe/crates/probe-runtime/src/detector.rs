//! Universal runtime detector - main entry point.

use crate::{
    AvailableDetector, AvailableRuntime, ContainerRuntime, InsideDetector, RuntimeInfo, available,
    inside,
};

/// Universal runtime detector that coordinates all detection methods.
pub struct UniversalRuntimeDetector {
    inside_detectors: Vec<Box<dyn InsideDetector>>,
    available_detectors: Vec<Box<dyn AvailableDetector>>,
}

impl Default for UniversalRuntimeDetector {
    fn default() -> Self {
        Self::new()
    }
}

impl UniversalRuntimeDetector {
    /// Create a new detector with all built-in detectors.
    #[must_use]
    pub fn new() -> Self {
        Self {
            inside_detectors: inside::all_detectors(),
            available_detectors: available::all_detectors(),
        }
    }

    /// Create a detector with custom detectors.
    #[must_use]
    pub fn with_detectors(
        inside_detectors: Vec<Box<dyn InsideDetector>>,
        available_detectors: Vec<Box<dyn AvailableDetector>>,
    ) -> Self {
        Self { inside_detectors, available_detectors }
    }

    /// Perform full runtime environment detection.
    #[must_use]
    pub fn detect(&self) -> RuntimeInfo {
        let mut info = RuntimeInfo::default();

        // Detect if we're inside a container
        for detector in &self.inside_detectors {
            log::trace!("Running inside detector: {}", detector.name());
            if let Some(inside) = detector.detect() {
                log::debug!("Inside detection matched: {} ({})", detector.name(), inside.runtime);
                info.is_containerized = true;
                info.container_runtime = Some(inside.runtime);
                info.orchestrator = inside.orchestrator;
                info.container_id = inside.container_id;
                info.workload_id = inside.workload_id;
                info.workload_name = inside.workload_name;
                info.namespace = inside.namespace;
                info.metadata = inside.metadata;
                break; // First match wins (sorted by priority)
            }
        }

        // Detect available runtimes on host
        for detector in &self.available_detectors {
            log::trace!("Running available detector: {}", detector.name());
            let detected = detector.detect();
            log::debug!("Available detector {} found {} runtimes", detector.name(), detected.len());
            info.available_runtimes.extend(detected);
        }

        // Deduplicate available runtimes (keep first of each type)
        deduplicate_available(&mut info.available_runtimes);

        info
    }

    /// Detect only if inside a container (faster, no host detection).
    #[must_use]
    pub fn detect_inside(&self) -> Option<RuntimeInfo> {
        for detector in &self.inside_detectors {
            if let Some(inside) = detector.detect() {
                return Some(RuntimeInfo {
                    is_containerized: true,
                    container_runtime: Some(inside.runtime),
                    orchestrator: inside.orchestrator,
                    container_id: inside.container_id,
                    workload_id: inside.workload_id,
                    workload_name: inside.workload_name,
                    namespace: inside.namespace,
                    metadata: inside.metadata,
                    available_runtimes: Vec::new(),
                });
            }
        }
        None
    }

    /// Detect only available runtimes on host (no inside detection).
    #[must_use]
    pub fn detect_available(&self) -> Vec<AvailableRuntime> {
        let mut available = Vec::new();

        for detector in &self.available_detectors {
            available.extend(detector.detect());
        }

        deduplicate_available(&mut available);
        available
    }
}

/// Deduplicate available runtimes by type, keeping first occurrence.
fn deduplicate_available(runtimes: &mut Vec<AvailableRuntime>) {
    let mut seen = std::collections::HashSet::new();
    runtimes.retain(|r| seen.insert(r.runtime));
}

/// Convenience function for quick detection.
#[must_use]
pub fn detect() -> RuntimeInfo {
    UniversalRuntimeDetector::new().detect()
}

/// Convenience function to check if containerized.
#[must_use]
pub fn is_containerized() -> bool {
    UniversalRuntimeDetector::new().detect_inside().is_some()
}

/// Convenience function to get container runtime.
#[must_use]
pub fn get_container_runtime() -> Option<ContainerRuntime> {
    UniversalRuntimeDetector::new().detect_inside().and_then(|info| info.container_runtime)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_detector_creation() {
        let detector = UniversalRuntimeDetector::new();
        assert!(!detector.inside_detectors.is_empty());
        assert!(!detector.available_detectors.is_empty());
    }

    #[test]
    fn test_detect() {
        let info = detect();
        // Should always return a RuntimeInfo, even if not containerized
        // The actual values depend on the environment
        let _ = info;
    }

    #[test]
    fn test_deduplicate() {
        let mut runtimes = vec![
            AvailableRuntime {
                runtime: ContainerRuntime::Docker,
                socket_path: Some("/var/run/docker.sock".into()),
                ..Default::default()
            },
            AvailableRuntime {
                runtime: ContainerRuntime::Docker,
                socket_path: Some("/run/docker.sock".into()),
                ..Default::default()
            },
            AvailableRuntime {
                runtime: ContainerRuntime::Podman,
                socket_path: Some("/run/podman/podman.sock".into()),
                ..Default::default()
            },
        ];

        deduplicate_available(&mut runtimes);

        assert_eq!(runtimes.len(), 2);
        assert_eq!(runtimes[0].runtime, ContainerRuntime::Docker);
        assert_eq!(runtimes[1].runtime, ContainerRuntime::Podman);
    }
}
