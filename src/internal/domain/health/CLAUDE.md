# Domain Health Package

Health checking entities, status aggregation, and probe abstractions.

## Files

| File | Purpose |
|------|---------|
| `status.go` | `Status` enum - Healthy, Unhealthy, Degraded, Unknown |
| `result.go` | `Result` - high-level health check result |
| `event.go` | `Event` - health state change event |
| `aggregation.go` | `AggregatedHealth` - combined health from multiple sources |
| `listener_status.go` | `ListenerStatus` - listener health status |
| `prober.go` | `Prober` port interface |
| `target.go` | `Target` - probe target configuration |
| `check_config.go` | `CheckConfig` - probe timing and thresholds |
| `check_result.go` | `CheckResult` - probe execution result |

## Key Types

### Status (Enum)
- `StatusUnknown`, `StatusHealthy`, `StatusUnhealthy`, `StatusDegraded`

### Result
- `Status`, `Message`, `Duration`, `Timestamp`, `Error`
- Factory: `NewHealthyResult(msg, dur)`, `NewUnhealthyResult(msg, dur, err)`

### AggregatedHealth
- Combines: `ProcessState`, `Listeners[]`, `CustomStatus`, `LastCheck`, `Latency`
- Status logic: Process running + All listeners ready + No custom degradation

### Prober (Port Interface)
```go
type Prober interface {
    Probe(ctx context.Context, target Target) CheckResult
    Type() string  // "tcp", "http", "grpc", "exec", "icmp"
}
```

### Target
- `Network`, `Address`, `Path`, `Service`, `Command`, `Args`, `Method`, `StatusCode`
- Factory: `NewTCPTarget(addr)`, `NewHTTPTarget(addr, method, code)`, `NewGRPCTarget(addr, svc)`, `NewExecTarget(cmd, args)`

### CheckConfig
- `Timeout` (5s), `Interval` (10s), `SuccessThreshold` (1), `FailureThreshold` (3)

### CheckResult
- `Success bool`, `Latency`, `Output`, `Error`
- Factory: `NewSuccessCheckResult(latency, output)`, `NewFailureCheckResult(latency, output, err)`

## Dependencies

- Depends on: `domain/process` (State)
- Used by: `application/health`, `infrastructure/observability/healthcheck`

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/process` | Process state for aggregation |
| `infrastructure/observability/healthcheck` | Implements Prober port |
| `application/health` | Orchestrates health checks |
