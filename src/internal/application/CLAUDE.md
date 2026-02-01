# Application Layer

Use case implementations coordinating domain and infrastructure.

## Structure

```
application/
├── config/       # Configuration port interface
├── health/       # Service health monitoring (renamed from healthcheck/)
├── lifecycle/    # Process lifecycle management (renamed from process/)
├── metrics/      # Process metrics tracking
└── supervisor/   # Service orchestration
```

## Packages

| Package | Role | See |
|---------|------|-----|
| `config` | Loader interface (port) | `config/CLAUDE.md` |
| `health` | ProbeMonitor coordinates service health checks | `health/CLAUDE.md` |
| `lifecycle` | Manager handles process lifecycle with restart | `lifecycle/CLAUDE.md` |
| `metrics` | Tracker monitors process CPU/memory metrics | `metrics/CLAUDE.md` |
| `supervisor` | Supervisor orchestrates multiple services | `supervisor/CLAUDE.md` |

## Terminology

| Term | Description |
|------|-------------|
| **health** | Service reachability monitoring via probes (TCP, HTTP, etc.) |
| **metrics** | Process metrics collection (CPU, RAM per process) |
| **lifecycle** | Process lifecycle management (start, stop, restart) |

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

## Data Flow

```
Supervisor
    │
    ├── Manager (per service)
    │       │
    │       └── Executor.Start()    [domain port]
    │
    ├── ProbeMonitor
    │       │
    │       └── Prober.Probe()      [domain port]
    │
    └── Tracker
            │
            └── Collector.Collect() [application port]
```

## Port Interfaces

Application layer defines these ports for infrastructure to implement:

| Port | Package | Implemented By |
|------|---------|----------------|
| `Loader` | `config` | `infrastructure/persistence/config/yaml` |
| `Creator` | `health` | `infrastructure/observability/healthcheck` |
| `Collector` | `metrics` | `infrastructure/probe` |
