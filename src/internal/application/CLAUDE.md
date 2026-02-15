<!-- updated: 2026-02-15T21:30:00Z -->
# Application Layer

Use case implementations coordinating domain and infrastructure.

## Structure

```
application/
├── config/       # Configuration port interface
├── health/       # Service health monitoring
├── lifecycle/    # Process lifecycle management
├── metrics/      # Process metrics tracking
├── monitoring/   # External target monitoring
└── supervisor/   # Service orchestration
```

## Packages

| Package | Role | See |
|---------|------|-----|
| `config` | Loader interface (port) | `config/CLAUDE.md` |
| `health` | ProbeMonitor coordinates service health checks | `health/CLAUDE.md` |
| `lifecycle` | Manager handles process lifecycle with restart | `lifecycle/CLAUDE.md` |
| `metrics` | Tracker monitors process CPU/memory metrics | `metrics/CLAUDE.md` |
| `monitoring` | ExternalMonitor for unmanaged targets | `monitoring/CLAUDE.md` |
| `supervisor` | Supervisor orchestrates multiple services | `supervisor/CLAUDE.md` |

## Terminology

| Term | Description |
|------|-------------|
| **health** | Service reachability monitoring via probes (TCP, HTTP, etc.) |
| **metrics** | Process metrics collection (CPU, RAM per process) |
| **lifecycle** | Process lifecycle management (start, stop, restart) |
| **monitoring** | External target observation (systemd, Docker, K8s, etc.) |

## Dependencies

- Depends on: `domain`
- Used by: `cmd/daemon`, `bootstrap`
- May use: `infrastructure` (via ports/interfaces)

## Key Types

| Package | Type | Description |
|---------|------|-------------|
| `supervisor` | `Supervisor` | Main service orchestrator |
| `lifecycle` | `Manager` | Per-service process lifecycle management |
| `health` | `ProbeMonitor` | Multi-protocol health probing |
| `metrics` | `Tracker` | Process metrics tracking |
| `config` | `Loader` | Configuration loading interface |
| `config` | `Reloader` | Configuration reloading interface |
| `health` | `Creator` | Prober factory interface |
| `metrics` | `Collector` | Metrics collection interface |
| `monitoring` | `ExternalMonitor` | External target monitoring |

## Data Flow

```
Supervisor
    │
    ├── Manager (per service)
    │       └── Executor.Start()    [domain port]
    │
    ├── ProbeMonitor
    │       └── Prober.Probe()      [domain port]
    │
    ├── ExternalMonitor
    │       └── Discoverer.Discover() [domain port]
    │
    └── Tracker
            └── Collector.Collect() [application port]
```

## Port Interfaces

Application layer defines these ports for infrastructure to implement:

| Port | Package | Implemented By |
|------|---------|----------------|
| `Loader` | `config` | `infrastructure/persistence/config/yaml` |
| `Creator` | `health` | `infrastructure/observability/healthcheck` |
| `Collector` | `metrics` | `infrastructure/probe` |
