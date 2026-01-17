# Persistence - Stockage de Données

Adaptateurs pour le stockage persistant et le chargement de configuration.

## Rôle

Abstraire l'accès aux données : fichiers de configuration, base de données key-value.

## Navigation

| Besoin | Package |
|--------|---------|
| Stocker des données clé-valeur | `storage/boltdb/` |
| Charger la configuration YAML | `config/yaml/` |

## Structure

```
persistence/
├── storage/           # Adaptateurs stockage
│   └── boltdb/        # BoltDB embedded database
│       └── store.go   # Implémente domain/storage.Store
│
└── config/            # Chargement configuration
    └── yaml/          # Parser YAML
        ├── loader.go  # Loader principal
        └── types.go   # Types de mapping YAML → domain
```

## Séparation des Responsabilités

- **storage/** : Persistance runtime (état, métriques, cache)
- **config/** : Configuration statique au démarrage
