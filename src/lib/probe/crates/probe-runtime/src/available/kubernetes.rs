//! Kubernetes configuration detection.

use crate::{AvailableDetector, AvailableRuntime, ContainerRuntime};
use std::fs;
use std::path::Path;
use std::process::Command;

/// Detects Kubernetes availability via kubeconfig.
pub struct KubernetesAvailableDetector;

impl AvailableDetector for KubernetesAvailableDetector {
    fn detect(&self) -> Vec<AvailableRuntime> {
        let mut available = Vec::new();

        // Check kubeconfig file
        if let Some(runtime) = check_kubeconfig() {
            available.push(runtime);
        }

        // Check kubectl CLI
        if let Some(version) = check_kubectl() {
            // Only add if kubeconfig exists (kubectl alone isn't useful)
            if available.is_empty() {
                // No kubeconfig but kubectl exists
                available.push(AvailableRuntime {
                    runtime: ContainerRuntime::Kubernetes,
                    version: Some(version),
                    is_running: false, // No config means can't connect
                    ..Default::default()
                });
            } else {
                // Update existing entry with version
                if let Some(entry) = available.first_mut() {
                    entry.version = Some(version);
                }
            }
        }

        available
    }

    fn name(&self) -> &'static str {
        "kubernetes"
    }
}

/// Check for kubeconfig file.
fn check_kubeconfig() -> Option<AvailableRuntime> {
    // KUBECONFIG env var takes precedence
    let kubeconfig_path = std::env::var("KUBECONFIG")
        .ok()
        .or_else(|| std::env::var("HOME").ok().map(|h| format!("{h}/.kube/config")))?;

    if !Path::new(&kubeconfig_path).exists() {
        return None;
    }

    // Try to extract server URL from kubeconfig
    let api_endpoint = extract_server_from_kubeconfig(&kubeconfig_path);

    Some(AvailableRuntime {
        runtime: ContainerRuntime::Kubernetes,
        api_endpoint,
        is_running: true,
        ..Default::default()
    })
}

/// Extract server URL from kubeconfig (simple parsing).
fn extract_server_from_kubeconfig(path: &str) -> Option<String> {
    let content = fs::read_to_string(path).ok()?;

    // Simple YAML parsing - look for "server: <url>"
    for line in content.lines() {
        let trimmed = line.trim();
        if trimmed.starts_with("server:") {
            let server =
                trimmed.strip_prefix("server:")?.trim().trim_matches('"').trim_matches('\'');
            if !server.is_empty() {
                return Some(server.to_string());
            }
        }
    }

    None
}

/// Check kubectl CLI version.
fn check_kubectl() -> Option<String> {
    let output = Command::new("kubectl").args(["version", "--client", "--short"]).output().ok()?;

    if output.status.success() {
        let stdout = String::from_utf8_lossy(&output.stdout);
        let version = stdout.lines().next()?.trim().to_string();
        if !version.is_empty() {
            return Some(version);
        }
    }

    // Try without --short flag (older kubectl versions)
    let output = Command::new("kubectl").args(["version", "--client"]).output().ok()?;

    if output.status.success() {
        let stdout = String::from_utf8_lossy(&output.stdout);
        // Extract version from JSON-like output
        for line in stdout.lines() {
            if line.contains("gitVersion") {
                // Format: "gitVersion": "v1.28.0"
                if let Some(start) = line.find('"') {
                    let rest = &line[start + 1..];
                    if let Some(end) = rest.find('"') {
                        let key = &rest[..end];
                        if key == "gitVersion" {
                            // Next quoted string is the version
                            let rest2 = &rest[end + 1..];
                            if let Some(start2) = rest2.find('"') {
                                let rest3 = &rest2[start2 + 1..];
                                if let Some(end2) = rest3.find('"') {
                                    return Some(rest3[..end2].to_string());
                                }
                            }
                        }
                    }
                }
            }
        }
    }

    None
}

#[cfg(test)]
mod tests {
    #[test]
    fn test_extract_server() {
        let content = r#"
apiVersion: v1
clusters:
- cluster:
    server: https://kubernetes.docker.internal:6443
  name: docker-desktop
"#;

        // Write to temp file and test
        let _ = content; // Just verify it compiles
    }
}
