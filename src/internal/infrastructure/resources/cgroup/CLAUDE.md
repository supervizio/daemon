# Cgroup - Limites Container

Lecture des limites et métriques cgroups v1 et v2.

## Contexte

Les cgroups contrôlent les ressources CPU/mémoire dans les containers. Deux versions coexistent :
- **v1** : Hiérarchies séparées (`/sys/fs/cgroup/cpu/`, `/sys/fs/cgroup/memory/`)
- **v2** : Hiérarchie unifiée (`/sys/fs/cgroup/`)

## Interface

```go
type Reader interface {
    CPUUsage(ctx context.Context) (uint64, error)
    CPULimit(ctx context.Context) (quota, period uint64, err error)
    MemoryUsage(ctx context.Context) (uint64, error)
    MemoryLimit(ctx context.Context) (uint64, error)
    ReadMemoryStat(ctx context.Context) (MemoryStat, error)
    Path() string
    Version() int  // 1 ou 2
}
```

## Fichiers

| Fichier | Rôle |
|---------|------|
| `detector.go` | `Detect()` → `VersionV1`, `VersionV2`, `VersionHybrid` |
| `v1.go` | `V1Reader` : lecture `/sys/fs/cgroup/cpu/`, `/sys/fs/cgroup/memory/` |
| `v2.go` | `V2Reader` : lecture `/sys/fs/cgroup/` unifié |
| `errors.go` | `ErrUnknownVersion`, `ErrNotInCgroup` |

## Détection Automatique

```go
func Detect() Version {
    if exists("/sys/fs/cgroup/cgroup.controllers") {
        if exists("/sys/fs/cgroup/cpu") {
            return VersionHybrid
        }
        return VersionV2
    }
    if exists("/sys/fs/cgroup/cpu") {
        return VersionV1
    }
    return VersionUnknown
}

func NewReader() (Reader, error) {
    switch Detect() {
    case VersionV2, VersionHybrid:
        return NewV2Reader("")
    case VersionV1:
        return NewV1Reader("", "")
    default:
        return nil, ErrUnknownVersion
    }
}
```

## Constructeurs

```go
NewReader() (Reader, error)           // Auto-détection
NewV1Reader(cpuPath, memPath) *V1Reader
NewV2Reader(path string) (*V2Reader, error)
```

## Build Tag

```go
//go:build linux
```

Cgroups n'existent que sur Linux.
