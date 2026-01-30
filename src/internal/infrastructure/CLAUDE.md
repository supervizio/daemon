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
| Stocker des données, charger config | `persistence/` |
| Logs, health checks | `observability/` |
| API gRPC | `transport/` |

## Structure

```
infrastructure/
├── probe/             # Métriques & quotas cross-platform (Rust FFI)
├── process/           # Processus OS (control, credentials, executor, reaper, signals)
├── persistence/       # Stockage (config/yaml, storage/boltdb)
├── observability/     # Monitoring (healthcheck, logging)
└── transport/         # Communication (grpc)
```

## Probe Package (NEW)

Le package `probe/` est l'adaptateur unifié pour toutes les ressources système :
- **Métriques** : CPU, mémoire, disque, réseau, I/O (cross-platform)
- **Quotas** : cgroups (Linux), launchd (macOS), jail (BSD)

Il remplace les anciens packages fragmentés :
- ~~`resources/metrics/linux/`~~ → `probe/`
- ~~`resources/metrics/darwin/`~~ → `probe/`
- ~~`resources/metrics/bsd/`~~ → `probe/`
- ~~`resources/metrics/scratch/`~~ → `probe/`
- ~~`resources/cgroup/`~~ → `probe/quota.go`

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
