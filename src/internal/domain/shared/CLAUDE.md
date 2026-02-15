<!-- updated: 2026-02-15T21:30:00Z -->
# Domain Shared Package

Common value objects, interfaces, and constants used across domain packages.

## Files

| File | Purpose |
|------|---------|
| `duration.go` | `Duration` value object - time duration wrapper |
| `size.go` | Size parsing and formatting (ParseSize, FormatSize) |
| `clock.go` | `Nower` interface, `RealClock` for time abstraction |
| `filesystem.go` | `FileSystem` interface for OS file operations |
| `constants.go` | Shared constants (network, numeric, unit conversion) |
| `errors.go` | Common domain errors |

## Key Types

### Duration
- `Duration` - Wrapper around `time.Duration`
- Factory: `Seconds(s int)`, `Minutes(m int)`, `FromTimeDuration(d)`
- Methods: `Duration()`, `Seconds()`, `Milliseconds()`, `String()`

### Size
- Constants: `Byte`, `Kilobyte`, `Megabyte`, `Gigabyte`
- `ParseSize(s string)` - Parse "100MB", "1GB" to bytes
- `FormatSize(bytes int64)` - Format bytes to human-readable

### Clock
- `Nower` interface with `Now() time.Time`
- `RealClock` - System time implementation
- `DefaultClock` - Global default

### FileSystem
- `FileSystem` interface: `Stat(name)`, `ReadFile(name)`
- `OSFileSystem` - Real OS implementation
- `DefaultFileSystem` - Global default

### Constants
- `MaxValidPort` (65535) - Maximum valid TCP/UDP port
- `Base10`, `BitSize64` - Numeric parsing constants
- `PercentMultiplier` (100.0) - Ratio to percentage conversion
- `BitsPerByte` (8) - Bits per byte

## Domain Errors

```go
var (
    ErrNotFound        // Resource not found
    ErrAlreadyExists   // Resource already exists
    ErrInvalidState    // Invalid state transition
    ErrInvalidArgument // Invalid function argument
    ErrEmptyCommand    // Empty command configuration
    ErrEmptySize       // Empty size string
    ErrNegativeSize    // Negative size value
)
```

## Dependencies

- Depends on: nothing (pure domain)
- Used by: `domain/config`, `domain/process`, `domain/health`, `domain/metrics`

## Related Packages

| Package | Usage |
|---------|-------|
| `domain/config` | Duration for timeouts, delays |
| `domain/health` | Duration for check intervals |
| `infrastructure/persistence/config/yaml` | Parses Duration and Size |
| `infrastructure/probe` | Uses FileSystem (quota.go) |
