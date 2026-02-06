# Infrastructure Layer

Adaptateurs techniques de l'architecture hexagonale.

## Rôle

Implémenter les interfaces (ports) du domaine avec des technologies concrètes : OS, base de données, réseau, fichiers.

**Règle** : Le domaine ne dépend jamais de l'infrastructure.

## Navigation

| Besoin | Package |
|--------|---------|
| Exécuter un processus, gérer signaux, credentials | `process/` |
| Métriques système & quotas (CPU, RAM, cgroups) | `probe/` |
| Découverte de cibles externes (Docker, systemd, K8s) | `discovery/` |
| Stocker des données, charger config | `persistence/` |
| Logs, health checks, events | `observability/` |
| API gRPC, TUI | `transport/` |

## Structure

```
infrastructure/
├── discovery/         # Target discovery (Docker, systemd, K8s, Nomad, etc.)
├── probe/             # Métriques & quotas cross-platform (Rust FFI)
├── process/           # Processus OS (control, credentials, executor, reaper, signals)
├── persistence/       # Stockage (config/yaml, storage/boltdb)
├── observability/     # Monitoring (healthcheck, logging, events)
└── transport/         # Communication (grpc, tui)
```

## Probe Package

Le package `probe/` est l'adaptateur unifié pour toutes les ressources système :
- **Métriques** : CPU, mémoire, disque, réseau, I/O (cross-platform)
- **Quotas** : cgroups (Linux), launchd (macOS), jail (BSD)

Voir `probe/CLAUDE.md` pour plus de détails.

## Conventions Globales

### Fichiers

| Pattern | Usage |
|---------|-------|
| `{concept}.go` | Interface + types publics |
| `{concept}_{platform}.go` | Code spécifique plateforme |
| `errors.go` | Erreurs sentinelles partagées |

### Constructeurs

| Pattern | Quand |
|---------|-------|
| `New()` | Création standard |
| `NewWithDeps(...)` | Injection Wire |
| `NewWithOptions(...)` | Tests avec mocks |

### Build Tags

```go
//go:build cgo          // Requires CGO (probe package)
//go:build linux        // Linux only
//go:build darwin       // macOS only
//go:build unix         // Linux + macOS + BSD
```

### Tests

| Suffix | Type |
|--------|------|
| `_external_test.go` | Black-box (package_test) |
| `_internal_test.go` | White-box (même package) |
