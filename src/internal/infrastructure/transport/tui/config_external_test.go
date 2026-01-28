package tui_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/stretchr/testify/assert"
)

func TestDefaultConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		version string
	}{
		{"empty_version", ""},
		{"semver", "1.0.0"},
		{"dev_version", "dev-123abc"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			cfg := tui.DefaultConfig(tt.version)
			assert.Equal(t, tt.version, cfg.Version)
			assert.Equal(t, tui.ModeRaw, cfg.Mode)
			assert.NotNil(t, cfg.Output)
			assert.Positive(t, cfg.RefreshInterval)
		})
	}
}
