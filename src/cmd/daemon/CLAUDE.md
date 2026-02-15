<!-- updated: 2026-02-15T21:30:00Z -->
# superviz.io CLI - Main Entry Point

Minimal entry point for the process supervisor.

## Structure

```
daemon/
└── main.go    # Single minimal entry point
```

## main.go

The main.go file is intentionally minimal (5 lines of code).
All dependency injection and application logic is handled by the `bootstrap` package.

```go
package main

import (
    "os"
    "github.com/kodflow/daemon/internal/bootstrap"
)

func main() {
    os.Exit(bootstrap.Run())
}
```

### Responsibilities

The `main` function only:
1. Calls `bootstrap.Run()` to start the application
2. Exits with the returned exit code

All other responsibilities are delegated to `internal/bootstrap/`:
- CLI flag parsing (`--config`, `--version`)
- Dependency injection via Wire
- Signal handling (SIGTERM, SIGINT, SIGHUP)
- Supervisor lifecycle management

## Build

```bash
# From src/
go build -o supervizio ./cmd/daemon

# With version
go build -ldflags "-X github.com/kodflow/daemon/internal/bootstrap.version=1.0.0" -o supervizio ./cmd/daemon
```

## CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | YAML config path | `/etc/daemon/config.yaml` |
| `--version` | Show version | - |

## Signal Handling

| Signal | Action |
|--------|--------|
| `SIGTERM`, `SIGINT` | Graceful shutdown |
| `SIGHUP` | Configuration reload |

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../../internal/bootstrap/` | All application logic |
| `../../internal/application/supervisor/` | Supervisor orchestration |
| `../../internal/infrastructure/` | Infrastructure adapters |
