# Executor - Exécution de Processus

Implémente `domain.Executor` : démarrer, arrêter, signaler des processus.

## Interface

```go
type Executor interface {
    Start(ctx, spec) (pid int, wait <-chan ExitResult, err error)
    Stop(pid, timeout) error
    Signal(pid, sig) error
}
```

## Fichiers

| Fichier | Rôle |
|---------|------|
| `executor.go` | Implémentation Start/Stop/Signal |
| `command.go` | `TrustedCommand()` - wrapper exec sécurisé |
| `os_process_wrapper.go` | Abstraction os.Process pour tests |

## Constructeurs

```go
New()                      // Standalone
NewWithDeps(creds, ctrl)   // Wire DI
NewWithOptions(...)        // Tests avec mocks
```

## Sécurité

Toute commande passe par `TrustedCommand()` :

```go
func TrustedCommand(ctx, name, args...) *exec.Cmd
```

**Modèle de confiance** : Commandes viennent de YAML admin, jamais d'input user.

## Dépendances

- `credentials.CredentialManager` : résolution user/group
- `control.ProcessControl` : configuration process group
