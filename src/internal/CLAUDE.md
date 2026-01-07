# Internal - Core Packages

Private internal packages following hexagonal architecture.

## Structure

```
internal/
├── application/          # Application layer (use cases)
│   ├── config/           # Configuration port interface
│   ├── health/           # Health monitoring orchestration
│   ├── process/          # Process lifecycle management
│   └── supervisor/       # Service orchestration
├── domain/               # Domain layer (entities, ports)
│   ├── health/           # Health check entities
│   ├── process/          # Process entities and ports
│   ├── service/          # Service configuration entities
│   └── shared/           # Shared value objects (Duration, Size)
└── infrastructure/       # Infrastructure layer (adapters)
    ├── config/yaml/      # YAML configuration loader
    ├── health/           # Health check adapters (HTTP, TCP, cmd)
    ├── kernel/           # OS abstraction layer
    │   ├── adapters/     # Platform-specific implementations
    │   └── ports/        # Kernel interfaces
    ├── logging/          # Log management (writers, capture, rotation)
    └── process/          # Process executor adapter
```

## Layer Responsibilities

| Layer | Package | Role |
|-------|---------|------|
| Application | `supervisor` | Service lifecycle orchestration |
| Application | `process` | Process manager with restart logic |
| Application | `health` | Health monitor coordination |
| Application | `config` | Configuration port interface |
| Domain | `service` | Service configuration entities |
| Domain | `process` | Process entities, states, events |
| Domain | `health` | Health status, results, events |
| Domain | `shared` | Duration, Size value objects |
| Infrastructure | `config/yaml` | YAML file parsing |
| Infrastructure | `health` | HTTP, TCP, command checkers |
| Infrastructure | `kernel` | OS abstraction (signals, reaper) |
| Infrastructure | `logging` | Writers, capture, rotation |
| Infrastructure | `process` | Unix process executor |

## Dependency Rules

```
application ──→ domain ←── infrastructure
     │              │           │
     │              │           ├── config/yaml
     │              │           ├── health
     │              │           ├── kernel
     │              │           ├── logging
     │              │           └── process
     │              │
     └──────────────┘
```

**Rules:**
- Application depends on Domain (never reverse)
- Infrastructure implements Domain ports
- Application ports (config.Loader) = bootstrap/orchestration concerns
- Domain ports (Executor, Checker) = business needs
- No circular dependencies

## Testing Strategy

- `*_external_test.go`: Black-box tests (package_test)
- `*_internal_test.go`: White-box tests (same package)
- All public functions must have external tests
- Race detection required (`go test -race`)

## Related Directories

| Directory | See |
|-----------|-----|
| application | `application/CLAUDE.md` |
| domain | `domain/CLAUDE.md` |
| infrastructure | `infrastructure/CLAUDE.md` |
