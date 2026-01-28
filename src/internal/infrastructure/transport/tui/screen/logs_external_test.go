package screen_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/screen"
	"github.com/stretchr/testify/assert"
)

func TestNewLogsRenderer(t *testing.T) {
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
			renderer := screen.NewLogsRenderer(tt.width)
			assert.NotNil(t, renderer)
		})
	}
}

func TestLogsRenderer_SetWidth(t *testing.T) {
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
			renderer := screen.NewLogsRenderer(tt.initial)
			renderer.SetWidth(tt.newWidth)
			assert.NotNil(t, renderer)
		})
	}
}

func TestLogsRenderer_Render(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "basic_render",
			snap: &model.Snapshot{
				Logs: model.LogSummary{},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewLogsRenderer(80)
			result := renderer.Render(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLogsRenderer_RenderBadge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "no_issues",
			snap: &model.Snapshot{
				Logs: model.LogSummary{
					WarnCount:  0,
					ErrorCount: 0,
				},
			},
		},
		{
			name: "with_warnings",
			snap: &model.Snapshot{
				Logs: model.LogSummary{
					WarnCount:  5,
					ErrorCount: 0,
				},
			},
		},
		{
			name: "with_errors",
			snap: &model.Snapshot{
				Logs: model.LogSummary{
					WarnCount:  0,
					ErrorCount: 3,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewLogsRenderer(80)
			result := renderer.RenderBadge(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestLogsRenderer_RenderInline(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "empty_logs",
			snap: &model.Snapshot{
				Logs: model.LogSummary{},
			},
		},
		{
			name: "with_counts",
			snap: &model.Snapshot{
				Logs: model.LogSummary{
					InfoCount:  10,
					WarnCount:  2,
					ErrorCount: 1,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := screen.NewLogsRenderer(80)
			result := renderer.RenderInline(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}
