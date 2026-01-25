// Package widget provides reusable TUI components.
package widget

import (
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
)

// StatusIndicator renders state/health indicators with colors.
type StatusIndicator struct {
	Theme ansi.Theme
	Icons ansi.StatusIcon
}

// NewStatusIndicator creates a new status indicator with default theme.
func NewStatusIndicator() *StatusIndicator {
	return &StatusIndicator{
		Theme: ansi.DefaultTheme(),
		Icons: ansi.DefaultIcons(),
	}
}

// ProcessState returns colored icon for process state.
func (s *StatusIndicator) ProcessState(state process.State) string {
	switch state {
	case process.StateRunning:
		return s.Theme.Success + s.Icons.Running + ansi.Reset
	case process.StateStarting:
		return s.Theme.Warning + s.Icons.Starting + ansi.Reset
	case process.StateStopped:
		return s.Theme.Muted + s.Icons.Stopped + ansi.Reset
	case process.StateStopping:
		return s.Theme.Warning + s.Icons.Starting + ansi.Reset
	case process.StateFailed:
		return s.Theme.Error + s.Icons.Failed + ansi.Reset
	default:
		return s.Theme.Muted + s.Icons.Unknown + ansi.Reset
	}
}

// ProcessStateText returns colored state text.
func (s *StatusIndicator) ProcessStateText(state process.State) string {
	switch state {
	case process.StateRunning:
		return s.Theme.Success + "running" + ansi.Reset
	case process.StateStarting:
		return s.Theme.Warning + "starting" + ansi.Reset
	case process.StateStopped:
		return s.Theme.Muted + "stopped" + ansi.Reset
	case process.StateStopping:
		return s.Theme.Warning + "stopping" + ansi.Reset
	case process.StateFailed:
		return s.Theme.Error + "failed" + ansi.Reset
	default:
		return s.Theme.Muted + "unknown" + ansi.Reset
	}
}

// ProcessStateShort returns short colored state text.
func (s *StatusIndicator) ProcessStateShort(state process.State) string {
	switch state {
	case process.StateRunning:
		return s.Theme.Success + "run" + ansi.Reset
	case process.StateStarting:
		return s.Theme.Warning + "start" + ansi.Reset
	case process.StateStopped:
		return s.Theme.Muted + "stop" + ansi.Reset
	case process.StateStopping:
		return s.Theme.Warning + "stopping" + ansi.Reset
	case process.StateFailed:
		return s.Theme.Error + "fail" + ansi.Reset
	default:
		return s.Theme.Muted + "?" + ansi.Reset
	}
}

// HealthStatus returns colored icon for health status.
func (s *StatusIndicator) HealthStatus(status health.Status) string {
	switch status {
	case health.StatusHealthy:
		return s.Theme.Success + s.Icons.Healthy + ansi.Reset
	case health.StatusUnhealthy:
		return s.Theme.Error + s.Icons.Failed + ansi.Reset
	case health.StatusDegraded:
		return s.Theme.Warning + s.Icons.Starting + ansi.Reset
	case health.StatusUnknown:
		return s.Theme.Muted + s.Icons.Unknown + ansi.Reset
	}
	return s.Theme.Muted + s.Icons.Unknown + ansi.Reset
}

// HealthStatusText returns colored health text.
func (s *StatusIndicator) HealthStatusText(status health.Status) string {
	switch status {
	case health.StatusHealthy:
		return s.Theme.Success + "healthy" + ansi.Reset
	case health.StatusUnhealthy:
		return s.Theme.Error + "unhealthy" + ansi.Reset
	case health.StatusDegraded:
		return s.Theme.Warning + "degraded" + ansi.Reset
	case health.StatusUnknown:
		return s.Theme.Muted + "unknown" + ansi.Reset
	}
	return s.Theme.Muted + "unknown" + ansi.Reset
}

// ListenerState returns colored icon for listener state.
func (s *StatusIndicator) ListenerState(state string) string {
	switch state {
	case "ready":
		return s.Theme.Success + s.Icons.Running + ansi.Reset
	case "listening":
		return s.Theme.Warning + s.Icons.Starting + ansi.Reset
	case "closed":
		return s.Theme.Muted + s.Icons.Stopped + ansi.Reset
	default:
		return s.Theme.Muted + s.Icons.Unknown + ansi.Reset
	}
}

// LogLevel returns colored log level.
func (s *StatusIndicator) LogLevel(level string) string {
	switch level {
	case "INFO", "INF":
		return s.Theme.Success + level + ansi.Reset
	case "WARN", "WRN":
		return s.Theme.Warning + level + ansi.Reset
	case "ERROR", "ERR":
		return s.Theme.Error + level + ansi.Reset
	case "DEBUG", "DBG":
		return s.Theme.Muted + level + ansi.Reset
	default:
		return level
	}
}

// Bool returns colored boolean indicator.
func (s *StatusIndicator) Bool(value bool) string {
	if value {
		return s.Theme.Success + s.Icons.Healthy + ansi.Reset
	}
	return s.Theme.Error + s.Icons.Failed + ansi.Reset
}

// Detected returns colored detection status.
func (s *StatusIndicator) Detected(detected bool) string {
	if detected {
		return s.Theme.Success + s.Icons.Running + ansi.Reset
	}
	return s.Theme.Muted + s.Icons.Stopped + ansi.Reset
}
