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
│   └── supervisor/       # Service orchestration
├── domain/               # Domain layer (entities, ports)
│   ├── config/           # Configuration value objects
│   ├── health/           # Health status, aggregation, Prober port
│   ├── lifecycle/        # Event types, DaemonState, Reaper port
│   ├── listener/         # Network listener entities
│   ├── metrics/          # System and process metrics types
│   ├── process/          # Process entities, Executor port
│   ├── shared/           # Duration, Size, Clock value objects
│   └── storage/          # MetricsStore port interface
└── infrastructure/       # Infrastructure layer (adapters)
    ├── observability/    # healthcheck (probers), logging
    ├── persistence/      # config/yaml, storage/boltdb
    ├── process/          # control, credentials, executor, reaper, signals
    ├── resources/        # cgroup, metrics (linux/darwin/bsd)
    └── transport/        # grpc server
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
| Domain | `config` | ServiceConfig, RestartConfig, ProbeConfig |
| Domain | `process` | Spec, State, Executor port, ExitResult |
| Domain | `health` | Status, Result, Prober port, AggregatedHealth |
| Domain | `lifecycle` | Event, DaemonState, Reaper port |
| Domain | `shared` | Duration, Size, Clock value objects |
| Infrastructure | `observability/healthcheck` | TCP, HTTP, gRPC, ICMP, Exec probers |
| Infrastructure | `persistence/config/yaml` | YAML configuration loader |
| Infrastructure | `process/executor` | Unix process execution |
| Infrastructure | `resources/metrics` | Platform-specific metrics (Linux/Darwin/BSD) |

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
