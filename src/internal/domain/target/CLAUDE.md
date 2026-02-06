# Domain Target Package

Domain entities for external monitoring targets.

## Purpose

External targets are services, containers, or hosts that supervizio monitors but does not manage. Unlike managed services (which supervizio spawns and controls), external targets have no lifecycle control - supervizio only observes their health.

## Files

| File | Purpose |
|------|---------|
| `type.go` | `Type` enum - target types (systemd, docker, kubernetes, etc.) |
| `source.go` | `Source` enum - static (config) or discovered |
| `target.go` | `ExternalTarget` entity |
| `status.go` | `Status`, `State` - health status tracking |
| `discovery.go` | `Discoverer` port interface, `DiscoveryConfig` |
| `watcher.go` | `Watcher` port interface, `Event` types |

## Key Types

### Type (Enum)
```go
const (
    TypeSystemd    Type = "systemd"
    TypeOpenRC     Type = "openrc"
    TypeBSDRC      Type = "bsd-rc"
    TypeDocker     Type = "docker"
    TypePodman     Type = "podman"
    TypeKubernetes Type = "kubernetes"
    TypeNomad      Type = "nomad"
    TypeRemote     Type = "remote"
    TypeCustom     Type = "custom"
)
```

### ExternalTarget (Entity)
```go
type ExternalTarget struct {
    ID               string
    Name             string
    Type             Type
    Source           Source
    Labels           map[string]string
    ProbeType        string
    ProbeTarget      health.Target
    Interval         time.Duration
    Timeout          time.Duration
    SuccessThreshold int
    FailureThreshold int
}
```

### Discoverer (Port Interface)
```go
type Discoverer interface {
    Discover(ctx context.Context) ([]ExternalTarget, error)
    Type() Type
}
```

### Watcher (Port Interface)
```go
type Watcher interface {
    Watch(ctx context.Context) (<-chan Event, error)
    Type() Type
}
```

## Factory Functions

- `NewExternalTarget(id, name, type, source)` - Generic target
- `NewRemoteTarget(name, address, probeType)` - Remote host/endpoint
- `NewDockerTarget(containerID, containerName)` - Docker container
- `NewSystemdTarget(unitName)` - systemd service
- `NewKubernetesTarget(namespace, resourceType, resourceName)` - K8s resource
- `NewNomadTarget(allocID, taskName, jobName)` - Nomad allocation

## Dependencies

- Depends on: `domain/health` (Target, CheckResult)
- Used by: `application/monitoring`, `infrastructure/discovery/*`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/health` | Reuses Target and CheckResult types |
| `application/monitoring` | Uses ExternalTarget for orchestration |
| `infrastructure/discovery/*` | Implements Discoverer port |
