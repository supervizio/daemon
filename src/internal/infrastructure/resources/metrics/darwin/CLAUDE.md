# Darwin Metrics - macOS

Collecte métriques via sysctl et Mach API.

## Status

Implémentation partielle. Certaines métriques retournent `ErrNotImplemented`.

## Sources de Données

| Métrique | API |
|----------|-----|
| CPU count | `sysctl hw.ncpu` |
| CPU time | `host_processor_info` Mach API |
| Memory total | `sysctl hw.memsize` |
| Memory stats | `host_statistics64` Mach API |

## Structure

| Fichier | Rôle |
|---------|------|
| `probe.go` | `Probe` implémentant `SystemCollector` |

## Constructeur

```go
NewProbe() *Probe
```

## Build Tag

```go
//go:build darwin
```

## Limitations

- Pas de PSI (Pressure Stall Information) sur macOS
- Certaines métriques process-level limitées
