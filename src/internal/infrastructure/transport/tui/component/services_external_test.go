// Package component_test provides black-box tests for the component package.
package component_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/component"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestNewServicesPanel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard_terminal", 80, 24},
		{"wide_terminal", 160, 50},
		{"narrow_terminal", 40, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(tt.width, tt.height)
			assert.Equal(t, tt.width, panel.Width())
			assert.Equal(t, tt.height, panel.Height())
		})
	}
}

func TestServicesPanel_SetSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		initW int
		initH int
		newW  int
		newH  int
	}{
		{"increase_size", 80, 24, 120, 30},
		{"decrease_size", 120, 30, 80, 24},
		{"width_only", 80, 24, 100, 24},
		{"height_only", 80, 24, 80, 30},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(tt.initW, tt.initH)
			panel.SetSize(tt.newW, tt.newH)
			assert.Equal(t, tt.newW, panel.Width())
			assert.Equal(t, tt.newH, panel.Height())
		})
	}
}

func TestServicesPanel_Focus(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		setFocused      bool
		wantFocused     bool
		toggleAgain     bool
		wantAfterToggle bool
	}{
		{"initial_unfocused_then_focus", true, true, false, true},
		{"initial_unfocused_then_focus_then_unfocus", true, true, true, false},
		{"stay_unfocused", false, false, false, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)

			// Initial state should be unfocused.
			assert.False(t, panel.Focused())

			// Apply first focus state.
			panel.SetFocused(tt.setFocused)
			assert.Equal(t, tt.wantFocused, panel.Focused())

			// Toggle if requested.
			if tt.toggleAgain {
				panel.SetFocused(!tt.setFocused)
				assert.Equal(t, tt.wantAfterToggle, panel.Focused())
			}
		})
	}
}

func TestServicesPanel_SetServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []model.ServiceSnapshot
	}{
		{
			name:     "empty_list",
			services: []model.ServiceSnapshot{},
		},
		{
			name: "single_running_service",
			services: []model.ServiceSnapshot{
				{
					Name:       "api",
					State:      process.StateRunning,
					PID:        12345,
					Uptime:     3600 * time.Second,
					CPUPercent: 25.5,
					MemoryRSS:  1024 * 1024 * 100,
				},
			},
		},
		{
			name: "multiple_services_mixed_states",
			services: []model.ServiceSnapshot{
				{Name: "api", State: process.StateRunning, PID: 100},
				{Name: "worker", State: process.StateStopped},
				{Name: "failed", State: process.StateFailed, LastError: "crash"},
				{Name: "starting", State: process.StateStarting},
			},
		},
		{
			name: "services_with_health_checks",
			services: []model.ServiceSnapshot{
				{
					Name:            "api",
					State:           process.StateRunning,
					HasHealthChecks: true,
					Health:          health.StatusHealthy,
				},
				{
					Name:            "db",
					State:           process.StateRunning,
					HasHealthChecks: true,
					Health:          health.StatusUnhealthy,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetServices(tt.services)
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_OptimalHeight(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceCount int
		minHeight    int
	}{
		{"no_services", 0, 3},
		{"few_services", 5, 3},
		{"many_services", 15, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)

			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{
					Name:  "svc",
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			height := panel.OptimalHeight()
			assert.GreaterOrEqual(t, height, tt.minHeight)
		})
	}
}

func TestServicesPanel_Init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard_size", 80, 24},
		{"wide_panel", 160, 50},
		{"narrow_panel", 40, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(tt.width, tt.height)
			cmd := panel.Init()
			assert.Nil(t, cmd)
		})
	}
}

func TestServicesPanel_Update_Unfocused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{"key_down", tea.KeyMsg{Type: tea.KeyDown}},
		{"key_up", tea.KeyMsg{Type: tea.KeyUp}},
		{"key_pgdn", tea.KeyMsg{Type: tea.KeyPgDown}},
		{"key_pgup", tea.KeyMsg{Type: tea.KeyPgUp}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetFocused(false)

			updatedPanel, cmd := panel.Update(tt.msg)
			assert.NotNil(t, updatedPanel)
			assert.Nil(t, cmd)
		})
	}
}

func TestServicesPanel_Update_Focused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{"key_down", tea.KeyMsg{Type: tea.KeyDown}},
		{"key_up", tea.KeyMsg{Type: tea.KeyUp}},
		{"key_pgdn", tea.KeyMsg{Type: tea.KeyPgDown}},
		{"key_pgup", tea.KeyMsg{Type: tea.KeyPgUp}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetFocused(true)

			// Add many services for scrolling.
			services := make([]model.ServiceSnapshot, 20)
			for i := range 20 {
				services[i] = model.ServiceSnapshot{
					Name:  "svc",
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			updatedPanel, cmd := panel.Update(tt.msg)
			assert.NotNil(t, updatedPanel)
			_ = cmd
		})
	}
}

func TestServicesPanel_View_Empty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{"standard_size", 80, 24},
		{"wide_panel", 160, 50},
		{"narrow_panel", 40, 20},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(tt.width, tt.height)
			view := panel.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Services")
			// The "No services configured" message is in the viewport content area
		})
	}
}

func TestServicesPanel_View_WithServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []model.ServiceSnapshot
	}{
		{
			name: "single_service",
			services: []model.ServiceSnapshot{
				{
					Name:       "api",
					State:      process.StateRunning,
					PID:        12345,
					Uptime:     3600 * time.Second,
					CPUPercent: 25.5,
					MemoryRSS:  1024 * 1024 * 100,
					Ports:      []int{8080, 8443},
				},
			},
		},
		{
			name: "multiple_services",
			services: []model.ServiceSnapshot{
				{Name: "api", State: process.StateRunning, PID: 100},
				{Name: "worker", State: process.StateStopped},
			},
		},
		{
			name: "service_with_ports",
			services: []model.ServiceSnapshot{
				{
					Name:  "web",
					State: process.StateRunning,
					Ports: []int{80, 443, 8080},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetServices(tt.services)

			view := panel.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Services")
		})
	}
}

func TestServicesPanel_View_Focused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		focused bool
	}{
		{"focused", true},
		{"unfocused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetFocused(tt.focused)
			panel.SetServices([]model.ServiceSnapshot{
				{Name: "test", State: process.StateRunning},
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_ServiceStates(t *testing.T) {
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
			panel := component.NewServicesPanel(80, 24)
			panel.SetServices([]model.ServiceSnapshot{
				{Name: "test", State: tt.state},
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_HealthStatuses(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		health health.Status
	}{
		{"healthy", health.StatusHealthy},
		{"unhealthy", health.StatusUnhealthy},
		{"degraded", health.StatusDegraded},
		{"unknown", health.StatusUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetServices([]model.ServiceSnapshot{
				{
					Name:            "test",
					State:           process.StateRunning,
					HasHealthChecks: true,
					Health:          tt.health,
				},
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_WithListeners(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		listeners []model.ListenerSnapshot
	}{
		{
			name:      "no_listeners",
			listeners: []model.ListenerSnapshot{},
		},
		{
			name: "single_listener_ok",
			listeners: []model.ListenerSnapshot{
				{Port: 8080, Status: model.PortStatusOK},
			},
		},
		{
			name: "multiple_listeners_mixed_status",
			listeners: []model.ListenerSnapshot{
				{Port: 8080, Status: model.PortStatusOK},
				{Port: 8443, Status: model.PortStatusWarning},
				{Port: 9000, Status: model.PortStatusError},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetServices([]model.ServiceSnapshot{
				{
					Name:      "api",
					State:     process.StateRunning,
					Listeners: tt.listeners,
				},
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_ScrollToTop(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceCount int
	}{
		{"few_services", 10},
		{"many_services", 50},
		{"large_list", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)

			// Add many services.
			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{
					Name:  "service",
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			panel.ScrollToTop()
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_ScrollToBottom(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceCount int
	}{
		{"few_services", 10},
		{"many_services", 50},
		{"large_list", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)

			// Add many services.
			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{
					Name:  "service",
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			panel.ScrollToBottom()
			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_ServiceCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		serviceCount int
		wantCount    int
	}{
		{"no_services", 0, 0},
		{"single_service", 1, 1},
		{"multiple_services", 3, 3},
		{"many_services", 10, 10},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)

			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{Name: "svc"}
			}
			panel.SetServices(services)

			assert.Equal(t, tt.wantCount, panel.ServiceCount())
		})
	}
}

func TestServicesPanel_WithMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		cpuPercent float64
		memoryRSS  uint64
		uptime     time.Duration
		wantCPU    string
	}{
		{"low_cpu", 5.5, 64 * 1024 * 1024, time.Hour, "5.5%"},
		{"medium_cpu", 45.2, 256 * 1024 * 1024, 2 * time.Hour, "45.2%"},
		{"high_cpu", 95.0, 1024 * 1024 * 1024, 24 * time.Hour, "95.0%"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetServices([]model.ServiceSnapshot{
				{
					Name:       "api",
					State:      process.StateRunning,
					PID:        12345,
					Uptime:     tt.uptime,
					CPUPercent: tt.cpuPercent,
					MemoryRSS:  tt.memoryRSS,
				},
			})

			view := panel.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, tt.wantCPU)
		})
	}
}

func TestServicesPanel_WithRestartCount(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		restartCount int
	}{
		{"no_restarts", 0},
		{"single_restart", 1},
		{"few_restarts", 5},
		{"many_restarts", 100},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetServices([]model.ServiceSnapshot{
				{
					Name:         "worker",
					State:        process.StateRunning,
					RestartCount: tt.restartCount,
				},
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_LongServiceNames(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		serviceName string
	}{
		{"short_name", "api"},
		{"medium_name", "user-service"},
		{"long_name", "very-long-service-name-that-exceeds-column-width"},
		{"empty_name", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetServices([]model.ServiceSnapshot{
				{
					Name:  tt.serviceName,
					State: process.StateRunning,
				},
			})

			view := panel.View()
			assert.NotEmpty(t, view)
		})
	}
}

func TestServicesPanel_SetFocused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		focused bool
	}{
		{"focus_panel", true},
		{"unfocus_panel", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetFocused(tt.focused)
			assert.Equal(t, tt.focused, panel.Focused())
		})
	}
}

func TestServicesPanel_Focused(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		focused bool
	}{
		{"when_focused", true},
		{"when_unfocused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetFocused(tt.focused)
			result := panel.Focused()
			assert.Equal(t, tt.focused, result)
		})
	}
}

func TestServicesPanel_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		focused bool
	}{
		{"update_when_focused", true},
		{"update_when_unfocused", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetFocused(tt.focused)

			// Add services.
			services := make([]model.ServiceSnapshot, 10)
			for i := range 10 {
				services[i] = model.ServiceSnapshot{
					Name:  "svc",
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			msg := tea.KeyMsg{Type: tea.KeyDown}
			updatedPanel, cmd := panel.Update(msg)
			assert.NotNil(t, updatedPanel)

			// When unfocused, cmd should be nil.
			if !tt.focused {
				assert.Nil(t, cmd)
			}
		})
	}
}

func TestServicesPanel_View(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		focused      bool
		serviceCount int
	}{
		{"empty_unfocused", false, 0},
		{"empty_focused", true, 0},
		{"with_services_unfocused", false, 3},
		{"with_services_focused", true, 3},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, 24)
			panel.SetFocused(tt.focused)

			// Add services.
			services := make([]model.ServiceSnapshot, tt.serviceCount)
			for i := range tt.serviceCount {
				services[i] = model.ServiceSnapshot{
					Name:  "svc" + string(rune('0'+i)),
					State: process.StateRunning,
				}
			}
			panel.SetServices(services)

			view := panel.View()
			assert.NotEmpty(t, view)
			assert.Contains(t, view, "Services")
			assert.Contains(t, view, "+")
		})
	}
}

func TestServicesPanel_Height(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		height int
	}{
		{"small", 10},
		{"standard", 24},
		{"large", 50},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			panel := component.NewServicesPanel(80, tt.height)
			assert.Equal(t, tt.height, panel.Height())
		})
	}
}

func TestServicesPanel_Width(t *testing.T) {
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
			panel := component.NewServicesPanel(tt.width, 24)
			assert.Equal(t, tt.width, panel.Width())
		})
	}
}
