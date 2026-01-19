# Infrastructure Layer

Adaptateurs techniques de l'architecture hexagonale.

## Rôle

Implémenter les interfaces (ports) du domaine avec des technologies concrètes : OS, base de données, réseau, fichiers.

**Règle** : Le domaine ne dépend jamais de l'infrastructure.

## Navigation

| Besoin | Package |
|--------|---------|
| Exécuter un processus, gérer signaux, credentials | `process/` |
| Lire métriques système (CPU, RAM, cgroups) | `resources/` |
| Stocker des données, charger config | `persistence/` |
| Logs, health checks | `observability/` |
| API gRPC | `transport/` |

## Structure

```
infrastructure/
├── process/           # Processus OS (exec, signals, reaper, credentials)
├── resources/         # Ressources système (cgroups, metrics)
├── persistence/       # Stockage (boltdb, yaml config)
├── observability/     # Monitoring (logs, healthcheck)
└── transport/         # Communication (grpc)
```

## Conventions Globales

### Fichiers

| Pattern | Usage |
|---------|-------|
| `{concept}.go` | Interface + types publics |
| `{concept}_{platform}.go` | Code spécifique plateforme |
| `{concept}_scratch.go` | Stub pour CI/tests |
| `errors.go` | Erreurs sentinelles partagées |

### Constructeurs

| Pattern | Quand |
|---------|-------|
| `New()` | Création standard |
| `NewWithDeps(...)` | Injection Wire |
| `NewWithOptions(...)` | Tests avec mocks |

### Build Tags

```go
//go:build linux        // Linux only
//go:build darwin       // macOS only
//go:build unix         // Linux + macOS + BSD
//go:build !unix        // Scratch/stub
```

### Tests

| Suffix | Type |
|--------|------|
| `_external_test.go` | Black-box (package_test) |
| `_internal_test.go` | White-box (même package) |
