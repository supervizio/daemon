//! NetBSD VM/hypervisor detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};

#[cfg(target_os = "netbsd")]
use crate::platform::netbsd::{HypervisorType, detect_virtualization};

/// Detector for NetBSD VM/hypervisor environments.
#[derive(Debug, Default)]
pub struct NetBsdVmInsideDetector;

impl InsideDetector for NetBsdVmInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        #[cfg(target_os = "netbsd")]
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
            metadata.insert("platform".to_string(), "netbsd".to_string());

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

        #[cfg(not(target_os = "netbsd"))]
        {
            None
        }
    }

    fn priority(&self) -> u8 {
        // Lower priority than containers, as VMs are less specific
        10
    }

    fn name(&self) -> &'static str {
        "netbsd-vm"
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_detector_creation() {
        let detector = NetBsdVmInsideDetector;
        assert_eq!(detector.name(), "netbsd-vm");
        assert_eq!(detector.priority(), 10);
    }

    #[test]
    fn test_detect() {
        let detector = NetBsdVmInsideDetector;
        // Should not panic, may or may not detect depending on environment
        let _ = detector.detect();
    }
}
