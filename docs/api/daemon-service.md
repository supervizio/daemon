# DaemonService

The `DaemonService` provides daemon state queries and process management via gRPC.

```protobuf
service DaemonService {
    rpc GetState(google.protobuf.Empty) returns (DaemonState);
    rpc StreamState(StreamStateRequest) returns (stream DaemonState);
    rpc ListProcesses(google.protobuf.Empty) returns (ListProcessesResponse);
    rpc GetProcess(GetProcessRequest) returns (ProcessMetrics);
    rpc StreamProcessMetrics(StreamProcessMetricsRequest) returns (stream ProcessMetrics);
}
```

---

## RPCs

### GetState

Returns the complete daemon state including version, uptime, health status, all processes, system metrics, and host information.

**Request**: `google.protobuf.Empty`

**Response**: `DaemonState`

```bash
grpcurl -plaintext localhost:50051 daemon.v1.DaemonService/GetState
```

### StreamState

Streams daemon state updates at a configurable interval.

**Request**: `StreamStateRequest`

| Field | Type | Description |
|-------|------|-------------|
| `interval` | `Duration` | Minimum interval between updates (default: 5s) |

**Response**: `stream DaemonState`

### ListProcesses

Returns metrics for all supervised processes.

**Request**: `google.protobuf.Empty`

**Response**: `ListProcessesResponse`

| Field | Type | Description |
|-------|------|-------------|
| `processes` | `repeated ProcessMetrics` | All supervised process metrics |

### GetProcess

Returns metrics for a specific process identified by service name.

**Request**: `GetProcessRequest`

| Field | Type | Description |
|-------|------|-------------|
| `service_name` | `string` | Service name from configuration |

**Response**: `ProcessMetrics`

### StreamProcessMetrics

Streams metrics for a specific process at a configurable interval.

**Request**: `StreamProcessMetricsRequest`

| Field | Type | Description |
|-------|------|-------------|
| `service_name` | `string` | Service name to stream |
| `interval` | `Duration` | Interval between snapshots (default: 5s) |

**Response**: `stream ProcessMetrics`

---

## Message Types

### DaemonState

Complete daemon state snapshot.

| Field | Type | Description |
|-------|------|-------------|
| `version` | `string` | Daemon version |
| `start_time` | `Timestamp` | Daemon start time |
| `uptime` | `Duration` | Daemon uptime |
| `healthy` | `bool` | Overall daemon health |
| `processes` | `repeated ProcessMetrics` | All supervised processes |
| `system` | `SystemMetrics` | System metrics snapshot |
| `host` | `HostInfo` | Host information |
| `kubernetes` | `KubernetesInfo` | K8s info (if applicable) |

### ProcessMetrics

Per-process metrics.

| Field | Type | Description |
|-------|------|-------------|
| `service_name` | `string` | Service name from configuration |
| `pid` | `int32` | Current process ID (0 if not running) |
| `state` | `ProcessState` | Current lifecycle state |
| `healthy` | `bool` | Health status |
| `cpu` | `ProcessCPU` | CPU metrics |
| `memory` | `ProcessMemory` | Memory metrics |
| `start_time` | `Timestamp` | Process start time |
| `uptime` | `Duration` | Process uptime |
| `restart_count` | `int32` | Number of restarts |
| `last_error` | `string` | Last error message (if failed) |
| `timestamp` | `Timestamp` | Collection timestamp |

### ProcessState

```protobuf
enum ProcessState {
    PROCESS_STATE_UNSPECIFIED = 0;
    PROCESS_STATE_STOPPED    = 1;
    PROCESS_STATE_STARTING   = 2;
    PROCESS_STATE_RUNNING    = 3;
    PROCESS_STATE_STOPPING   = 4;
    PROCESS_STATE_FAILED     = 5;
}
```

### HostInfo

| Field | Type | Description |
|-------|------|-------------|
| `hostname` | `string` | System hostname |
| `os` | `string` | Operating system |
| `arch` | `string` | CPU architecture |
| `num_cpus` | `int32` | Number of CPUs |

### KubernetesInfo

| Field | Type | Description |
|-------|------|-------------|
| `pod_name` | `string` | Pod name |
| `namespace` | `string` | K8s namespace |
| `node_name` | `string` | Node name |
| `labels` | `map<string, string>` | Pod labels |

### ProcessCPU

| Field | Type | Description |
|-------|------|-------------|
| `user_time_ns` | `uint64` | User mode CPU time (ns) |
| `system_time_ns` | `uint64` | System mode CPU time (ns) |
| `total_time_ns` | `uint64` | Total CPU time (ns) |
| `usage_percent` | `double` | CPU usage (0-100 per core) |

### ProcessMemory

| Field | Type | Description |
|-------|------|-------------|
| `rss_bytes` | `uint64` | Resident set size |
| `vms_bytes` | `uint64` | Virtual memory size |
| `swap_bytes` | `uint64` | Swap usage |
| `shared_bytes` | `uint64` | Shared memory |
| `data_bytes` | `uint64` | Data segment size |
| `stack_bytes` | `uint64` | Stack size |
