// Package bootstrap provides dependency injection wiring using Google Wire.
// It isolates all dependency construction from the main entry point,
// allowing for a minimal main.go and better testability.
package bootstrap

import (
	"context"
	"os"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
)

// tuiModeConfig holds configuration for TUI mode execution.
type tuiModeConfig struct {
	ctx             context.Context
	cancel          context.CancelFunc
	sigCh           <-chan os.Signal
	tui             Runner
	bufferedConsole Flusher
	tuiMode         tui.Mode
	sup             SignalHandler
}
