# Domain Shared Package

Common value objects, interfaces, and constants used across multiple domain packages.

This package provides foundational types that are shared across the domain layer,
ensuring consistent handling of durations, sizes, time, and common errors without
creating circular dependencies.

## Files

| File | Purpose |
|------|---------|
| `duration.go` | `Duration` value object - time duration wrapper |
| `size.go` | Size parsing and formatting utilities |
| `clock.go` | `Nower` interface, `RealClock` for time abstraction |
| `constants.go` | Shared constants (e.g., MaxValidPort) |
| `errors.go` | Common domain errors |

## Key Types

### Duration (Value Object)

Domain wrapper around `time.Duration` for clean domain modeling.

```go
type Duration time.Duration
```

**Methods:**
- `Duration()` - Get underlying time.Duration
- `Seconds()` - Get duration in seconds (float64)
- `Milliseconds()` - Get duration in milliseconds (int64)
- `String()` - Human-readable format

**Factory Functions:**
- `Seconds(s int)` - Create Duration from seconds
- `Minutes(m int)` - Create Duration from minutes
- `FromTimeDuration(d)` - Convert from time.Duration

### Size Utilities

Parse and format human-readable size strings.

**Constants:**
```go
const (
    Byte     int64 = 1
    Kilobyte int64 = 1024
    Megabyte int64 = 1024 * Kilobyte
    Gigabyte int64 = 1024 * Megabyte
)
```

**Functions:**
- `ParseSize(s string)` - Parse "100MB", "1GB" to bytes
- `FormatSize(bytes int64)` - Format bytes to human-readable

**Supported Formats:**
- `"100"` - bytes (no suffix)
- `"100B"` - bytes
- `"100KB"` - kilobytes
- `"100MB"` - megabytes
- `"100GB"` - gigabytes

**Size Errors:**
```go
var (
    ErrEmptySize    // Empty size string
    ErrNegativeSize // Negative size value
)
```

### Clock (Interface + Implementation)

Abstraction for time operations, enabling deterministic testing.

```go
type Nower interface {
    Now() time.Time
}

type RealClock struct{}  // Uses system time
var DefaultClock Nower   // Global default
```

**Functions:**
- `NewRealClock()` - Create real clock instance

### Constants

```go
const (
    MaxValidPort int = 65535  // Maximum TCP/UDP port
)
```

### Domain Errors

Common sentinel errors for domain operations.

```go
var (
    ErrNotFound        // Resource not found
    ErrAlreadyExists   // Resource already exists
    ErrInvalidState    // Invalid state transition
    ErrInvalidArgument // Invalid function argument
    ErrEmptyCommand    // Empty command configuration
)
```

## Usage Examples

### Duration
```go
// Create durations
timeout := shared.Seconds(30)
interval := shared.Minutes(5)

// Use in time operations
time.Sleep(timeout.Duration())

// Get numeric values
fmt.Printf("Timeout: %.0f seconds\n", timeout.Seconds())
```

### Size Parsing
```go
// Parse size strings
bytes, err := shared.ParseSize("100MB")
// bytes = 104857600

// Format sizes
str := shared.FormatSize(1073741824)
// str = "1GB"
```

### Clock for Testing
```go
// Production code
type Service struct {
    clock shared.Nower
}

func (s *Service) GetCurrentTime() time.Time {
    return s.clock.Now()
}

// Test code
type MockClock struct {
    fixedTime time.Time
}

func (m MockClock) Now() time.Time {
    return m.fixedTime
}
```

## Dependencies

- Depends on: nothing (pure domain)
- Used by: `domain/config`, `domain/process`, `domain/health`, and most domain packages

## Related Packages

| Package | Relation |
|---------|----------|
| `domain/config` | Uses Duration for restart delays, timeouts |
| `domain/health` | Uses Duration for check intervals |
| `domain/process` | Uses Duration for stop timeouts |
| `infrastructure/persistence/config/yaml` | Parses Duration and Size from YAML |
