# Monitoring - External Target Monitoring

Application service for monitoring external targets (services, containers, hosts).

## Purpose

Unlike `health/ProbeMonitor` which monitors managed service listeners, `ExternalMonitor` observes external resources that supervizio does not manage. External targets include:

- **systemd/OpenRC/BSD rc.d services** - OS-level services
- **Docker/Podman containers** - Container runtimes
- **Kubernetes pods/services** - Orchestrated workloads
- **Nomad allocations** - HashiCorp Nomad jobs
- **Remote hosts** - ICMP ping, TCP/UDP port checks

## Structure

```
monitoring/
├── monitor.go            # ExternalMonitor - main orchestrator
├── monitor_config.go     # Config, DefaultsConfig, DiscoveryModeConfig
├── registry.go           # Registry - thread-safe target storage
├── ports.go              # ProberCreator interface, callbacks
└── errors.go             # Sentinel errors
```

## Key Types

| Type | Description |
|------|-------------|
| `ExternalMonitor` | Main orchestrator managing target probing |
| `Config` | Monitor configuration with defaults and callbacks |
| `Registry` | Thread-safe storage for targets and their status |
| `HealthSummary` | Aggregated counts by state and type |

## ExternalMonitor Methods

| Method | Description |
|--------|-------------|
| `NewExternalMonitor(config)` | Create a new monitor |
| `AddTarget(target)` | Add a target to monitor |
| `RemoveTarget(id)` | Remove a target |
| `Start(ctx)` | Start probing goroutines |
| `Stop()` | Stop all probing |
| `Health()` | Return health summary |
| `Registry()` | Access the target registry |
| `GetStatus(id)` | Get status for a target |
| `AllStatuses()` | Get all target statuses |

## Usage

```go
// Create monitor with factory
config := monitoring.NewConfig().
    WithFactory(proberFactory).
    WithDiscoverers(systemdDiscoverer, dockerDiscoverer)

monitor := monitoring.NewExternalMonitor(config)

// Add static targets
monitor.AddTarget(target.NewRemoteTarget("db", "db.internal:5432", "tcp"))
monitor.AddTarget(target.NewRemoteTarget("cache", "redis.internal:6379", "tcp"))

// Start monitoring
monitor.Start(ctx)
defer monitor.Stop()

// Check health
summary := monitor.Health()
fmt.Printf("Healthy: %d, Unhealthy: %d\n", summary.HealthyCount(), summary.UnhealthyCount())
```

## Configuration

```go
config := monitoring.Config{
    Defaults: monitoring.DefaultsConfig{
        Interval:         30 * time.Second,
        Timeout:          5 * time.Second,
        SuccessThreshold: 1,
        FailureThreshold: 3,
    },
    Discovery: monitoring.DiscoveryModeConfig{
        Enabled:     true,
        Interval:    60 * time.Second,
        Discoverers: []target.Discoverer{...},
        Watchers:    []target.Watcher{...},
    },
    Factory:        proberFactory,
    Events:         eventsChan,
    OnHealthChange: onHealthChange,
    OnUnhealthy:    onUnhealthy,
    OnHealthy:      onHealthy,
}
```

## Dependencies

- Depends on: `domain/target`, `domain/health`
- Used by: `application/supervisor`, `bootstrap`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/target` | ExternalTarget, Status, Discoverer, Watcher |
| `domain/health` | Prober interface, CheckResult, Target |
| `infrastructure/discovery/*` | Discoverer implementations |
| `infrastructure/observability/healthcheck` | Prober implementations |

## Errors

| Error | Description |
|-------|-------------|
| `ErrProberFactoryMissing` | Factory not configured |
| `ErrEmptyProbeType` | Target has no probe type |
| `ErrTargetNotFound` | Target not in registry |
| `ErrTargetExists` | Target already exists |
