<!-- updated: 2026-02-15T21:30:00Z -->
# probe-ffi - C ABI Interface

FFI layer exposing Rust probe functions as C ABI for Go CGO bindings.

## Structure

```
src/
└── lib.rs      # All extern "C" function exports
```

## Role

Bridge between Go (CGO) and Rust. Exports functions as `staticlib` for linking.

## Output

- Library type: `staticlib` (libprobe.a)
- Header: `../../include/probe.h`

## Related

| Location | Relation |
|----------|----------|
| `../../include/probe.h` | C header for Go CGO |
| `internal/infrastructure/probe/` | Go CGO bindings |
