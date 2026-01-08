# Application Layer

Use case implementations coordinating domain and infrastructure.

## Structure

```
application/
├── config/       # Configuration port interface
├── health/       # Health monitoring orchestration
├── process/      # Process lifecycle management
└── supervisor/   # Service orchestration
```

## Packages

| Package | Role |
|---------|------|
| `config` | Loader interface (port) |
| `health` | HealthMonitor coordinates checks |
| `process` | ProcessManager handles lifecycle |
| `supervisor` | Supervisor orchestrates services |

## Dependencies

- Depends on: `domain`
- Used by: `cmd/daemon`
- May use: `infrastructure/kernel` (for OS abstractions via ports)

## Key Types

- `Supervisor` - Main service orchestrator
- `ProcessManager` - Per-service process management
- `HealthMonitor` - Periodic health checking
- `Loader` - Config loading interface
