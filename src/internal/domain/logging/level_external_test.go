package logging_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevel_String(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level    logging.Level
		expected string
	}{
		{logging.LevelDebug, "DEBUG"},
		{logging.LevelInfo, "INFO"},
		{logging.LevelWarn, "WARN"},
		{logging.LevelError, "ERROR"},
		{logging.Level(99), "UNKNOWN"},
	}

	for _, tt := range tests {
		t.Run(tt.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.expected, tt.level.String())
		})
	}
}

func TestParseLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected logging.Level
		wantErr  bool
	}{
		{"debug lowercase", "debug", logging.LevelDebug, false},
		{"DEBUG uppercase", "DEBUG", logging.LevelDebug, false},
		{"info lowercase", "info", logging.LevelInfo, false},
		{"INFO uppercase", "INFO", logging.LevelInfo, false},
		{"warn lowercase", "warn", logging.LevelWarn, false},
		{"warning lowercase", "warning", logging.LevelWarn, false},
		{"WARN uppercase", "WARN", logging.LevelWarn, false},
		{"error lowercase", "error", logging.LevelError, false},
		{"ERROR uppercase", "ERROR", logging.LevelError, false},
		{"with spaces", "  info  ", logging.LevelInfo, false},
		{"invalid", "invalid", logging.LevelInfo, true},
		{"empty", "", logging.LevelInfo, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			level, err := logging.ParseLevel(tt.input)

			if tt.wantErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, logging.ErrInvalidLevel)
			} else {
				require.NoError(t, err)
			}
			assert.Equal(t, tt.expected, level)
		})
	}
}
