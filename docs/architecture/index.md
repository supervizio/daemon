# Architecture

superviz.io implements **hexagonal architecture** (ports & adapters) with domain-driven design principles. The codebase enforces strict dependency rules: the domain layer has zero external dependencies, the application layer depends only on domain, and infrastructure implements domain ports.

---

## C4 Context Diagram

```mermaid
C4Context
    title System Context - superviz.io

    Person(admin, "System Administrator", "Manages services and monitors health")
    Person(dev, "Developer", "Builds and tests services")

    System(daemon, "superviz.io Daemon", "Process supervisor with health monitoring, metrics collection, and gRPC API")

    System_Ext(docker, "Docker Engine", "Container runtime")
    System_Ext(k8s, "Kubernetes", "Container orchestration")
    System_Ext(systemd, "systemd", "Linux init system")
    System_Ext(nomad, "HashiCorp Nomad", "Workload orchestrator")

    Rel(admin, daemon, "Controls via gRPC/TUI")
    Rel(dev, daemon, "Configures via YAML")
    Rel(daemon, docker, "Discovers containers")
    Rel(daemon, k8s, "Discovers pods/services")
    Rel(daemon, systemd, "Discovers services")
    Rel(daemon, nomad, "Discovers allocations")
```

---

## C4 Container Diagram

```mermaid
C4Container
    title Container Diagram - superviz.io

    Person(client, "Client", "gRPC client or CLI user")

    System_Boundary(daemon, "superviz.io Daemon") {
        Container(cli, "CLI Entry Point", "Go", "Parses flags, starts bootstrap")
        Container(bootstrap, "Bootstrap", "Go + Wire", "Dependency injection, signal handling")
        Container(supervisor, "Supervisor", "Go", "Service orchestration, lifecycle management")
        Container(grpc_server, "gRPC Server", "Go + gRPC", "API with streaming support")
        Container(tui, "Terminal UI", "Go + Bubble Tea", "Raw mode MOTD and interactive dashboard")
        Container(probe, "System Probe", "Go + Rust FFI", "Cross-platform metrics collection")
        ContainerDb(boltdb, "BoltDB", "Embedded KV Store", "Metrics persistence")
    }

    System_Ext(os, "Operating System", "Process management, /proc, /sys")

    Rel(client, grpc_server, "gRPC", "protobuf")
    Rel(client, tui, "Terminal", "ANSI")
    Rel(cli, bootstrap, "Initializes")
    Rel(bootstrap, supervisor, "Creates via Wire DI")
    Rel(supervisor, grpc_server, "Provides state/metrics")
    Rel(supervisor, probe, "Collects metrics")
    Rel(probe, os, "Reads /proc, /sys, syscalls")
    Rel(supervisor, boltdb, "Persists metrics")
```

---

## Layer Overview

### Domain Layer (Pure Business Logic)

Zero external dependencies. Defines entities, value objects, and port interfaces.

| Package | Purpose |
|---------|---------|
| `process` | Process `Spec`, `State`, `Executor` port, `ExitResult` |
| `config` | `ServiceConfig`, `RestartConfig`, `ProbeConfig` value objects |
| `health` | `Status`, `Result`, `Prober` port, `AggregatedHealth` |
| `lifecycle` | `Event` types, `DaemonState`, `Reaper` port |
| `metrics` | `SystemCPU`, `SystemMemory`, `ProcessMetrics` types |
| `logging` | `LogEvent`, `LogLevel`, `Logger`/`Writer` ports |
| `shared` | `Duration`, `Size`, `Clock` value objects |
| `storage` | `MetricsStore` port interface |
| `target` | `ExternalTarget`, `Discoverer`/`Watcher` ports |

### Application Layer (Use Cases)

Orchestrates domain entities through port interfaces.

| Package | Purpose |
|---------|---------|
| `supervisor` | Main service orchestrator, coordinates all managers |
| `lifecycle` | Per-service process lifecycle with restart logic |
| `health` | `ProbeMonitor` coordinates multi-protocol health checks |
| `metrics` | `Tracker` monitors per-process CPU/memory |
| `monitoring` | `ExternalMonitor` for unmanaged targets |
| `config` | `Loader`/`Reloader` port interfaces |

### Infrastructure Layer (Adapters)

Implements domain ports with concrete technologies.

| Package | Implements | Technology |
|---------|-----------|------------|
| `process/executor` | `domain.Executor` | `exec.Cmd`, signals |
| `process/reaper` | `domain.Reaper` | `waitpid` (PID1 mode) |
| `persistence/config/yaml` | `app.Loader` | `gopkg.in/yaml.v3` |
| `persistence/storage/boltdb` | `domain.MetricsStore` | `go.etcd.io/bbolt` |
| `observability/healthcheck` | `domain.Prober` | TCP, HTTP, gRPC, ICMP, exec |
| `probe` | `app.Collector` | Rust FFI via CGO |
| `discovery` | `domain.Discoverer` | Docker, systemd, K8s APIs |
| `transport/grpc` | gRPC server | `google.golang.org/grpc` |
| `transport/tui` | Terminal UI | `charmbracelet/bubbletea` |

---

## Dependency Rules

```mermaid
graph TD
    CMD["cmd/daemon"] --> BOOT["bootstrap (Wire DI)"]
    BOOT --> APP["Application Layer"]
    APP --> DOM["Domain Layer"]
    INFRA["Infrastructure Layer"] -.->|implements ports| DOM

    style CMD fill:#6c76931a,stroke:#6c7693,color:#d4d8e0
    style BOOT fill:#fbdf411a,stroke:#fbdf41,color:#d4d8e0
    style APP fill:#df41fb1a,stroke:#df41fb,color:#d4d8e0
    style DOM fill:#41fbdf1a,stroke:#41fbdf,color:#d4d8e0
    style INFRA fill:#fbdf411a,stroke:#fbdf41,color:#d4d8e0
```

- Application depends on Domain (never reverse)
- Infrastructure implements Domain ports (dependency inversion)
- No circular dependencies between packages
- Bootstrap wires everything together via Google Wire
