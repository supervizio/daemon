<!-- updated: 2026-02-15T21:30:00Z -->
# Monitoring - External Target Monitoring

Monitors external resources supervizio does not manage (unlike `health/ProbeMonitor` for managed services).

## Purpose

Observe external targets: systemd/OpenRC/BSD services, Docker/Podman containers, Kubernetes pods, Nomad allocations, remote hosts (ICMP/TCP/UDP).

## Structure

```
monitoring/
├── monitor.go                     # ExternalMonitor orchestrator
├── monitor_config.go              # Config value object
├── defaults_config.go             # DefaultsConfig (intervals, thresholds)
├── discovery_config.go            # DiscoveryModeConfig
├── registry.go                    # Thread-safe target storage
├── health_summary.go              # Aggregated counts by state/type
├── ports.go                       # ProberCreator interface, callbacks
└── errors.go                      # Sentinel errors
```

## Key Types

| Type | Description |
|------|-------------|
| `ExternalMonitor` | Main orchestrator: AddTarget, RemoveTarget, Start, Stop, Health |
| `Config` | Monitor configuration with defaults, discovery, and callbacks |
| `Registry` | Thread-safe target storage with status tracking |
| `HealthSummary` | Aggregated counts (HealthyCount, UnhealthyCount) |

## Dependencies

- Depends on: `domain/target`, `domain/health`
- Used by: `application/supervisor`, `bootstrap`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/target` | ExternalTarget, Status, Discoverer, Watcher |
| `domain/health` | Prober interface, CheckResult, Target |
| `infrastructure/discovery` | Discoverer implementations |
| `infrastructure/observability/healthcheck` | Prober implementations |

## Errors

| Error | Description |
|-------|-------------|
| `ErrProberFactoryMissing` | Factory not configured |
| `ErrEmptyProbeType` | Target has no probe type |
| `ErrTargetNotFound` | Target not in registry |
| `ErrTargetExists` | Target already exists |
