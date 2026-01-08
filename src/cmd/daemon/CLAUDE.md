# superviz.io CLI - Main Entry Point

Main binary for the process supervisor.

## Structure

```
daemon/
└── main.go    # Single entry point
```

## main.go

### Responsibilities

1. Parse CLI arguments
2. Create infrastructure components (config loader, executor)
3. Load YAML configuration
4. Initialize Supervisor with dependencies
5. Set up signal handlers
6. Manage main loop

### Dependency Injection

```go
// Infrastructure layer
loader := infraconfig.NewLoader()
executor := infraprocess.NewUnixExecutor()
reaper := kernel.Default.Reaper

// Application layer
cfg, _ := loader.Load(configPath)
sup, _ := appsupervisor.NewSupervisor(cfg, loader, executor, reaper)
```

### Signal Handling

```go
sigCh := make(chan os.Signal, 1)
signal.Notify(sigCh, syscall.SIGTERM, syscall.SIGINT, syscall.SIGHUP)

for sig := range sigCh {
    switch sig {
    case syscall.SIGHUP:
        sup.Reload()
    case syscall.SIGTERM, syscall.SIGINT:
        sup.Stop()
        return
    }
}
```

## Build

```bash
# From src/
go build -o supervizio ./cmd/daemon

# With version
go build -ldflags "-X main.version=1.0.0" -o supervizio ./cmd/daemon
```

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../../internal/infrastructure/config/yaml/` | Config loader |
| `../../internal/infrastructure/process/` | Process executor |
| `../../internal/application/supervisor/` | Supervisor |
| `../../internal/infrastructure/kernel/` | OS abstraction |
