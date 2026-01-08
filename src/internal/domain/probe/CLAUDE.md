# Domain Probe Package

Core domain abstractions for service probing.

## Files

| File | Purpose |
|------|---------|
| `port.go` | `Prober` interface - port for infrastructure adapters |
| `target.go` | `Target` struct - probe target configuration |
| `result.go` | `Result` struct - probe execution result |
| `config.go` | `Config` struct - probe timing/threshold settings |
| `errors.go` | Domain errors for probing |

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

- `NewTCPTarget(address)` - Create TCP probe target
- `NewUDPTarget(address)` - Create UDP probe target
- `NewHTTPTarget(address, method, statusCode)` - Create HTTP probe target
- `NewGRPCTarget(address, service)` - Create gRPC probe target
- `NewExecTarget(command, args...)` - Create exec probe target
- `NewICMPTarget(address)` - Create ICMP probe target
- `NewConfig()` - Create config with defaults
- `NewSuccessResult(latency, output)` - Create success result
- `NewFailureResult(latency, output, err)` - Create failure result

## Infrastructure Adapters

See `infrastructure/probe/CLAUDE.md` for implementations.
