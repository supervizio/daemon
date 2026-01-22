# Domain Lifecycle Package

Domain types for daemon lifecycle management, event publishing, and zombie reaping.

## Files

| File | Purpose |
|------|---------|
| `event.go` | Event, Type enum - lifecycle event types |
| `publisher.go` | Publisher port interface |
| `daemon.go` | DaemonState, SystemState snapshots |
| `host.go` | HostInfo - system information |
| `reaper.go` | Reaper port interface (zombie cleanup) |

## Event Types (Type enum)

| Category | Events |
|----------|--------|
| Process | Started, Stopped, Failed, Restarted, Healthy, Unhealthy |
| Mesh | NodeUp, NodeDown, LeaderChanged, TopologyChanged |
| Kubernetes | PodCreated, PodDeleted, PodReady, PodFailed |
| System | HighCPU, HighMemory, DiskFull |
| Daemon | Started, Stopping, ConfigReloaded |

## Key Types

| Type | Fields |
|------|--------|
| `Event` | ID, Type, Timestamp, ServiceName, NodeID, PodName, Message, Data |
| `DaemonState` | Timestamp, Host, Processes, System, Mesh?, Kubernetes? |
| `HostInfo` | Hostname, OS, Arch, KernelVersion, DaemonPID, DaemonVersion, StartTime |

## Port Interfaces

| Interface | Methods | Purpose |
|-----------|---------|---------|
| `Publisher` | Publish, Subscribe, Unsubscribe | Event distribution |
| `Reaper` | Start, Stop, ReapOnce, IsPID1 | Zombie process cleanup |

## Filter Functions

- `FilterByType(types ...Type)` - Filter by event types
- `FilterByCategory(category string)` - Filter by category
- `FilterByServiceName(name string)` - Filter by service

## Dependencies

- Depends on: `domain/metrics` (ProcessMetrics in DaemonState)
- Used by: `application/supervisor`, `infrastructure/process/reaper`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/metrics` | ProcessMetrics in DaemonState |
| `infrastructure/process/reaper` | Implements Reaper port |
