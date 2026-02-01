// Package component provides internal white-box tests.
package component

import (
	"strings"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestServicesPanel_formatServiceName(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
	}{
		{"short_name", "api"},
		{"exact_width", "service12345"},
		{"long_name", "very-long-service-name-that-exceeds-limit"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			result := panel.formatServiceName(tt.serviceName)
			assert.NotEmpty(t, result)
			assert.LessOrEqual(t, len([]rune(result)), nameColWidth)
		})
	}
}

func TestServicesPanel_getStateIcon(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
	}{
		{"running", process.StateRunning},
		{"stopped", process.StateStopped},
		{"failed", process.StateFailed},
		{"starting", process.StateStarting},
		{"stopping", process.StateStopping},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			icon := panel.getStateIcon(tt.state)
			assert.NotEmpty(t, icon)
			assert.Contains(t, icon, "o")
		})
	}
}

func TestServicesPanel_getStateText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		state    process.State
		wantText string
	}{
		{"running", process.StateRunning, "running"},
		{"stopped", process.StateStopped, "stopped"},
		{"failed", process.StateFailed, "failed"},
		{"starting", process.StateStarting, "starting"},
		{"stopping", process.StateStopping, "stopping"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			text := panel.getStateText(tt.state)
			assert.Contains(t, text, tt.wantText)
		})
	}
}

func TestServicesPanel_getHealthText(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
		want    string
	}{
		{
			name: "not_running",
			service: model.ServiceSnapshot{
				State: process.StateStopped,
			},
			want: "-",
		},
		{
			name: "no_health_checks",
			service: model.ServiceSnapshot{
				State:           process.StateRunning,
				HasHealthChecks: false,
			},
			want: "-",
		},
		{
			name: "healthy",
			service: model.ServiceSnapshot{
				State:           process.StateRunning,
				HasHealthChecks: true,
				Health:          health.StatusHealthy,
			},
			want: "healthy",
		},
		{
			name: "unhealthy",
			service: model.ServiceSnapshot{
				State:           process.StateRunning,
				HasHealthChecks: true,
				Health:          health.StatusUnhealthy,
			},
			want: "unhealthy",
		},
		{
			name: "degraded",
			service: model.ServiceSnapshot{
				State:           process.StateRunning,
				HasHealthChecks: true,
				Health:          health.StatusDegraded,
			},
			want: "degraded",
		},
		{
			name: "unknown",
			service: model.ServiceSnapshot{
				State:           process.StateRunning,
				HasHealthChecks: true,
				Health:          health.StatusUnknown,
			},
			want: "pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			text := panel.getHealthText(tt.service)
			assert.Contains(t, text, tt.want)
		})
	}
}

func TestServicesPanel_formatUptime(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
		want    string
	}{
		{
			name: "not_running",
			service: model.ServiceSnapshot{
				State: process.StateStopped,
			},
			want: "-",
		},
		{
			name: "zero_uptime",
			service: model.ServiceSnapshot{
				State:  process.StateRunning,
				Uptime: 0,
			},
			want: "-",
		},
		{
			name: "with_uptime",
			service: model.ServiceSnapshot{
				State:  process.StateRunning,
				Uptime: 3665 * time.Second,
			},
			want: "1h", // Should contain some uptime indicator.
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			result := panel.formatUptime(tt.service)
			assert.Contains(t, result, tt.want)
		})
	}
}

func TestServicesPanel_formatPID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
		want    string
	}{
		{
			name:    "no_pid",
			service: model.ServiceSnapshot{PID: 0},
			want:    "-",
		},
		{
			name:    "with_pid",
			service: model.ServiceSnapshot{PID: 12345},
			want:    "12345",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			result := panel.formatPID(tt.service)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestServicesPanel_formatRestarts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
		want    string
	}{
		{
			name:    "no_restarts",
			service: model.ServiceSnapshot{RestartCount: 0},
			want:    "-",
		},
		{
			name:    "with_restarts",
			service: model.ServiceSnapshot{RestartCount: 5},
			want:    "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			result := panel.formatRestarts(tt.service)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestServicesPanel_formatCPU(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
		want    string
	}{
		{
			name: "not_running",
			service: model.ServiceSnapshot{
				State: process.StateStopped,
			},
			want: "-",
		},
		{
			name: "with_cpu",
			service: model.ServiceSnapshot{
				State:      process.StateRunning,
				CPUPercent: 25.5,
			},
			want: "25.5%",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			result := panel.formatCPU(tt.service)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestServicesPanel_formatMemory(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
	}{
		{
			name: "not_running",
			service: model.ServiceSnapshot{
				State: process.StateStopped,
			},
		},
		{
			name: "zero_memory",
			service: model.ServiceSnapshot{
				State:     process.StateRunning,
				MemoryRSS: 0,
			},
		},
		{
			name: "with_memory",
			service: model.ServiceSnapshot{
				State:     process.StateRunning,
				MemoryRSS: 1024 * 1024 * 100, // 100 MB
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			result := panel.formatMemory(tt.service)
			assert.NotEmpty(t, result)
		})
	}
}

func TestFormatPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		ports []int
		want  string
	}{
		{"no_ports", []int{}, "-"},
		{"single_port", []int{8080}, "8080"},
		{"multiple_ports", []int{8080, 8443, 9000}, "8080,8443,9000"},
		{"many_ports", []int{80, 443, 8000, 8080, 8443, 9000, 9090}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatPorts(tt.ports)
			if tt.want != "" {
				assert.Equal(t, tt.want, result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestServicesPanel_formatPortsWithStatus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
	}{
		{
			name: "no_listeners",
			service: model.ServiceSnapshot{
				Listeners: []model.ListenerSnapshot{},
				Ports:     []int{},
			},
		},
		{
			name: "no_listeners_with_detected_ports",
			service: model.ServiceSnapshot{
				Listeners: []model.ListenerSnapshot{},
				Ports:     []int{8080, 8443},
			},
		},
		{
			name: "with_listeners",
			service: model.ServiceSnapshot{
				Listeners: []model.ListenerSnapshot{
					{Port: 8080, Status: model.PortStatusOK},
					{Port: 8443, Status: model.PortStatusWarning},
				},
			},
		},
		{
			name: "mixed_status",
			service: model.ServiceSnapshot{
				Listeners: []model.ListenerSnapshot{
					{Port: 8080, Status: model.PortStatusOK},
					{Port: 8443, Status: model.PortStatusError},
					{Port: 9000, Status: model.PortStatusUnknown},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			result := panel.formatPortsWithStatus(tt.service)
			assert.NotEmpty(t, result)
		})
	}
}

func TestServicesPanel_getPortStatusColor(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status model.PortStatus
	}{
		{"ok", model.PortStatusOK},
		{"warning", model.PortStatusWarning},
		{"error", model.PortStatusError},
		{"unknown", model.PortStatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			color := panel.getPortStatusColor(tt.status)
			assert.NotEmpty(t, color)
		})
	}
}

func TestServicesPanel_renderVerticalScrollbar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceCount int
		height       int
	}{
		{"no_scroll", 5, 20},
		{"scroll_needed", 50, 10},
		{"exact_fit", 10, 13},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, tt.height)

			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{
					Name:  "service" + string(rune('0'+(i%10))),
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			scrollbar := panel.renderVerticalScrollbar()
			assert.NotEmpty(t, scrollbar)
		})
	}
}

// mockServiceKeyMsg is a mock for tea.KeyMsg implementing Stringer.
type mockServiceKeyMsg struct {
	str string
}

func (m mockServiceKeyMsg) String() string {
	return m.str
}

func TestServicesPanel_handleKeyMsg(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		key  string
	}{
		{"home", "home"},
		{"g_key", "g"},
		{"end", "end"},
		{"G_key", "G"},
		{"page_up", "pgup"},
		{"ctrl_u", "ctrl+u"},
		{"page_down", "pgdown"},
		{"ctrl_d", "ctrl+d"},
		{"up_arrow", "up"},
		{"k_key", "k"},
		{"down_arrow", "down"},
		{"j_key", "j"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			msg := mockServiceKeyMsg{str: tt.key}
			cmd := panel.handleKeyMsg(msg)
			_ = cmd
		})
	}
}

func TestServicesPanel_updateContent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceCount int
	}{
		{"empty_services", 0},
		{"single_service", 1},
		{"multiple_services", 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)

			// Add services.
			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{
					Name:  "service" + string(rune('0'+i)),
					State: process.StateRunning,
				}
			}
			panel.services = services

			// Call updateContent.
			panel.updateContent()

			// Verify viewport has content set.
			view := panel.viewport.View()
			if tt.serviceCount == 0 {
				assert.Contains(t, view, "No services configured")
			} else {
				assert.NotEmpty(t, view)
			}
		})
	}
}

func TestServicesPanel_formatServiceLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
	}{
		{
			name: "running_service",
			service: model.ServiceSnapshot{
				Name:       "api",
				State:      process.StateRunning,
				PID:        12345,
				Uptime:     time.Hour,
				CPUPercent: 25.5,
				MemoryRSS:  1024 * 1024 * 100,
			},
		},
		{
			name: "stopped_service",
			service: model.ServiceSnapshot{
				Name:  "worker",
				State: process.StateStopped,
			},
		},
		{
			name: "service_with_health",
			service: model.ServiceSnapshot{
				Name:            "db",
				State:           process.StateRunning,
				HasHealthChecks: true,
				Health:          health.StatusHealthy,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			line := panel.formatServiceLine(tt.service)
			assert.NotEmpty(t, line)
			assert.Contains(t, line, "o") // State icon.
		})
	}
}

func TestServicesPanel_collectServiceColumns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service model.ServiceSnapshot
	}{
		{
			name: "complete_service",
			service: model.ServiceSnapshot{
				Name:         "api",
				State:        process.StateRunning,
				PID:          12345,
				Uptime:       time.Hour,
				CPUPercent:   25.5,
				MemoryRSS:    1024 * 1024 * 100,
				RestartCount: 3,
				Ports:        []int{8080},
			},
		},
		{
			name: "minimal_service",
			service: model.ServiceSnapshot{
				Name:  "worker",
				State: process.StateStopped,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			cols := panel.collectServiceColumns(tt.service)
			assert.NotEmpty(t, cols.stateIcon)
			assert.NotEmpty(t, cols.stateText)
			assert.NotEmpty(t, cols.healthText)
		})
	}
}

func TestServicesPanel_buildServiceLineString(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "full_service",
			svc: model.ServiceSnapshot{
				Name:       "api",
				State:      process.StateRunning,
				PID:        12345,
				CPUPercent: 25.5,
			},
		},
		{
			name: "minimal_service",
			svc: model.ServiceSnapshot{
				Name:  "worker",
				State: process.StateStopped,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			name := panel.formatServiceName(tt.svc.Name)
			cols := panel.collectServiceColumns(tt.svc)
			line := panel.buildServiceLineString(name, &cols)
			assert.NotEmpty(t, line)
		})
	}
}

func TestServicesPanel_renderTopBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		width        int
		serviceCount int
	}{
		{"narrow_empty", 40, 0},
		{"standard_empty", 80, 0},
		{"wide_with_services", 120, 5},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(tt.width, 24)

			// Add services.
			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{
					Name:  "svc",
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			var sb strings.Builder
			borderColor := panel.theme.Muted
			innerWidth := tt.width - borderWidth
			panel.renderTopBorder(&sb, borderColor, innerWidth)

			result := sb.String()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "+")
			assert.Contains(t, result, "Services")
		})
	}
}

func TestServicesPanel_renderHeaderRow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{"narrow", 60},
		{"standard", 80},
		{"wide", 160},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(tt.width, 24)

			var sb strings.Builder
			borderColor := panel.theme.Muted
			innerWidth := tt.width - borderWidth
			panel.renderHeaderRow(&sb, borderColor, innerWidth)

			result := sb.String()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "|")
		})
	}
}

func TestServicesPanel_renderContentLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceCount int
		height       int
	}{
		{"empty_content", 0, 10},
		{"few_services", 3, 10},
		{"many_services", 20, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, tt.height)

			// Add services.
			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{
					Name:  "svc",
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			var sb strings.Builder
			borderColor := panel.theme.Muted
			innerWidth := 80 - borderWidth
			panel.renderContentLines(&sb, borderColor, innerWidth)

			result := sb.String()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "|")
		})
	}
}

func TestServicesPanel_renderBottomBorder(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{"narrow", 40},
		{"standard", 80},
		{"wide", 160},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(tt.width, 24)

			var sb strings.Builder
			borderColor := panel.theme.Muted
			innerWidth := tt.width - borderWidth
			panel.renderBottomBorder(&sb, borderColor, innerWidth)

			result := sb.String()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "+")
		})
	}
}

func TestServicesPanel_renderHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{"narrow", 60},
		{"standard", 80},
		{"wide", 160},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(tt.width, 24)

			header := panel.renderHeader()
			assert.NotEmpty(t, header)
			assert.Contains(t, header, "NAME")
			assert.Contains(t, header, "STATE")
			assert.Contains(t, header, "HEALTH")
		})
	}
}

func TestServicesPanel_countIndicator(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		services       []model.ServiceSnapshot
		expectedTotal  int
		runningWantMin int
	}{
		{
			name:           "no_services",
			services:       []model.ServiceSnapshot{},
			expectedTotal:  0,
			runningWantMin: 0,
		},
		{
			name: "all_running",
			services: []model.ServiceSnapshot{
				{Name: "a", State: process.StateRunning},
				{Name: "b", State: process.StateRunning},
			},
			expectedTotal:  2,
			runningWantMin: 2,
		},
		{
			name: "mixed_states",
			services: []model.ServiceSnapshot{
				{Name: "a", State: process.StateRunning},
				{Name: "b", State: process.StateStopped},
				{Name: "c", State: process.StateFailed},
			},
			expectedTotal:  3,
			runningWantMin: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := NewServicesPanel(80, 24)
			panel.SetServices(tt.services)

			indicator := panel.countIndicator()
			assert.NotEmpty(t, indicator)
			assert.Contains(t, indicator, "[")
			assert.Contains(t, indicator, "]")
			assert.Contains(t, indicator, "/")
		})
	}
}
