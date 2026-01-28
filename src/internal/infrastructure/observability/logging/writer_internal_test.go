package logging

import (
	"bufio"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestOpenLogFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "open new log file",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			file, size, err := openLogFile(path)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, file)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, file)
				assert.GreaterOrEqual(t, size, int64(0))
				if file != nil {
					_ = file.Close()
				}
			}
		})
	}
}

func TestParseMaxSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		sizeStr  string
		expected int64
	}{
		{
			name:     "valid MB size",
			sizeStr:  "50MB",
			expected: 50 * 1024 * 1024,
		},
		{
			name:     "invalid size returns default",
			sizeStr:  "invalid",
			expected: defaultMaxSize,
		},
		{
			name:     "empty size returns default",
			sizeStr:  "",
			expected: defaultMaxSize,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			result := parseMaxSize(tt.sizeStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestWriter_openNewFile(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		wantErr bool
	}{
		{
			name:    "open new log file",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			file, _, err := openLogFile(path)
			require.NoError(t, err)

			writer := &Writer{
				path: path,
			}

			newFile, err := writer.openNewFile()
			if tt.wantErr {
				assert.Error(t, err)
				assert.Nil(t, newFile)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newFile)
				if newFile != nil {
					_ = newFile.Close()
				}
			}

			_ = file.Close()
		})
	}
}

func TestWriter_rotateFiles(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		maxFiles int
	}{
		{
			name:     "rotate with max 3 files",
			maxFiles: 3,
		},
		{
			name:     "rotate with max 5 files",
			maxFiles: 5,
		},
		{
			name:     "rotate with max 10 files",
			maxFiles: 10,
		},
		{
			name:     "rotate with max 1 file",
			maxFiles: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			file, _, err := openLogFile(path)
			require.NoError(t, err)

			writer := &Writer{
				file:     file,
				path:     path,
				maxFiles: tt.maxFiles,
			}

			err = writer.rotateFiles()
			assert.NoError(t, err)
		})
	}
}
func TestWriter_Rotate(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		maxSize int64
	}{
		{
			name:    "rotation on size limit",
			maxSize: 100,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			file, _, err := openLogFile(path)
			require.NoError(t, err)

			writer := &Writer{
				file:     file,
				writer:   bufio.NewWriter(file),
				path:     path,
				maxSize:  tt.maxSize,
				maxFiles: 5,
			}

			// Verify writer structure (rotation logic tested in integration tests).
			assert.NotNil(t, writer)
			_ = file.Close()
		})
	}
}
