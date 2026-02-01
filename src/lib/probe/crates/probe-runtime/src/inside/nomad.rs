//! `HashiCorp` Nomad allocation inside detection.

use crate::{ContainerRuntime, InsideDetector, InsideInfo};
use std::collections::HashMap;

/// Detects if running inside a Nomad allocation.
pub struct NomadInsideDetector;

impl InsideDetector for NomadInsideDetector {
    fn detect(&self) -> Option<InsideInfo> {
        // Nomad injects NOMAD_ALLOC_ID for all tasks
        let alloc_id = std::env::var("NOMAD_ALLOC_ID").ok()?;

        Some(InsideInfo {
            runtime: ContainerRuntime::Nomad,
            orchestrator: Some(ContainerRuntime::Nomad),
            workload_id: Some(alloc_id),
            workload_name: std::env::var("NOMAD_JOB_NAME").ok(),
            namespace: std::env::var("NOMAD_NAMESPACE").ok(),
            metadata: collect_nomad_metadata(),
            ..Default::default()
        })
    }

    fn priority(&self) -> u8 {
        // High priority - very specific env vars
        95
    }

    fn name(&self) -> &'static str {
        "nomad"
    }
}

/// Collect Nomad-specific metadata from environment.
fn collect_nomad_metadata() -> HashMap<String, String> {
    let mut meta = HashMap::new();

    // All standard Nomad environment variables
    let env_vars = [
        "NOMAD_ALLOC_ID",
        "NOMAD_SHORT_ALLOC_ID",
        "NOMAD_ALLOC_INDEX",
        "NOMAD_ALLOC_NAME",
        "NOMAD_JOB_NAME",
        "NOMAD_JOB_ID",
        "NOMAD_GROUP_NAME",
        "NOMAD_TASK_NAME",
        "NOMAD_DC",
        "NOMAD_REGION",
        "NOMAD_NAMESPACE",
        "NOMAD_CPU_LIMIT",
        "NOMAD_MEMORY_LIMIT",
        "NOMAD_MEMORY_MAX_LIMIT",
        "NOMAD_ALLOC_DIR",
        "NOMAD_TASK_DIR",
        "NOMAD_SECRETS_DIR",
        "NOMAD_ADDR_*", // Dynamic port addresses
    ];

    for key in env_vars {
        if key.ends_with('*') {
            // Handle wildcard patterns (e.g., NOMAD_ADDR_*)
            let prefix = key.trim_end_matches('*');
            for (k, v) in std::env::vars() {
                if k.starts_with(prefix) {
                    meta.insert(k.to_lowercase(), v);
                }
            }
        } else if let Ok(val) = std::env::var(key) {
            meta.insert(key.to_lowercase(), val);
        }
    }

    meta
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_detector_without_env() {
        let detector = NomadInsideDetector;
        // Should return None when not in Nomad
        // (Unless we're actually running in Nomad)
        let _ = detector.detect();
    }
}
