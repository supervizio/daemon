<!-- updated: 2026-02-15T21:30:00Z -->
# Health - Health Monitoring

Application service for monitoring service health via probes. This package was renamed from `healthcheck/` to better align with domain terminology.

## Role

Coordinate health probing for services, aggregate health status from multiple listeners, and emit health events. Uses the domain's `Prober` interface for actual probe execution.

## Structure

```
health/
├── monitor.go                       # ProbeMonitor - main health orchestrator
├── monitor_config.go                # ProbeMonitorConfig configuration
├── monitor_external_test.go         # Black-box tests
├── monitor_internal_test.go         # White-box tests
├── monitor_config_external_test.go  # Config black-box tests
├── monitor_config_internal_test.go  # Config white-box tests
├── listener.go                      # ListenerProbe - listener with probe
├── listener_external_test.go        # Listener black-box tests
├── listener_internal_test.go        # Listener white-box tests
├── ports.go                         # Creator port interface
└── errors.go                        # Sentinel errors
```

## Key Types

| Type | Description |
|------|-------------|
| `ProbeMonitor` | Main health orchestrator managing multiple listeners |
| `ProbeMonitorConfig` | Configuration for ProbeMonitor |
| `ListenerProbe` | Combines a listener with its associated prober |
| `Creator` | Port interface for creating probers |

## ProbeMonitor Methods

| Method | Description |
|--------|-------------|
| `NewProbeMonitor(config)` | Create a new probe-based health monitor |
| `AddListener(listener)` | Add a listener to monitor |
| `Start(ctx)` | Start periodic probing goroutines |
| `Stop()` | Stop all probing and cleanup |
| `SetProcessState(state)` | Update the process state |
| `SetCustomStatus(status)` | Set a custom status string |
| `Status()` | Return current aggregated health status |
| `Health()` | Return full aggregated health with listener details |
| `IsHealthy()` | Return true if all checks are healthy |
| `Latency()` | Return latest probe latency |

## Port Interface

```go
// Creator creates probers based on type.
// Infrastructure adapters implement this for prober creation.
type Creator interface {
    Create(proberType string, timeout time.Duration) (health.Prober, error)
}
```

## Errors

| Error | Description |
|-------|-------------|
| `ErrProberFactoryMissing` | Prober factory was not configured |
| `ErrEmptyProbeType` | Listener has probe config but no probe type |

## Dependencies

- Depends on: `domain/health`, `domain/listener`, `domain/process`
- Used by: `application/supervisor`, `cmd/daemon`

## Related Packages

| Package | Role |
|---------|------|
| `domain/health` | AggregatedHealth, Status, Event, Prober, Target, CheckConfig, CheckResult |
| `domain/listener` | Listener entity and states |
| `infrastructure/observability/healthcheck` | Prober implementations (TCP, HTTP, etc.) |
