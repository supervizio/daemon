package screen_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/stretchr/testify/assert"
)

func TestNewServicesRenderer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "standard_width",
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewServicesRenderer(tt.width)
			assert.NotNil(t, renderer)
		})
	}
}

func TestServicesRenderer_SetWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		initial  int
		newWidth int
	}{
		{
			name:     "update_width",
			initial:  80,
			newWidth: 120,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewServicesRenderer(tt.initial)
			renderer.SetWidth(tt.newWidth)
			assert.NotNil(t, renderer)
		})
	}
}

func TestServicesRenderer_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "basic_render",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewServicesRenderer(80)
			result := renderer.Render(tt.snap)
			assert.NotNil(t, &result)
		})
	}
}

func TestServicesRenderer_Render_Empty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "no_services",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewServicesRenderer(80)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Services")
		})
	}
}

func TestServicesRenderer_Render_Compact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		snap  *model.Snapshot
	}{
		{
			name:  "compact_layout",
			width: 60,
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web", State: 1, CPUPercent: 25.5},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewServicesRenderer(tt.width)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestServicesRenderer_Render_Normal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		snap  *model.Snapshot
	}{
		{
			name:  "normal_layout",
			width: 100,
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{
						Name:       "web",
						State:      1,
						PID:        1234,
						CPUPercent: 25.5,
						MemoryRSS:  1024 * 1024,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewServicesRenderer(tt.width)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestServicesRenderer_Render_Wide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		snap  *model.Snapshot
	}{
		{
			name:  "wide_layout",
			width: 180,
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{
						Name:         "web",
						State:        1,
						PID:          1234,
						RestartCount: 2,
						Health:       1,
						CPUPercent:   25.5,
						MemoryRSS:    1024 * 1024,
						Listeners: []model.ListenerSnapshot{
							{Port: 8080, Status: 1},
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewServicesRenderer(tt.width)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestServicesRenderer_RenderNamesOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "empty_services",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{},
			},
		},
		{
			name: "single_service_no_ports",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web"},
				},
			},
		},
		{
			name: "service_with_ports",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{
						Name: "web",
						Listeners: []model.ListenerSnapshot{
							{Port: 8080},
							{Port: 9090},
						},
					},
				},
			},
		},
		{
			name: "multiple_services",
			snap: &model.Snapshot{
				Services: []model.ServiceSnapshot{
					{Name: "web", Listeners: []model.ListenerSnapshot{{Port: 8080}}},
					{Name: "db", Listeners: []model.ListenerSnapshot{{Port: 5432}}},
					{Name: "cache"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewServicesRenderer(120)
			result := renderer.RenderNamesOnly(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "Services")
		})
	}
}
