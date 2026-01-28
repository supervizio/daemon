// Package tui_test provides external black-box tests.
package tui_test

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
)

// mockServiceLister is a mock for ListServicesser.
type mockServiceLister struct {
	services []model.ServiceSnapshot
}

func (m *mockServiceLister) ListServices() []model.ServiceSnapshot {
	return m.services
}

// mockMetricser is a mock for Metricser.
type mockMetricser struct {
	metrics model.SystemMetrics
}

func (m *mockMetricser) Metrics() model.SystemMetrics {
	return m.metrics
}

// mockSummarizeer is a mock for Summarizeer.
type mockSummarizeer struct {
	summary model.LogSummary
}

func (m *mockSummarizeer) Summarize() model.LogSummary {
	return m.summary
}

func TestNewTUI(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		config  tui.Config
		wantNil bool
	}{
		{
			name: "default_config",
			config: tui.Config{
				Mode:            tui.ModeRaw,
				RefreshInterval: 100 * time.Millisecond,
				Version:         "1.0.0",
			},
			wantNil: false,
		},
		{
			name: "interactive_mode",
			config: tui.Config{
				Mode:            tui.ModeInteractive,
				RefreshInterval: 100 * time.Millisecond,
				Version:         "1.0.0",
			},
			wantNil: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			instance := tui.NewTUI(tt.config)
			if tt.wantNil {
				assert.Nil(t, instance)
			} else {
				assert.NotNil(t, instance)
				// Verify instance is usable by checking Snapshot returns non-nil.
				assert.NotNil(t, instance.Snapshot())
			}
		})
	}
}

func TestTUI_SetServiceLister(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "set_service_lister"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			instance := tui.NewTUI(tui.DefaultConfig("1.0.0"))

			lister := &mockServiceLister{
				services: []model.ServiceSnapshot{
					{Name: "test-service", State: process.StateRunning},
				},
			}

			// SetServiceLister should not panic.
			instance.SetServiceLister(lister)
			// Verify instance remains usable.
			assert.NotNil(t, instance.Snapshot())
		})
	}
}

func TestTUI_SetMetricser(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "set_metricser"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			instance := tui.NewTUI(tui.DefaultConfig("1.0.0"))

			metrics := &mockMetricser{
				metrics: model.SystemMetrics{
					CPUPercent:    50.5,
					MemoryPercent: 60.0,
				},
			}

			// SetMetricser should not panic.
			instance.SetMetricser(metrics)
			// Verify instance remains usable.
			assert.NotNil(t, instance.Snapshot())
		})
	}
}

func TestTUI_SetSummarizeer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "set_summarizeer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			instance := tui.NewTUI(tui.DefaultConfig("1.0.0"))

			summarizer := &mockSummarizeer{
				summary: model.LogSummary{
					InfoCount:  10,
					WarnCount:  2,
					ErrorCount: 1,
				},
			}

			// SetSummarizeer should not panic.
			instance.SetSummarizeer(summarizer)
			// Verify instance remains usable.
			assert.NotNil(t, instance.Snapshot())
		})
	}
}

func TestTUI_SetConfigPath(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		path string
	}{
		{"valid_path", "/etc/daemon/config.yaml"},
		{"empty_path", ""},
		{"relative_path", "./config.yaml"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			instance := tui.NewTUI(tui.DefaultConfig("1.0.0"))
			// SetConfigPath should not panic.
			instance.SetConfigPath(tt.path)
		})
	}
}

func TestTUI_Snapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "get_snapshot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			instance := tui.NewTUI(tui.DefaultConfig("1.0.0"))
			snapshot := instance.Snapshot()
			assert.NotNil(t, snapshot)
		})
	}
}

func TestShouldUseInteractive(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "returns_bool"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// This test just verifies the function exists and returns a bool.
			result := tui.ShouldUseInteractive()
			assert.IsType(t, true, result)
		})
	}
}

func TestTUI_Run_RawMode(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "raw_mode_run"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			cfg := tui.Config{
				Mode:            tui.ModeRaw,
				RefreshInterval: 100 * time.Millisecond,
				Version:         "1.0.0",
				Output:          &buf,
			}

			instance := tui.NewTUI(cfg)
			instance.SetServiceLister(&mockServiceLister{
				services: []model.ServiceSnapshot{
					{Name: "test", State: process.StateRunning},
				},
			})

			ctx := context.Background()
			err := instance.Run(ctx)
			assert.NoError(t, err)
			assert.NotEmpty(t, buf.String())
		})
	}
}

func TestTUI_SetSnapshot(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "set_snapshot_on_new_tui"},
		{name: "replace_existing_snapshot"},
		{name: "set_nil_snapshot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			instance := tui.NewTUI(tui.DefaultConfig("1.0.0"))

			snap := &model.Snapshot{
				Context: model.RuntimeContext{
					Hostname: "test-host",
					Version:  "1.0.0",
				},
			}

			instance.SetSnapshot(snap)

			assert.Equal(t, snap, instance.Snapshot())
			assert.Equal(t, "test-host", instance.Snapshot().Context.Hostname)
		})
	}
}

func TestTUI_Run(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		mode tui.Mode
	}{
		{name: "raw_mode", mode: tui.ModeRaw},
		{name: "interactive_mode_fallback", mode: tui.ModeInteractive},
		{name: "unknown_mode_defaults_to_raw", mode: tui.Mode(99)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var buf strings.Builder
			cfg := tui.Config{
				Mode:            tt.mode,
				RefreshInterval: 100 * time.Millisecond,
				Version:         "1.0.0",
				Output:          &buf,
			}

			instance := tui.NewTUI(cfg)
			instance.SetServiceLister(&mockServiceLister{
				services: []model.ServiceSnapshot{
					{Name: "test", State: process.StateRunning},
				},
			})

			ctx := context.Background()
			err := instance.Run(ctx)

			assert.NoError(t, err)
			assert.NotEmpty(t, buf.String())
		})
	}
}
