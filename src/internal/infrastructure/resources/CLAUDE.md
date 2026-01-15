# Resources - Ressources Système

Lecture des ressources système : cgroups et métriques (CPU, RAM).

## Rôle

Abstraire l'accès aux métriques système selon la plateforme (Linux, macOS, BSD) et l'environnement (container, VM, bare metal).

## Navigation

| Besoin | Package |
|--------|---------|
| Limites cgroups (CPU, mémoire) | `cgroup/` |
| Métriques système (CPU, RAM) | `metrics/` |

## Structure

```
resources/
├── cgroup/            # Lecture cgroups v1/v2
│   ├── detector.go    # Auto-détection version
│   ├── v1.go          # Legacy cgroups
│   ├── v2.go          # Unified hierarchy
│   └── errors.go
│
└── metrics/           # Collecte métriques
    ├── factory.go     # Factory multi-plateforme
    ├── linux/         # /proc parsing
    ├── darwin/        # sysctl macOS
    ├── bsd/           # BSD systems
    └── scratch/       # Stub CI/tests
```

## Choix Plateforme

```go
// metrics/factory.go
func NewCollector() Collector {
    switch runtime.GOOS {
    case "linux":
        return linux.NewCollector()
    case "darwin":
        return darwin.NewProbe()
    case "freebsd", "openbsd":
        return bsd.NewProbe()
    default:
        return scratch.NewProbe()
    }
}
```
