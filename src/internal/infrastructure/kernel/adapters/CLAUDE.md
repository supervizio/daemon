# Adapters - Platform Implementations

Platform-specific implementations.

## Structure

```
adapters/
├── signals_linux.go         # Linux signal handling
├── signals_unix.go          # Unix base implementation
├── signals_bsd.go           # BSD variants
├── signals_darwin.go        # macOS specific
├── reaper_unix.go           # Zombie reaping (all Unix)
├── process_unix.go          # Process groups (all Unix)
├── credentials_unix.go      # User/group (standard Unix with /etc/passwd)
└── credentials_scratch.go   # User/group (scratch containers, numeric only)
```

## Build Tags

Each file uses build constraints:

```go
//go:build linux
//go:build darwin
//go:build freebsd || openbsd || netbsd
```

## Implementations

### signals_linux.go

- `SetSubreaper()`: prctl(PR_SET_CHILD_SUBREAPER)
- Allows orphan process adoption without being PID 1

### signals_bsd.go

- Signal handling via syscall.Wait4
- No subreaper (Linux-only feature)

### signals_darwin.go

- Adapted for macOS specifics
- Kqueue-based signal handling

### reaper_unix.go

```go
func (r *UnixReaper) ReapOnce() (int, error) {
    var status syscall.WaitStatus
    pid, err := syscall.Wait4(-1, &status, syscall.WNOHANG, nil)
    return pid, err
}
```

### credentials_unix.go

```go
func ResolveCredentials(user, group string) (*Credentials, error) {
    u, _ := user.Lookup(user)
    g, _ := user.LookupGroup(group)
    return &Credentials{Uid: u.Uid, Gid: g.Gid}, nil
}
```

### credentials_scratch.go

For scratch containers without `/etc/passwd`:

```go
// Only supports numeric UIDs/GIDs
func (m *ScratchCredentialManager) LookupUser(nameOrID string) (*ports.User, error) {
    uid, err := strconv.ParseUint(nameOrID, 10, 32)
    if err != nil {
        return nil, ErrScratchNameLookup // name lookup not available
    }
    return &ports.User{UID: uint32(uid), GID: uint32(uid)}, nil
}

// Detect scratch environment
func IsScratchEnvironment() bool {
    _, err := os.Stat("/etc/passwd")
    return err != nil
}
```

## Related Directories

| Directory | Relation |
|-----------|----------|
| `../ports/` | Implements these interfaces |
| `../../process/` | Uses these adapters |
| `../../../application/process/` | Application layer consumer |
