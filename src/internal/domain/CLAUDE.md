# Domain Layer

Core business entities, value objects, and port interfaces.

## Structure

```
domain/
├── health/       # Health check entities
├── process/      # Process entities and ports
├── service/      # Service configuration entities
└── shared/       # Shared value objects
```

## Packages

| Package | Role |
|---------|------|
| `health` | Status, Result, Event types |
| `process` | Spec, State, Event, Executor port |
| `service` | Config, ServiceConfig, HealthCheck |
| `shared` | Duration, Size value objects |

## Dependencies

- Depends on: nothing (pure domain)
- Used by: `application`, `infrastructure`

## Key Types

### health
- `Status` - Healthy, Unhealthy, Unknown
- `Result` - Check result with status and message
- `Event` - Health state change events

### process
- `Spec` - Process specification
- `State` - Running, Stopped, Failed states
- `Executor` - Port for process execution

### service
- `Config` - Root configuration
- `ServiceConfig` - Service definition
- `HealthCheckConfig` - Health check settings

### shared
- `Duration` - Time duration wrapper
- `Size` - Byte size wrapper
