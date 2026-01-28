// Package tui provides terminal user interface for superviz.io.
package tui

import (
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockSnapshot creates a test snapshot with reasonable defaults.
func mockSnapshot() *model.Snapshot {
	snap := model.NewSnapshot()
	snap.Context = model.RuntimeContext{
		Hostname:         "testhost",
		Version:          "1.2.3",
		OS:               "linux",
		Arch:             "amd64",
		Kernel:           "6.1.0",
		Mode:             model.ModeContainer,
		ContainerRuntime: "docker",
		DaemonPID:        1234,
		StartTime:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
		Uptime:           2 * time.Hour,
		PrimaryIP:        "10.0.0.5",
		ConfigPath:       "/etc/supervizio/config.yaml",
	}
	snap.System = model.SystemMetrics{
		CPUPercent:    45.5,
		LoadAvg1:      2.3,
		LoadAvg5:      1.8,
		MemoryTotal:   16 * 1024 * 1024 * 1024,
		MemoryUsed:    8 * 1024 * 1024 * 1024,
		MemoryPercent: 50.0,
		SwapTotal:     4 * 1024 * 1024 * 1024,
		SwapUsed:      1 * 1024 * 1024 * 1024,
		SwapPercent:   25.0,
		DiskTotal:     500 * 1024 * 1024 * 1024,
		DiskUsed:      200 * 1024 * 1024 * 1024,
		DiskPercent:   40.0,
		DiskPath:      "/",
	}
	snap.Limits = model.ResourceLimits{
		HasLimits:   true,
		CPUQuota:    2.5,
		CPUSet:      "0-3",
		MemoryMax:   8 * 1024 * 1024 * 1024,
		CPUQuotaRaw: 250000,
		CPUPeriod:   100000,
	}
	snap.Sandboxes = []model.SandboxInfo{
		{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock", Version: "24.0.0"},
		{Name: "podman", Detected: false, Endpoint: "", Version: ""},
	}
	return snap
}

// Test_renderHeader tests the header rendering with different terminal widths.
func Test_RawRenderer_renderHeader(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		version    string
		hostname   string
		wantSubstr []string
	}{
		{
			name:     "standard_width",
			width:    80,
			version:  "1.0.0",
			hostname: "testhost",
			wantSubstr: []string{
				"superviz",
				".io",
				"v1.0.0",
				"testhost",
				"linux/amd64",
			},
		},
		{
			name:     "narrow_terminal",
			width:    60,
			version:  "2.0.0-beta",
			hostname: "myhost",
			wantSubstr: []string{
				"superviz",
				"v2.0.0-beta",
				"myhost",
			},
		},
		{
			name:     "wide_terminal",
			width:    120,
			version:  "3.5.1",
			hostname: "production-server",
			wantSubstr: []string{
				"superviz",
				"v3.5.1",
				"production-server",
			},
		},
		{
			name:     "version_without_v_prefix",
			width:    80,
			version:  "1.0.0",
			hostname: "testhost",
			wantSubstr: []string{
				"v1.0.0",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := mockSnapshot()
			snap.Context.Version = tt.version
			snap.Context.Hostname = tt.hostname

			r := &RawRenderer{
				width:  tt.width,
				height: 24,
				theme:  defaultTheme(),
			}

			result := r.renderHeader(snap)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr, "expected substring not found")
			}
		})
	}
}

// Test_buildTitleLine tests title line formatting.
func Test_RawRenderer_buildTitleLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		version    string
		width      int
		wantSubstr []string
	}{
		{
			name:       "with_v_prefix",
			version:    "v1.0.0",
			width:      80,
			wantSubstr: []string{"superviz", ".io", "v1.0.0"},
		},
		{
			name:       "without_v_prefix",
			version:    "2.3.4",
			width:      80,
			wantSubstr: []string{"superviz", ".io", "v2.3.4"},
		},
		{
			name:       "empty_version",
			version:    "",
			width:      80,
			wantSubstr: []string{"superviz", ".io"},
		},
		{
			name:       "narrow_width",
			version:    "1.0.0",
			width:      40,
			wantSubstr: []string{"superviz", ".io", "v1.0.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &RawRenderer{
				width: tt.width,
				theme: defaultTheme(),
			}

			result := r.buildTitleLine(tt.version)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr)
			}
		})
	}
}

// Test_buildHeaderContentLines tests header content line generation.
func Test_RawRenderer_buildHeaderContentLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ctx        model.RuntimeContext
		wantSubstr []string
		wantLines  int
	}{
		{
			name: "full_context",
			ctx: model.RuntimeContext{
				Hostname:         "myhost",
				OS:               "linux",
				Arch:             "amd64",
				Mode:             model.ModeContainer,
				ContainerRuntime: "docker",
				ConfigPath:       "/custom/config.yaml",
				StartTime:        time.Date(2024, 1, 15, 10, 30, 0, 0, time.UTC),
			},
			wantSubstr: []string{
				"myhost",
				"linux/amd64",
				"container (docker)",
				"/custom/config.yaml",
				"2024-01-15T10:30:00Z",
			},
			wantLines: 5,
		},
		{
			name: "host_mode",
			ctx: model.RuntimeContext{
				Hostname:   "barehost",
				OS:         "darwin",
				Arch:       "arm64",
				Mode:       model.ModeHost,
				ConfigPath: "",
				StartTime:  time.Date(2024, 2, 1, 8, 0, 0, 0, time.UTC),
			},
			wantSubstr: []string{
				"barehost",
				"darwin/arm64",
				"host",
				"/etc/supervizio/config.yaml",
				"2024-02-01T08:00:00Z",
			},
			wantLines: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &RawRenderer{
				width: 80,
				theme: defaultTheme(),
			}

			result := r.buildHeaderContentLines(tt.ctx)

			assert.Len(t, result, tt.wantLines)
			fullText := strings.Join(result, "\n")
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, fullText, substr)
			}
		})
	}
}

// Test_buildSystemContentLines tests system section content generation.
func Test_RawRenderer_buildSystemContentLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		system    model.SystemMetrics
		limits    model.ResourceLimits
		width     int
		wantLines int
	}{
		{
			name: "with_limits",
			system: model.SystemMetrics{
				CPUPercent:    50.0,
				LoadAvg1:      2.5,
				LoadAvg5:      2.0,
				MemoryTotal:   8 * 1024 * 1024 * 1024,
				MemoryUsed:    4 * 1024 * 1024 * 1024,
				MemoryPercent: 50.0,
				SwapTotal:     2 * 1024 * 1024 * 1024,
				SwapUsed:      500 * 1024 * 1024,
				SwapPercent:   25.0,
				DiskTotal:     100 * 1024 * 1024 * 1024,
				DiskUsed:      50 * 1024 * 1024 * 1024,
				DiskPercent:   50.0,
			},
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  2.0,
				CPUSet:    "0-1",
				MemoryMax: 4 * 1024 * 1024 * 1024,
			},
			width:     80,
			wantLines: 5, // 4 metrics + 1 limits line
		},
		{
			name: "without_limits",
			system: model.SystemMetrics{
				CPUPercent:    25.0,
				LoadAvg1:      1.0,
				LoadAvg5:      0.8,
				MemoryTotal:   16 * 1024 * 1024 * 1024,
				MemoryUsed:    2 * 1024 * 1024 * 1024,
				MemoryPercent: 12.5,
				SwapTotal:     0,
				SwapUsed:      0,
				SwapPercent:   0,
				DiskTotal:     500 * 1024 * 1024 * 1024,
				DiskUsed:      100 * 1024 * 1024 * 1024,
				DiskPercent:   20.0,
			},
			limits: model.ResourceLimits{
				HasLimits: false,
			},
			width:     100,
			wantLines: 4, // 4 metrics only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := &model.Snapshot{
				System: tt.system,
				Limits: tt.limits,
			}

			r := &RawRenderer{
				width: tt.width,
				theme: defaultTheme(),
			}

			result := r.buildSystemContentLines(snap, tt.width)

			assert.Len(t, result, tt.wantLines)
			for _, line := range result {
				assert.NotEmpty(t, line)
			}
		})
	}
}

// Test_buildSystemMetricLines tests system metric line generation.
func Test_RawRenderer_buildSystemMetricLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		system     model.SystemMetrics
		barWidth   int
		wantSubstr []string
	}{
		{
			name: "normal_metrics",
			system: model.SystemMetrics{
				CPUPercent:    45.5,
				LoadAvg1:      2.3,
				LoadAvg5:      1.8,
				MemoryTotal:   8 * 1024 * 1024 * 1024,
				MemoryUsed:    4 * 1024 * 1024 * 1024,
				MemoryPercent: 50.0,
				SwapTotal:     2 * 1024 * 1024 * 1024,
				SwapUsed:      500 * 1024 * 1024,
				SwapPercent:   25.0,
				DiskTotal:     100 * 1024 * 1024 * 1024,
				DiskUsed:      40 * 1024 * 1024 * 1024,
				DiskPercent:   40.0,
			},
			barWidth: 30,
			wantSubstr: []string{
				"CPU",
				"RAM",
				"Swap",
				"Disk",
				"Load",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &RawRenderer{
				theme: defaultTheme(),
			}

			result := r.buildSystemMetricLines(tt.system, tt.barWidth)

			assert.Len(t, result, 4) // CPU, RAM, Swap, Disk
			fullText := strings.Join(result, "\n")
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, fullText, substr)
			}
		})
	}
}

// Test_buildLimitsLine tests resource limits line generation.
func Test_RawRenderer_buildLimitsLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		limits     model.ResourceLimits
		wantSubstr []string
		wantEmpty  bool
	}{
		{
			name: "all_limits",
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  2.5,
				CPUSet:    "0-3",
				MemoryMax: 8 * 1024 * 1024 * 1024,
			},
			wantSubstr: []string{"Limits", "CPUSet: 0-3", "CPU: 2.5 cores", "MEM:"},
			wantEmpty:  false,
		},
		{
			name: "cpu_only",
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  1.5,
			},
			wantSubstr: []string{"Limits", "CPU: 1.5 cores"},
			wantEmpty:  false,
		},
		{
			name: "memory_only",
			limits: model.ResourceLimits{
				HasLimits: true,
				MemoryMax: 4 * 1024 * 1024 * 1024,
			},
			wantSubstr: []string{"Limits", "MEM:"},
			wantEmpty:  false,
		},
		{
			name: "no_limits",
			limits: model.ResourceLimits{
				HasLimits: false,
			},
			wantEmpty: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &RawRenderer{
				theme: defaultTheme(),
			}

			result := r.buildLimitsLine(tt.limits)

			if tt.wantEmpty {
				assert.Empty(t, result)
			} else {
				require.NotEmpty(t, result)
				for _, substr := range tt.wantSubstr {
					assert.Contains(t, result, substr)
				}
			}
		})
	}
}

// Test_hasAnySandboxes tests sandbox detection.
func Test_RawRenderer_hasAnySandboxes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		sandboxes []model.SandboxInfo
		want      bool
	}{
		{
			name: "has_sandboxes",
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true},
			},
			want: true,
		},
		{
			name:      "no_sandboxes",
			sandboxes: []model.SandboxInfo{},
			want:      false,
		},
		{
			name:      "nil_sandboxes",
			sandboxes: nil,
			want:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := &model.Snapshot{
				Sandboxes: tt.sandboxes,
			}

			r := &RawRenderer{}

			result := r.hasAnySandboxes(snap)

			assert.Equal(t, tt.want, result)
		})
	}
}

// Test_buildSandboxContentLines tests sandbox content line generation.
func Test_RawRenderer_buildSandboxContentLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		sandboxes  []model.SandboxInfo
		width      int
		wantLines  int
		wantSubstr []string
	}{
		{
			name: "detected_and_not_detected",
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
				{Name: "podman", Detected: false, Endpoint: ""},
				{Name: "lxc", Detected: true, Endpoint: "unix:///var/run/lxc.sock"},
			},
			width:     80,
			wantLines: 3,
			wantSubstr: []string{
				"docker",
				"unix:///var/run/docker.sock",
				"podman",
				"not detected",
				"lxc",
			},
		},
		{
			name: "long_endpoint_truncation",
			sandboxes: []model.SandboxInfo{
				{Name: "kubernetes", Detected: true, Endpoint: "https://very-long-kubernetes-endpoint-url.example.com:6443/api/v1/namespaces/default"},
			},
			width:      50,
			wantLines:  1,
			wantSubstr: []string{"kubernetes", "..."},
		},
		{
			name:      "empty_sandboxes",
			sandboxes: []model.SandboxInfo{},
			width:     80,
			wantLines: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := &model.Snapshot{
				Sandboxes: tt.sandboxes,
			}

			r := &RawRenderer{
				theme: defaultTheme(),
			}

			result := r.buildSandboxContentLines(snap, tt.width)

			assert.Len(t, result, tt.wantLines)
			if len(tt.wantSubstr) > 0 {
				fullText := strings.Join(result, "\n")
				for _, substr := range tt.wantSubstr {
					assert.Contains(t, fullText, substr)
				}
			}
		})
	}
}

// Test_formatSandboxLine tests sandbox line formatting.
func Test_RawRenderer_formatSandboxLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		icon        string
		sandboxName string
		endpoint    string
		wantSubstr  []string
	}{
		{
			name:        "short_name",
			icon:        "✓",
			sandboxName: "docker",
			endpoint:    "unix:///var/run/docker.sock",
			wantSubstr:  []string{"✓", "docker", "unix:///var/run/docker.sock"},
		},
		{
			name:        "long_name",
			icon:        "✗",
			sandboxName: "kubernetes",
			endpoint:    "not detected",
			wantSubstr:  []string{"✗", "kubernetes", "not detected"},
		},
		{
			name:        "empty_endpoint",
			icon:        "✓",
			sandboxName: "podman",
			endpoint:    "",
			wantSubstr:  []string{"✓", "podman"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &RawRenderer{}

			result := r.formatSandboxLine(tt.icon, tt.sandboxName, tt.endpoint)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr)
			}
		})
	}
}

// Test_calculateMaxVisibleWidth tests visible width calculation.
func Test_calculateMaxVisibleWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		lines []string
		want  int
	}{
		{
			name: "simple_lines",
			lines: []string{
				"short",
				"longer line",
				"medium",
			},
			want: 11, // "longer line"
		},
		{
			name: "with_ansi_codes",
			lines: []string{
				"\x1b[31mred\x1b[0m",          // 3 visible chars
				"\x1b[1;32mbold green\x1b[0m", // 10 visible chars
				"plain text",                  // 10 visible chars
			},
			want: 10,
		},
		{
			name:  "empty_lines",
			lines: []string{},
			want:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := calculateMaxVisibleWidth(tt.lines)

			assert.Equal(t, tt.want, result)
		})
	}
}

// Test_removeTrailingEmptyLines tests trailing empty line removal.
func Test_removeTrailingEmptyLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		in   []string
		want []string
	}{
		{
			name: "with_trailing_empty",
			in:   []string{"line1", "line2", "", ""},
			want: []string{"line1", "line2"},
		},
		{
			name: "no_trailing_empty",
			in:   []string{"line1", "line2", "line3"},
			want: []string{"line1", "line2", "line3"},
		},
		{
			name: "all_empty",
			in:   []string{"", "", ""},
			want: []string{},
		},
		{
			name: "empty_in_middle",
			in:   []string{"line1", "", "line3", ""},
			want: []string{"line1", "", "line3"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := removeTrailingEmptyLines(tt.in)

			assert.Equal(t, tt.want, result)
		})
	}
}

// Test_getLineOrEmpty tests line retrieval helper.
func Test_getLineOrEmpty(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		lines []string
		idx   int
		want  string
	}{
		{
			name:  "valid_index",
			lines: []string{"line0", "line1", "line2"},
			idx:   1,
			want:  "line1",
		},
		{
			name:  "out_of_bounds",
			lines: []string{"line0", "line1"},
			idx:   5,
			want:  "",
		},
		{
			name:  "empty_slice",
			lines: []string{},
			idx:   0,
			want:  "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := getLineOrEmpty(tt.lines, tt.idx)

			assert.Equal(t, tt.want, result)
		})
	}
}

// Test_padLineToWidth tests line padding.
func Test_padLineToWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		line       string
		width      int
		wantMinLen int
		wantSpace  bool
	}{
		{
			name:       "needs_padding",
			line:       "short",
			width:      10,
			wantMinLen: 10,
			wantSpace:  true,
		},
		{
			name:       "no_padding_needed",
			line:       "exact length",
			width:      12,
			wantMinLen: 12,
			wantSpace:  false,
		},
		{
			name:       "already_longer",
			line:       "this is a longer line",
			width:      10,
			wantMinLen: 21,
			wantSpace:  false,
		},
		{
			name:       "with_ansi_codes",
			line:       "\x1b[31mred\x1b[0m",
			width:      10,
			wantMinLen: 10,
			wantSpace:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := padLineToWidth(tt.line, tt.width)

			// Check minimum length (visible width).
			assert.GreaterOrEqual(t, len(result), tt.wantMinLen)
			if tt.wantSpace {
				assert.Contains(t, result, " ")
			}
		})
	}
}

// Test_mergeLines tests line merging for side-by-side layout.
func Test_mergeLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		leftLines  []string
		rightLines []string
		leftWidth  int
		separator  string
		wantLines  int
	}{
		{
			name:       "equal_lengths",
			leftLines:  []string{"L1", "L2", "L3"},
			rightLines: []string{"R1", "R2", "R3"},
			leftWidth:  10,
			separator:  " ",
			wantLines:  3,
		},
		{
			name:       "left_longer",
			leftLines:  []string{"L1", "L2", "L3", "L4"},
			rightLines: []string{"R1", "R2"},
			leftWidth:  10,
			separator:  " | ",
			wantLines:  4,
		},
		{
			name:       "right_longer",
			leftLines:  []string{"L1"},
			rightLines: []string{"R1", "R2", "R3"},
			leftWidth:  10,
			separator:  " ",
			wantLines:  3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mergeLines(tt.leftLines, tt.rightLines, tt.leftWidth, tt.separator)

			lines := strings.Split(result, "\n")
			assert.Len(t, lines, tt.wantLines)

			// Verify separator is present in each line.
			for _, line := range lines {
				assert.Contains(t, line, tt.separator)
			}
		})
	}
}

// Test_mergeSideBySide tests full side-by-side merging.
func Test_mergeSideBySide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		left      string
		right     string
		separator string
		wantLines int
	}{
		{
			name:      "simple_merge",
			left:      "Left1\nLeft2\nLeft3",
			right:     "Right1\nRight2\nRight3",
			separator: " ",
			wantLines: 3,
		},
		{
			name:      "with_trailing_newlines",
			left:      "Left1\nLeft2\n\n",
			right:     "Right1\nRight2\n\n",
			separator: " | ",
			wantLines: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := mergeSideBySide(tt.left, tt.right, tt.separator)

			lines := strings.Split(result, "\n")
			assert.Equal(t, tt.wantLines, len(lines))
		})
	}
}

// Test_renderHeaderCompact tests compact header rendering.
func Test_RawRenderer_renderHeaderCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		width            int
		hostname         string
		version          string
		containerRuntime string
		mode             model.RuntimeMode
		wantSubstr       []string
	}{
		{
			name:             "container_mode_with_runtime",
			width:            80,
			hostname:         "prodhost",
			version:          "2.0.0",
			containerRuntime: "docker",
			mode:             model.ModeContainer,
			wantSubstr:       []string{"superviz", ".io", "v2.0.0", "prodhost", "docker"},
		},
		{
			name:             "host_mode_no_runtime",
			width:            80,
			hostname:         "barehost",
			version:          "1.5.0",
			containerRuntime: "",
			mode:             model.ModeHost,
			wantSubstr:       []string{"superviz", ".io", "v1.5.0", "barehost", "host"},
		},
		{
			name:             "narrow_terminal",
			width:            60,
			hostname:         "testhost",
			version:          "1.0.0",
			containerRuntime: "",
			mode:             model.ModeHost,
			wantSubstr:       []string{"superviz", "testhost"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := mockSnapshot()
			snap.Context.Hostname = tt.hostname
			snap.Context.Version = tt.version
			snap.Context.ContainerRuntime = tt.containerRuntime
			snap.Context.Mode = tt.mode

			r := &RawRenderer{
				width:  tt.width,
				height: 24,
				theme:  defaultTheme(),
			}

			result := r.renderHeaderCompact(snap)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr, "expected substring not found")
			}
		})
	}
}

// Test_renderSystemSection tests system section rendering.
func Test_RawRenderer_renderSystemSection(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		wantSubstr []string
	}{
		{
			name:       "standard_width",
			width:      80,
			wantSubstr: []string{"System"},
		},
		{
			name:       "wide_terminal",
			width:      120,
			wantSubstr: []string{"System"},
		},
		{
			name:       "narrow_terminal",
			width:      60,
			wantSubstr: []string{"System"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := mockSnapshot()

			r := &RawRenderer{
				width:  tt.width,
				height: 24,
				theme:  defaultTheme(),
			}

			result := r.renderSystemSection(snap)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr, "expected substring not found")
			}
		})
	}
}

// Test_renderSystemAndSandboxesSideBySide tests side-by-side layout.
func Test_RawRenderer_renderSystemAndSandboxesSideBySide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		sandboxes  []model.SandboxInfo
		limits     model.ResourceLimits
		wantSubstr []string
	}{
		{
			name:  "wide_terminal_with_sandboxes",
			width: 160,
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
				{Name: "podman", Detected: false, Endpoint: ""},
			},
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  2.0,
				CPUSet:    "0-1",
				MemoryMax: 4 * 1024 * 1024 * 1024,
			},
			wantSubstr: []string{"System (at start)", "Sandboxes", "docker"},
		},
		{
			name:      "wide_terminal_no_sandboxes",
			width:     160,
			sandboxes: []model.SandboxInfo{},
			limits: model.ResourceLimits{
				HasLimits: false,
			},
			wantSubstr: []string{"System (at start)", "Sandboxes"},
		},
		{
			name:  "extra_wide_terminal",
			width: 200,
			sandboxes: []model.SandboxInfo{
				{Name: "kubernetes", Detected: true, Endpoint: "https://k8s.example.com:6443"},
			},
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  4.0,
			},
			wantSubstr: []string{"System (at start)", "Sandboxes", "kubernetes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := mockSnapshot()
			snap.Sandboxes = tt.sandboxes
			snap.Limits = tt.limits

			r := &RawRenderer{
				width:  tt.width,
				height: 40,
				theme:  defaultTheme(),
			}

			result := r.renderSystemAndSandboxesSideBySide(snap)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr, "expected substring not found")
			}
		})
	}
}

// Test_renderSandboxes tests sandboxes section rendering.
func Test_RawRenderer_renderSandboxes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		sandboxes  []model.SandboxInfo
		wantSubstr []string
	}{
		{
			name:  "detected_sandboxes",
			width: 80,
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
				{Name: "podman", Detected: true, Endpoint: "unix:///var/run/podman.sock"},
			},
			wantSubstr: []string{"Sandboxes", "docker", "podman"},
		},
		{
			name:  "mixed_detection",
			width: 80,
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
				{Name: "lxc", Detected: false, Endpoint: ""},
			},
			wantSubstr: []string{"Sandboxes", "docker", "lxc", "not detected"},
		},
		{
			name:  "narrow_terminal",
			width: 60,
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
			},
			wantSubstr: []string{"Sandboxes", "docker"},
		},
		{
			name:       "empty_sandboxes",
			width:      80,
			sandboxes:  []model.SandboxInfo{},
			wantSubstr: []string{"Sandboxes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := &model.Snapshot{
				Sandboxes: tt.sandboxes,
			}

			r := &RawRenderer{
				width:  tt.width,
				height: 24,
				theme:  defaultTheme(),
			}

			result := r.renderSandboxes(snap)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr, "expected substring not found")
			}
		})
	}
}

// Test_renderSandboxesWithWidth tests sandboxes rendering with explicit width.
func Test_RawRenderer_renderSandboxesWithWidth(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		sandboxes  []model.SandboxInfo
		wantSubstr []string
	}{
		{
			name:  "standard_width",
			width: 80,
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
			},
			wantSubstr: []string{"Sandboxes", "docker"},
		},
		{
			name:  "narrow_width_truncates_endpoint",
			width: 50,
			sandboxes: []model.SandboxInfo{
				{Name: "kubernetes", Detected: true, Endpoint: "https://very-long-kubernetes-endpoint.example.com:6443"},
			},
			wantSubstr: []string{"Sandboxes", "kubernetes", "..."},
		},
		{
			name:  "wide_width",
			width: 120,
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
				{Name: "podman", Detected: false, Endpoint: ""},
			},
			wantSubstr: []string{"Sandboxes", "docker", "podman", "not detected"},
		},
		{
			name:       "empty_sandboxes_with_width",
			width:      80,
			sandboxes:  []model.SandboxInfo{},
			wantSubstr: []string{"Sandboxes"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := &model.Snapshot{
				Sandboxes: tt.sandboxes,
			}

			r := &RawRenderer{
				width:  100, // Different from the passed width.
				height: 24,
				theme:  defaultTheme(),
			}

			result := r.renderSandboxesWithWidth(snap, tt.width)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr, "expected substring not found")
			}
		})
	}
}

// defaultTheme returns the default ANSI theme for testing.
func defaultTheme() ansi.Theme {
	return ansi.DefaultTheme()
}

// Test_Render tests the main Render method with different terminal widths and data.
// Test_Render_WriterError tests Render error handling with failing writer.
func Test_RawRenderer_Render_WriterError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap func() *model.Snapshot
		out  io.Writer
	}{
		{
			name: "standard_snapshot_write_error",
			snap: mockSnapshot,
			out:  &failingWriter{},
		},
		{
			name: "minimal_snapshot_write_error",
			snap: func() *model.Snapshot {
				snap := model.NewSnapshot()
				snap.Context = model.RuntimeContext{
					Hostname: "minimal",
					Version:  "1.0.0",
					OS:       "linux",
					Arch:     "amd64",
				}
				return snap
			},
			out: &failingWriter{},
		},
		{
			name: "full_snapshot_write_error",
			snap: func() *model.Snapshot {
				snap := mockSnapshot()
				snap.Sandboxes = []model.SandboxInfo{
					{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
					{Name: "podman", Detected: true, Endpoint: "unix:///var/run/podman.sock"},
				}
				return snap
			},
			out: &failingWriter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &RawRenderer{
				width:  80,
				height: 24,
				out:    tt.out,
				theme:  defaultTheme(),
			}

			err := r.Render(tt.snap())

			assert.Error(t, err, "expected error from failing writer")
		})
	}
}

// Test_RenderCompact_WriterError tests RenderCompact error handling with failing writer.
func Test_RawRenderer_RenderCompact_WriterError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		snap func() *model.Snapshot
		out  io.Writer
	}{
		{
			name: "standard_snapshot_write_error",
			snap: mockSnapshot,
			out:  &failingWriter{},
		},
		{
			name: "host_mode_write_error",
			snap: func() *model.Snapshot {
				snap := mockSnapshot()
				snap.Context.Mode = model.ModeHost
				snap.Context.ContainerRuntime = ""
				return snap
			},
			out: &failingWriter{},
		},
		{
			name: "container_mode_write_error",
			snap: func() *model.Snapshot {
				snap := mockSnapshot()
				snap.Context.Mode = model.ModeContainer
				snap.Context.ContainerRuntime = "docker"
				return snap
			},
			out: &failingWriter{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := &RawRenderer{
				width:  80,
				height: 24,
				out:    tt.out,
				theme:  defaultTheme(),
			}

			err := r.RenderCompact(tt.snap())

			assert.Error(t, err, "expected error from failing writer")
		})
	}
}

// failingWriter always returns an error on Write.
type failingWriter struct{}

// Write always returns an error.
func (fw *failingWriter) Write(p []byte) (n int, err error) {
	return 0, fmt.Errorf("write failed")
}

// Test_Render_LayoutDecision tests that Render chooses the correct layout based on width.
func Test_RawRenderer_Render_LayoutDecision(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		width          int
		wantSideBySide bool
	}{
		{
			name:           "narrow_stacked",
			width:          79,
			wantSideBySide: false,
		},
		{
			name:           "at_threshold_stacked",
			width:          159,
			wantSideBySide: false,
		},
		{
			name:           "wide_side_by_side",
			width:          160,
			wantSideBySide: true,
		},
		{
			name:           "extra_wide_side_by_side",
			width:          200,
			wantSideBySide: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			r := &RawRenderer{
				width:  tt.width,
				height: 40,
				out:    &buf,
				theme:  defaultTheme(),
			}

			snap := mockSnapshot()
			err := r.Render(snap)
			require.NoError(t, err)

			output := buf.String()

			if tt.wantSideBySide {
				// In wide mode, both "System (at start)" and "Sandboxes" appear as titles.
				assert.Contains(t, output, "System (at start)", "expected side-by-side layout")
			} else {
				// In stacked mode, "System" appears alone (not with "(at start)").
				assert.Contains(t, output, "System", "expected stacked layout")
			}
		})
	}
}

// Test_Render_EmptySnapshot tests Render with minimal snapshot data.
func Test_RawRenderer_Render_EmptySnapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hostname   string
		version    string
		os         string
		arch       string
		wantSubstr []string
	}{
		{
			name:     "minimal_snapshot_linux_amd64",
			hostname: "emptyhost",
			version:  "0.0.0",
			os:       "linux",
			arch:     "amd64",
			wantSubstr: []string{
				"superviz",
				"emptyhost",
				"linux/amd64",
			},
		},
		{
			name:     "minimal_snapshot_darwin_arm64",
			hostname: "machost",
			version:  "1.0.0",
			os:       "darwin",
			arch:     "arm64",
			wantSubstr: []string{
				"superviz",
				"machost",
				"darwin/arm64",
			},
		},
		{
			name:     "minimal_snapshot_no_version",
			hostname: "testhost",
			version:  "",
			os:       "linux",
			arch:     "amd64",
			wantSubstr: []string{
				"superviz",
				"testhost",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			r := &RawRenderer{
				width:  80,
				height: 24,
				out:    &buf,
				theme:  defaultTheme(),
			}

			snap := model.NewSnapshot()
			snap.Context = model.RuntimeContext{
				Hostname:  tt.hostname,
				Version:   tt.version,
				OS:        tt.os,
				Arch:      tt.arch,
				Mode:      model.ModeHost,
				StartTime: time.Now(),
			}

			err := r.Render(snap)

			require.NoError(t, err)
			output := buf.String()
			require.NotEmpty(t, output)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, output, substr)
			}
		})
	}
}

// Test_RenderCompact_EmptySnapshot tests RenderCompact with minimal snapshot data.
func Test_RawRenderer_RenderCompact_EmptySnapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		hostname   string
		version    string
		os         string
		arch       string
		mode       model.RuntimeMode
		runtime    string
		wantSubstr []string
	}{
		{
			name:     "minimal_snapshot_host_mode",
			hostname: "compacthost",
			version:  "1.0.0",
			os:       "linux",
			arch:     "arm64",
			mode:     model.ModeContainer,
			runtime:  "",
			wantSubstr: []string{
				"superviz",
				"compacthost",
			},
		},
		{
			name:     "minimal_snapshot_container_mode",
			hostname: "dockerhost",
			version:  "2.0.0",
			os:       "linux",
			arch:     "amd64",
			mode:     model.ModeContainer,
			runtime:  "docker",
			wantSubstr: []string{
				"superviz",
				"dockerhost",
				"docker",
			},
		},
		{
			name:     "minimal_snapshot_vm_mode",
			hostname: "vmhost",
			version:  "3.0.0",
			os:       "linux",
			arch:     "amd64",
			mode:     model.ModeVM,
			runtime:  "qemu",
			wantSubstr: []string{
				"superviz",
				"vmhost",
				"qemu",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			r := &RawRenderer{
				width:  80,
				height: 24,
				out:    &buf,
				theme:  defaultTheme(),
			}

			snap := model.NewSnapshot()
			snap.Context = model.RuntimeContext{
				Hostname:         tt.hostname,
				Version:          tt.version,
				OS:               tt.os,
				Arch:             tt.arch,
				Mode:             tt.mode,
				ContainerRuntime: tt.runtime,
				StartTime:        time.Now(),
			}

			err := r.RenderCompact(snap)

			require.NoError(t, err)
			output := buf.String()
			require.NotEmpty(t, output)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, output, substr)
			}
		})
	}
}

// Test_renderHeaderCompact_EdgeCases tests edge cases for compact header rendering.
func Test_RawRenderer_renderHeaderCompact_EdgeCases(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		hostname   string
		version    string
		mode       model.RuntimeMode
		runtime    string
		wantSubstr []string
	}{
		{
			name:       "empty_hostname",
			width:      80,
			hostname:   "",
			version:    "1.0.0",
			mode:       model.ModeHost,
			runtime:    "",
			wantSubstr: []string{"superviz", "v1.0.0"},
		},
		{
			name:       "empty_version",
			width:      80,
			hostname:   "testhost",
			version:    "",
			mode:       model.ModeHost,
			runtime:    "",
			wantSubstr: []string{"superviz", "testhost"},
		},
		{
			name:       "very_long_hostname",
			width:      80,
			hostname:   "this-is-a-very-long-hostname-that-should-still-render",
			version:    "1.0.0",
			mode:       model.ModeHost,
			runtime:    "",
			wantSubstr: []string{"superviz", "this-is-a-very-long-hostname"},
		},
		{
			name:       "unknown_mode",
			width:      80,
			hostname:   "testhost",
			version:    "1.0.0",
			mode:       model.ModeUnknown,
			runtime:    "",
			wantSubstr: []string{"superviz", "testhost", "unknown"},
		},
		{
			name:       "vm_mode",
			width:      80,
			hostname:   "vmhost",
			version:    "2.0.0",
			mode:       model.ModeVM,
			runtime:    "qemu",
			wantSubstr: []string{"superviz", "vmhost", "qemu"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := mockSnapshot()
			snap.Context.Hostname = tt.hostname
			snap.Context.Version = tt.version
			snap.Context.Mode = tt.mode
			snap.Context.ContainerRuntime = tt.runtime

			r := &RawRenderer{
				width:  tt.width,
				height: 24,
				theme:  defaultTheme(),
			}

			result := r.renderHeaderCompact(snap)

			require.NotEmpty(t, result)
			for _, substr := range tt.wantSubstr {
				assert.Contains(t, result, substr, "expected substring not found")
			}
		})
	}
}

// Test_renderSystemAndSandboxesSideBySide_HeightEqualization tests height matching.
func Test_RawRenderer_renderSystemAndSandboxesSideBySide_HeightEqualization(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		sandboxes    []model.SandboxInfo
		limits       model.ResourceLimits
		wantBalanced bool
		minLineCount int
	}{
		{
			name: "many_sandboxes_few_limits",
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
				{Name: "podman", Detected: true, Endpoint: "unix:///var/run/podman.sock"},
				{Name: "lxc", Detected: false, Endpoint: ""},
				{Name: "systemd", Detected: false, Endpoint: ""},
			},
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  2.0,
			},
			wantBalanced: true,
			minLineCount: 4,
		},
		{
			name: "few_sandboxes_many_limits",
			sandboxes: []model.SandboxInfo{
				{Name: "docker", Detected: true, Endpoint: "unix:///var/run/docker.sock"},
			},
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  4.0,
				CPUSet:    "0-7",
				MemoryMax: 16 * 1024 * 1024 * 1024,
			},
			wantBalanced: true,
			minLineCount: 4,
		},
		{
			name:      "no_sandboxes",
			sandboxes: []model.SandboxInfo{},
			limits: model.ResourceLimits{
				HasLimits: true,
				CPUQuota:  2.0,
			},
			wantBalanced: true,
			minLineCount: 4,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			snap := mockSnapshot()
			snap.Sandboxes = tt.sandboxes
			snap.Limits = tt.limits

			r := &RawRenderer{
				width:  160,
				height: 40,
				theme:  defaultTheme(),
			}

			result := r.renderSystemAndSandboxesSideBySide(snap)

			require.NotEmpty(t, result)
			lines := strings.Split(result, "\n")
			assert.GreaterOrEqual(t, len(lines), tt.minLineCount, "expected minimum line count")
		})
	}
}
