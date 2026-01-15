# Healthcheck - Probers de Santé

Vérification de la santé des services via différents protocoles.

## Rôle

Tester qu'un service est accessible et répond correctement.

**Note** : Différent de `resources/metrics/` qui collecte des métriques système (CPU, RAM).

## Probers Disponibles

| Type | Fichier | Description |
|------|---------|-------------|
| TCP | `tcp.go` | Connexion TCP réussie |
| UDP | `udp.go` | Envoi paquet (connectionless) |
| HTTP | `http.go` | GET/HEAD, validation status |
| gRPC | `grpc.go` | Protocole health/v1 |
| Exec | `exec.go` | Commande exit code 0 |
| ICMP | `icmp.go` | Ping (fallback TCP si pas CAP_NET_RAW) |

## Factory

```go
factory := NewFactory(5 * time.Second)  // Timeout par défaut
prober, err := factory.Create("http", 10*time.Second)

// Ou méthodes typées
tcpProber := factory.CreateTCP(5 * time.Second)
httpProber := factory.CreateHTTP(10 * time.Second)
```

## Interface Implémentée

```go
// domain/healthcheck/prober.go
type Prober interface {
    Probe(ctx context.Context, target Target) Result
}
```

## Constructeurs

```go
NewTCPProber(timeout time.Duration) *TCPProber
NewHTTPProber(timeout time.Duration) *HTTPProber
NewGRPCProber(timeout time.Duration) *GRPCProber
NewExecProber(timeout time.Duration) *ExecProber
NewICMPProber(timeout time.Duration) *ICMPProber
NewUDPProber(timeout time.Duration) *UDPProber
```

## Sécurité

`ExecProber` utilise `executor.TrustedCommand()` - voir `process/executor/CLAUDE.md`.
