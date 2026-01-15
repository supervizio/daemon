# Infrastructure Layer

Adapters implementing domain ports for external systems.

## Structure

```
infrastructure/
├── kernel/                    # OS abstraction aggregator
│   └── ports/                 # Kernel interface definitions
│
├── process/                   # Process management
│   ├── execution/             # Process executor (Start, Stop, Signal)
│   ├── signals/               # Signal forwarding (SIGTERM, SIGHUP, etc.)
│   ├── reaper/                # Zombie process cleanup
│   ├── credentials/           # User/Group resolution
│   └── control/               # Process group control
│
├── resources/                 # System resources
│   ├── cgroup/                # Container resource limits (v1 + v2)
│   └── metrics/               # System metrics collection
│       ├── linux/             # Linux /proc parsing
│       ├── darwin/            # macOS implementation
│       ├── bsd/               # BSD implementation
│       └── scratch/           # Stub implementation
│
├── persistence/               # Data storage
│   ├── storage/               # Storage adapters
│   │   └── boltdb/            # BoltDB key-value adapter
│   └── config/                # Configuration loading
│       └── yaml/              # YAML file parser
│
├── observability/             # Monitoring & logging
│   ├── logging/               # Log management
│   └── healthcheck/           # Service health probing
│
└── transport/                 # Network communication
    └── grpc/                  # gRPC daemon server
```

## Packages

### kernel/ - OS Abstraction Aggregator

| Package | Role |
|---------|------|
| `kernel.go` | Aggregates all platform-specific adapters |
| `ports/` | Interface definitions for OS operations |

### process/ - Process Management

| Package | Role |
|---------|------|
| `execution` | Unix process executor (Start, Stop, Signal) |
| `signals` | Signal notification and forwarding |
| `reaper` | Zombie process cleanup |
| `credentials` | User/group credential resolution |
| `control` | Process group operations |

### resources/ - System Resources

| Package | Role |
|---------|------|
| `cgroup` | cgroups v1/v2 CPU/memory metrics |
| `metrics` | System telemetry collection |
| `metrics/linux` | Linux /proc parsing |
| `metrics/darwin` | macOS sysctl |
| `metrics/bsd` | BSD systems |
| `metrics/scratch` | Stub implementation |

### persistence/ - Data Storage

| Package | Role |
|---------|------|
| `storage/boltdb` | BoltDB key-value storage adapter |
| `config/yaml` | YAML configuration loader |

### observability/ - Monitoring

| Package | Role |
|---------|------|
| `logging` | File writers, capture, rotation |
| `healthcheck` | TCP, UDP, HTTP, gRPC, Exec, ICMP probers |

### transport/ - Network Communication

| Package | Role |
|---------|------|
| `grpc` | gRPC server implementing daemon API |

## Dependencies

- Depends on: `domain`
- Implements: domain port interfaces
- Never imported by: `domain`

## Key Types

### resources/cgroup
- `Reader` - Interface for cgroup metrics
- `V1Reader` - Legacy cgroups implementation
- `V2Reader` - Unified cgroups implementation
- `Detect()` - Auto-detect cgroup version

### process/signals
- `UnixSignalManager` - Unix signal handling
- Signal forwarding, subreaper support

### process/credentials
- `UnixCredentialManager` - User/group resolution
- `ScratchCredentialManager` - Numeric-only fallback

### process/execution
- `UnixExecutor` - Implements domain.Executor
- `TrustedCommand` - Secure wrapper for exec.CommandContext

### observability/healthcheck
- `TCPProber`, `UDPProber`, `HTTPProber`
- `GRPCProber`, `ExecProber`, `ICMPProber`
- `Factory` - Creates probers by type

### observability/logging
- `Writer` - Base log file writer with rotation
- `Capture` - Stdout/stderr capture coordinator
- `LineWriter` - Line-buffered output
- `MultiWriter` - Multiple destination writer

## Security Notes

### Command Execution
All command execution (`exec.CommandContext`) is centralized in `process/execution.TrustedCommand()`.
This function is intended for commands from validated configuration files only, not user input.
The security model assumes administrator-controlled YAML configurations loaded at startup.
