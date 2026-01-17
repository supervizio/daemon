# Reaper - Récupération des Zombies

Boucle de nettoyage des processus zombies quand on est PID1.

## Contexte

Quand le daemon tourne comme PID1 (dans un container), les processus orphelins lui sont réassignés. Sans reaper, ils deviennent zombies.

## Interface

```go
type ZombieReaper interface {
    Start()           // Démarre la boucle background
    Stop()            // Arrête la boucle
    ReapOnce() int    // Un cycle manuel (pour tests)
    IsPID1() bool     // Vrai si PID == 1
}
```

## Fichiers

| Fichier | Rôle |
|---------|------|
| `zombie_reaper.go` | Interface `ZombieReaper` |
| `reaper_unix.go` | Implémentation avec `waitpid(-1, WNOHANG)` |

## Fonctionnement

```go
func (r *Reaper) Start() {
    go func() {
        for {
            select {
            case <-r.stopCh:
                return
            case <-ticker.C:
                r.ReapOnce()
            }
        }
    }()
}

func (r *Reaper) ReapOnce() int {
    count := 0
    for {
        pid, _ := syscall.Wait4(-1, nil, syscall.WNOHANG, nil)
        if pid <= 0 {
            break
        }
        count++
    }
    return count
}
```

## Constructeur

```go
New() *Reaper
```

## Activation Conditionnelle

Dans `bootstrap/providers.go` :

```go
func ProvideReaper(r *reaper.Reaper) reaper.ZombieReaper {
    if r.IsPID1() {
        return r
    }
    return nil  // Pas de reaper si pas PID1
}
```
