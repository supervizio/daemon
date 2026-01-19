# Daemon API Protocol Buffers

Generated Protocol Buffers and gRPC service definitions for the daemon API.

## Files

| File | Purpose |
|------|---------|
| `daemon.proto` | Protocol Buffer definitions |
| `daemon.pb.go` | Generated Go message types |
| `daemon_grpc.pb.go` | Generated gRPC service stubs |

## Services

### DaemonService

Process lifecycle and daemon state management.

| RPC | Description |
|-----|-------------|
| `GetState` | Get current daemon state |
| `StreamState` | Stream daemon state updates |
| `ListProcesses` | List all managed processes |
| `GetProcess` | Get specific process metrics |
| `StreamProcessMetrics` | Stream process metrics updates |

### MetricsService

System and process metrics streaming.

| RPC | Description |
|-----|-------------|
| `GetSystemMetrics` | Get current system metrics |
| `StreamSystemMetrics` | Stream system metrics updates |
| `StreamProcessMetrics` | Stream specific process metrics |
| `StreamAllProcessMetrics` | Stream all process metrics |

## Message Types

### Core Types

- `DaemonState` - Complete daemon state snapshot
- `ProcessMetrics` - Per-process CPU, memory, health
- `SystemMetrics` - System-wide CPU, memory usage
- `HostInfo` - Hostname, OS, architecture
- `KubernetesInfo` - Pod name, namespace, node

### Metrics Types

- `ProcessCPU` - User/system CPU time
- `ProcessMemory` - RSS, VMS, swap, shared, data, stack
- `SystemCPU` - User, nice, system, idle, iowait, irq
- `SystemMemory` - Total, available, used, free, swap

## Code Generation

Requires `protoc` with Go plugins:

```bash
protoc --go_out=. --go_opt=paths=source_relative \
       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
       daemon.proto
```

## Related Packages

| Package | Relation |
|---------|----------|
| `internal/infrastructure/grpc` | Server implementation |
| `internal/domain/state` | Domain state types |
| `internal/domain/metrics` | Domain metric types |
