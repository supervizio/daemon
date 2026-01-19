# Process - Gestion Processus OS

Tout ce qui touche à l'exécution et au contrôle des processus Unix.

## Rôle

Abstraire les opérations OS : démarrer/arrêter des processus, envoyer des signaux, gérer les credentials, nettoyer les zombies.

## Navigation

| Besoin | Package |
|--------|---------|
| Démarrer/arrêter un processus | `executor/` |
| Envoyer des signaux (SIGTERM, SIGHUP) | `signals/` |
| Récupérer les processus zombies (PID1) | `reaper/` |
| Résoudre user/group vers UID/GID | `credentials/` |
| Gérer les process groups | `control/` |

## Structure

```
process/
├── errors.go       # Erreurs partagées par tous les sous-packages
├── executor/       # Start(), Stop(), Signal()
├── signals/        # Notification, forwarding, subreaper
├── reaper/         # Boucle waitpid() pour PID1
├── credentials/    # LookupUser(), ApplyCredentials()
└── control/        # SetProcessGroup(), GetProcessGroup()
```

## Erreurs Partagées (errors.go)

```go
var (
    ErrProcessNotFound  = errors.New("process not found")
    ErrPermissionDenied = errors.New("permission denied")
    ErrNotSupported     = errors.New("operation not supported")
)

func WrapError(op string, err error) error  // Ajoute contexte
```

## Flux Principal

```
Supervisor
    │
    ▼
executor.Start(spec)
    ├── credentials.ResolveCredentials(user, group)
    ├── credentials.ApplyCredentials(cmd, uid, gid)
    ├── control.SetProcessGroup(cmd)
    └── cmd.Start()

executor.Stop(pid, timeout)
    └── signals.Forward(pid, SIGTERM)
        └── si timeout → Kill()

reaper.Start()  ← Tourne en background si PID1
    └── waitpid(-1) en boucle
```
