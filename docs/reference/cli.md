# CLI Reference

superviz.io daemon command-line interface.

---

## Usage

```bash
supervizio [flags]
```

---

## Flags

| Flag | Type | Default | Description |
|------|------|---------|-------------|
| `--config` | `string` | `/etc/supervizio/config.yaml` | Path to configuration file |
| `--tui` | `bool` | `false` | Enable interactive TUI mode |

---

## Examples

```bash
# Run with default config location
supervizio

# Run with custom config
supervizio --config /path/to/config.yaml

# Run with interactive TUI
supervizio --config config.yaml --tui
```

---

## Exit Codes

| Code | Description |
|------|-------------|
| `0` | Clean shutdown |
| `1` | Startup failure (config error, port in use) |
| `2` | Runtime error |

---

## Signals

| Signal | Behavior |
|--------|----------|
| `SIGTERM` | Graceful shutdown |
| `SIGINT` | Graceful shutdown (Ctrl+C) |
| `SIGHUP` | Reload configuration |

---

## Environment Variables

| Variable | Description |
|----------|-------------|
| `CGO_ENABLED` | Must be `1` for probe library (build-time) |
| `GOMAXPROCS` | Go runtime parallelism (default: number of CPUs) |

---

## gRPC Client

Use `grpcurl` to interact with the running daemon:

```bash
# List available services
grpcurl -plaintext localhost:50051 list

# Get daemon state
grpcurl -plaintext localhost:50051 daemon.v1.DaemonService/GetState

# Get specific process
grpcurl -plaintext -d '{"service_name": "my-app"}' \
  localhost:50051 daemon.v1.DaemonService/GetProcess

# Stream system metrics (5 second interval)
grpcurl -plaintext -d '{"interval": "5s"}' \
  localhost:50051 daemon.v1.MetricsService/StreamSystemMetrics

# Health check
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```
