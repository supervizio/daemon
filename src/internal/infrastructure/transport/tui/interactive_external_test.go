// Package tui_test provides external tests.
package tui_test

import (
	"testing"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/ansi"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/component"
	"github.com/stretchr/testify/assert"
)

func TestInteractiveCompiles(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "package compiles"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Compilation test
		})
	}
}

// Test_Model_Init verifies model initialization.
func Test_Model_Init(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		interval time.Duration
	}{
		{name: "default interval", interval: 1 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := createTestModel(tt.interval)

			cmd := m.Init()

			assert.NotNil(t, cmd, "Init should return command batch")
		})
	}
}

// Test_Model_Update verifies message handling.
func Test_Model_Update(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{"window size change", tea.WindowSizeMsg{Width: 100, Height: 30}},
		{"key message", tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune("x")}},
		{"mouse message", tea.MouseMsg{Type: tea.MouseMotion}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := createTestModel(1 * time.Second)

			result, _ := m.Update(tt.msg)

			assert.NotNil(t, result, "Update should return model")
		})
	}
}

// Test_Model_View verifies view rendering.
func Test_Model_View(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{"normal view"},
		{"view with snapshot"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			m := createTestModel(1 * time.Second)

			output := m.View()

			// NewTUI always initializes with a non-nil snapshot,
			// so View() should produce output (not "Loading...").
			assert.NotEmpty(t, output, "view should produce output")
		})
	}
}

// Test_NewModel verifies model creation.
func Test_NewModel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		width  int
		height int
	}{
		{name: "standard dimensions", width: 80, height: 24},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			theme := ansi.DefaultTheme()
			logsPanel := component.NewLogsPanel(tt.width, 10)
			servicesPanel := component.NewServicesPanel(tt.width, 6)
			cfg := tui.ModelConfig{
				TUI:           createTestTUI(1 * time.Second),
				Width:         tt.width,
				Height:        tt.height,
				Theme:         &theme,
				LogsPanel:     &logsPanel,
				ServicesPanel: &servicesPanel,
			}

			m := tui.NewModel(&cfg)

			assert.NotNil(t, m, "NewModel should return a model")
		})
	}
}

// Test_FocusPanel_constants verifies focus panel constants.
func Test_FocusPanel_constants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		got      tui.FocusPanel
		expected tui.FocusPanel
		desc     string
	}{
		{name: "FocusServices is 0", got: tui.FocusPanel(0), expected: tui.FocusServices, desc: "FocusServices should be 0"},
		{name: "FocusLogs is 1", got: tui.FocusPanel(1), expected: tui.FocusLogs, desc: "FocusLogs should be 1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.expected, tt.got, tt.desc)
		})
	}
}

// Helper functions for test setup.

// createTestTUI creates a minimal TUI instance for testing.
func createTestTUI(interval time.Duration) *tui.TUI {
	cfg := tui.Config{
		RefreshInterval: interval,
	}
	return tui.NewTUI(cfg)
}

// createTestModel creates a minimal test model without snapshot.
func createTestModel(interval time.Duration) tui.Model {
	theme := ansi.DefaultTheme()
	logsPanel := component.NewLogsPanel(80, 10)
	servicesPanel := component.NewServicesPanel(80, 6)
	return tui.NewModel(&tui.ModelConfig{
		TUI:           createTestTUI(interval),
		Width:         80,
		Height:        24,
		Theme:         &theme,
		LogsPanel:     &logsPanel,
		ServicesPanel: &servicesPanel,
	})
}
