# probe-runtime - Universal Runtime Detection

Cross-platform detection of container runtimes and orchestrators.

## Role

Detect container/orchestrator environments through:
- **Inside detection**: Determine if running inside a container
- **Available detection**: Find what runtimes are available on the host

## Structure

```
src/
├── lib.rs              # Types: ContainerRuntime, RuntimeInfo, traits
├── detector.rs         # UniversalRuntimeDetector main entry point
├── inside/             # Inside detection (am I in a container?)
│   ├── mod.rs          # all_detectors()
│   ├── docker.rs       # /.dockerenv, cgroup
│   ├── podman.rs       # /run/.containerenv
│   ├── kubernetes.rs   # KUBERNETES_SERVICE_HOST, service account
│   ├── nomad.rs        # NOMAD_ALLOC_ID, env vars
│   ├── lxc.rs          # /proc/1/environ, cgroup
│   ├── containerd.rs   # cri-containerd- cgroup
│   ├── crio.rs         # crio- cgroup
│   ├── cloud.rs        # AWS ECS/Fargate, GKE, AKS
│   ├── swarm.rs        # Docker Swarm
│   ├── systemd_nspawn.rs # machine.slice, /run/host/
│   └── freebsd_jail.rs # sysctl security.jail.jailed (FreeBSD only)
├── available/          # Available detection (what's on host?)
│   ├── mod.rs          # all_detectors()
│   ├── sockets.rs      # Unix socket detection
│   ├── cli.rs          # CLI tool detection
│   ├── kubernetes.rs   # kubeconfig detection
│   └── nomad.rs        # NOMAD_ADDR, nomad CLI
└── platform/           # Platform utilities
    ├── mod.rs
    ├── linux.rs        # cgroup detection
    ├── freebsd.rs      # jail utilities
    └── darwin.rs       # Docker Desktop, Colima
```

## Usage

```rust
use probe_runtime::detector::{detect, is_containerized};

// Full detection
let info = detect();
if info.is_containerized {
    println!("Running in: {}", info.container_runtime.unwrap());
    if let Some(id) = info.container_id {
        println!("Container ID: {}", id);
    }
}

// Available runtimes on host
for rt in &info.available_runtimes {
    println!("Available: {} ({})", rt.runtime, rt.is_running);
}

// Fast check
if is_containerized() {
    println!("In container!");
}
```

## Supported Runtimes

| Runtime | Inside Detection | Available Detection |
|---------|-----------------|---------------------|
| Docker | /.dockerenv, cgroup | /var/run/docker.sock |
| Podman | /run/.containerenv | /run/podman/podman.sock |
| Kubernetes | KUBERNETES_SERVICE_HOST | ~/.kube/config |
| Nomad | NOMAD_ALLOC_ID | NOMAD_ADDR, CLI |
| LXC/LXD | /proc/1/environ, cgroup | /var/lib/lxd/unix.socket |
| containerd | cri-containerd- cgroup | /run/containerd/*.sock |
| CRI-O | crio- cgroup | /run/crio/crio.sock |
| AWS ECS | ECS_CONTAINER_METADATA_URI | ECS_AGENT_URI |
| AWS Fargate | AWS_EXECUTION_ENV | (same as ECS) |
| GKE | K8s + GOOGLE_CLOUD_PROJECT | gcloud CLI |
| AKS | K8s + AZURE_* | az CLI |
| Docker Swarm | DOCKER_SWARM_* | docker info --swarm |
| systemd-nspawn | /run/host/, machine.slice | machinectl |
| FreeBSD Jail | sysctl jail.jailed | jls command |

## Platform Support

Linux (full), macOS (VM-based*), FreeBSD (jail). *macOS containers run in Linux VMs.

## Related

| Location | Description |
|----------|-------------|
| `probe-ffi/src/lib.rs` | FFI exports |
| `include/probe.h` | C header |
| `infrastructure/probe/runtime.go` | Go bindings |
