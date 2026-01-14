# Internal - Core Packages

Private internal packages following hexagonal architecture.

## Structure

```
internal/
├── application/          # Application layer (use cases)
│   ├── config/           # Configuration port interface
│   ├── healthcheck/      # Health monitoring with ProbeMonitor
│   ├── process/          # Process lifecycle management
│   └── supervisor/       # Service orchestration
├── domain/               # Domain layer (entities, ports)
│   ├── health/           # Health status, aggregation
│   ├── listener/         # Network listener entities
│   ├── probe/            # Prober port, Target, Result, Config
│   ├── process/          # Process entities and ports
│   ├── service/          # Service configuration entities
│   └── shared/           # Shared value objects (Duration, Size)
└── infrastructure/       # Infrastructure layer (adapters)
    ├── config/yaml/      # YAML configuration loader
    ├── kernel/           # OS abstraction layer
    │   ├── adapters/     # Platform-specific implementations
    │   └── ports/        # Kernel interfaces
    ├── logging/          # Log management (writers, capture, rotation)
    ├── probe/            # Protocol probers (TCP, UDP, HTTP, gRPC, Exec, ICMP)
    └── process/          # Process executor adapter
```

## Layer Responsibilities

| Layer | Package | Role |
|-------|---------|------|
| Application | `supervisor` | Service lifecycle orchestration |
| Application | `process` | Process manager with restart logic |
| Application | `healthcheck` | ProbeMonitor - health orchestration |
| Application | `config` | Configuration port interface |
| Domain | `service` | Service configuration entities |
| Domain | `process` | Process entities, states, events |
| Domain | `health` | Health status, AggregatedHealth |
| Domain | `listener` | Listener entity, state machine |
| Domain | `probe` | Prober port, Target, Result |
| Domain | `shared` | Duration, Size value objects |
| Infrastructure | `config/yaml` | YAML file parsing |
| Infrastructure | `kernel` | OS abstraction (signals, reaper) |
| Infrastructure | `logging` | Writers, capture, rotation |
| Infrastructure | `probe` | TCP, UDP, HTTP, gRPC, Exec, ICMP probers |
| Infrastructure | `process` | Unix process executor |

## Dependency Rules

```
application ──→ domain ←── infrastructure
     │              │           │
     │              │           ├── config/yaml
     │              │           ├── kernel
     │              │           ├── logging
     │              │           ├── probe
     │              │           └── process
     │              │
     └──────────────┘
```

**Rules:**
- Application depends on Domain (never reverse)
- Infrastructure implements Domain ports
- Application ports (config.Loader) = bootstrap/orchestration concerns
- Domain ports (Executor, Prober) = business needs
- No circular dependencies

## Testing Strategy

- `*_external_test.go`: Black-box tests (package_test)
- `*_internal_test.go`: White-box tests (same package)
- All public functions must have external tests
- Race detection required (`go test -race`)

## Security Model

Command execution is centralized in `infrastructure/process.TrustedCommand()`.
Commands come from validated YAML configurations, not user input.
See `infrastructure/CLAUDE.md` for details.

## Related Directories

| Directory | See |
|-----------|-----|
| application | `application/CLAUDE.md` |
| domain | `domain/CLAUDE.md` |
| infrastructure | `infrastructure/CLAUDE.md` |
