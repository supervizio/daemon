# Observability - Monitoring

Logging et health checking des services.

## Rôle

Fournir les outils de monitoring : capture des logs, vérification de santé des services.

## Navigation

| Besoin | Package |
|--------|---------|
| Capturer stdout/stderr des processus | `logging/` |
| Vérifier la santé des services (TCP, HTTP, etc.) | `healthcheck/` |

## Structure

```
observability/
├── logging/           # Capture et formatage logs
│   ├── capture.go     # Coordination stdout/stderr
│   ├── linewriter.go  # Écriture ligne par ligne
│   ├── multiwriter.go # Destinations multiples
│   └── timestamp.go   # Préfixage horodatage
│
└── healthcheck/       # Probers de santé
    ├── factory.go     # Factory par type
    ├── tcp.go         # TCP connect
    ├── http.go        # HTTP GET/HEAD
    ├── grpc.go        # gRPC health/v1
    ├── exec.go        # Command execution
    └── icmp.go        # Ping (fallback TCP)
```

## Terminologie

| Terme | Signification |
|-------|---------------|
| **healthcheck** | Vérification qu'un service répond (TCP, HTTP...) |
| **probe/metrics** | Collecte de métriques système (CPU, RAM) - voir `resources/` |
