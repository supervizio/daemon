package screen

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/widget"
	"github.com/stretchr/testify/assert"
)

func TestServicesRenderer_renderEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{
			name: "empty_state",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.renderEmpty()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Services")
		})
	}
}

func TestServicesRenderer_renderCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "single_service",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web", State: process.StateRunning, CPUPercent: 25.5},
				},
			},
		},
		{
			name: "multiple_services",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web", State: process.StateRunning, CPUPercent: 25.5},
					{Name: "db", State: process.StateStopped},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.renderCompact(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestServicesRenderer_createNormalTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard_width",
			width: 80,
		},
		{
			name:  "wide_width",
			width: 160,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			table := renderer.createNormalTable()
			assert.NotNil(t, table)
		})
	}
}

func TestServicesRenderer_populateNormalRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []model.ServiceSnapshot
	}{
		{
			name: "running_service",
			services: []model.ServiceSnapshot{
				{
					Name:       "web",
					State:      process.StateRunning,
					PID:        1234,
					Uptime:     time.Hour,
					CPUPercent: 25.5,
					MemoryRSS:  1024 * 1024,
				},
			},
		},
		{
			name: "stopped_service",
			services: []model.ServiceSnapshot{
				{Name: "db", State: process.StateStopped},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			table := widget.NewTable(70)
			renderer.populateNormalRows(table, tt.services)
			assert.NotNil(t, table)
		})
	}
}

func TestServicesRenderer_formatPID(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		pid  int
		want string
	}{
		{
			name: "valid_pid",
			pid:  1234,
			want: "1234",
		},
		{
			name: "zero_pid",
			pid:  0,
			want: "-",
		},
		{
			name: "negative_pid",
			pid:  -1,
			want: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.formatPID(tt.pid)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestServicesRenderer_formatUptimeShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
		want  string
	}{
		{
			name:  "running_service",
			state: process.StateRunning,
			want:  "1h",
		},
		{
			name:  "starting_service",
			state: process.StateStarting,
			want:  "30s",
		},
		{
			name:  "stopped_service",
			state: process.StateStopped,
			want:  "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			uptime := time.Hour
			if tt.state == process.StateStarting {
				uptime = 30 * time.Second
			}
			result := renderer.formatUptimeShort(tt.state, uptime)
			if tt.state == process.StateRunning || tt.state == process.StateStarting {
				assert.NotEqual(t, "-", result)
			} else {
				assert.Equal(t, "-", result)
			}
		})
	}
}

func TestServicesRenderer_formatUptimeLong(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
	}{
		{
			name:  "running_service",
			state: process.StateRunning,
		},
		{
			name:  "stopped_service",
			state: process.StateStopped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.formatUptimeLong(tt.state, time.Hour)
			if tt.state == process.StateRunning {
				assert.NotEqual(t, "-", result)
			} else {
				assert.Equal(t, "-", result)
			}
		})
	}
}

func TestServicesRenderer_formatMetrics(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		state      process.State
		wantCPU    string
		wantMemory string
	}{
		{
			name:       "running_service",
			state:      process.StateRunning,
			wantCPU:    "25.5%",
			wantMemory: "1.0 MB",
		},
		{
			name:       "stopped_service",
			state:      process.StateStopped,
			wantCPU:    "-",
			wantMemory: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			cpu, mem := renderer.formatMetrics(tt.state, 25.5, 1024*1024)
			if tt.state == process.StateRunning {
				assert.NotEqual(t, "-", cpu)
				assert.NotEqual(t, "-", mem)
			} else {
				assert.Equal(t, "-", cpu)
				assert.Equal(t, "-", mem)
			}
		})
	}
}

func TestServicesRenderer_formatRestarts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		count int
		want  string
	}{
		{
			name:  "no_restarts",
			count: 0,
			want:  "-",
		},
		{
			name:  "with_restarts",
			count: 5,
			want:  "5",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.formatRestarts(tt.count)
			assert.Equal(t, tt.want, result)
		})
	}
}

func TestServicesRenderer_renderSummary(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "all_running",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{State: process.StateRunning},
					{State: process.StateRunning},
				},
			},
		},
		{
			name: "mixed_states",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{State: process.StateRunning},
					{State: process.StateStopped},
					{State: process.StateFailed},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.renderSummary(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "services")
		})
	}
}

func TestServicesRenderer_stateShort(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		state process.State
	}{
		{
			name:  "running",
			state: process.StateRunning,
		},
		{
			name:  "stopped",
			state: process.StateStopped,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.stateShort(tt.state)
			assert.NotEmpty(t, result)
		})
	}
}

func TestServicesRenderer_formatPorts(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		listeners []model.ListenerSnapshot
		want      string
	}{
		{
			name:      "no_listeners",
			listeners: []model.ListenerSnapshot{},
			want:      "-",
		},
		{
			name: "single_port_ok",
			listeners: []model.ListenerSnapshot{
				{Port: 8080, Status: model.PortStatusOK},
			},
		},
		{
			name: "multiple_ports",
			listeners: []model.ListenerSnapshot{
				{Port: 8080, Status: model.PortStatusOK},
				{Port: 9090, Status: model.PortStatusWarning},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.formatPorts(tt.listeners)
			if len(tt.listeners) == 0 {
				assert.Equal(t, "-", result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func Test_prefixLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		lines  []string
		prefix string
		want   int
	}{
		{
			name:   "empty_lines",
			lines:  []string{},
			prefix: "  ",
			want:   0,
		},
		{
			name:   "multiple_lines",
			lines:  []string{"line1", "line2", "line3"},
			prefix: "  ",
			want:   3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := prefixLines(tt.lines, tt.prefix)
			assert.Len(t, result, tt.want)
			for _, line := range result {
				assert.Contains(t, line, tt.prefix)
			}
		})
	}
}

func TestServicesRenderer_buildServiceEntries(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []model.ServiceSnapshot
		want     int
	}{
		{
			name:     "empty_services",
			services: []model.ServiceSnapshot{},
			want:     0,
		},
		{
			name: "single_service",
			services: []model.ServiceSnapshot{
				{Name: "web"},
			},
			want: 1,
		},
		{
			name: "service_with_ports",
			services: []model.ServiceSnapshot{
				{Name: "web", Listeners: []model.ListenerSnapshot{{Port: 8080}}},
			},
			want: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.buildServiceEntries(tt.services)
			assert.Len(t, result, tt.want)
		})
	}
}

func TestServicesRenderer_formatServiceEntry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  model.ServiceSnapshot
	}{
		{
			name: "service_no_ports",
			svc:  model.ServiceSnapshot{Name: "web"},
		},
		{
			name: "service_with_single_port",
			svc: model.ServiceSnapshot{
				Name:      "web",
				Listeners: []model.ListenerSnapshot{{Port: 8080}},
			},
		},
		{
			name: "service_with_multiple_ports",
			svc: model.ServiceSnapshot{
				Name: "web",
				Listeners: []model.ListenerSnapshot{
					{Port: 8080},
					{Port: 9090},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.formatServiceEntry(tt.svc)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, tt.svc.Name)
		})
	}
}

func TestServicesRenderer_calculateColumnLayout(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entries []serviceEntry
		width   int
	}{
		{
			name: "single_entry",
			entries: []serviceEntry{
				{display: "web", visibleLen: 3},
			},
			width: 80,
		},
		{
			name: "multiple_entries",
			entries: []serviceEntry{
				{display: "web", visibleLen: 3},
				{display: "database", visibleLen: 8},
			},
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			colWidth, cols := renderer.calculateColumnLayout(tt.entries, 6, 10, 2)
			assert.Greater(t, colWidth, 0)
			assert.Greater(t, cols, 0)
		})
	}
}

func TestServicesRenderer_createWideTable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "wide_width",
			width: 160,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			table := renderer.createWideTable()
			assert.NotNil(t, table)
		})
	}
}

func TestServicesRenderer_populateWideRows(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		services []model.ServiceSnapshot
	}{
		{
			name: "running_service_with_health",
			services: []model.ServiceSnapshot{
				{
					Name:         "web",
					State:        process.StateRunning,
					PID:          1234,
					Uptime:       time.Hour,
					RestartCount: 2,
					Health:       health.StatusHealthy,
					CPUPercent:   25.5,
					MemoryRSS:    1024 * 1024,
					Listeners:    []model.ListenerSnapshot{{Port: 8080}},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  160,
				status: widget.NewStatusIndicator(),
			}
			table := widget.NewTable(150)
			renderer.populateWideRows(table, tt.services)
			assert.NotNil(t, table)
		})
	}
}

func TestServicesRenderer_renderNormal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		snap  *model.Snapshot
		width int
	}{
		{
			name: "single_running_service",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{
						Name:       "web",
						State:      process.StateRunning,
						PID:        1234,
						Uptime:     time.Hour,
						CPUPercent: 25.5,
						MemoryRSS:  1024 * 1024,
					},
				},
			},
			width: 100,
		},
		{
			name: "multiple_services_mixed_states",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web", State: process.StateRunning, PID: 1234, Uptime: time.Hour, CPUPercent: 25.5, MemoryRSS: 1024 * 1024},
					{Name: "db", State: process.StateStopped},
					{Name: "cache", State: process.StateFailed},
				},
			},
			width: 100,
		},
		{
			name: "narrow_width",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "api", State: process.StateStarting},
				},
			},
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.renderNormal(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Services")
		})
	}
}

func TestServicesRenderer_renderTableInBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		snap  *model.Snapshot
		width int
	}{
		{
			name: "single_service",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web", State: process.StateRunning},
				},
			},
			width: 100,
		},
		{
			name: "multiple_services",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web", State: process.StateRunning},
					{Name: "db", State: process.StateRunning},
					{Name: "cache", State: process.StateStopped},
				},
			},
			width: 120,
		},
		{
			name: "services_with_health",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "api", State: process.StateRunning, Health: health.StatusHealthy},
				},
			},
			width: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			table := renderer.createNormalTable()
			renderer.populateNormalRows(table, tt.snap.Services)
			result := renderer.renderTableInBox(table, tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Services")
			assert.Contains(t, result, "services")
		})
	}
}

func TestServicesRenderer_renderWide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		snap  *model.Snapshot
		width int
	}{
		{
			name: "single_service",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{
						Name:         "web",
						State:        process.StateRunning,
						PID:          1234,
						Uptime:       time.Hour,
						RestartCount: 0,
						Health:       health.StatusHealthy,
						CPUPercent:   25.5,
						MemoryRSS:    1024 * 1024,
						Listeners:    []model.ListenerSnapshot{{Port: 8080, Status: model.PortStatusOK}},
					},
				},
			},
			width: 160,
		},
		{
			name: "multiple_services_with_ports",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{
						Name:       "api",
						State:      process.StateRunning,
						PID:        1000,
						Uptime:     time.Hour * 2,
						CPUPercent: 10.0,
						MemoryRSS:  512 * 1024,
						Listeners:  []model.ListenerSnapshot{{Port: 3000, Status: model.PortStatusOK}},
					},
					{
						Name:         "db",
						State:        process.StateRunning,
						PID:          2000,
						Uptime:       time.Hour * 5,
						RestartCount: 2,
						Health:       health.StatusUnhealthy,
						CPUPercent:   50.0,
						MemoryRSS:    2048 * 1024,
						Listeners:    []model.ListenerSnapshot{{Port: 5432, Status: model.PortStatusWarning}},
					},
				},
			},
			width: 180,
		},
		{
			name: "ultra_wide",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "worker", State: process.StateStopped},
				},
			},
			width: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.renderWide(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Services")
		})
	}
}

func TestServicesRenderer_renderEmptyServices(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard_width",
			width: 100,
		},
		{
			name:  "narrow_width",
			width: 60,
		},
		{
			name:  "wide_width",
			width: 160,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.renderEmptyServices()
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Services")
			assert.Contains(t, result, "0 configured")
			assert.Contains(t, result, "No services configured")
		})
	}
}

func TestServicesRenderer_layoutEntriesInColumns(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		entries []serviceEntry
		width   int
	}{
		{
			name: "single_entry",
			entries: []serviceEntry{
				{display: "web", visibleLen: 3},
			},
			width: 80,
		},
		{
			name: "multiple_entries_single_column",
			entries: []serviceEntry{
				{display: "web", visibleLen: 3},
				{display: "database", visibleLen: 8},
				{display: "cache", visibleLen: 5},
			},
			width: 40,
		},
		{
			name: "multiple_entries_multi_column",
			entries: []serviceEntry{
				{display: "web", visibleLen: 3},
				{display: "api", visibleLen: 3},
				{display: "db", visibleLen: 2},
				{display: "cache", visibleLen: 5},
			},
			width: 120,
		},
		{
			name: "entries_with_ports",
			entries: []serviceEntry{
				{display: "web :8080", visibleLen: 9},
				{display: "api :3000,:3001", visibleLen: 15},
			},
			width: 100,
		},
		{
			name:    "empty_entries",
			entries: []serviceEntry{},
			width:   80,
		},
		{
			name: "long_service_names",
			entries: []serviceEntry{
				{display: "very-long-service-name :8080", visibleLen: 28},
				{display: "another-long-service :9090", visibleLen: 26},
			},
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.layoutEntriesInColumns(tt.entries)
			if len(tt.entries) == 0 {
				assert.Empty(t, result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestServicesRenderer_renderNamesOnlyBox(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		snap  *model.Snapshot
		lines []string
		width int
	}{
		{
			name: "single_service",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web"},
				},
			},
			lines: []string{"  web"},
			width: 80,
		},
		{
			name: "multiple_services",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web"},
					{Name: "api"},
					{Name: "db"},
				},
			},
			lines: []string{"  web  api  db"},
			width: 100,
		},
		{
			name: "services_with_ports",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web", Listeners: []model.ListenerSnapshot{{Port: 8080}}},
					{Name: "api", Listeners: []model.ListenerSnapshot{{Port: 3000}}},
				},
			},
			lines: []string{"  web :8080", "  api :3000"},
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  tt.width,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.renderNamesOnlyBox(tt.snap, tt.lines)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Services")
			assert.Contains(t, result, "configured")
		})
	}
}

func TestServicesRenderer_buildCompactServiceLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		svc  *model.ServiceSnapshot
	}{
		{
			name: "running_service",
			svc:  &model.ServiceSnapshot{Name: "web", State: process.StateRunning, CPUPercent: 25.5},
		},
		{
			name: "stopped_service",
			svc:  &model.ServiceSnapshot{Name: "db", State: process.StateStopped},
		},
		{
			name: "long_name",
			svc:  &model.ServiceSnapshot{Name: "very-long-service-name", State: process.StateRunning},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &ServicesRenderer{
				theme:  ansi.DefaultTheme(),
				icons:  ansi.DefaultIcons(),
				width:  80,
				status: widget.NewStatusIndicator(),
			}
			result := renderer.buildCompactServiceLine(tt.svc)
			assert.NotEmpty(t, result)
		})
	}
}
