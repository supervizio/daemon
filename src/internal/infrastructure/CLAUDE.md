# Infrastructure Layer

Adapters implementing domain ports for external systems.

## Structure

```
infrastructure/
├── config/
│   └── yaml/     # YAML configuration loader
├── health/       # Health check adapters
├── kernel/       # OS abstraction layer
│   ├── adapters/ # Platform-specific implementations
│   └── ports/    # Kernel interfaces
├── logging/      # Log management (writers, capture, rotation)
└── process/      # Process executor adapter
```

## Packages

| Package | Role |
|---------|------|
| `config/yaml` | Loads and parses YAML config files |
| `health` | HTTP, TCP, Command health checkers |
| `kernel` | OS abstraction (signals, reaper, credentials) |
| `logging` | File writers, capture, rotation, timestamps |
| `process` | Unix process execution |

## Dependencies

- Depends on: `domain`
- Implements: domain port interfaces
- Never imported by: `domain`

## Key Types

### config/yaml
- `Loader` - YAML file loader
- `ConfigDTO` - YAML data transfer objects
- `Duration` - YAML duration parsing

### health
- `HTTPChecker` - HTTP endpoint checks
- `TCPChecker` - TCP port connectivity
- `CommandChecker` - Command execution checks
- `Factory` - Creates checkers by type

### kernel
- `SignalManager` - Signal forwarding (ports)
- `Reaper` - Zombie process reaping (ports)
- `UnixSignalManager` - Unix signal implementation (adapters)
- `UnixReaper` - Unix reaper implementation (adapters)
- `UnixCredentials` - User/Group resolution (adapters)

### logging
- `Writer` - Base log file writer with rotation
- `Capture` - Stdout/stderr capture coordinator
- `LineWriter` - Line-buffered output
- `MultiWriter` - Multiple destination writer
- `TimestampWriter` - Timestamp prefix formatting

### process
- `UnixExecutor` - Implements domain.Executor
