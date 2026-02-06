// Package tui provides terminal user interface for superviz.io.
package tui

import (
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/component"
)

// ModelConfig holds configuration for creating a new Model.
// Grouping parameters in a struct to comply with FUNC-MAXPARAM (max 5 params).
// Large fields use pointers to avoid copies when passed by value.
type ModelConfig struct {
	TUI           *TUI
	Width, Height int
	Theme         *ansi.Theme
	LogsPanel     *component.LogsPanel
	ServicesPanel *component.ServicesPanel
}
