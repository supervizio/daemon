# Infrastructure Layer

Adapters implementing domain ports for external systems.

## Structure

```
infrastructure/
├── boltdb/       # BoltDB storage adapter
├── cgroup/       # cgroup v2 metrics collection
├── config/
│   └── yaml/     # YAML configuration loader
├── grpc/         # gRPC server for daemon API
├── healthcheck/  # Health check probers (TCP, UDP, HTTP, gRPC, Exec, ICMP)
├── kernel/       # OS abstraction layer
│   ├── adapters/ # Platform-specific implementations
│   └── ports/    # Kernel interfaces
├── logging/      # Log management (writers, capture, rotation)
├── metrics/      # System and process metrics collection
│   ├── linux/    # Linux-specific collectors
│   └── scratch/  # Scratch/stub collectors
└── process/      # Process executor adapter
```

## Packages

| Package | Role |
|---------|------|
| `boltdb` | BoltDB key-value storage adapter |
| `cgroup` | cgroup v2 CPU/memory metrics |
| `config/yaml` | Loads and parses YAML config files |
| `grpc` | gRPC server implementing daemon API |
| `healthcheck` | Protocol probers implementing domain.Prober |
| `kernel` | OS abstraction (signals, reaper, credentials) |
| `logging` | File writers, capture, rotation, timestamps |
| `metrics` | System and process metrics collection |
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

### healthcheck
- `TCPProber` - TCP connection probes
- `UDPProber` - UDP packet probes
- `HTTPProber` - HTTP endpoint probes
- `GRPCProber` - gRPC health probes (health/v1 protocol)
- `ExecProber` - Command execution probes
- `ICMPProber` - ICMP ping probes (TCP fallback)
- `Factory` - Creates probers by type

### grpc
- `Server` - gRPC server for daemon API
- `MetricsProvider` - Interface for process metrics
- `StateProvider` - Interface for daemon state

### process
- `UnixExecutor` - Implements domain.Executor
- `TrustedCommand` - Secure wrapper for exec.CommandContext (trusted config sources only)

## Security Notes

### Command Execution
All command execution (`exec.CommandContext`) is centralized in `process.TrustedCommand()`.
This function is intended for commands from validated configuration files only, not user input.
The security model assumes administrator-controlled YAML configurations loaded at startup.
