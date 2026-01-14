# Application Layer

Use case implementations coordinating domain and infrastructure.

## Structure

```
application/
├── config/       # Configuration port interface
├── healthcheck/  # Service reachability monitoring (TCP, HTTP, etc.)
├── probe/        # System metrics tracking (CPU, RAM, etc.)
├── process/      # Process lifecycle management
└── supervisor/   # Service orchestration
```

## Packages

| Package | Role |
|---------|------|
| `config` | Loader interface (port) |
| `healthcheck` | ProbeMonitor coordinates service health checks |
| `probe` | Tracker monitors process/system metrics |
| `process` | ProcessManager handles lifecycle |
| `supervisor` | Supervisor orchestrates services |

## Terminology

- **healthcheck**: Service reachability verification (TCP, HTTP, ICMP, gRPC, UDP, Exec)
- **probe**: System metrics collection (CPU, RAM, DISK, NET, I/O)

## Dependencies

- Depends on: `domain`
- Used by: `cmd/daemon`
- May use: `infrastructure/kernel` (for OS abstractions via ports)

## Key Types

- `Supervisor` - Main service orchestrator
- `ProcessManager` - Per-service process management
- `ProbeMonitor` - Multi-protocol health checking
- `Tracker` - Process metrics tracking
- `Creator` - Port for creating health checkers
- `Loader` - Config loading interface
