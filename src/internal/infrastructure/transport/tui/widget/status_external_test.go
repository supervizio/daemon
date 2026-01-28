package widget_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

func TestNewStatusIndicator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "creates_status_indicator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			assert.NotNil(t, indicator)
		})
	}
}

func TestStatusIndicator_ProcessState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
	}{
		{
			name:  "running_state",
			state: process.StateRunning,
		},
		{
			name:  "starting_state",
			state: process.StateStarting,
		},
		{
			name:  "stopped_state",
			state: process.StateStopped,
		},
		{
			name:  "stopping_state",
			state: process.StateStopping,
		},
		{
			name:  "failed_state",
			state: process.StateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.ProcessState(tt.state)
			assert.NotEmpty(t, result)
		})
	}
}

func TestStatusIndicator_ProcessStateText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
	}{
		{
			name:  "running_text",
			state: process.StateRunning,
		},
		{
			name:  "starting_text",
			state: process.StateStarting,
		},
		{
			name:  "stopped_text",
			state: process.StateStopped,
		},
		{
			name:  "stopping_text",
			state: process.StateStopping,
		},
		{
			name:  "failed_text",
			state: process.StateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.ProcessStateText(tt.state)
			assert.NotEmpty(t, result)
		})
	}
}

func TestStatusIndicator_ProcessStateShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
	}{
		{
			name:  "running_short",
			state: process.StateRunning,
		},
		{
			name:  "starting_short",
			state: process.StateStarting,
		},
		{
			name:  "stopped_short",
			state: process.StateStopped,
		},
		{
			name:  "stopping_short",
			state: process.StateStopping,
		},
		{
			name:  "failed_short",
			state: process.StateFailed,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.ProcessStateShort(tt.state)
			assert.NotEmpty(t, result)
		})
	}
}

func TestStatusIndicator_HealthStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status health.Status
	}{
		{
			name:   "healthy_status",
			status: health.StatusHealthy,
		},
		{
			name:   "unhealthy_status",
			status: health.StatusUnhealthy,
		},
		{
			name:   "degraded_status",
			status: health.StatusDegraded,
		},
		{
			name:   "unknown_status",
			status: health.StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.HealthStatus(tt.status)
			assert.NotEmpty(t, result)
		})
	}
}

func TestStatusIndicator_HealthStatusText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status health.Status
	}{
		{
			name:   "healthy_text",
			status: health.StatusHealthy,
		},
		{
			name:   "unhealthy_text",
			status: health.StatusUnhealthy,
		},
		{
			name:   "degraded_text",
			status: health.StatusDegraded,
		},
		{
			name:   "unknown_text",
			status: health.StatusUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.HealthStatusText(tt.status)
			assert.NotEmpty(t, result)
		})
	}
}

func TestStatusIndicator_ListenerState(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state string
	}{
		{
			name:  "ready_state",
			state: "ready",
		},
		{
			name:  "listening_state",
			state: "listening",
		},
		{
			name:  "closed_state",
			state: "closed",
		},
		{
			name:  "unknown_state",
			state: "unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.ListenerState(tt.state)
			assert.NotEmpty(t, result)
		})
	}
}

func TestStatusIndicator_LogLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level string
	}{
		{
			name:  "info_level",
			level: "INFO",
		},
		{
			name:  "warn_level",
			level: "WARN",
		},
		{
			name:  "error_level",
			level: "ERROR",
		},
		{
			name:  "debug_level",
			level: "DEBUG",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.LogLevel(tt.level)
			assert.NotEmpty(t, result)
		})
	}
}

func TestStatusIndicator_Bool(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value bool
	}{
		{
			name:  "true_value",
			value: true,
		},
		{
			name:  "false_value",
			value: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.Bool(tt.value)
			assert.NotEmpty(t, result)
		})
	}
}

func TestStatusIndicator_Detected(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		detected bool
	}{
		{
			name:     "detected",
			detected: true,
		},
		{
			name:     "not_detected",
			detected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			indicator := widget.NewStatusIndicator()
			result := indicator.Detected(tt.detected)
			assert.NotEmpty(t, result)
		})
	}
}
