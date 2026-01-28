// Package tui provides terminal user interface rendering for superviz.io.
package tui

import "github.com/kodflow/daemon/internal/domain/process"

// TUISnapshotData contains service data for TUI display.
// It provides a minimal view of service state optimized for terminal rendering.
type TUISnapshotData struct {
	Name   string
	State  process.State
	PID    int
	Uptime int64
}
