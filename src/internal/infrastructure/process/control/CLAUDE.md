# Control - Process Groups

Gestion des process groups pour le signal forwarding.

## Contexte

Quand on envoie SIGTERM à un processus, on veut aussi toucher ses enfants. Les process groups permettent ça.

## Interface

```go
type ProcessControl interface {
    SetProcessGroup(cmd *exec.Cmd)
    GetProcessGroup(pid int) (int, error)
}
```

## Fichiers

| Fichier | Rôle |
|---------|------|
| `control.go` | Interface `ProcessControl` |
| `process_unix.go` | Implémentation via `syscall.Setpgid` |

## Implémentation

```go
func (c *Control) SetProcessGroup(cmd *exec.Cmd) {
    cmd.SysProcAttr = &syscall.SysProcAttr{
        Setpgid: true,  // Nouveau process group
    }
}

func (c *Control) GetProcessGroup(pid int) (int, error) {
    pgid, err := syscall.Getpgid(pid)
    if err != nil {
        return 0, process.WrapError("getpgid", err)
    }
    return pgid, nil
}
```

## Usage dans Executor

```go
func (e *Executor) buildCommand(ctx, spec) (*exec.Cmd, error) {
    cmd := TrustedCommand(ctx, parts[0], args...)
    e.process.SetProcessGroup(cmd)  // ← Ici
    return cmd, nil
}
```

## Constructeur

```go
New() *Control
```
