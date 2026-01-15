# Storage - Adaptateurs de Stockage

Implémentations de stockage persistant.

## Rôle

Fournir des adaptateurs pour stocker et récupérer des données de manière persistante.

## Navigation

| Backend | Package |
|---------|---------|
| BoltDB (embedded) | `boltdb/` |

## Structure

```
storage/
└── boltdb/           # Base de données embedded
    └── store.go      # Store implémentant domain/storage.Store
```

## Interface Implémentée

```go
// domain/storage/port.go
type Store interface {
    Get(bucket, key string) ([]byte, error)
    Put(bucket, key string, value []byte) error
    Delete(bucket, key string) error
    Close() error
}
```
