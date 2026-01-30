//! CLI tool detection for container runtimes.

use crate::{AvailableDetector, AvailableRuntime, ContainerRuntime};
use std::process::Command;

/// CLI tools to check with their version arguments.
const CLI_TOOLS: &[(&str, ContainerRuntime, &[&str])] = &[
    ("docker", ContainerRuntime::Docker, &["--version"]),
    ("podman", ContainerRuntime::Podman, &["--version"]),
    ("nerdctl", ContainerRuntime::Containerd, &["--version"]),
    ("crictl", ContainerRuntime::CriO, &["--version"]),
    ("lxc-info", ContainerRuntime::Lxc, &["--version"]),
    ("lxc", ContainerRuntime::Lxd, &["--version"]),
];

/// Detects available runtimes via CLI tools in PATH.
pub struct CliDetector;

impl AvailableDetector for CliDetector {
    fn detect(&self) -> Vec<AvailableRuntime> {
        let mut available = Vec::new();

        for (cmd, runtime, args) in CLI_TOOLS {
            if let Some(version) = check_cli_tool(cmd, args) {
                available.push(AvailableRuntime {
                    runtime: *runtime,
                    version: Some(version),
                    is_running: true, // If CLI works, it's "running"
                    ..Default::default()
                });
            }
        }

        available
    }

    fn name(&self) -> &'static str {
        "cli"
    }
}

/// Check if a CLI tool is available and get its version.
fn check_cli_tool(cmd: &str, args: &[&str]) -> Option<String> {
    let output = Command::new(cmd).args(args).output().ok()?;

    if output.status.success() {
        let stdout = String::from_utf8_lossy(&output.stdout);
        // Extract first line as version
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
    fn test_cli_detector() {
        let detector = CliDetector;
        let available = detector.detect();
        // Returns whatever CLIs are available on the system
        for runtime in &available {
            assert!(runtime.version.is_some());
        }
    }
}
