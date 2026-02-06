//! OpenBSD-specific utilities for VM/hypervisor detection.

use std::fs;
use std::process::Command;

/// Hypervisor type detected on OpenBSD.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum HypervisorType {
    /// VMware virtualization.
    VMware,
    /// QEMU/KVM virtualization.
    Qemu,
    /// VirtualBox virtualization.
    VirtualBox,
    /// Microsoft Hyper-V.
    HyperV,
    /// bhyve (FreeBSD hypervisor).
    Bhyve,
    /// Xen hypervisor.
    Xen,
    /// Parallels Desktop.
    Parallels,
    /// Unknown hypervisor detected.
    Unknown,
}

impl HypervisorType {
    /// Returns the string name of the hypervisor.
    #[must_use]
    pub fn as_str(&self) -> &'static str {
        match self {
            Self::VMware => "vmware",
            Self::Qemu => "qemu",
            Self::VirtualBox => "virtualbox",
            Self::HyperV => "hyper-v",
            Self::Bhyve => "bhyve",
            Self::Xen => "xen",
            Self::Parallels => "parallels",
            Self::Unknown => "unknown",
        }
    }
}

/// Detect virtualization/hypervisor on OpenBSD.
///
/// # Examples
///
/// ```rust,no_run
/// use probe_runtime::platform::openbsd::detect_virtualization;
///
/// if let Some(hypervisor) = detect_virtualization() {
///     println!("Running on: {}", hypervisor.as_str());
/// }
/// ```
#[must_use]
pub fn detect_virtualization() -> Option<HypervisorType> {
    // Try multiple detection methods
    detect_via_sysctl().or_else(detect_via_dmesg).or_else(detect_via_device)
}

/// Detect virtualization using sysctl hw.* variables.
#[must_use]
fn detect_via_sysctl() -> Option<HypervisorType> {
    // Check hw.product
    if let Some(product) = sysctl_string("hw.product") {
        let product_lower = product.to_lowercase();
        if product_lower.contains("vmware") || product_lower.contains("vmw") {
            return Some(HypervisorType::VMware);
        }
        if product_lower.contains("virtualbox") || product_lower.contains("vbox") {
            return Some(HypervisorType::VirtualBox);
        }
        if product_lower.contains("qemu") {
            return Some(HypervisorType::Qemu);
        }
        if product_lower.contains("bhyve") {
            return Some(HypervisorType::Bhyve);
        }
        if product_lower.contains("parallels") {
            return Some(HypervisorType::Parallels);
        }
        if product_lower.contains("xen") {
            return Some(HypervisorType::Xen);
        }
    }

    // Check hw.vendor
    if let Some(vendor) = sysctl_string("hw.vendor") {
        let vendor_lower = vendor.to_lowercase();
        if vendor_lower.contains("vmware") {
            return Some(HypervisorType::VMware);
        }
        if vendor_lower.contains("qemu") {
            return Some(HypervisorType::Qemu);
        }
        if vendor_lower.contains("microsoft") && vendor_lower.contains("virtual") {
            return Some(HypervisorType::HyperV);
        }
        if vendor_lower.contains("innotek") || vendor_lower.contains("oracle") {
            // VirtualBox was originally by Innotek, now Oracle
            if let Some(product) = sysctl_string("hw.product") {
                if product.to_lowercase().contains("virtualbox") {
                    return Some(HypervisorType::VirtualBox);
                }
            }
        }
    }

    None
}

/// Detect virtualization by parsing dmesg output.
#[must_use]
fn detect_via_dmesg() -> Option<HypervisorType> {
    // Try /var/run/dmesg.boot first (persisted across reboots)
    if let Ok(dmesg) = fs::read_to_string("/var/run/dmesg.boot") {
        return parse_dmesg(&dmesg);
    }

    // Fall back to dmesg command
    if let Ok(output) = Command::new("dmesg").output() {
        if output.status.success() {
            let dmesg = String::from_utf8_lossy(&output.stdout);
            return parse_dmesg(&dmesg);
        }
    }

    None
}

/// Parse dmesg output for hypervisor signatures.
#[must_use]
fn parse_dmesg(dmesg: &str) -> Option<HypervisorType> {
    let dmesg_lower = dmesg.to_lowercase();

    // Check for hypervisor signatures in dmesg
    if dmesg_lower.contains("vmware") || dmesg_lower.contains("vmw") {
        return Some(HypervisorType::VMware);
    }
    if dmesg_lower.contains("qemu") {
        return Some(HypervisorType::Qemu);
    }
    if dmesg_lower.contains("virtualbox") || dmesg_lower.contains("vbox") {
        return Some(HypervisorType::VirtualBox);
    }
    if dmesg_lower.contains("hyper-v") || dmesg_lower.contains("hyperv") {
        return Some(HypervisorType::HyperV);
    }
    if dmesg_lower.contains("bhyve") {
        return Some(HypervisorType::Bhyve);
    }
    if dmesg_lower.contains("xen") {
        return Some(HypervisorType::Xen);
    }
    if dmesg_lower.contains("parallels") {
        return Some(HypervisorType::Parallels);
    }

    // Check for specific device drivers
    if dmesg.contains("vmt0") {
        // VMware Tools device
        return Some(HypervisorType::VMware);
    }
    if dmesg.contains("vboxguest") {
        return Some(HypervisorType::VirtualBox);
    }
    if dmesg.contains("hvn") || dmesg.contains("vmbus") {
        return Some(HypervisorType::HyperV);
    }

    None
}

/// Detect virtualization by checking for specific device files or drivers.
#[must_use]
fn detect_via_device() -> Option<HypervisorType> {
    // On OpenBSD, VMware typically shows up as vmt0 network device
    if device_exists("vmt0") {
        return Some(HypervisorType::VMware);
    }

    None
}

/// Check if a device exists (crude check via dmesg or ifconfig).
fn device_exists(device_name: &str) -> bool {
    // Try ifconfig to see if network device exists
    if let Ok(output) = Command::new("ifconfig").arg(device_name).output() {
        if output.status.success() {
            return true;
        }
    }

    false
}

/// Get a sysctl string value.
///
/// # Errors
///
/// Returns `None` if the sysctl command fails or the value cannot be parsed.
#[must_use]
fn sysctl_string(name: &str) -> Option<String> {
    let output = Command::new("sysctl").args(["-n", name]).output().ok()?;

    if output.status.success() {
        let value = String::from_utf8_lossy(&output.stdout);
        let trimmed = value.trim();
        if !trimmed.is_empty() {
            return Some(trimmed.to_string());
        }
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_hypervisor_type_as_str() {
        assert_eq!(HypervisorType::VMware.as_str(), "vmware");
        assert_eq!(HypervisorType::Qemu.as_str(), "qemu");
        assert_eq!(HypervisorType::VirtualBox.as_str(), "virtualbox");
        assert_eq!(HypervisorType::HyperV.as_str(), "hyper-v");
        assert_eq!(HypervisorType::Bhyve.as_str(), "bhyve");
    }

    #[test]
    fn test_detect_virtualization() {
        // This test only makes sense on OpenBSD
        #[cfg(target_os = "openbsd")]
        {
            let _ = detect_virtualization();
        }
    }

    #[test]
    fn test_parse_dmesg() {
        let vmware_dmesg = "vmt0 at mainbus0\nOpenBSD 7.4 (GENERIC) #1397: Mon Oct 9";
        assert_eq!(parse_dmesg(vmware_dmesg), Some(HypervisorType::VMware));

        let qemu_dmesg = "cpu0: QEMU Virtual CPU version 2.5+";
        assert_eq!(parse_dmesg(qemu_dmesg), Some(HypervisorType::Qemu));

        let bare_metal = "cpu0: Intel(R) Core(TM) i7-9750H CPU @ 2.60GHz";
        assert_eq!(parse_dmesg(bare_metal), None);
    }

    #[test]
    fn test_sysctl_string() {
        // On OpenBSD, hw.model should always exist
        #[cfg(target_os = "openbsd")]
        {
            let model = sysctl_string("hw.model");
            // Should return something, even if not virtualized
            assert!(model.is_some());
        }
    }
}
