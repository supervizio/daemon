package screen_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/stretchr/testify/assert"
)

func TestNewHeaderRenderer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{"standard_width", 80},
		{"small_width", 40},
		{"large_width", 200},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewHeaderRenderer(tt.width)
			assert.NotNil(t, renderer)
		})
	}
}

func TestHeaderRenderer_SetWidth(t *testing.T) {
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
			renderer := screen.NewHeaderRenderer(tt.initial)
			renderer.SetWidth(tt.newWidth)
			assert.NotNil(t, renderer)
		})
	}
}

func TestHeaderRenderer_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		width       int
		containsStr []string
	}{
		{"compact_width", 60, []string{"superviz"}},
		{"normal_width", 100, []string{"superviz", ".io"}},
		{"wide_width", 180, []string{"superviz"}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewHeaderRenderer(tt.width)
			snap := createHeaderTestSnapshot()
			result := renderer.Render(snap)
			assert.NotEmpty(t, result)
			for _, s := range tt.containsStr {
				assert.Contains(t, result, s)
			}
		})
	}
}

func TestHeaderRenderer_RenderBrandOnly(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		width       int
		containsStr []string
	}{
		{
			name:        "brand_only",
			width:       80,
			containsStr: []string{"superviz", ".io"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewHeaderRenderer(tt.width)
			result := renderer.RenderBrandOnly()
			assert.NotEmpty(t, result)
			for _, s := range tt.containsStr {
				assert.Contains(t, result, s)
			}
		})
	}
}

func createHeaderTestSnapshot() *model.Snapshot {
	return &model.Snapshot{
		Context: model.RuntimeContext{
			Hostname: "testhost",
			Version:  "1.0.0",
			OS:       "linux",
			Arch:     "amd64",
			Kernel:   "5.15.0",
			Mode:     model.ModeHost,
			Uptime:   time.Hour,
		},
	}
}
