//! Example demonstrating VM/hypervisor detection on OpenBSD and NetBSD.
//!
//! Usage:
//!   cargo run --example detect_vm

use probe_runtime::detector::detect;

fn main() {
    println!("=== VM/Container Runtime Detection ===\n");

    let info = detect();

    if info.is_containerized {
        println!("✓ Virtualized/Containerized Environment Detected");

        if let Some(runtime) = info.container_runtime {
            println!("  Runtime: {}", runtime);
            println!(
                "  Runtime Type: {}",
                if runtime.is_vm() {
                    "Virtual Machine"
                } else if runtime.is_orchestrator() {
                    "Orchestrator"
                } else {
                    "Container"
                }
            );

            if let Some(container_id) = &info.container_id {
                println!("  Container ID: {}", container_id);
            }

            if let Some(orchestrator) = info.orchestrator {
                println!("  Orchestrator: {}", orchestrator);
            }

            if !info.metadata.is_empty() {
                println!("\n  Metadata:");
                for (key, value) in &info.metadata {
                    println!("    {}: {}", key, value);
                }
            }
        }
    } else {
        println!("✗ Running on Bare Metal (no virtualization detected)");
    }

    if !info.available_runtimes.is_empty() {
        println!("\n=== Available Runtimes on Host ===");
        for runtime in &info.available_runtimes {
            println!("\n  Runtime: {}", runtime.runtime);
            println!("    Running: {}", runtime.is_running);

            if let Some(socket) = &runtime.socket_path {
                println!("    Socket: {}", socket);
            }

            if let Some(version) = &runtime.version {
                println!("    Version: {}", version);
            }
        }
    }

    println!("\n=== Platform Information ===");
    println!("  Platform: {}", probe_runtime::platform::current_platform());
    println!("  Supports cgroups: {}", probe_runtime::platform::supports_cgroups());
    println!("  Supports jails: {}", probe_runtime::platform::supports_jails());
    println!(
        "  Supports hypervisor detection: {}",
        probe_runtime::platform::supports_hypervisor_detection()
    );
}
