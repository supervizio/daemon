# YAML Config - Parser Configuration

Chargement et parsing des fichiers YAML de configuration.

## Rôle

Lire un fichier YAML et le convertir en `*service.Config` du domaine.

## Structure

| Fichier | Rôle |
|---------|------|
| `loader.go` | `Loader` avec `Load(path)` |
| `types.go` | Types YAML intermédiaires |

## Flux

```
config.yaml
    │
    ▼
yaml.Unmarshal() → types.go (YAMLConfig)
    │
    ▼
Mapping → domain/service.Config
```

## Types Intermédiaires

```go
// types.go
type YAMLConfig struct {
    Version  string        `yaml:"version"`
    Services []YAMLService `yaml:"services"`
}

type YAMLService struct {
    Name    string `yaml:"name"`
    Command string `yaml:"command"`
    // ...
}
```

## Constructeur

```go
NewLoader() *Loader
```

## Usage

```go
loader := yaml.NewLoader()
cfg, err := loader.Load("/etc/daemon/config.yaml")
```

## Validation

Le `Loader` valide :
- Syntaxe YAML
- Champs requis
- Valeurs acceptables

Erreurs retournées avec contexte (ligne, champ).
