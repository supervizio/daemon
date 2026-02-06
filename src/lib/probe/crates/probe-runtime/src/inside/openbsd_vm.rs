//! OpenBSD VM/hypervisor detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};

#[cfg(target_os = "openbsd")]
use crate::platform::openbsd::{HypervisorType, detect_virtualization};

/// Detector for OpenBSD VM/hypervisor environments.
#[derive(Debug, Default)]
pub struct OpenBsdVmInsideDetector;

impl InsideDetector for OpenBsdVmInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        #[cfg(target_os = "openbsd")]
        {
            let hypervisor = detect_virtualization()?;

            let runtime = match hypervisor {
                HypervisorType::VMware => ContainerRuntime::VMware,
                HypervisorType::Qemu => ContainerRuntime::Qemu,
                HypervisorType::VirtualBox => ContainerRuntime::VirtualBox,
                HypervisorType::HyperV => ContainerRuntime::HyperV,
                HypervisorType::Bhyve => ContainerRuntime::Bhyve,
                HypervisorType::Xen => ContainerRuntime::Xen,
                HypervisorType::Parallels => ContainerRuntime::Parallels,
                HypervisorType::Unknown => ContainerRuntime::Unknown,
            };

            let mut metadata = std::collections::HashMap::new();
            metadata.insert("hypervisor".to_string(), hypervisor.as_str().to_string());
            metadata.insert("platform".to_string(), "openbsd".to_string());

            Some(InsideInfo {
                runtime,
                orchestrator: None,
                container_id: None,
                workload_id: None,
                workload_name: None,
                namespace: None,
                metadata,
            })
        }

        #[cfg(not(target_os = "openbsd"))]
        {
            None
        }
    }

    fn priority(&self) -> u8 {
        // Lower priority than containers, as VMs are less specific
        10
    }

    fn name(&self) -> &'static str {
        "openbsd-vm"
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_detector_creation() {
        let detector = OpenBsdVmInsideDetector;
        assert_eq!(detector.name(), "openbsd-vm");
        assert_eq!(detector.priority(), 10);
    }

    #[test]
    fn test_detect() {
        let detector = OpenBsdVmInsideDetector;
        // Should not panic, may or may not detect depending on environment
        let _ = detector.detect();
    }
}
