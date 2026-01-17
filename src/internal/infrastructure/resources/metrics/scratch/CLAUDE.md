# Scratch Metrics - Stub

Implémentation stub pour environnements sans métriques.

## Usage

- CI/CD sans accès OS
- Tests unitaires
- Containers scratch
- Plateformes inconnues

## Comportement

Toutes les méthodes retournent des valeurs zéro sans erreur (ou `ErrNotSupported`).

## Structure

| Fichier | Rôle |
|---------|------|
| `probe.go` | `Probe` retournant des zéros |

## Constructeur

```go
NewProbe() *Probe
```

## Build Tag

```go
//go:build !linux && !darwin && !freebsd && !openbsd && !netbsd && !dragonfly
```

## Détection

```go
cpu, err := collector.CPU().CollectSystem(ctx)
if errors.Is(err, scratch.ErrNotSupported) {
    // Plateforme non supportée
}
```
