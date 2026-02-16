<!-- updated: 2026-02-15T21:30:00Z -->
# Domain Layer

Core business entities, value objects, and port interfaces.

## Structure

```
domain/
├── config/       # Configuration value objects (ServiceConfig, RestartConfig)
├── health/       # Health status, aggregation, Prober port
├── lifecycle/    # Daemon lifecycle: events, state, Reaper port
├── listener/     # Network listener entities
├── logging/      # Daemon event logging: Level, LogEvent, Writer/Logger ports
├── metrics/      # System and process metrics types
├── process/      # Process entities, Executor port
├── shared/       # Common value objects (Duration, Size, Clock)
├── storage/      # MetricsStore port interface
└── target/       # External target entities, Discoverer/Watcher ports
```

## Packages

| Package | Key Types |
|---------|-----------|
| `config` | Config, ServiceConfig, RestartConfig, LoggingConfig, DaemonLogging, ProbeConfig |
| `health` | Status, Result, AggregatedHealth, Prober port, Target, CheckConfig |
| `lifecycle` | Event, Type, Publisher port, DaemonState, HostInfo, Reaper port |
| `listener` | Listener entity, State enum |
| `logging` | Level, LogEvent, Writer port, Logger port |
| `metrics` | SystemCPU, SystemMemory, ProcessMetrics, Collector interfaces |
| `process` | Spec, State, Executor port, ExitResult, RestartTracker |
| `shared` | Duration, Size, Clock (Nower), RealClock |
| `storage` | MetricsStore port, StoreConfig |
| `target` | ExternalTarget, Status, Discoverer/Watcher ports, Type enum |

## Port Interfaces

| Port | Package | Purpose |
|------|---------|---------|
| `Executor` | process | OS process execution |
| `Prober` | health | Health probing |
| `Publisher` | lifecycle | Event publishing |
| `Reaper` | lifecycle | Zombie process cleanup |
| `Logger` | logging | Daemon event logging |
| `Writer` | logging | Log output destinations |
| `MetricsStore` | storage | Metrics persistence |
| `Discoverer` | target | External target discovery |
| `Watcher` | target | External target watching |
| Collectors | metrics | System metrics collection |

## Dependencies

- Depends on: nothing (pure domain)
- Used by: `application`, `infrastructure`

```
shared ←── config, process, health, metrics, lifecycle
process ←── metrics, health
config ←── process (RestartTracker)
```

## Related Directories

| Directory | See |
|-----------|-----|
| config | `config/CLAUDE.md` |
| health | `health/CLAUDE.md` |
| lifecycle | `lifecycle/CLAUDE.md` |
| listener | `listener/CLAUDE.md` |
| logging | `logging/CLAUDE.md` |
| metrics | `metrics/CLAUDE.md` |
| process | `process/CLAUDE.md` |
| shared | `shared/CLAUDE.md` |
| storage | `storage/CLAUDE.md` |
| target | `target/CLAUDE.md` |
