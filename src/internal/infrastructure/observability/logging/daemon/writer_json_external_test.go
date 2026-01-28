package daemon_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewJSONWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "create new json writer",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.json")

			writer, err := daemon.NewJSONWriter(path)
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

func TestJSONWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    logging.Level
		service  string
		message  string
		metadata map[string]any
	}{
		{
			name:     "write json event with metadata",
			level:    logging.LevelInfo,
			service:  "nginx",
			message:  "service started",
			metadata: map[string]any{"pid": 1234},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.json")

			writer, err := daemon.NewJSONWriter(path)
			require.NoError(t, err)
			defer writer.Close()

			event := logging.NewLogEvent(tt.level, tt.service, "event", tt.message).
				WithMetadata(tt.metadata)
			err = writer.Write(event)
			assert.NoError(t, err)

			content, err := os.ReadFile(path)
			require.NoError(t, err)

			var parsed map[string]any
			err = json.Unmarshal(content, &parsed)
			assert.NoError(t, err)
			assert.Equal(t, tt.level.String(), parsed["level"])
			assert.Equal(t, tt.service, parsed["service"])
		})
	}
}

func TestJSONWriter_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "close json writer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.json")

			writer, err := daemon.NewJSONWriter(path)
			require.NoError(t, err)

			err = writer.Close()
			assert.NoError(t, err)
		})
	}
}
