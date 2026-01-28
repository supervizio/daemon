package logging_test

import (
	"path/filepath"
	"testing"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWriterConfig struct {
	file      string
	timestamp string
	rotation  config.RotationConfig
}

func (m *mockWriterConfig) File() string {
	return m.file
}

func (m *mockWriterConfig) TimestampFormat() string {
	return m.timestamp
}

func (m *mockWriterConfig) Rotation() config.RotationConfig {
	return m.rotation
}

func TestNewWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "create new writer",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := &mockWriterConfig{
				file:      "test.log",
				timestamp: "",
				rotation:  config.RotationConfig{},
			}

			writer, err := logging.NewWriter(path, cfg)
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

func TestNewWriterFromConfig(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "create writer from config interface",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := &mockWriterConfig{
				file:      "test.log",
				timestamp: "",
				rotation: config.RotationConfig{
					MaxSize:  "100MB",
					MaxFiles: 5,
				},
			}

			writer, err := logging.NewWriterFromConfig(path, cfg)
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

func TestWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		data    string
		wantErr bool
	}{
		{
			name:    "write to file",
			data:    "test log message\n",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := &mockWriterConfig{
				file:      "test.log",
				timestamp: "",
				rotation:  config.RotationConfig{},
			}

			writer, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer writer.Close()

			n, err := writer.Write([]byte(tt.data))
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, len(tt.data), n)
			}
		})
	}
}

func TestWriter_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "close writer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := &mockWriterConfig{
				file:      "test.log",
				timestamp: "",
				rotation:  config.RotationConfig{},
			}

			writer, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)

			err = writer.Close()
			assert.NoError(t, err)
		})
	}
}

func TestWriter_Sync(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "sync writer to disk"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := &mockWriterConfig{
				file:      "test.log",
				timestamp: "",
				rotation:  config.RotationConfig{},
			}

			writer, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer writer.Close()

			_, _ = writer.Write([]byte("test"))
			err = writer.Sync()
			assert.NoError(t, err)
		})
	}
}

func TestWriter_Path(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "get writer path"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := &mockWriterConfig{
				file:      "test.log",
				timestamp: "",
				rotation:  config.RotationConfig{},
			}

			writer, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer writer.Close()

			assert.Equal(t, path, writer.Path())
		})
	}
}

func TestWriter_Size(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		data string
	}{
		{
			name: "get writer size",
			data: "test data\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := &mockWriterConfig{
				file:      "test.log",
				timestamp: "",
				rotation:  config.RotationConfig{},
			}

			writer, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer writer.Close()

			initialSize := writer.Size()
			assert.Zero(t, initialSize)

			_, err = writer.Write([]byte(tt.data))
			require.NoError(t, err)

			newSize := writer.Size()
			assert.Greater(t, newSize, initialSize)
		})
	}
}
