# Metrics - Collecte Métriques Système

Métriques CPU et mémoire, multi-plateforme.

## Rôle

Collecter les métriques système de manière portable. Chaque plateforme a son implémentation.

## Navigation

| Plateforme | Package | Source |
|------------|---------|--------|
| Linux | `linux/` | `/proc/stat`, `/proc/meminfo` |
| macOS | `darwin/` | sysctl |
| BSD | `bsd/` | sysctl variantes |
| Autres | `scratch/` | Stub (retourne zéros) |

## Structure

```
metrics/
├── factory.go      # NewCollector() auto-détection
├── linux/          # Parsing /proc
├── darwin/         # Mach API + sysctl
├── bsd/            # FreeBSD, OpenBSD, NetBSD
└── scratch/        # Fallback CI/tests
```

## Factory

```go
collector := metrics.NewSystemCollector()  // Auto-détection plateforme
cpu, err := collector.CPU().CollectSystem(ctx)
mem, err := collector.Memory().CollectSystem(ctx)
```

## Constructeurs par Plateforme

| Package | Constructeur | Build Tag |
|---------|--------------|-----------|
| `linux` | `NewCollector()` | `//go:build linux` |
| `darwin` | `NewProbe()` | `//go:build darwin` |
| `bsd` | `NewProbe()` | `//go:build freebsd \|\| openbsd...` |
| `scratch` | `NewProbe()` | `//go:build !unix` |
