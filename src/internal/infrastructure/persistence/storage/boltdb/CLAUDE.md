# BoltDB - Stockage Key-Value

Adaptateur BoltDB pour stockage embedded.

## Contexte

BoltDB est une base de données key-value embedded, parfaite pour un daemon :
- Pas de serveur externe
- ACID transactions
- Fichier unique

## Interface Implémentée

```go
type Store interface {
    Get(bucket, key string) ([]byte, error)
    Put(bucket, key string, value []byte) error
    Delete(bucket, key string) error
    Close() error
}
```

## Structure

| Fichier | Rôle |
|---------|------|
| `store.go` | `Store` wrappant `*bolt.DB` |

## Constructeur

```go
New(path string) (*Store, error)
NewWithOptions(path string, opts Options) (*Store, error)
```

## Usage

```go
store, err := boltdb.New("/var/lib/daemon/state.db")
defer store.Close()

err = store.Put("services", "nginx", []byte(`{"status":"running"}`))
data, err := store.Get("services", "nginx")
```

## Buckets

BoltDB organise les données en "buckets" (équivalent de tables) :

```go
store.Put("services", "nginx", ...)   // bucket=services, key=nginx
store.Put("metrics", "cpu", ...)      // bucket=metrics, key=cpu
```
