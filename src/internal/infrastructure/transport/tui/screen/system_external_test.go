package screen_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/stretchr/testify/assert"
)

func TestNewSystemRenderer(t *testing.T) {
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
			name:  "small_width",
			width: 40,
		},
		{
			name:  "large_width",
			width: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewSystemRenderer(tt.width)
			assert.NotNil(t, renderer)
		})
	}
}

func TestSystemRenderer_SetWidth(t *testing.T) {
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
			renderer := screen.NewSystemRenderer(tt.initial)
			renderer.SetWidth(tt.newWidth)
			assert.NotNil(t, renderer)
		})
	}
}

func TestSystemRenderer_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		snap  *model.Snapshot
	}{
		{
			name:  "compact_width",
			width: 60,
			snap:  createSystemTestSnapshot(),
		},
		{
			name:  "normal_width",
			width: 100,
			snap:  createSystemTestSnapshot(),
		},
		{
			name:  "wide_width",
			width: 180,
			snap:  createSystemTestSnapshot(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewSystemRenderer(tt.width)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "System")
		})
	}
}

func TestSystemRenderer_RenderInline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "basic_snapshot",
			snap: createSystemTestSnapshot(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewSystemRenderer(80)
			result := renderer.RenderInline(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "CPU")
			assert.Contains(t, result, "RAM")
			assert.Contains(t, result, "Load")
		})
	}
}

func TestSystemRenderer_RenderForInteractive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		snap  *model.Snapshot
	}{
		{
			name:  "with_limits",
			width: 120,
			snap: &model.Snapshot{
				System: model.SystemMetrics{
					CPUPercent:    25.5,
					MemoryPercent: 50.0,
					SwapPercent:   10.0,
					DiskPercent:   75.0,
					LoadAvg1:      1.5,
					LoadAvg5:      1.2,
					LoadAvg15:     1.0,
					MemoryUsed:    1024 * 1024 * 1024,
					MemoryTotal:   2048 * 1024 * 1024,
					SwapUsed:      512 * 1024 * 1024,
					SwapTotal:     4096 * 1024 * 1024,
					DiskUsed:      100 * 1024 * 1024 * 1024,
					DiskTotal:     500 * 1024 * 1024 * 1024,
				},
				Limits: model.ResourceLimits{
					HasLimits: true,
					CPUQuota:  2.0,
					MemoryMax: 1024 * 1024 * 1024,
				},
			},
		},
		{
			name:  "without_limits",
			width: 120,
			snap:  createSystemTestSnapshot(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewSystemRenderer(tt.width)
			result := renderer.RenderForInteractive(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "System")
		})
	}
}

func TestSystemRenderer_RenderForRaw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
		snap  *model.Snapshot
	}{
		{
			name:  "with_limits",
			width: 120,
			snap: &model.Snapshot{
				System: model.SystemMetrics{
					CPUPercent:    25.5,
					MemoryPercent: 50.0,
					SwapPercent:   10.0,
					DiskPercent:   75.0,
					LoadAvg1:      1.5,
					LoadAvg5:      1.2,
					LoadAvg15:     1.0,
					MemoryUsed:    1024 * 1024 * 1024,
					MemoryTotal:   2048 * 1024 * 1024,
					SwapUsed:      512 * 1024 * 1024,
					SwapTotal:     4096 * 1024 * 1024,
					DiskUsed:      100 * 1024 * 1024 * 1024,
					DiskTotal:     500 * 1024 * 1024 * 1024,
				},
				Limits: model.ResourceLimits{
					HasLimits: true,
					CPUQuota:  2.0,
					MemoryMax: 1024 * 1024 * 1024,
					CPUSet:    "0-3",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewSystemRenderer(tt.width)
			result := renderer.RenderForRaw(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "System (at start)")
		})
	}
}

func createSystemTestSnapshot() *model.Snapshot {
	return &model.Snapshot{
		System: model.SystemMetrics{
			CPUPercent:    25.5,
			MemoryPercent: 50.0,
			SwapPercent:   10.0,
			LoadAvg1:      1.5,
			LoadAvg5:      1.2,
			LoadAvg15:     1.0,
			MemoryUsed:    1024 * 1024 * 1024,
			MemoryTotal:   2048 * 1024 * 1024,
			SwapUsed:      512 * 1024 * 1024,
			SwapTotal:     4096 * 1024 * 1024,
		},
		Limits: model.ResourceLimits{
			HasLimits: false,
		},
	}
}
