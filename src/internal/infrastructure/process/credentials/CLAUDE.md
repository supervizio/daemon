<!-- updated: 2026-02-15T21:30:00Z -->
# Credentials - User/Group Resolution

Resolve user/group names to UID/GID and apply to commands.

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

## Files

| File | Role |
|------|------|
| `manager.go` | Interface + types + errors |
| `credentials_unix.go` | Lookup via `os/user` |
| `credentials_scratch.go` | Numeric fallback for minimal containers |

## Implementations

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

For containers without `/etc/passwd`:

```go
func (m *Manager) LookupUser(nameOrID string) (*User, error) {
    uid, err := strconv.ParseUint(nameOrID, 10, 32)
    if err != nil {
        return nil, ErrUserNotFound  // Names not supported
    }
    return &User{UID: uint32(uid)}, nil
}
```

## Errors

```go
var (
    ErrUserNotFound  = errors.New("user not found")
    ErrGroupNotFound = errors.New("group not found")
)
```

## Constructor

```go
New() *Manager
```
