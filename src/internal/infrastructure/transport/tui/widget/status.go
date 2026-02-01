// Package widget provides reusable TUI components.
package widget

import (
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

// StatusIndicator renders state/health indicators with colors.
// It provides themed icons and text for process states, health status, and log levels.
type StatusIndicator struct {
	Theme ansi.Theme
	Icons ansi.StatusIcon
}

// NewStatusIndicator creates a new status indicator with default theme.
//
// Returns:
//   - *StatusIndicator: configured indicator with default theme
func NewStatusIndicator() *StatusIndicator {
	// Return configured status indicator with defaults.
	return &StatusIndicator{
		Theme: ansi.DefaultTheme(),
		Icons: ansi.DefaultIcons(),
	}
}

// ProcessState returns colored icon for process state.
//
// Params:
//   - state: the process state to render
//
// Returns:
//   - string: colored icon representing the state
func (s *StatusIndicator) ProcessState(state process.State) string {
	// Map process state to appropriate icon and color.
	switch state {
	// Running state.
	case process.StateRunning:
		// Return green success icon.
		return s.Theme.Success + s.Icons.Running + ansi.Reset
	// Starting state.
	case process.StateStarting:
		// Return yellow warning icon.
		return s.Theme.Warning + s.Icons.Starting + ansi.Reset
	// Stopped state.
	case process.StateStopped:
		// Return muted stopped icon.
		return s.Theme.Muted + s.Icons.Stopped + ansi.Reset
	// Stopping state.
	case process.StateStopping:
		// Return yellow warning icon.
		return s.Theme.Warning + s.Icons.Starting + ansi.Reset
	// Failed state.
	case process.StateFailed:
		// Return red error icon.
		return s.Theme.Error + s.Icons.Failed + ansi.Reset
	// Unknown state.
	default:
		// Return muted unknown icon.
		return s.Theme.Muted + s.Icons.Unknown + ansi.Reset
	}
}

// ProcessStateText returns colored state text.
//
// Params:
//   - state: the process state to render
//
// Returns:
//   - string: colored text representing the state
func (s *StatusIndicator) ProcessStateText(state process.State) string {
	// Map process state to colored text representation.
	switch state {
	// Running state.
	case process.StateRunning:
		// Return green text.
		return s.Theme.Success + "running" + ansi.Reset
	// Starting state.
	case process.StateStarting:
		// Return yellow text.
		return s.Theme.Warning + "starting" + ansi.Reset
	// Stopped state.
	case process.StateStopped:
		// Return muted text.
		return s.Theme.Muted + "stopped" + ansi.Reset
	// Stopping state.
	case process.StateStopping:
		// Return yellow text.
		return s.Theme.Warning + "stopping" + ansi.Reset
	// Failed state.
	case process.StateFailed:
		// Return red text.
		return s.Theme.Error + "failed" + ansi.Reset
	// Unknown state.
	default:
		// Return muted text.
		return s.Theme.Muted + "unknown" + ansi.Reset
	}
}

// ProcessStateShort returns short colored state text.
//
// Params:
//   - state: the process state to render
//
// Returns:
//   - string: short colored abbreviation
func (s *StatusIndicator) ProcessStateShort(state process.State) string {
	// Map process state to short colored abbreviation.
	switch state {
	// Running state.
	case process.StateRunning:
		// Return green abbreviation.
		return s.Theme.Success + "run" + ansi.Reset
	// Starting state.
	case process.StateStarting:
		// Return yellow abbreviation.
		return s.Theme.Warning + "start" + ansi.Reset
	// Stopped state.
	case process.StateStopped:
		// Return muted abbreviation.
		return s.Theme.Muted + "stop" + ansi.Reset
	// Stopping state.
	case process.StateStopping:
		// Return yellow text.
		return s.Theme.Warning + "stopping" + ansi.Reset
	// Failed state.
	case process.StateFailed:
		// Return red abbreviation.
		return s.Theme.Error + "fail" + ansi.Reset
	// Unknown state.
	default:
		// Return muted placeholder.
		return s.Theme.Muted + "?" + ansi.Reset
	}
}

// HealthStatus returns colored icon for health status.
//
// Params:
//   - status: the health status to render
//
// Returns:
//   - string: colored icon representing the status
func (s *StatusIndicator) HealthStatus(status health.Status) string {
	// Map health status to appropriate icon and color.
	switch status {
	// Healthy status.
	case health.StatusHealthy:
		// Return green success icon.
		return s.Theme.Success + s.Icons.Healthy + ansi.Reset
	// Unhealthy status.
	case health.StatusUnhealthy:
		// Return red error icon.
		return s.Theme.Error + s.Icons.Failed + ansi.Reset
	// Degraded status.
	case health.StatusDegraded:
		// Return yellow warning icon.
		return s.Theme.Warning + s.Icons.Starting + ansi.Reset
	// Unknown status.
	case health.StatusUnknown:
		// Return muted unknown icon.
		return s.Theme.Muted + s.Icons.Unknown + ansi.Reset
	// handle default case.
	default:
		// Default fallback: muted unknown icon.
		return s.Theme.Muted + s.Icons.Unknown + ansi.Reset
	}
}

// HealthStatusText returns colored health text.
//
// Params:
//   - status: the health status to render
//
// Returns:
//   - string: colored text representing the status
func (s *StatusIndicator) HealthStatusText(status health.Status) string {
	// Map health status to colored text representation.
	switch status {
	// Healthy status.
	case health.StatusHealthy:
		// Return green text.
		return s.Theme.Success + "healthy" + ansi.Reset
	// Unhealthy status.
	case health.StatusUnhealthy:
		// Return red text.
		return s.Theme.Error + "unhealthy" + ansi.Reset
	// Degraded status.
	case health.StatusDegraded:
		// Return yellow text.
		return s.Theme.Warning + "degraded" + ansi.Reset
	// Unknown status - explicit case for exhaustive switch.
	case health.StatusUnknown:
		// Return muted "unknown" text.
		return s.Theme.Muted + "unknown" + ansi.Reset
	// Future status values default to unknown display.
	default:
		// Fallback for any future status values.
		return s.Theme.Muted + "unknown" + ansi.Reset
	}
}

// ListenerState returns colored icon for listener state.
//
// Params:
//   - state: the listener state string
//
// Returns:
//   - string: colored icon representing the state
func (s *StatusIndicator) ListenerState(state string) string {
	// Map listener state to appropriate icon and color.
	switch state {
	// Ready state.
	case "ready":
		// Return green success icon.
		return s.Theme.Success + s.Icons.Running + ansi.Reset
	// Listening state.
	case "listening":
		// Return yellow warning icon.
		return s.Theme.Warning + s.Icons.Starting + ansi.Reset
	// Closed state.
	case "closed":
		// Return muted stopped icon.
		return s.Theme.Muted + s.Icons.Stopped + ansi.Reset
	// Unknown state.
	default:
		// Return muted unknown icon.
		return s.Theme.Muted + s.Icons.Unknown + ansi.Reset
	}
}

// LogLevel returns colored log level.
//
// Params:
//   - level: the log level string
//
// Returns:
//   - string: colored log level text
func (s *StatusIndicator) LogLevel(level string) string {
	// Map log level to appropriate color.
	switch level {
	// Info level.
	case "INFO", "INF":
		// Return green text.
		return s.Theme.Success + level + ansi.Reset
	// Warning level.
	case "WARN", "WRN":
		// Return yellow text.
		return s.Theme.Warning + level + ansi.Reset
	// Error level.
	case "ERROR", "ERR":
		// Return red text.
		return s.Theme.Error + level + ansi.Reset
	// Debug level.
	case "DEBUG", "DBG":
		// Return muted text.
		return s.Theme.Muted + level + ansi.Reset
	// Unknown level.
	default:
		// Return without color.
		return level
	}
}

// Bool returns colored boolean indicator.
//
// Params:
//   - value: boolean value to render
//
// Returns:
//   - string: colored icon for true/false
func (s *StatusIndicator) Bool(value bool) string {
	// Map boolean to colored icon.
	if value {
		// True: green success icon.
		return s.Theme.Success + s.Icons.Healthy + ansi.Reset
	}
	// False: red error icon.
	return s.Theme.Error + s.Icons.Failed + ansi.Reset
}

// Detected returns colored detection status.
//
// Params:
//   - detected: whether the item was detected
//
// Returns:
//   - string: colored icon for detection status
func (s *StatusIndicator) Detected(detected bool) string {
	// Map detection status to colored icon.
	if detected {
		// Detected: green running icon.
		return s.Theme.Success + s.Icons.Running + ansi.Reset
	}
	// Not detected: muted stopped icon.
	return s.Theme.Muted + s.Icons.Stopped + ansi.Reset
}
