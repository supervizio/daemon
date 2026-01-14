# Domain Healthcheck Package

Core domain abstractions for service health checking.

## Purpose

This package provides the domain model for **service health checks** - verifying
that services are reachable and responding correctly via TCP, HTTP, gRPC, etc.

Note: This is different from **metrics/probes** which collect system telemetry (CPU, RAM, DISK).

## Files

| File | Purpose |
|------|---------|
| `port.go` | `Prober` interface - port for infrastructure adapters |
| `target.go` | `Target` struct - health check target configuration |
| `result.go` | `Result` struct - health check execution result |
| `config.go` | `Config` struct - timing/threshold settings |
| `errors.go` | Domain errors for health checking |

## Key Types

### Prober (Port Interface)
```go
type Prober interface {
    Probe(ctx context.Context, target Target) Result
    Type() string  // "tcp", "http", "grpc", "exec", "icmp"
}
```

### Target (Value Object)
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

### Result (Value Object)
```go
type Result struct {
    Success bool
    Latency time.Duration
    Output  string
    Error   error
}
```

### Config (Value Object)
```go
type Config struct {
    Timeout          time.Duration
    Interval         time.Duration
    SuccessThreshold int
    FailureThreshold int
}
```

## Factory Functions

- `NewTCPTarget(address)` - Create TCP health check target
- `NewUDPTarget(address)` - Create UDP health check target
- `NewHTTPTarget(address, method, statusCode)` - Create HTTP health check target
- `NewGRPCTarget(address, service)` - Create gRPC health check target
- `NewExecTarget(command, args...)` - Create exec health check target
- `NewICMPTarget(address)` - Create ICMP health check target
- `NewConfig()` - Create config with defaults
- `NewSuccessResult(latency, output)` - Create success result
- `NewFailureResult(latency, output, err)` - Create failure result

## Infrastructure Adapters

See `infrastructure/healthcheck/CLAUDE.md` for implementations.

## Related Packages

| Package | Relation |
|---------|----------|
| `infrastructure/healthcheck/` | Implements Prober interface |
| `application/healthcheck/` | Orchestrates health checks |
| `domain/probe/` | System metrics (CPU, RAM) - different concern |
