package screen_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/stretchr/testify/assert"
)

func TestNewContextRenderer(t *testing.T) {
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
			renderer := screen.NewContextRenderer(tt.width)
			assert.NotNil(t, renderer)
		})
	}
}

func TestContextRenderer_SetWidth(t *testing.T) {
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
			renderer := screen.NewContextRenderer(tt.initial)
			renderer.SetWidth(tt.newWidth)
			assert.NotNil(t, renderer)
		})
	}
}

func TestContextRenderer_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "basic_render",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Hostname: "test-host",
					OS:       "linux",
					Kernel:   "5.10.0",
					Arch:     "x86_64",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewContextRenderer(80)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestContextRenderer_RenderLimits(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "with_limits",
			snap: &model.Snapshot{
				Limits: model.ResourceLimits{
					HasLimits: true,
					CPUQuota:  2.0,
					MemoryMax: 1024 * 1024 * 1024,
				},
			},
		},
		{
			name: "without_limits",
			snap: &model.Snapshot{
				Limits: model.ResourceLimits{
					HasLimits: false,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewContextRenderer(80)
			result := renderer.RenderLimits(tt.snap)
			if tt.snap.Limits.HasLimits {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}

func TestContextRenderer_RenderSandboxes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "with_sandboxes",
			snap: &model.Snapshot{
				Sandboxes: []model.SandboxInfo{
					{Name: "docker", Detected: true, Endpoint: "/var/run/docker.sock"},
				},
			},
		},
		{
			name: "without_sandboxes",
			snap: &model.Snapshot{
				Sandboxes: []model.SandboxInfo{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewContextRenderer(80)
			result := renderer.RenderSandboxes(tt.snap)
			if len(tt.snap.Sandboxes) > 0 {
				assert.NotEmpty(t, result)
			} else {
				assert.Empty(t, result)
			}
		})
	}
}
