# Internal - Core Packages

Private internal packages following hexagonal architecture.

## Structure

```
internal/
├── bootstrap/            # Wire DI, app lifecycle, CLI flags, signals
├── application/          # Application layer (use cases)
│   ├── config/           # Config loader port interface
│   ├── health/           # ProbeMonitor, health orchestration
│   ├── lifecycle/        # Per-service process lifecycle management
│   ├── metrics/          # Process metrics tracking
│   ├── monitoring/       # External target monitoring
│   └── supervisor/       # Service orchestration
├── domain/               # Domain layer (entities, ports)
│   ├── config/           # Configuration value objects
│   ├── health/           # Health status, aggregation, Prober port
│   ├── lifecycle/        # Event types, DaemonState, Reaper port
│   ├── listener/         # Network listener entities
│   ├── logging/          # Log levels, events, Logger/Writer ports
│   ├── metrics/          # System and process metrics types
│   ├── process/          # Process entities, Executor port
│   ├── shared/           # Duration, Size, Clock value objects
│   ├── storage/          # MetricsStore port interface
│   └── target/           # External target entities, Discoverer port
└── infrastructure/       # Infrastructure layer (adapters)
    ├── discovery/        # Target discovery (Docker, systemd, K8s, Nomad, etc.)
    ├── observability/    # healthcheck (probers), logging, events
    ├── persistence/      # config/yaml, storage/boltdb
    ├── probe/            # System metrics & quotas (cross-platform Rust FFI)
    ├── process/          # control, credentials, executor, reaper, signals
    └── transport/        # grpc, tui
```

## Layer Responsibilities

| Layer | Package | Role |
|-------|---------|------|
| Bootstrap | `bootstrap` | Wire DI, App.Run(), signal handling |
| Application | `supervisor` | Service lifecycle orchestration |
| Application | `lifecycle` | Per-service process manager, restart logic |
| Application | `health` | ProbeMonitor - health check coordination |
| Application | `metrics` | Process metrics tracking |
| Application | `config` | Configuration loader port |
| Application | `monitoring` | External target monitoring orchestration |
| Domain | `config` | ServiceConfig, RestartConfig, ProbeConfig |
| Domain | `process` | Spec, State, Executor port, ExitResult |
| Domain | `health` | Status, Result, Prober port, AggregatedHealth |
| Domain | `lifecycle` | Event, DaemonState, Reaper port |
| Domain | `logging` | LogEvent, LogLevel, Logger/Writer ports |
| Domain | `shared` | Duration, Size, Clock value objects |
| Domain | `target` | ExternalTarget, Discoverer/Watcher ports |
| Infrastructure | `discovery` | Discoverer adapters (Docker, systemd, K8s, etc.) |
| Infrastructure | `observability/healthcheck` | TCP, HTTP, gRPC, ICMP, Exec probers |
| Infrastructure | `persistence/config/yaml` | YAML configuration loader |
| Infrastructure | `process/executor` | Unix process execution |
| Infrastructure | `probe` | Cross-platform system metrics & quotas (Rust FFI) |

## Dependency Rules

- Application depends on Domain (never reverse)
- Infrastructure implements Domain ports
- No circular dependencies

## Testing Strategy

- `*_external_test.go`: Black-box tests (package_test)
- `*_internal_test.go`: White-box tests (same package)
- Race detection required (`go test -race`)

## Related Directories

| Directory | See |
|-----------|-----|
| bootstrap | `bootstrap/CLAUDE.md` |
| application | `application/CLAUDE.md` |
| domain | `domain/CLAUDE.md` |
| infrastructure | `infrastructure/CLAUDE.md` |
