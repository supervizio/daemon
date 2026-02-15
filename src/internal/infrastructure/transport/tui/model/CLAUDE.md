<!-- updated: 2026-02-15T21:30:00Z -->
# Model - TUI Data Types

Data structures for TUI state and rendering.

## Structure

```
model/
└── snapshot.go    # Snapshot and related types
```

## Key Types

| Type | Description |
|------|-------------|
| `Snapshot` | Complete TUI state for one frame |
| `ServiceSnapshot` | Per-service display data |
| `SandboxInfo` | Container runtime detection result |
| `SystemMetrics` | CPU, RAM, swap, disk usage |
| `NetworkStats` | Network interface statistics |
| `CgroupLimits` | Cgroup resource limits |

## Snapshot Structure

```go
type Snapshot struct {
    Hostname      string
    Version       string
    Uptime        time.Duration
    Services      []ServiceSnapshot
    System        SystemMetrics
    Network       NetworkStats
    Cgroups       CgroupLimits
    Sandboxes     []SandboxInfo
    CollectedAt   time.Time
}
```

## Usage

Collectors populate Snapshot → Renderers consume it.

Immutable during render cycle (1Hz refresh).
