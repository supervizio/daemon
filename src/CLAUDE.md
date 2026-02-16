<!-- updated: 2026-02-15T21:30:00Z -->
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
├── lib/probe/                # Rust system metrics library (CGO)
├── api/proto/                # gRPC protobuf definitions
├── go.mod                    # Module github.com/kodflow/daemon
└── go.sum                    # Dependency checksums
```

## Go Module

```go
module github.com/kodflow/daemon
go 1.25.6
```

## Dependencies

| Package | Usage |
|---------|-------|
| `gopkg.in/yaml.v3` | Configuration parsing |
| `github.com/stretchr/testify` | Unit tests (assert, require) |
| `github.com/charmbracelet/bubbletea` | Interactive TUI framework |
| `github.com/google/wire` | Compile-time dependency injection |
| `go.etcd.io/bbolt` | Embedded key-value storage |
| `google.golang.org/grpc` | gRPC server and client |
| `google.golang.org/protobuf` | Protocol buffer support |

## Commands

```bash
go build ./cmd/daemon         # Build binary
go test -race ./...           # Tests with race detection
golangci-lint run -c ../.golangci.yml   # Standard linting
ktn-linter lint -c ../.ktn-linter.yaml ./...  # KTN convention linting
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
│  observability (healthcheck, logging), probe (metrics, CGO) │
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
| `lib/` | Rust libraries | `lib/CLAUDE.md` |
| `api/proto/` | gRPC definitions | `api/proto/v1/daemon/CLAUDE.md` |
