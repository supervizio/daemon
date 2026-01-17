# Config - Chargement Configuration

Adaptateurs pour charger la configuration au démarrage.

## Rôle

Parser les fichiers de configuration et les mapper vers les types du domaine.

## Navigation

| Format | Package |
|--------|---------|
| YAML | `yaml/` |

## Structure

```
config/
└── yaml/              # Parser YAML
    ├── loader.go      # Loader principal
    └── types.go       # Types intermédiaires
```

## Interface Implémentée

```go
// application/config/loader.go
type Loader interface {
    Load(path string) (*service.Config, error)
}
```
