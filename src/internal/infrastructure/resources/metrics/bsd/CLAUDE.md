# BSD Metrics - FreeBSD/OpenBSD/NetBSD

Collecte métriques pour la famille BSD.

## Status

Stub - retourne `ErrNotImplemented`.

## Plateformes Supportées

- FreeBSD
- OpenBSD
- NetBSD
- DragonFly BSD

## Sources de Données (à implémenter)

| Métrique | API |
|----------|-----|
| CPU count | `sysctl hw.ncpu` |
| CPU time | `sysctl kern.cp_time` |
| Memory | `sysctl hw.physmem`, `vm.stats` |

## Structure

| Fichier | Rôle |
|---------|------|
| `probe.go` | `Probe` stub |

## Constructeur

```go
NewProbe() *Probe
```

## Build Tag

```go
//go:build freebsd || openbsd || netbsd || dragonfly
```
