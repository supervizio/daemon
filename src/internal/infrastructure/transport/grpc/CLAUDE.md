<!-- updated: 2026-02-15T21:30:00Z -->
# gRPC - Daemon API

gRPC server exposing control and monitoring services.

## Role

Enable remote daemon control: service management, metrics streaming.

## Structure

| File | Role |
|------|------|
| `server.go` | `Server` implementing gRPC services |

## Services

### DaemonService

```protobuf
service DaemonService {
    rpc Start(StartRequest) returns (StartResponse);
    rpc Stop(StopRequest) returns (StopResponse);
    rpc Restart(RestartRequest) returns (RestartResponse);
    rpc Status(StatusRequest) returns (StatusResponse);
    rpc StreamStatus(StatusRequest) returns (stream StatusResponse);
}
```

### MetricsService

```protobuf
service MetricsService {
    rpc GetMetrics(MetricsRequest) returns (MetricsResponse);
    rpc StreamMetrics(MetricsRequest) returns (stream MetricsResponse);
}
```

## Providers Requis

```go
type MetricsProvider interface {
    GetProcessMetrics(serviceName string) (metrics.ProcessMetrics, error)
    GetAllProcessMetrics() []metrics.ProcessMetrics
}

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

## Health Checks

Enregistre le protocole gRPC health/v1 pour :
- `daemon.v1.DaemonService`
- `daemon.v1.MetricsService`
- Server global (service name vide)

## Streaming

Intervalle par défaut : 5 secondes. Gère :
- Envoi initial immédiat
- Tick régulier
- Annulation context
