# Infrastructure gRPC Package

gRPC server implementation for the daemon API services.

## Files

| File | Purpose |
|------|---------|
| `server.go` | gRPC server implementation |

## Types

### Server

Implements both `DaemonService` and `MetricsService` from the proto definitions.

```go
type Server struct {
    grpcServer      *grpc.Server
    healthServer    *health.Server
    metricsProvider MetricsProvider
    stateProvider   StateProvider
}
```

### MetricsProvider

Interface for accessing process metrics.

```go
type MetricsProvider interface {
    GetProcessMetrics(serviceName string) (metrics.ProcessMetrics, error)
    GetAllProcessMetrics() []metrics.ProcessMetrics
    Subscribe() <-chan metrics.ProcessMetrics
    Unsubscribe(ch <-chan metrics.ProcessMetrics)
}
```

### StateProvider

Interface for accessing daemon state.

```go
type StateProvider interface {
    GetState() state.DaemonState
}
```

## Usage

```go
server := grpc.NewServer(metricsProvider, stateProvider)
if err := server.Serve(":50051"); err != nil {
    log.Fatal(err)
}
defer server.Stop()
```

## Streaming

All streaming methods use a generic `streamLoop` helper that:
1. Sends initial value immediately
2. Ticks at configured interval
3. Handles context cancellation
4. Continues on emit errors (skip tick)

Default stream interval: 5 seconds.

## Health Checks

Registers gRPC health/v1 protocol for:
- `daemon.v1.DaemonService`
- `daemon.v1.MetricsService`
- Overall server health (empty service name)

## Related Packages

| Package | Relation |
|---------|----------|
| `api/proto/v1/daemon` | Proto definitions |
| `internal/domain/state` | DaemonState type |
| `internal/domain/metrics` | ProcessMetrics type |
| `internal/infrastructure/healthcheck` | gRPC health prober |
