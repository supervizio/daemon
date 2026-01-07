# Command - Entry Points

CLI entry points for the supervisor.

## Structure

```
cmd/
└── daemon/
    └── main.go    # Main entry point
```

## Daemon

The main supervisor binary.

### CLI Flags

| Flag | Description | Default |
|------|-------------|---------|
| `--config` | YAML config path | `/etc/daemon/config.yaml` |
| `--version` | Show version | - |

### Signal Handling

| Signal | Action |
|--------|--------|
| `SIGTERM`, `SIGINT` | Graceful shutdown |
| `SIGHUP` | Configuration reload |

### Execution Flow

```
main()
  ├─→ parseFlags()
  ├─→ infraconfig.NewLoader().Load()
  ├─→ appsupervisor.NewSupervisor()
  ├─→ setupSignals()
  ├─→ supervisor.Start()
  └─→ waitForSignals()
        ├─→ SIGHUP: supervisor.Reload()
        └─→ SIGTERM: supervisor.Stop()
```

## Build

```bash
# From src/
go build -o supervizio ./cmd/daemon

# With version
go build -ldflags "-X main.version=1.0.0" -o supervizio ./cmd/daemon
```

## Related Directories

| Directory | See |
|-----------|-----|
| daemon | `daemon/CLAUDE.md` |
