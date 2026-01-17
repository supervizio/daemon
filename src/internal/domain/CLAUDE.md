# Domain Layer

Core business entities, value objects, and port interfaces.

## Structure

```
domain/
├── lifecycle/    # Daemon lifecycle: events, state, reaper port
├── config/       # Service configuration value objects
├── health/       # Health status, aggregation, prober port
├── healthcheck/  # Standalone probe abstractions (avoids cycles)
├── process/      # Process entities, executor port
├── listener/     # Network listener entities
├── metrics/      # System and process metrics
├── shared/       # Common value objects (Duration, Size, Clock)
└── storage/      # MetricsStore port interface
```

## Packages

| Package | Role |
|---------|------|
| `lifecycle` | Event types, Publisher port, DaemonState, HostInfo, Reaper port |
| `config` | Config, ServiceConfig, ListenerConfig, RestartConfig, LoggingConfig |
| `health` | Status, Result, Event, AggregatedHealth, Prober port, Target, CheckConfig |
| `healthcheck` | Prober port, Target, Config, Result (standalone, avoids import cycles) |
| `process` | Spec, State, Executor port, ExitResult, RestartTracker, Event |
| `listener` | Listener entity, State enum |
| `metrics` | SystemCPU, SystemMemory, ProcessMetrics, collector interfaces |
| `shared` | Duration, Size, Clock (Nower), constants, errors |
| `storage` | MetricsStore port interface |

## Dependencies

- Depends on: nothing (pure domain, except internal dependencies)
- Used by: `application`, `infrastructure`

**Internal Dependencies:**
```
shared ←── config, process, health, metrics, lifecycle
process ←── metrics, health
listener ←── health
metrics ←── lifecycle, storage
config ←── process (RestartTracker)
```

## Key Types by Package

### lifecycle
- `Type` - Event type enum (process, mesh, k8s, system, daemon)
- `Event` - Lifecycle event with timestamp and data
- `Publisher` - Port for event publishing
- `DaemonState` - Complete daemon state snapshot
- `HostInfo` - Host system information
- `Reaper` - Port for zombie process cleanup

### config
- `Config` - Root configuration
- `ServiceConfig` - Service definition with command, env, restart, listeners
- `ListenerConfig` - Network listener with probe
- `RestartConfig` - Restart policy, retries, delays
- `LoggingConfig` - Global logging defaults
- `ProbeConfig` - Health probe settings

### health
- `Status` - Healthy, Unhealthy, Unknown, Degraded
- `Result` - Health check result with status and message
- `Event` - Health state change events
- `AggregatedHealth` - Combined health from process, listeners, custom status
- `Prober` - Port interface for health probing
- `Target` - Probe target configuration
- `CheckConfig` - Probe timing and thresholds
- `CheckResult` - Probe execution result

### healthcheck
- `Prober` - Port interface (standalone version)
- `Target` - Probe target
- `Config` - Probe configuration
- `Result` - Probe result

### process
- `Spec` - Process execution specification
- `State` - Stopped, Starting, Running, Stopping, Failed
- `Executor` - Port for OS process execution
- `ExitResult` - Process exit code and error
- `RestartTracker` - Restart attempts with exponential backoff
- `Event` - Process lifecycle event

### listener
- `Listener` - Network listener entity
- `State` - Closed, Listening, Ready

### metrics
- `SystemCPU`, `ProcessCPU` - CPU metrics
- `SystemMemory`, `ProcessMemory` - Memory metrics
- `DiskUsage`, `DiskIOStats` - Disk metrics
- `NetInterface`, `NetStats` - Network metrics
- `ProcessMetrics` - Aggregated process metrics
- `Collector` interfaces - CPU, Memory, Disk, Network, IO

### shared
- `Duration` - Time duration wrapper
- `Size` - Byte size utilities (ParseSize, FormatSize)
- `Nower` - Clock interface
- `RealClock` - System time implementation
- Common errors and constants

### storage
- `MetricsStore` - Port for metrics persistence
- `StoreConfig` - Storage configuration

## Port Interfaces

The domain defines these port interfaces for infrastructure to implement:

| Port | Package | Purpose |
|------|---------|---------|
| `Executor` | process | OS process execution |
| `Prober` | health, healthcheck | Health probing |
| `Publisher` | lifecycle | Event publishing |
| `Reaper` | lifecycle | Zombie process cleanup |
| `MetricsStore` | storage | Metrics persistence |
| Collectors | metrics | System metrics collection |

## Related Directories

| Directory | See |
|-----------|-----|
| lifecycle | `lifecycle/CLAUDE.md` |
| config | `config/CLAUDE.md` |
| health | `health/CLAUDE.md` |
| healthcheck | `healthcheck/CLAUDE.md` |
| process | `process/CLAUDE.md` |
| listener | `listener/CLAUDE.md` |
| metrics | `metrics/CLAUDE.md` |
| shared | `shared/CLAUDE.md` |
| storage | `storage/CLAUDE.md` |
