package screen

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func Test_formatFloat2(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"zero", 0.0, "0.00"},
		{"integer", 5.0, "5.00"},
		{"decimal", 3.14, "3.14"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatFloat2(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_formatFloat1(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"zero", 0.0, "0.0"},
		{"integer", 5.0, "5.0"},
		{"decimal", 3.14, "3.1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatFloat1(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func Test_formatFloat0(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		value    float64
		expected string
	}{
		{"zero", 0.0, "0"},
		{"integer", 5.0, "5"},
		{"round_down", 3.4, "3"},
		{"round_up", 3.6, "4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			result := formatFloat0(tt.value)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSystemRenderer_createProgressBar(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		width   int
		percent float64
		label   string
		want    string
	}{
		{
			name:    "cpu_50_percent",
			width:   20,
			percent: 50,
			label:   "CPU",
			want:    "CPU",
		},
		{
			name:    "memory_75_percent",
			width:   30,
			percent: 75,
			label:   "RAM",
			want:    "RAM",
		},
		{
			name:    "zero_percent",
			width:   20,
			percent: 0,
			label:   "DSK",
			want:    "DSK",
		},
		{
			name:    "full_percent",
			width:   20,
			percent: 100,
			label:   "SWP",
			want:    "SWP",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 80,
			}
			bar := renderer.createProgressBar(tt.width, tt.percent, tt.label)
			assert.NotNil(t, bar)
			rendered := bar.Render()
			assert.NotEmpty(t, rendered)
			assert.Contains(t, rendered, tt.want)
		})
	}
}

func TestSystemRenderer_appendLimitsNormal(t *testing.T) {
	t.Parallel()

	renderer := &SystemRenderer{
		theme: ansi.DefaultTheme(),
		width: 80,
	}

	tests := []struct {
		name       string
		limits     model.ResourceLimits
		expectMore bool
	}{
		{"no_limits", model.ResourceLimits{HasLimits: false}, false},
		{"with_cpu_limit", model.ResourceLimits{HasLimits: true, CPUQuota: 2.0}, true},
		{"with_memory_limit", model.ResourceLimits{HasLimits: true, MemoryMax: 4 * 1024 * 1024 * 1024}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			lines := []string{"line1", "line2"}
			result := renderer.appendLimitsNormal(lines, tt.limits)
			if tt.expectMore {
				assert.Greater(t, len(result), len(lines))
			} else {
				assert.Equal(t, len(lines), len(result))
			}
		})
	}
}

func TestSystemRenderer_renderCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "basic_metrics",
			snap: &model.Snapshot{
				System: model.SystemMetrics{
					CPUPercent:    25.5,
					MemoryPercent: 50.0,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 60,
			}
			result := renderer.renderCompact(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "CPU")
			assert.Contains(t, result, "RAM")
		})
	}
}

func TestSystemRenderer_renderNormal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "basic_metrics",
			snap: &model.Snapshot{
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
			},
		},
		{
			name: "with_limits",
			snap: &model.Snapshot{
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
					HasLimits: true,
					CPUQuota:  2.0,
					MemoryMax: 1024 * 1024 * 1024,
					PIDsMax:   1000,
					PIDsCurrent: 100,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 100,
			}
			result := renderer.renderNormal(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestSystemRenderer_renderWide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap *model.Snapshot
	}{
		{
			name: "basic_metrics",
			snap: &model.Snapshot{
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
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 180,
			}
			result := renderer.renderWide(tt.snap)
			assert.NotEmpty(t, result)
		})
	}
}

func TestSystemRenderer_createInteractiveBars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		sys    model.SystemMetrics
	}{
		{
			name: "basic_metrics",
			sys: model.SystemMetrics{
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 120,
			}
			bars, infos := renderer.createInteractiveBars(40, tt.sys)
			assert.Len(t, bars, metricBarCount)
			assert.Len(t, infos, metricBarCount)
			for i := range metricBarCount {
				assert.NotNil(t, bars[i])
				assert.NotEmpty(t, infos[i])
			}
		})
	}
}

func TestSystemRenderer_createRawBars(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		sys    model.SystemMetrics
	}{
		{
			name: "basic_metrics",
			sys: model.SystemMetrics{
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
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 120,
			}
			bars, infos := renderer.createRawBars(40, tt.sys)
			assert.Len(t, bars, metricBarCount)
			assert.Len(t, infos, metricBarCount)
		})
	}
}

func TestSystemRenderer_appendLimitsWide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		limits     model.ResourceLimits
		expectMore bool
	}{
		{
			name:       "no_limits",
			limits:     model.ResourceLimits{HasLimits: false},
			expectMore: false,
		},
		{
			name: "with_cpu_and_memory",
			limits: model.ResourceLimits{
				HasLimits:     true,
				CPUQuota:      2.0,
				CPUSet:        "0-3",
				MemoryMax:     1024 * 1024 * 1024,
				MemoryCurrent: 512 * 1024 * 1024,
			},
			expectMore: true,
		},
		{
			name: "with_pids",
			limits: model.ResourceLimits{
				HasLimits:   true,
				PIDsMax:     1000,
				PIDsCurrent: 100,
			},
			expectMore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 180,
			}
			lines := []string{"line1"}
			result := renderer.appendLimitsWide(lines, tt.limits)
			if tt.expectMore {
				assert.Greater(t, len(result), len(lines))
			} else {
				assert.Equal(t, len(lines), len(result))
			}
		})
	}
}

func TestSystemRenderer_appendLimitsInteractive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		limits     model.ResourceLimits
		expectMore bool
	}{
		{
			name:       "no_limits",
			limits:     model.ResourceLimits{HasLimits: false},
			expectMore: false,
		},
		{
			name: "with_all_limits",
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  2.0,
				MemoryMax: 1024 * 1024 * 1024,
				CPUSet:    "0-3",
			},
			expectMore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 120,
			}
			lines := []string{"line1"}
			result := renderer.appendLimitsInteractive(lines, tt.limits)
			if tt.expectMore {
				assert.Greater(t, len(result), len(lines))
			} else {
				assert.Equal(t, len(lines), len(result))
			}
		})
	}
}

func TestSystemRenderer_appendLimitsRaw(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		limits     model.ResourceLimits
		expectMore bool
	}{
		{
			name:       "no_limits",
			limits:     model.ResourceLimits{HasLimits: false},
			expectMore: false,
		},
		{
			name: "with_limits",
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  2.0,
				MemoryMax: 1024 * 1024 * 1024,
				CPUSet:    "0-3",
			},
			expectMore: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &SystemRenderer{
				theme: ansi.DefaultTheme(),
				width: 120,
			}
			lines := []string{"line1"}
			result := renderer.appendLimitsRaw(lines, tt.limits)
			if tt.expectMore {
				assert.Greater(t, len(result), len(lines))
			} else {
				assert.Equal(t, len(lines), len(result))
			}
		})
	}
}
