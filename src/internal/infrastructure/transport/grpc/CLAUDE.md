# gRPC - API du Daemon

Serveur gRPC exposant les services de contrôle et monitoring.

## Rôle

Permettre le contrôle à distance du daemon : gestion des services, streaming de métriques.

## Structure

| Fichier | Rôle |
|---------|------|
| `server.go` | `Server` implémentant les services gRPC |

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
