// Package logging_test provides external tests for writer.go.
// It tests the public API of the Writer type using black-box testing.
package logging_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/logging"
)

// readFileAsStringForWriter reads a file and returns its content as string.
//
// Params:
//   - path: the file path to read.
//
// Returns:
//   - string: the file content as string.
//   - error: nil on success, error on failure.
func readFileAsStringForWriter(path string) (string, error) {
	content, err := os.ReadFile(path)
	// Check if read succeeded.
	if err != nil {
		// Return empty string and error on failure.
		return "", err
	}
	// Convert bytes to string and return.
	return string(content), nil
}

// mockWriterConfig implements writerConfig interface for testing.
type mockWriterConfig struct {
	filePath        string
	timestampFormat string
	rotation        service.RotationConfig
}

// File returns the file path.
func (m *mockWriterConfig) File() string {
	return m.filePath
}

// TimestampFormat returns the timestamp format.
func (m *mockWriterConfig) TimestampFormat() string {
	return m.timestampFormat
}

// Rotation returns the rotation configuration.
func (m *mockWriterConfig) Rotation() service.RotationConfig {
	return m.rotation
}

// createWriterConfig creates a mockWriterConfig for writer tests.
//
// Params:
//   - file: the log file name.
//   - maxSize: the max size string.
//   - maxFiles: the max files count.
//   - timestampFormat: the timestamp format.
//
// Returns:
//   - *mockWriterConfig: the configured mock config.
func createWriterConfig(file, maxSize string, maxFiles int, timestampFormat string) *mockWriterConfig {
	return &mockWriterConfig{
		filePath:        file,
		timestampFormat: timestampFormat,
		rotation: service.RotationConfig{
			MaxSize:  maxSize,
			MaxFiles: maxFiles,
		},
	}
}

// TestNewWriter tests the NewWriter constructor.
//
// Params:
//   - t: the testing context.
func TestNewWriter(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// maxSize is the max size configuration.
		maxSize string
		// maxFiles is the max files configuration.
		maxFiles int
	}{
		{
			name:     "default_configuration",
			maxSize:  "1MB",
			maxFiles: 3,
		},
		{
			name:     "large_max_size",
			maxSize:  "100MB",
			maxFiles: 10,
		},
		{
			name:     "small_max_size",
			maxSize:  "100",
			maxFiles: 1,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := createWriterConfig("test.log", tt.maxSize, tt.maxFiles, "")

			w, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer func() { _ = w.Close() }()

			assert.Equal(t, path, w.Path())
			assert.Equal(t, int64(0), w.Size())
		})
	}
}

// TestWriter_Write tests the Write method of Writer.
//
// Params:
//   - t: the testing context.
func TestWriter_Write(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// message is the message to write.
		message string
	}{
		{
			name:    "simple_message",
			message: "Hello, World!\n",
		},
		{
			name:    "empty_message",
			message: "",
		},
		{
			name:    "multiline_message",
			message: "line1\nline2\nline3\n",
		},
		{
			name:    "long_message",
			message: "This is a longer test message that spans multiple words and should be written correctly.\n",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Convert message to bytes once for consistent usage.
			msgBytes := []byte(tt.message)
			msgLen := len(msgBytes)

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := createWriterConfig("", "1MB", 3, "")

			w, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer func() { _ = w.Close() }()

			n, err := w.Write(msgBytes)
			require.NoError(t, err)
			assert.Equal(t, msgLen, n)

			// Verify content.
			contentStr, err := readFileAsStringForWriter(path)
			require.NoError(t, err)
			assert.Equal(t, tt.message, contentStr)
		})
	}
}

// TestWriterWithTimestamp tests writing with timestamp formatting.
//
// Params:
//   - t: the testing context.
func TestWriterWithTimestamp(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// timestampFormat is the timestamp format to use.
		timestampFormat string
		// message is the message to write.
		message string
	}{
		{
			name:            "iso8601_format",
			timestampFormat: logging.FormatISO8601,
			message:         "test message\n",
		},
		{
			name:            "rfc3339_format",
			timestampFormat: logging.FormatRFC3339,
			message:         "test message\n",
		},
		{
			name:            "custom_format",
			timestampFormat: "2006-01-02",
			message:         "test message\n",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			// Convert message to bytes once for consistent usage.
			msgBytes := []byte(tt.message)
			msgLen := len(msgBytes)

			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := createWriterConfig("", "1MB", 3, tt.timestampFormat)

			w, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer func() { _ = w.Close() }()

			_, err = w.Write(msgBytes)
			require.NoError(t, err)

			// Should contain timestamp and message.
			contentStr, err := readFileAsStringForWriter(path)
			require.NoError(t, err)
			assert.Contains(t, contentStr, "test message")
			assert.True(t, len(contentStr) > msgLen) // Timestamp adds length.
		})
	}
}

// TestWriterRotation tests the log rotation functionality.
//
// Params:
//   - t: the testing context.
func TestWriterRotation(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// maxSize is the max file size.
		maxSize string
		// maxFiles is the max number of files.
		maxFiles int
		// writeCount is the number of writes to perform.
		writeCount int
	}{
		{
			name:       "rotation_triggered",
			maxSize:    "100",
			maxFiles:   3,
			writeCount: 10,
		},
		{
			name:       "single_backup",
			maxSize:    "100",
			maxFiles:   1,
			writeCount: 5,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := createWriterConfig("", tt.maxSize, tt.maxFiles, "")

			w, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer func() { _ = w.Close() }()

			// Write enough to trigger rotation.
			for i := range tt.writeCount {
				_ = i
				_, err := w.Write([]byte("This is a test line that will trigger rotation\n"))
				require.NoError(t, err)
			}

			// Check that rotated files exist.
			files, err := filepath.Glob(path + "*")
			require.NoError(t, err)
			assert.True(t, len(files) >= 1)
		})
	}
}

// TestWriter_Close tests the Close method.
//
// Params:
//   - t: the testing context.
func TestWriter_Close(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// writeBeforeClose indicates whether to write before closing.
		writeBeforeClose bool
	}{
		{
			name:             "close_empty_file",
			writeBeforeClose: false,
		},
		{
			name:             "close_after_write",
			writeBeforeClose: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := createWriterConfig("", "1MB", 3, "")

			w, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)

			// Write if requested.
			if tt.writeBeforeClose {
				_, err := w.Write([]byte("test message\n"))
				require.NoError(t, err)
			}

			err = w.Close()
			assert.NoError(t, err)
		})
	}
}

// TestWriter_Sync tests the Sync method.
//
// Params:
//   - t: the testing context.
func TestWriter_Sync(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// writeBeforeSync indicates whether to write before syncing.
		writeBeforeSync bool
	}{
		{
			name:            "sync_empty_file",
			writeBeforeSync: false,
		},
		{
			name:            "sync_after_write",
			writeBeforeSync: true,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := createWriterConfig("", "1MB", 3, "")

			w, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer func() { _ = w.Close() }()

			// Write if requested.
			if tt.writeBeforeSync {
				_, err := w.Write([]byte("test message\n"))
				require.NoError(t, err)
			}

			err = w.Sync()
			assert.NoError(t, err)
		})
	}
}

// TestWriter_Path tests the Path method.
//
// Params:
//   - t: the testing context.
func TestWriter_Path(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// filename is the log file name.
		filename string
	}{
		{
			name:     "simple_path",
			filename: "test.log",
		},
		{
			name:     "nested_path",
			filename: "subdir/test.log",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, tt.filename)

			cfg := createWriterConfig("", "1MB", 3, "")

			w, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer func() { _ = w.Close() }()

			assert.Equal(t, path, w.Path())
		})
	}
}

// TestWriter_Size tests the Size method.
//
// Params:
//   - t: the testing context.
func TestWriter_Size(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// message is the message to write.
		message string
		// expectedSize is the expected size after write.
		expectedSize int64
	}{
		{
			name:         "empty_file",
			message:      "",
			expectedSize: 0,
		},
		{
			name:         "after_write",
			message:      "test",
			expectedSize: 4,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := createWriterConfig("", "1MB", 3, "")

			w, err := logging.NewWriter(path, cfg)
			require.NoError(t, err)
			defer func() { _ = w.Close() }()

			// Write if message provided.
			if tt.message != "" {
				_, err := w.Write([]byte(tt.message))
				require.NoError(t, err)
			}

			assert.Equal(t, tt.expectedSize, w.Size())
		})
	}
}

// TestNewWriterFromConfig tests the NewWriterFromConfig constructor.
//
// Params:
//   - t: the testing context.
func TestNewWriterFromConfig(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// maxSize is the max size configuration.
		maxSize string
		// maxFiles is the max files configuration.
		maxFiles int
		// timestampFormat is the timestamp format.
		timestampFormat string
	}{
		{
			name:            "default_configuration",
			maxSize:         "1MB",
			maxFiles:        3,
			timestampFormat: "",
		},
		{
			name:            "with_timestamp",
			maxSize:         "1MB",
			maxFiles:        3,
			timestampFormat: logging.FormatISO8601,
		},
		{
			name:            "zero_max_files_uses_default",
			maxSize:         "1MB",
			maxFiles:        0,
			timestampFormat: "",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			path := filepath.Join(tmpDir, "test.log")

			cfg := createWriterConfig("test.log", tt.maxSize, tt.maxFiles, tt.timestampFormat)

			w, err := logging.NewWriterFromConfig(path, cfg)
			require.NoError(t, err)
			defer func() { _ = w.Close() }()

			assert.Equal(t, path, w.Path())
		})
	}
}

// TestNewWriterFromConfig_DirectoryError tests NewWriterFromConfig when directory creation fails.
//
// Params:
//   - t: the testing context.
func TestNewWriterFromConfig_DirectoryError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// path is the invalid path that will cause an error.
		path string
	}{
		{
			name: "directory_creation_fails",
			path: "/nonexistent/readonly/path/test.log",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createWriterConfig("test.log", "1MB", 3, "")

			w, err := logging.NewWriterFromConfig(tt.path, cfg)
			assert.Error(t, err)
			assert.Nil(t, w)
		})
	}
}

// TestNewWriter_DirectoryError tests NewWriter when directory creation fails.
//
// Params:
//   - t: the testing context.
func TestNewWriter_DirectoryError(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// path is the invalid path that will cause an error.
		path string
	}{
		{
			name: "directory_creation_fails",
			path: "/nonexistent/readonly/path/test.log",
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			cfg := createWriterConfig("test.log", "1MB", 3, "")

			w, err := logging.NewWriter(tt.path, cfg)
			assert.Error(t, err)
			assert.Nil(t, w)
		})
	}
}
