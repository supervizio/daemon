# Credentials - Résolution User/Group

Résolution des noms d'utilisateurs/groupes vers UID/GID et application aux commandes.

## Interface

```go
type CredentialManager interface {
    LookupUser(nameOrID string) (*User, error)
    LookupGroup(nameOrID string) (*Group, error)
    ResolveCredentials(username, groupname string) (uid, gid uint32, err error)
    ApplyCredentials(cmd *exec.Cmd, uid, gid uint32) error
}
```

## Types

```go
type User struct {
    UID      uint32
    GID      uint32
    Username string
    HomeDir  string
}

type Group struct {
    GID  uint32
    Name string
}
```

## Fichiers

| Fichier | Rôle |
|---------|------|
| `manager.go` | Interface + types + erreurs |
| `credentials_unix.go` | Lookup via `os/user` |
| `credentials_scratch.go` | Fallback numérique pour containers minimaux |

## Implémentations

### Unix (credentials_unix.go)

```go
func (m *Manager) LookupUser(nameOrID string) (*User, error) {
    u, err := user.Lookup(nameOrID)
    if err != nil {
        u, err = user.LookupId(nameOrID)
    }
    // ...
}
```

### Scratch (credentials_scratch.go)

Pour containers sans `/etc/passwd` :

```go
func (m *Manager) LookupUser(nameOrID string) (*User, error) {
    uid, err := strconv.ParseUint(nameOrID, 10, 32)
    if err != nil {
        return nil, ErrUserNotFound  // Noms non supportés
    }
    return &User{UID: uint32(uid)}, nil
}
```

## Erreurs

```go
var (
    ErrUserNotFound  = errors.New("user not found")
    ErrGroupNotFound = errors.New("group not found")
)
```

## Constructeur

```go
New() *Manager
```
