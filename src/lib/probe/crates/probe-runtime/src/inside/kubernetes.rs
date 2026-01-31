//! Kubernetes pod inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::collections::HashMap;
use std::fs;
use std::path::Path;

/// Detects if running inside a Kubernetes pod.
pub struct KubernetesInsideDetector;

impl InsideDetector for KubernetesInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Method 1: Check KUBERNETES_SERVICE_HOST env var (fastest, most reliable)
        if std::env::var("KUBERNETES_SERVICE_HOST").is_ok() {
            let mut info = InsideInfo {
                runtime: ContainerRuntime::Kubernetes,
                orchestrator: Some(ContainerRuntime::Kubernetes),
                namespace: get_namespace(),
                workload_name: std::env::var("POD_NAME").ok(),
                workload_id: std::env::var("POD_UID").ok(),
                container_id: get_container_id_from_cgroup(),
                metadata: collect_k8s_metadata(),
            };

            // Try to get namespace from file if not in env
            if info.namespace.is_none() {
                info.namespace = read_namespace_file();
            }

            return Some(info);
        }

        // Method 2: Check service account token exists
        if Path::new("/var/run/secrets/kubernetes.io/serviceaccount/token").exists() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Kubernetes,
                orchestrator: Some(ContainerRuntime::Kubernetes),
                namespace: read_namespace_file(),
                container_id: get_container_id_from_cgroup(),
                metadata: collect_k8s_metadata(),
                ..Default::default()
            });
        }

        // Method 3: Check cgroup for kubepods pattern
        if check_cgroup_kubepods() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::Kubernetes,
                orchestrator: Some(ContainerRuntime::Kubernetes),
                container_id: get_container_id_from_cgroup(),
                metadata: collect_k8s_metadata(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        // High priority - K8s is specific
        100
    }

    fn name(&self) -> &'static str {
        "kubernetes"
    }
}

/// Get namespace from environment or file.
fn get_namespace() -> Option<String> {
    std::env::var("POD_NAMESPACE")
        .ok()
        .or_else(read_namespace_file)
}

/// Read namespace from service account file.
fn read_namespace_file() -> Option<String> {
    fs::read_to_string("/var/run/secrets/kubernetes.io/serviceaccount/namespace")
        .ok()
        .map(|s| s.trim().to_string())
}

/// Check if cgroup contains kubepods.
fn check_cgroup_kubepods() -> bool {
    if let Ok(content) = fs::read_to_string("/proc/self/cgroup") {
        return content.contains("/kubepods") || content.contains("kubepods-");
    }
    false
}

/// Get container ID from cgroup.
fn get_container_id_from_cgroup() -> Option<String> {
    let content = fs::read_to_string("/proc/self/cgroup").ok()?;

    for line in content.lines() {
        // Patterns for K8s containers:
        // /kubepods/pod<uid>/crio-<id>
        // /kubepods/pod<uid>/<id> (containerd)
        // /kubepods.slice/kubepods-pod<uid>.slice/crio-<id>.scope

        // Extract the last 64-char hex segment
        let segments: Vec<&str> = line.split('/').collect();
        for segment in segments.iter().rev() {
            // Remove common prefixes
            let cleaned = segment
                .trim_start_matches("crio-")
                .trim_start_matches("docker-")
                .trim_start_matches("containerd-")
                .trim_end_matches(".scope");

            // Check if it's a 64-char hex ID
            let id: String = cleaned
                .chars()
                .take_while(|c| c.is_ascii_hexdigit())
                .collect();

            if id.len() == 64 {
                return Some(id);
            }
        }
    }

    None
}

/// Collect Kubernetes-specific metadata.
fn collect_k8s_metadata() -> HashMap<String, String> {
    let mut meta = HashMap::new();

    // Downward API environment variables
    let env_vars = [
        ("POD_NAME", "pod_name"),
        ("POD_NAMESPACE", "namespace"),
        ("POD_IP", "pod_ip"),
        ("POD_UID", "pod_uid"),
        ("NODE_NAME", "node_name"),
        ("SERVICE_ACCOUNT", "service_account"),
        ("KUBERNETES_SERVICE_HOST", "api_host"),
        ("KUBERNETES_SERVICE_PORT", "api_port"),
    ];

    for (env_key, meta_key) in env_vars {
        if let Ok(val) = std::env::var(env_key) {
            meta.insert(meta_key.to_string(), val);
        }
    }

    // Try to read Downward API files
    let files = [
        ("/etc/podinfo/labels", "labels"),
        ("/etc/podinfo/annotations", "annotations"),
    ];

    for (path, key) in files {
        if let Ok(content) = fs::read_to_string(path) {
            meta.insert(key.to_string(), content.trim().to_string());
        }
    }

    meta
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_extract_container_id() {
        // This test just verifies the function doesn't panic
        let _ = get_container_id_from_cgroup();
    }
}
