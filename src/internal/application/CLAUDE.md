# Application Layer

Use case implementations coordinating domain and infrastructure.

## Structure

```
application/
├── config/       # Configuration port interface
├── health/       # Health monitoring with ProbeMonitor
├── process/      # Process lifecycle management
└── supervisor/   # Service orchestration
```

## Packages

| Package | Role |
|---------|------|
| `config` | Loader interface (port) |
| `health` | ProbeMonitor coordinates listener probes |
| `process` | ProcessManager handles lifecycle |
| `supervisor` | Supervisor orchestrates services |

## Dependencies

- Depends on: `domain`
- Used by: `cmd/daemon`
- May use: `infrastructure/kernel` (for OS abstractions via ports)

## Key Types

- `Supervisor` - Main service orchestrator
- `ProcessManager` - Per-service process management
- `ProbeMonitor` - Multi-protocol health probing
- `ProberFactory` - Port for creating probers
- `Loader` - Config loading interface
