// Package component provides internal tests for the services component.
package component

import (
	"strings"
	"testing"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// TestStateColorAndText tests the stateColorAndText helper function.
func TestStateColorAndText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		state        process.State
		expectedText string
	}{
		{
			name:         "running state returns running text",
			state:        process.StateRunning,
			expectedText: "running",
		},
		{
			name:         "stopped state returns stopped text",
			state:        process.StateStopped,
			expectedText: "stopped",
		},
		{
			name:         "failed state returns failed text",
			state:        process.StateFailed,
			expectedText: "failed",
		},
		{
			name:         "starting state returns starting text",
			state:        process.StateStarting,
			expectedText: "starting",
		},
		{
			name:         "stopping state returns stopping text",
			state:        process.StateStopping,
			expectedText: "stopping",
		},
		{
			name:         "unknown state returns unknown text",
			state:        process.State(99),
			expectedText: "unknown",
		},
	}

	theme := ansi.DefaultTheme()

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			colorFn, text := stateColorAndText(tc.state)

			assert.Equal(t, tc.expectedText, text)
			assert.NotNil(t, colorFn)

			// Verify color function returns valid theme color.
			color := colorFn(&theme)
			assert.NotEmpty(t, color)
		})
	}
}

// TestUpdateContent tests the updateContent method.
func TestServicesPanel_updateContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []model.ServiceSnapshot
	}{
		{
			name:     "empty services",
			services: nil,
		},
		{
			name: "single service",
			services: []model.ServiceSnapshot{
				{Name: "test", State: process.StateRunning},
			},
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 20)
			panel.services = tc.services
			panel.updateContent()
			// Verify method completes without error.
			assert.NotNil(t, panel.viewport)
		})
	}
}

// TestFormatServiceLine tests the formatServiceLine method.
func TestServicesPanel_formatServiceLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "running service",
			svc:  model.ServiceSnapshot{Name: "test", State: process.StateRunning, PID: 123},
		},
		{
			name: "stopped service",
			svc:  model.ServiceSnapshot{Name: "stopped", State: process.StateStopped},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.formatServiceLine(tc.svc)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFormatServiceName tests the formatServiceName method.
func TestServicesPanel_formatServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "short name",
			input:    "test",
			expected: "test",
		},
		{
			name:  "long name gets truncated",
			input: "very_long_service_name_here",
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.formatServiceName(tc.input)
			assert.NotEmpty(t, result)
		})
	}
}

// TestCollectServiceColumns tests the collectServiceColumns method.
func TestServicesPanel_collectServiceColumns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "running service",
			svc:  model.ServiceSnapshot{Name: "test", State: process.StateRunning, PID: 123},
		},
		{
			name: "stopped service with health checks",
			svc:  model.ServiceSnapshot{Name: "test", State: process.StateStopped, HasHealthChecks: true},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.collectServiceColumns(tc.svc)
			assert.NotEmpty(t, result.stateIcon)
			assert.NotEmpty(t, result.stateText)
		})
	}
}

// TestBuildServiceLineString tests the buildServiceLineString method.
func TestServicesPanel_buildServiceLineString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		cols serviceColumns
	}{
		{
			name: "basic columns",
			cols: serviceColumns{
				stateIcon:  "o",
				stateText:  "running",
				healthText: "healthy",
				uptime:     "1h",
				pid:        "123",
				restarts:   "0",
				cpu:        "5%",
				mem:        "1MB",
				ports:      ":8080",
			},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.buildServiceLineString("test", &tc.cols)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFormatUptime tests the formatUptime method.
func TestServicesPanel_formatUptime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "running with uptime",
			svc:  model.ServiceSnapshot{State: process.StateRunning, Uptime: 3600000000000},
		},
		{
			name: "stopped",
			svc:  model.ServiceSnapshot{State: process.StateStopped},
		},
		{
			name: "running no uptime",
			svc:  model.ServiceSnapshot{State: process.StateRunning, Uptime: 0},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.formatUptime(tc.svc)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFormatPID tests the formatPID method.
func TestServicesPanel_formatPID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "with PID",
			svc:  model.ServiceSnapshot{PID: 1234},
		},
		{
			name: "no PID",
			svc:  model.ServiceSnapshot{PID: 0},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.formatPID(tc.svc)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFormatRestarts tests the formatRestarts method.
func TestServicesPanel_formatRestarts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "with restarts",
			svc:  model.ServiceSnapshot{RestartCount: 5},
		},
		{
			name: "no restarts",
			svc:  model.ServiceSnapshot{RestartCount: 0},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.formatRestarts(tc.svc)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFormatCPU tests the formatCPU method.
func TestServicesPanel_formatCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "running with CPU",
			svc:  model.ServiceSnapshot{State: process.StateRunning, CPUPercent: 5.5},
		},
		{
			name: "stopped",
			svc:  model.ServiceSnapshot{State: process.StateStopped},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.formatCPU(tc.svc)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFormatMemory tests the formatMemory method.
func TestServicesPanel_formatMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "running with memory",
			svc:  model.ServiceSnapshot{State: process.StateRunning, MemoryRSS: 1024000},
		},
		{
			name: "stopped",
			svc:  model.ServiceSnapshot{State: process.StateStopped},
		},
		{
			name: "running no memory",
			svc:  model.ServiceSnapshot{State: process.StateRunning, MemoryRSS: 0},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.formatMemory(tc.svc)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFormatPorts tests the formatPorts package function.
func Test_formatPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		ports []int
	}{
		{
			name:  "no ports",
			ports: []int{},
		},
		{
			name:  "single port",
			ports: []int{8080},
		},
		{
			name:  "multiple ports",
			ports: []int{80, 443, 8080},
		},
		{
			name:  "many ports truncated",
			ports: []int{80, 443, 8080, 9000, 9001, 9002},
		},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := formatPorts(tc.ports)
			assert.NotEmpty(t, result)
		})
	}
}

// TestFormatPortsWithStatus tests the formatPortsWithStatus method.
func TestServicesPanel_formatPortsWithStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "no listeners no ports",
			svc:  model.ServiceSnapshot{},
		},
		{
			name: "detected ports only",
			svc:  model.ServiceSnapshot{Ports: []int{8080}},
		},
		{
			name: "with listeners",
			svc: model.ServiceSnapshot{
				Listeners: []model.ListenerSnapshot{
					{Port: 8080, Status: model.PortStatusOK},
				},
			},
		},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.formatPortsWithStatus(tc.svc)
			assert.NotEmpty(t, result)
		})
	}
}

// TestGetPortStatusColor tests the getPortStatusColor method.
func TestServicesPanel_getPortStatusColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status model.PortStatus
	}{
		{name: "ok", status: model.PortStatusOK},
		{name: "warning", status: model.PortStatusWarning},
		{name: "error", status: model.PortStatusError},
		{name: "unknown", status: model.PortStatusUnknown},
		{name: "default", status: model.PortStatus(99)},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.getPortStatusColor(tc.status)
			assert.NotEmpty(t, result)
		})
	}
}

// TestGetStateIcon tests the getStateIcon method.
func TestServicesPanel_getStateIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
	}{
		{name: "running", state: process.StateRunning},
		{name: "stopped", state: process.StateStopped},
		{name: "failed", state: process.StateFailed},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.getStateIcon(tc.state)
			assert.NotEmpty(t, result)
		})
	}
}

// TestGetStateText tests the getStateText method.
func TestServicesPanel_getStateText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
	}{
		{name: "running", state: process.StateRunning},
		{name: "stopped", state: process.StateStopped},
		{name: "failed", state: process.StateFailed},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.getStateText(tc.state)
			assert.NotEmpty(t, result)
		})
	}
}

// TestGetHealthText tests the getHealthText method.
func TestServicesPanel_getHealthText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{name: "not running", svc: model.ServiceSnapshot{State: process.StateStopped}},
		{name: "running no health", svc: model.ServiceSnapshot{State: process.StateRunning, HasHealthChecks: false}},
		{name: "healthy", svc: model.ServiceSnapshot{State: process.StateRunning, HasHealthChecks: true, Health: health.StatusHealthy}},
		{name: "unhealthy", svc: model.ServiceSnapshot{State: process.StateRunning, HasHealthChecks: true, Health: health.StatusUnhealthy}},
		{name: "degraded", svc: model.ServiceSnapshot{State: process.StateRunning, HasHealthChecks: true, Health: health.StatusDegraded}},
		{name: "unknown", svc: model.ServiceSnapshot{State: process.StateRunning, HasHealthChecks: true, Health: health.StatusUnknown}},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.getHealthText(tc.svc)
			assert.NotEmpty(t, result)
		})
	}
}

// TestHandleKeyMsg tests the handleKeyMsg method.
func TestServicesPanel_handleKeyMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{name: "home key", key: "home"},
		{name: "end key", key: "end"},
		{name: "up key", key: "up"},
		{name: "down key", key: "down"},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			msg := mockKeyMsg{str: tc.key}
			cmd := panel.handleKeyMsg(msg)
			// Verify function completes without error.
			_ = cmd
		})
	}
}

// TestRenderTopBorder tests the renderTopBorder method.
func TestServicesPanel_renderTopBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "renders top border"},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var sb strings.Builder
			panel.renderTopBorder(&sb, ansi.DefaultTheme().Primary, 70)
			result := sb.String()
			assert.NotEmpty(t, result)
		})
	}
}

// TestRenderHeaderRow tests the renderHeaderRow method.
func TestServicesPanel_renderHeaderRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "renders header row"},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var sb strings.Builder
			panel.renderHeaderRow(&sb, ansi.DefaultTheme().Primary, 70)
			result := sb.String()
			assert.NotEmpty(t, result)
		})
	}
}

// TestRenderContentLines tests the renderContentLines method.
func TestServicesPanel_renderContentLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "renders content lines"},
	}

	panel := NewServicesPanel(80, 20)
	panel.services = []model.ServiceSnapshot{{Name: "test", State: process.StateRunning}}
	panel.updateContent()

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var sb strings.Builder
			panel.renderContentLines(&sb, ansi.DefaultTheme().Primary, 70)
			result := sb.String()
			assert.NotEmpty(t, result)
		})
	}
}

// TestRenderBottomBorder tests the renderBottomBorder method.
func TestServicesPanel_renderBottomBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "renders bottom border"},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			var sb strings.Builder
			panel.renderBottomBorder(&sb, ansi.DefaultTheme().Primary, 70)
			result := sb.String()
			assert.NotEmpty(t, result)
		})
	}
}

// TestRenderHeader tests the renderHeader method.
func TestServicesPanel_renderHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "renders header"},
	}

	panel := NewServicesPanel(80, 20)

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := panel.renderHeader()
			assert.NotEmpty(t, result)
		})
	}
}

// TestCountIndicator tests the countIndicator method.
func TestServicesPanel_countIndicator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []model.ServiceSnapshot
	}{
		{name: "no services", services: nil},
		{name: "some running", services: []model.ServiceSnapshot{
			{State: process.StateRunning},
			{State: process.StateStopped},
		}},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 20)
			panel.services = tc.services
			result := panel.countIndicator()
			assert.NotEmpty(t, result)
		})
	}
}

// TestRenderVerticalScrollbar tests the renderVerticalScrollbar method.
func TestServicesPanel_renderVerticalScrollbar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []model.ServiceSnapshot
	}{
		{name: "few services", services: []model.ServiceSnapshot{{Name: "a"}, {Name: "b"}}},
		{name: "many services", services: func() []model.ServiceSnapshot {
			svcs := make([]model.ServiceSnapshot, 20)
			for i := range svcs {
				svcs[i] = model.ServiceSnapshot{Name: "svc"}
			}
			return svcs
		}()},
	}

	// Execute test cases.
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 20)
			panel.services = tc.services
			result := panel.renderVerticalScrollbar()
			assert.NotEmpty(t, result)
		})
	}
}
