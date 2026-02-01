//! Nomad configuration detection.

use crate::{AvailableDetector, AvailableRuntime, ContainerRuntime};
use std::process::Command;

/// Detects Nomad availability via environment and CLI.
pub struct NomadAvailableDetector;

impl AvailableDetector for NomadAvailableDetector {
    fn detect(&self) -> Vec<AvailableRuntime> {
        let mut available = Vec::new();

        // Check NOMAD_ADDR environment variable
        if let Ok(addr) = std::env::var("NOMAD_ADDR") {
            available.push(AvailableRuntime {
                runtime: ContainerRuntime::Nomad,
                api_endpoint: Some(addr),
                is_running: true,
                version: check_nomad_cli(),
                ..Default::default()
            });
            return available;
        }

        // Check nomad CLI
        if let Some(version) = check_nomad_cli() {
            available.push(AvailableRuntime {
                runtime: ContainerRuntime::Nomad,
                version: Some(version),
                is_running: true,
                ..Default::default()
            });
        }

        available
    }

    fn name(&self) -> &'static str {
        "nomad"
    }
}

/// Check nomad CLI version.
fn check_nomad_cli() -> Option<String> {
    let output = Command::new("nomad").args(["--version"]).output().ok()?;

    if output.status.success() {
        let stdout = String::from_utf8_lossy(&output.stdout);
        // Format: "Nomad v1.6.0 (..."
        let version = stdout.lines().next()?.trim().to_string();
        if !version.is_empty() {
            return Some(version);
        }
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_nomad_detector() {
        let detector = NomadAvailableDetector;
        let available = detector.detect();
        // Returns whatever is available on the system
        let _ = available;
    }
}
