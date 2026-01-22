# Source Code - superviz.io

Go source code for the process supervisor using hexagonal architecture.

## Structure

```
src/
├── cmd/daemon/               # CLI entry point (→ bootstrap.Run())
├── internal/                 # Private internal packages
│   ├── bootstrap/            # Wire DI, app lifecycle, signals
│   ├── application/          # Application layer (use cases)
│   ├── domain/               # Domain layer (entities, ports)
│   └── infrastructure/       # Infrastructure layer (adapters)
├── api/proto/                # gRPC protobuf definitions
├── go.mod                    # Module github.com/kodflow/daemon
└── go.sum                    # Dependency checksums
```

## Go Module

```go
module github.com/kodflow/daemon
go 1.25
```

## Dependencies

| Package | Usage |
|---------|-------|
| `gopkg.in/yaml.v3` | Configuration parsing |
| `github.com/stretchr/testify` | Unit tests (assert, require) |

## Commands

```bash
go build ./cmd/daemon         # Build binary
go test -race ./...           # Tests with race detection
golangci-lint run             # Standard linting
ktn-linter lint ./...         # KTN convention linting
```

## Hexagonal Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Application Layer                       │
│  (supervisor, process manager, health monitor, config port) │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                        Domain Layer                          │
│   (entities, value objects, domain ports/interfaces)        │
└─────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌─────────────────────────────────────────────────────────────┐
│                    Infrastructure Layer                      │
│  persistence (yaml, boltdb), process (exec, signals, reaper)│
│  observability (healthcheck, logging), resources (metrics)  │
└─────────────────────────────────────────────────────────────┘
```

## Data Flow

```
main.go
    ↓
config.Loader.Load()           # Infrastructure: YAML parsing
    ↓
Supervisor.New(cfg)            # Application: orchestration
    ↓
ProcessManager per service     # Application: process lifecycle
    ↓
Executor.Start()               # Infrastructure: exec.Cmd
    ↓
HealthMonitor                  # Application: health coordination
    ↓
Checker.Check()                # Infrastructure: HTTP/TCP/cmd
```

## Related Directories

| Directory | Description | See |
|-----------|-------------|-----|
| `cmd/` | Entry point | `cmd/CLAUDE.md` |
| `internal/` | Core logic | `internal/CLAUDE.md` |
| `api/proto/` | gRPC definitions | `api/proto/v1/daemon/CLAUDE.md` |
