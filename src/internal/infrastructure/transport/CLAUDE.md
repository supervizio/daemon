# Transport - Communication Réseau

Serveurs et clients pour communication externe.

## Rôle

Exposer des APIs pour contrôler le daemon (start/stop services, métriques, état).

## Navigation

| Protocole | Package |
|-----------|---------|
| gRPC | `grpc/` |

## Structure

```
transport/
└── grpc/              # API gRPC du daemon
    └── server.go      # Serveur gRPC
```

## Services Exposés

| Service | Description |
|---------|-------------|
| `DaemonService` | Start, Stop, Restart, Status des services |
| `MetricsService` | Métriques processus (CPU, RAM) |
| Health | gRPC health/v1 standard |
