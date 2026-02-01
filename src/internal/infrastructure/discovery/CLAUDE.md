# Discovery - Target Discovery Adapters

Infrastructure adapters for discovering external monitoring targets.

## Purpose

Implement the `domain/target.Discoverer` interface for various platforms:
- **systemd** - Linux service manager (via systemctl)
- **Docker** - Container runtime (via Docker Engine API)
- **Static** - Targets from configuration file

## Structure

```
discovery/
├── factory.go           # Factory for creating discoverers
├── factory_unix.go      # Unix-specific factory methods
├── factory_linux.go     # Linux-specific factory (systemd)
├── docker.go            # Docker discoverer (Unix)
├── systemd.go           # Systemd discoverer (Linux)
└── static.go            # Static configuration discoverer
```

## Key Types

| Type | Description |
|------|-------------|
| `Factory` | Creates discoverers from configuration |
| `DockerDiscoverer` | Discovers Docker containers via API |
| `SystemdDiscoverer` | Discovers systemd services via systemctl |
| `StaticDiscoverer` | Creates targets from static configuration |

## Usage

```go
// Create factory from config
factory := discovery.NewFactory(cfg.Discovery)

// Get all enabled discoverers
discoverers := factory.CreateDiscoverers()

// Use with ExternalMonitor
monitorCfg := monitoring.NewConfig().
    WithDiscoverers(discoverers...)
```

## Discoverer Interface

All discoverers implement `domain/target.Discoverer`:

```go
type Discoverer interface {
    Discover(ctx context.Context) ([]ExternalTarget, error)
    Type() Type
}
```

## Platform Support

| Discoverer | Platform | Build Tag |
|------------|----------|-----------|
| Docker | Unix | `//go:build unix` |
| Systemd | Linux | `//go:build linux` |
| Static | All | (none) |

## Dependencies

- Depends on: `domain/target`, `domain/config`, `domain/health`
- Used by: `application/monitoring`, `bootstrap`

## Docker Discovery

Connects to Docker socket and queries running containers:

```go
discoverer := discovery.NewDockerDiscoverer(
    "/var/run/docker.sock",
    map[string]string{"supervizio.monitor": "true"},
)
```

Features:
- Label-based filtering
- Automatic TCP probe configuration from exposed ports
- Container state tracking

## Systemd Discovery

Queries systemd via systemctl for running services:

```go
discoverer := discovery.NewSystemdDiscoverer([]string{
    "nginx.service",
    "postgresql*.service",
})
```

Features:
- Glob pattern matching for service names
- Exec probe using `systemctl is-active`

## Static Discovery

Creates targets from YAML configuration:

```go
discoverer := discovery.NewStaticDiscoverer(cfg.Targets)
```

Supports all probe types: TCP, UDP, HTTP, ICMP, Exec
