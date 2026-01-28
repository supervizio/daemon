package daemon_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFileWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setupDir bool
		wantErr  bool
	}{
		{
			name:     "create new file",
			setupDir: true,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			writer, err := daemon.NewFileWriter(path, config.RotationConfig{})
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, writer)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, writer)
				if writer != nil {
					_ = writer.Close()
				}
			}
		})
	}
}

func TestFileWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		level   logging.Level
		message string
	}{
		{
			name:    "write info event",
			level:   logging.LevelInfo,
			message: "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			writer, err := daemon.NewFileWriter(path, config.RotationConfig{})
			require.NoError(t, err)
			defer writer.Close()

			event := logging.NewLogEvent(tt.level, "test", "event", tt.message)
			err = writer.Write(event)
			assert.NoError(t, err)

			content, err := os.ReadFile(path)
			assert.NoError(t, err)
			assert.Contains(t, string(content), tt.message)
		})
	}
}

func TestFileWriter_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "close file writer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			writer, err := daemon.NewFileWriter(path, config.RotationConfig{})
			require.NoError(t, err)

			err = writer.Close()
			assert.NoError(t, err)
		})
	}
}
