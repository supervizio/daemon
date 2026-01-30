//! Cloud provider container detection (AWS ECS/Fargate, GKE, AKS).

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::collections::HashMap;

/// Detects if running in AWS Fargate (serverless ECS).
pub struct AwsFargateDetector;

impl InsideDetector for AwsFargateDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // AWS_EXECUTION_ENV contains "Fargate" for Fargate tasks
        if let Ok(env) = std::env::var("AWS_EXECUTION_ENV") {
            if env.contains("Fargate") {
                return Some(InsideInfo {
                    runtime: ContainerRuntime::AwsFargate,
                    orchestrator: Some(ContainerRuntime::AwsFargate),
                    metadata: collect_ecs_metadata(),
                    ..Default::default()
                });
            }
        }

        None
    }

    fn priority(&self) -> u8 {
        // Highest - most specific cloud detection
        99
    }

    fn name(&self) -> &'static str {
        "aws-fargate"
    }
}

/// Detects if running in AWS ECS (Elastic Container Service).
pub struct AwsEcsDetector;

impl InsideDetector for AwsEcsDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // ECS Task Metadata endpoint v4
        if let Ok(uri) = std::env::var("ECS_CONTAINER_METADATA_URI_V4") {
            let mut metadata = collect_ecs_metadata();
            metadata.insert("metadata_uri_v4".to_string(), uri);

            return Some(InsideInfo {
                runtime: ContainerRuntime::AwsEcs,
                orchestrator: Some(ContainerRuntime::AwsEcs),
                metadata,
                ..Default::default()
            });
        }

        // ECS Task Metadata endpoint v3 (legacy)
        if let Ok(uri) = std::env::var("ECS_CONTAINER_METADATA_URI") {
            let mut metadata = collect_ecs_metadata();
            metadata.insert("metadata_uri".to_string(), uri);

            return Some(InsideInfo {
                runtime: ContainerRuntime::AwsEcs,
                orchestrator: Some(ContainerRuntime::AwsEcs),
                metadata,
                ..Default::default()
            });
        }

        // ECS Agent URI (alternative detection)
        if std::env::var("ECS_AGENT_URI").is_ok() {
            return Some(InsideInfo {
                runtime: ContainerRuntime::AwsEcs,
                orchestrator: Some(ContainerRuntime::AwsEcs),
                metadata: collect_ecs_metadata(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        98
    }

    fn name(&self) -> &'static str {
        "aws-ecs"
    }
}

/// Detects if running in Google GKE (Kubernetes + GKE-specific labels).
pub struct GoogleGkeDetector;

impl InsideDetector for GoogleGkeDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // GKE sets specific environment variables
        let is_gke = std::env::var("KUBERNETES_SERVICE_HOST").is_ok()
            && (std::env::var("GKE_CLUSTER_NAME").is_ok()
                || std::env::var("CLOUDSDK_CORE_PROJECT").is_ok()
                || std::env::var("GOOGLE_CLOUD_PROJECT").is_ok()
                || check_gce_metadata());

        if is_gke {
            return Some(InsideInfo {
                runtime: ContainerRuntime::GoogleGke,
                orchestrator: Some(ContainerRuntime::GoogleGke),
                metadata: collect_gke_metadata(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        // Before generic K8s
        97
    }

    fn name(&self) -> &'static str {
        "google-gke"
    }
}

/// Detects if running in Azure AKS (Kubernetes + AKS-specific markers).
pub struct AzureAksDetector;

impl InsideDetector for AzureAksDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // AKS sets specific environment variables
        let is_aks = std::env::var("KUBERNETES_SERVICE_HOST").is_ok()
            && (std::env::var("AKS_NODE_NAME").is_ok()
                || std::env::var("AZURE_CLIENT_ID").is_ok()
                || check_azure_metadata());

        if is_aks {
            return Some(InsideInfo {
                runtime: ContainerRuntime::AzureAks,
                orchestrator: Some(ContainerRuntime::AzureAks),
                metadata: collect_aks_metadata(),
                ..Default::default()
            });
        }

        None
    }

    fn priority(&self) -> u8 {
        // Before generic K8s
        96
    }

    fn name(&self) -> &'static str {
        "azure-aks"
    }
}

/// Collect AWS ECS/Fargate metadata from environment.
fn collect_ecs_metadata() -> HashMap<String, String> {
    let mut meta = HashMap::new();

    let env_vars = [
        "ECS_CONTAINER_METADATA_URI_V4",
        "ECS_CONTAINER_METADATA_URI",
        "ECS_AGENT_URI",
        "AWS_EXECUTION_ENV",
        "AWS_REGION",
        "AWS_DEFAULT_REGION",
        "ECS_CLUSTER",
        "ECS_ENABLE_TASK_IAM_ROLE",
    ];

    for key in env_vars {
        if let Ok(val) = std::env::var(key) {
            meta.insert(key.to_lowercase(), val);
        }
    }

    meta
}

/// Collect GKE metadata from environment.
fn collect_gke_metadata() -> HashMap<String, String> {
    let mut meta = HashMap::new();

    let env_vars = [
        "GOOGLE_CLOUD_PROJECT",
        "CLOUDSDK_CORE_PROJECT",
        "GKE_CLUSTER_NAME",
        "GKE_NODE_NAME",
        "CLOUDSDK_COMPUTE_ZONE",
        "CLOUDSDK_COMPUTE_REGION",
    ];

    for key in env_vars {
        if let Ok(val) = std::env::var(key) {
            meta.insert(key.to_lowercase(), val);
        }
    }

    meta
}

/// Collect AKS metadata from environment.
fn collect_aks_metadata() -> HashMap<String, String> {
    let mut meta = HashMap::new();

    let env_vars = [
        "AKS_NODE_NAME",
        "AZURE_CLIENT_ID",
        "AZURE_TENANT_ID",
        "AZURE_SUBSCRIPTION_ID",
        "AZURE_RESOURCE_GROUP",
    ];

    for key in env_vars {
        if let Ok(val) = std::env::var(key) {
            meta.insert(key.to_lowercase(), val);
        }
    }

    meta
}

/// Check if GCE metadata server is reachable (indicates GCP environment).
fn check_gce_metadata() -> bool {
    // The GCE metadata server is at 169.254.169.254
    // We just check for the magic header file that GCE injects
    std::path::Path::new("/var/run/secrets/google").exists()
        || std::path::Path::new("/etc/google").exists()
}

/// Check if Azure IMDS is configured (indicates Azure environment).
fn check_azure_metadata() -> bool {
    // Azure Instance Metadata Service markers
    std::path::Path::new("/var/run/secrets/azure").exists()
        || std::path::Path::new("/etc/kubernetes/azure.json").exists()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_detectors_without_cloud_env() {
        // These should return None when not in cloud
        assert!(AwsFargateDetector.detect().is_none());
        assert!(AwsEcsDetector.detect().is_none());
        // GKE and AKS might detect if KUBERNETES_SERVICE_HOST is set
    }
}
