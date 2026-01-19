# Bootstrap - Dependency Injection with Wire

Wire-based dependency injection for the superviz.io daemon.

## Structure

```
bootstrap/
├── app.go           # App struct, Run(), signal handling
├── providers.go     # Custom Wire providers
├── wire.go          # Wire injector (build tag: wireinject)
└── wire_gen.go      # Generated code (DO NOT EDIT)
```

## Files

| File | Purpose |
|------|---------|
| `app.go` | Entry point (`Run()`), App struct, signal handling |
| `providers.go` | Custom providers for complex dependencies |
| `wire.go` | Wire injector declaration (ignored at build) |
| `wire_gen.go` | Auto-generated initialization code |

## Key Types

### App

```go
type App struct {
    Supervisor *appsupervisor.Supervisor
    Cleanup    func()
}
```

Root object holding all wired dependencies.

### Providers

| Provider | Purpose |
|----------|---------|
| `ProvideReaper` | Returns ZombieReaper only if PID 1 |
| `LoadConfig` | Loads config from path via Loader |
| `NewApp` | Creates final App struct |

## Dependency Graph

```
configPath (string)
    │
    ▼
infraconfig.NewLoader() ──────► appconfig.Loader
    │                                │
    ▼                                ▼
LoadConfig() ◄───────────────── service.Config
                                     │
kernel.New() ──────────────────────► │
    │                                │
    ├─► infraprocess.NewUnixExecutorWithKernel()
    │       │
    │       ▼
    │   domain.Executor
    │       │
    └─► ProvideReaper() ─────────────│
            │                        │
            ▼                        ▼
        ZombieReaper            Supervisor
                                     │
                                     ▼
                                   App
```

## Commands

```bash
# Generate wire_gen.go
wire ./internal/bootstrap/

# Or via go generate (if //go:generate directive added)
go generate ./internal/bootstrap/...

# Build (uses wire_gen.go, ignores wire.go)
go build ./cmd/daemon
```

## Wire Build Tags

| File | Build Tag | When Used |
|------|-----------|-----------|
| `wire.go` | `//go:build wireinject` | Only by Wire tool |
| `wire_gen.go` | `//go:build !wireinject` | Normal builds |

## Interface Bindings

| Interface | Implementation |
|-----------|----------------|
| `appconfig.Loader` | `*infraconfig.Loader` |
| `domain.Executor` | `*infraprocess.UnixExecutor` |
| `domainkernel.ZombieReaper` | `*reaper.UnixZombieReaper` (or nil) |

## Usage

```go
// cmd/daemon/main.go
package main

import (
    "os"
    "github.com/kodflow/daemon/internal/bootstrap"
)

func main() {
    os.Exit(bootstrap.Run())
}
```

## Dependencies

- `github.com/google/wire` - Compile-time DI

## Related

| Directory | Relation |
|-----------|----------|
| `../application/supervisor/` | Supervisor dependency |
| `../infrastructure/kernel/` | Kernel, executor, reaper |
| `../infrastructure/persistence/config/yaml/` | Config loader |
| `../../cmd/daemon/` | Entry point using bootstrap |
