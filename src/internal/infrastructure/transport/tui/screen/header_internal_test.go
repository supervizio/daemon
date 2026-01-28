package screen

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

func TestHeaderRenderer_ThemeDefaults(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		width int
	}{
		{
			name:  "default_theme",
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &HeaderRenderer{
				theme: ansi.DefaultTheme(),
				width: tt.width,
			}
			defaultTheme := ansi.DefaultTheme()
			assert.Equal(t, defaultTheme.Primary, renderer.theme.Primary)
			assert.Equal(t, defaultTheme.Accent, renderer.theme.Accent)
		})
	}
}

func TestHeaderRenderer_buildCompactLogoLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "with_version",
			version: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &HeaderRenderer{
				theme: ansi.DefaultTheme(),
				width: 60,
			}
			result := renderer.buildCompactLogoLine(tt.version)
			assert.NotEmpty(t, result)
		})
	}
}

func TestHeaderRenderer_buildCompactInfoLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  model.RuntimeContext
	}{
		{
			name: "basic_context",
			ctx: model.RuntimeContext{
				Hostname: "testhost",
				OS:       "linux",
				Arch:     "amd64",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &HeaderRenderer{
				theme: ansi.DefaultTheme(),
				width: 60,
			}
			result := renderer.buildCompactInfoLine(tt.ctx)
			assert.NotEmpty(t, result)
		})
	}
}

func TestHeaderRenderer_buildNormalTitleLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
	}{
		{
			name:    "with_version",
			version: "1.0.0",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &HeaderRenderer{
				theme: ansi.DefaultTheme(),
				width: 100,
			}
			result := renderer.buildNormalTitleLine(tt.version)
			assert.NotEmpty(t, result)
		})
	}
}

func TestHeaderRenderer_buildNormalContentLines(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		ctx  model.RuntimeContext
	}{
		{
			name: "basic_context",
			ctx: model.RuntimeContext{
				Hostname: "testhost",
				OS:       "linux",
				Arch:     "amd64",
				Kernel:   "5.15.0",
				Mode:     model.ModeHost,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &HeaderRenderer{
				theme: ansi.DefaultTheme(),
				width: 100,
			}
			result := renderer.buildNormalContentLines(tt.ctx)
			assert.NotEmpty(t, result)
		})
	}
}

func TestHeaderRenderer_renderCompact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		snap  *model.Snapshot
		width int
	}{
		{
			name: "basic_snapshot",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:  "1.0.0",
					Hostname: "testhost",
					OS:       "linux",
					Arch:     "amd64",
					Mode:     model.ModeHost,
				},
			},
			width: 60,
		},
		{
			name: "with_container_runtime",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:          "2.0.0",
					Hostname:         "container-host",
					OS:               "linux",
					Arch:             "arm64",
					Mode:             model.ModeContainer,
					ContainerRuntime: "docker",
				},
			},
			width: 80,
		},
		{
			name: "version_without_v_prefix",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:  "3.1.0",
					Hostname: "myhost",
					OS:       "linux",
					Arch:     "amd64",
					Mode:     model.ModeHost,
				},
			},
			width: 70,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &HeaderRenderer{
				theme: ansi.DefaultTheme(),
				width: tt.width,
			}
			result := renderer.renderCompact(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "superviz")
		})
	}
}

func TestHeaderRenderer_renderNormal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		snap  *model.Snapshot
		width int
	}{
		{
			name: "basic_snapshot",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:  "1.0.0",
					Hostname: "testhost",
					OS:       "linux",
					Arch:     "amd64",
					Mode:     model.ModeHost,
				},
			},
			width: 100,
		},
		{
			name: "with_config_path",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:    "1.0.0",
					Hostname:   "testhost",
					OS:         "linux",
					Arch:       "amd64",
					Mode:       model.ModeHost,
					ConfigPath: "/custom/config.yaml",
				},
			},
			width: 120,
		},
		{
			name: "with_container_runtime",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:          "2.0.0",
					Hostname:         "container-host",
					OS:               "linux",
					Arch:             "arm64",
					Mode:             model.ModeContainer,
					ContainerRuntime: "podman",
				},
			},
			width: 100,
		},
		{
			name: "narrow_width",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:  "1.0.0",
					Hostname: "testhost",
					OS:       "linux",
					Arch:     "amd64",
					Mode:     model.ModeHost,
				},
			},
			width: 80,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &HeaderRenderer{
				theme: ansi.DefaultTheme(),
				width: tt.width,
			}
			result := renderer.renderNormal(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "superviz")
			assert.Contains(t, result, tt.snap.Context.Hostname)
		})
	}
}

func TestHeaderRenderer_renderWide(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		snap  *model.Snapshot
		width int
	}{
		{
			name: "basic_snapshot",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:  "1.0.0",
					Hostname: "testhost",
					OS:       "linux",
					Arch:     "amd64",
					Mode:     model.ModeHost,
				},
			},
			width: 160,
		},
		{
			name: "ultra_wide",
			snap: &model.Snapshot{
				Context: model.RuntimeContext{
					Version:  "2.0.0",
					Hostname: "widehost",
					OS:       "linux",
					Arch:     "amd64",
					Mode:     model.ModeHost,
				},
			},
			width: 200,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			renderer := &HeaderRenderer{
				theme: ansi.DefaultTheme(),
				width: tt.width,
			}
			result := renderer.renderWide(tt.snap)
			assert.NotEmpty(t, result)
			assert.Contains(t, result, "superviz")
		})
	}
}
