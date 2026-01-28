// Package tui provides terminal user interface for superviz.io.
package tui

import (
	"io"
	"os"
	"time"
)

// Config holds TUI configuration.
// It specifies the operating mode, refresh interval, version, and output writer.
type Config struct {
	// Mode is the operating mode (raw or interactive).
	Mode Mode
	// RefreshInterval is the update frequency for interactive mode.
	// Minimum: 1 second.
	RefreshInterval time.Duration
	// Version is the daemon version string.
	Version string
	// Output is the writer for raw mode output.
	Output io.Writer
}

// DefaultConfig returns the default configuration.
//
// Params:
//   - version: the daemon version string.
//
// Returns:
//   - Config: the default configuration.
func DefaultConfig(version string) Config {
	// Return config with default values.
	return Config{
		Mode:            ModeRaw,
		RefreshInterval: defaultRefreshInterval,
		Version:         version,
		Output:          os.Stdout,
	}
}
