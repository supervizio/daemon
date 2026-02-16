<!-- updated: 2026-02-15T21:30:00Z -->
# probe-quota - Resource Quota Management

Cross-platform resource quota detection and management.

## Structure

```
src/
├── lib.rs        # QuotaCollector trait and types
├── linux.rs      # cgroups v1/v2 quota detection
├── freebsd.rs    # FreeBSD rctl/jail quota detection
└── rlimit.rs     # POSIX getrlimit (macOS, OpenBSD, NetBSD)
```

## Platform Support

| Platform | Mechanism | Detection |
|----------|-----------|-----------|
| Linux | cgroups v1/v2 | `/sys/fs/cgroup` |
| macOS | getrlimit | POSIX rlimit |
| FreeBSD | rctl/jail | `sysctl` |
| OpenBSD | getrlimit | POSIX rlimit |
| NetBSD | getrlimit | POSIX rlimit |

## Related

| Crate | Relation |
|-------|----------|
| `probe-platform` | Uses quota info for resource limits |
| `probe-ffi` | Exposes quota data via FFI |
