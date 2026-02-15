<!-- updated: 2026-02-15T21:30:00Z -->
# Terminal - Size Detection

Cross-platform terminal size detection using ioctl syscalls.

## Files

| File | Purpose |
|------|---------|
| `size.go` | Size type, detection logic, resize handling |
| `size_linux.go` | Linux ioctl constants (TIOCGWINSZ) |
| `size_darwin.go` | macOS ioctl constants |
| `size_bsd.go` | BSD ioctl constants |
| `size_unix.go` | Unix implementation (shared logic) |

## Key Types

| Type | Description |
|------|-------------|
| `Size` | Terminal dimensions (Width, Height) |
| `SizeWatcher` | Monitors terminal resize events (SIGWINCH) |

## Build Tags

```go
//go:build linux   // size_linux.go
//go:build darwin  // size_darwin.go
//go:build freebsd || openbsd || netbsd  // size_bsd.go
//go:build unix    // size_unix.go (shared)
```

## Usage

```go
size := terminal.GetSize()
fmt.Printf("Terminal: %dx%d\n", size.Width, size.Height)

watcher := terminal.NewSizeWatcher()
for newSize := range watcher.Changes() {
    // Handle resize
}
```

## Related

| Package | Relation |
|---------|----------|
| `tui/layout` | Uses Size for layout calculations |
| `tui/screen` | Uses SizeWatcher for redraw |
