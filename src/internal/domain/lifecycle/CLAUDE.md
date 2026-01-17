# Domain Lifecycle Package

Domain types for daemon lifecycle management, event publishing, and zombie reaping.

This package is a fusion of event, state, and reaper concerns, providing a cohesive
API for managing daemon lifecycle across process supervision, mesh networking, and
Kubernetes orchestration contexts.

## Files

| File | Purpose |
|------|---------|
| `event.go` | `Event`, `Type` - lifecycle event types and structures |
| `publisher.go` | `Publisher` port interface for event publishing |
| `daemon.go` | `DaemonState`, `SystemState` - daemon state snapshots |
| `host.go` | `HostInfo` - host system information |
| `reaper.go` | `Reaper` port interface for zombie process cleanup |

## Key Types

### Event Types (Type enum)

Event categories:
- **Process**: `TypeProcessStarted`, `TypeProcessStopped`, `TypeProcessFailed`, `TypeProcessRestarted`, `TypeProcessHealthy`, `TypeProcessUnhealthy`
- **Mesh**: `TypeMeshNodeUp`, `TypeMeshNodeDown`, `TypeMeshLeaderChanged`, `TypeMeshTopologyChanged`
- **Kubernetes**: `TypeK8sPodCreated`, `TypeK8sPodDeleted`, `TypeK8sPodReady`, `TypeK8sPodFailed`
- **System**: `TypeSystemHighCPU`, `TypeSystemHighMemory`, `TypeSystemDiskFull`
- **Daemon**: `TypeDaemonStarted`, `TypeDaemonStopping`, `TypeDaemonConfigReloaded`

### Event (Value Object)

```go
type Event struct {
    ID          string
    Type        Type
    Timestamp   time.Time
    ServiceName string
    NodeID      string
    PodName     string
    Message     string
    Data        map[string]any
}
```

### Publisher (Port Interface)

```go
type Publisher interface {
    Publish(event Event)
    Subscribe() <-chan Event
    Unsubscribe(ch <-chan Event)
}
```

### DaemonState (Value Object)

```go
type DaemonState struct {
    Timestamp   time.Time
    Host        HostInfo
    Processes   []metrics.ProcessMetrics
    System      SystemState
    Mesh        *MeshTopology       // optional
    Kubernetes  *KubernetesState    // optional
}
```

### HostInfo (Value Object)

```go
type HostInfo struct {
    Hostname      string
    OS            string
    Arch          string
    KernelVersion string
    DaemonPID     int
    DaemonVersion string
    StartTime     time.Time
}
```

### Reaper (Port Interface)

```go
type Reaper interface {
    Start()
    Stop()
    ReapOnce() int
    IsPID1() bool
}
```

The Reaper interface handles zombie process cleanup for PID1 scenarios in containers.
When running as PID1, orphaned child processes become zombies if not reaped.

## Filter Functions

- `FilterByType(types ...Type)` - Filter events by specific types
- `FilterByCategory(category string)` - Filter events by category
- `FilterByServiceName(serviceName string)` - Filter events by service name

## Dependencies

- Depends on: `domain/metrics` (for ProcessMetrics in DaemonState)
- Used by: `application/supervisor`, `infrastructure/process/reaper`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/metrics` | ProcessMetrics used in DaemonState |
| `domain/process` | Process events complement lifecycle events |
| `infrastructure/process/reaper` | Implements Reaper port |
