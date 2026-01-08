# Domain Layer

Core business entities, value objects, and port interfaces.

## Structure

```
domain/
├── health/       # Health check entities and aggregation
├── listener/     # Network listener entities
├── probe/        # Probe abstractions and port interface
├── process/      # Process entities and ports
├── service/      # Service configuration entities
└── shared/       # Shared value objects
```

## Packages

| Package | Role |
|---------|------|
| `health` | Status, Result, Event, AggregatedHealth |
| `listener` | Listener entity, State enum |
| `probe` | Prober port, Target, Result, Config |
| `process` | Spec, State, Event, Executor port |
| `service` | Config, ServiceConfig, ListenerConfig |
| `shared` | Duration, Size value objects |

## Dependencies

- Depends on: nothing (pure domain)
- Used by: `application`, `infrastructure`

## Key Types

### health
- `Status` - Healthy, Unhealthy, Unknown, Degraded
- `Result` - Check result with status and message
- `Event` - Health state change events
- `AggregatedHealth` - Combined health from process, listeners, custom status

### listener
- `Listener` - Network listener entity
- `State` - Closed, Listening, Ready states

### probe
- `Prober` - Port interface for probing
- `Target` - Probe target (address, path, command, etc.)
- `Result` - Probe result with latency
- `Config` - Timeout, interval, thresholds

### process
- `Spec` - Process specification
- `State` - Running, Stopped, Failed states
- `Executor` - Port for process execution

### service
- `Config` - Root configuration
- `ServiceConfig` - Service definition
- `ListenerConfig` - Listener with probe config
- `HealthCheckConfig` - Legacy health check settings

### shared
- `Duration` - Time duration wrapper
- `Size` - Byte size wrapper
