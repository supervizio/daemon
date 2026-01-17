# Domain Health Package

Domain entities and value objects for health checking and status aggregation.

This package provides both high-level health status aggregation (combining process
state, listener states, and custom status) and low-level probe abstractions
(Prober port, Target, CheckConfig, CheckResult).

## Files

| File | Purpose |
|------|---------|
| `status.go` | `Status` enum - health status values |
| `result.go` | `Result` - high-level health check result |
| `event.go` | `Event` - health state change event |
| `aggregation.go` | `AggregatedHealth` - combined health from multiple sources |
| `listener_status.go` | `ListenerStatus` - listener health status |
| `prober.go` | `Prober` port interface for infrastructure adapters |
| `target.go` | `Target` - probe target configuration |
| `check_config.go` | `CheckConfig` - probe timing and thresholds |
| `check_result.go` | `CheckResult` - probe execution result |
| `check_errors.go` | Domain errors for health checking |

## Key Types

### Status (Enum)

```go
const (
    StatusUnknown   Status = iota  // Health not yet determined
    StatusHealthy                   // All checks pass
    StatusUnhealthy                 // Checks are failing
    StatusDegraded                  // Some checks are failing
)
```

### Result (High-Level Health Check Result)

```go
type Result struct {
    Status    Status
    Message   string
    Duration  time.Duration
    Timestamp time.Time
    Error     error
}
```

### Event (Health State Change)

```go
type Event struct {
    Checker   string
    Status    Status
    Result    Result
    Timestamp time.Time
}
```

### AggregatedHealth (Combined Status)

```go
type AggregatedHealth struct {
    ProcessState process.State
    Listeners    []ListenerStatus
    CustomStatus string           // "DRAINING", "DEGRADED", etc.
    LastCheck    time.Time
    Latency      time.Duration
}
```

Status computation logic:
1. Process must be running
2. All listeners must be ready
3. CustomStatus must be empty or "HEALTHY"

### Prober (Port Interface)

```go
type Prober interface {
    Probe(ctx context.Context, target Target) CheckResult
    Type() string  // "tcp", "udp", "http", "grpc", "exec", "icmp"
}
```

### Target (Probe Target)

```go
type Target struct {
    Network    string   // "tcp", "udp", "icmp"
    Address    string   // "localhost:8080"
    Path       string   // HTTP path
    Service    string   // gRPC service name
    Command    string   // Exec command
    Args       []string // Exec args
    Method     string   // HTTP method
    StatusCode int      // Expected HTTP status
}
```

### CheckConfig (Probe Configuration)

```go
type CheckConfig struct {
    Timeout          time.Duration  // Default: 5s
    Interval         time.Duration  // Default: 10s
    SuccessThreshold int            // Default: 1
    FailureThreshold int            // Default: 3
}
```

### CheckResult (Probe Execution Result)

```go
type CheckResult struct {
    Success bool
    Latency time.Duration
    Output  string
    Error   error
}
```

## Factory Functions

### High-Level Results
- `NewHealthyResult(message, duration)` - Create healthy result
- `NewHealthyResultAt(message, duration, timestamp)` - With specific timestamp
- `NewUnhealthyResult(message, duration, err)` - Create unhealthy result
- `NewUnhealthyResultAt(message, duration, err, timestamp)` - With specific timestamp

### Events
- `NewEvent(checker, status, result)` - Create health event
- `NewEventAt(checker, status, result, timestamp)` - With specific timestamp

### Aggregation
- `NewAggregatedHealth(processState)` - Create aggregated health

### Probe Targets
- `NewTarget(network, address)` - Generic target
- `NewTCPTarget(address)` - TCP probe target
- `NewUDPTarget(address)` - UDP probe target
- `NewHTTPTarget(address, method, statusCode)` - HTTP probe target
- `NewGRPCTarget(address, service)` - gRPC probe target
- `NewExecTarget(command, args...)` - Exec probe target
- `NewICMPTarget(address)` - ICMP ping target

### Probe Results
- `NewCheckConfig()` - Default probe config
- `NewCheckResult(success, latency, output, err)` - Generic result
- `NewSuccessCheckResult(latency, output)` - Success result
- `NewFailureCheckResult(latency, output, err)` - Failure result

## Dependencies

- Depends on: `domain/process` (State)
- Used by: `application/health`, `infrastructure/observability/healthcheck`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/process` | Process state for aggregation |
| `infrastructure/observability/healthcheck` | Implements Prober port |
| `application/health` | Orchestrates health checks |
