# Protobuf Reference

Complete Protocol Buffer definitions for the superviz.io gRPC API.

**Source**: `src/api/proto/v1/daemon/daemon.proto`

**Package**: `daemon.v1`

**Go Package**: `github.com/kodflow/daemon/api/proto/v1/daemon;daemonpb`

---

## Services

### DaemonService

```protobuf
service DaemonService {
    rpc GetState(google.protobuf.Empty) returns (DaemonState);
    rpc StreamState(StreamStateRequest) returns (stream DaemonState);
    rpc ListProcesses(google.protobuf.Empty) returns (ListProcessesResponse);
    rpc GetProcess(GetProcessRequest) returns (ProcessMetrics);
    rpc StreamProcessMetrics(StreamProcessMetricsRequest) returns (stream ProcessMetrics);
}
```

### MetricsService

```protobuf
service MetricsService {
    rpc GetSystemMetrics(google.protobuf.Empty) returns (SystemMetrics);
    rpc StreamSystemMetrics(StreamMetricsRequest) returns (stream SystemMetrics);
    rpc StreamProcessMetrics(StreamProcessMetricsRequest) returns (stream ProcessMetrics);
    rpc StreamAllProcessMetrics(StreamMetricsRequest) returns (stream ProcessMetrics);
}
```

---

## Request Messages

### StreamStateRequest

```protobuf
message StreamStateRequest {
    google.protobuf.Duration interval = 1;
}
```

### StreamMetricsRequest

```protobuf
message StreamMetricsRequest {
    google.protobuf.Duration interval = 1;
}
```

### StreamProcessMetricsRequest

```protobuf
message StreamProcessMetricsRequest {
    string service_name = 1;
    google.protobuf.Duration interval = 2;
}
```

### GetProcessRequest

```protobuf
message GetProcessRequest {
    string service_name = 1;
}
```

---

## Response Messages

### DaemonState

```protobuf
message DaemonState {
    string version = 1;
    google.protobuf.Timestamp start_time = 2;
    google.protobuf.Duration uptime = 3;
    bool healthy = 4;
    repeated ProcessMetrics processes = 5;
    SystemMetrics system = 6;
    HostInfo host = 7;
    KubernetesInfo kubernetes = 8;
}
```

### ListProcessesResponse

```protobuf
message ListProcessesResponse {
    repeated ProcessMetrics processes = 1;
}
```

### ProcessMetrics

```protobuf
message ProcessMetrics {
    string service_name = 1;
    int32 pid = 2;
    ProcessState state = 3;
    bool healthy = 4;
    ProcessCPU cpu = 5;
    ProcessMemory memory = 6;
    google.protobuf.Timestamp start_time = 7;
    google.protobuf.Duration uptime = 8;
    int32 restart_count = 9;
    string last_error = 10;
    google.protobuf.Timestamp timestamp = 11;
}
```

### SystemMetrics

```protobuf
message SystemMetrics {
    SystemCPU cpu = 1;
    SystemMemory memory = 2;
    LoadAverage load = 3;
    google.protobuf.Timestamp timestamp = 4;
}
```

---

## Enums

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

---

## Nested Types

### HostInfo

```protobuf
message HostInfo {
    string hostname = 1;
    string os = 2;
    string arch = 3;
    int32 num_cpus = 4;
}
```

### KubernetesInfo

```protobuf
message KubernetesInfo {
    string pod_name = 1;
    string namespace = 2;
    string node_name = 3;
    map<string, string> labels = 4;
}
```

### ProcessCPU

```protobuf
message ProcessCPU {
    uint64 user_time_ns = 1;
    uint64 system_time_ns = 2;
    uint64 total_time_ns = 3;
    double usage_percent = 4;
}
```

### ProcessMemory

```protobuf
message ProcessMemory {
    uint64 rss_bytes = 1;
    uint64 vms_bytes = 2;
    uint64 swap_bytes = 3;
    uint64 shared_bytes = 4;
    uint64 data_bytes = 5;
    uint64 stack_bytes = 6;
}
```

### SystemCPU

```protobuf
message SystemCPU {
    uint64 user_ns = 1;
    uint64 nice_ns = 2;
    uint64 system_ns = 3;
    uint64 idle_ns = 4;
    uint64 iowait_ns = 5;
    uint64 irq_ns = 6;
    uint64 softirq_ns = 7;
    uint64 steal_ns = 8;
    double usage_percent = 9;
}
```

### SystemMemory

```protobuf
message SystemMemory {
    uint64 total_bytes = 1;
    uint64 available_bytes = 2;
    uint64 used_bytes = 3;
    uint64 free_bytes = 4;
    uint64 buffers_bytes = 5;
    uint64 cached_bytes = 6;
    uint64 shared_bytes = 7;
    uint64 swap_total_bytes = 8;
    uint64 swap_used_bytes = 9;
    uint64 swap_free_bytes = 10;
    double usage_percent = 11;
}
```

### LoadAverage

```protobuf
message LoadAverage {
    double load1 = 1;
    double load5 = 2;
    double load15 = 3;
}
```

---

## Imports

```protobuf
import "google/protobuf/timestamp.proto";
import "google/protobuf/duration.proto";
import "google/protobuf/empty.proto";
```
