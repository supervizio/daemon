// Package logging_test provides external tests for multiwriter.go.
// It tests the public API of the MultiWriter type using black-box testing.
package logging_test

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kodflow/daemon/internal/domain/service"
	"github.com/kodflow/daemon/internal/logging"
)

// readFileAsString reads a file and returns its content as string.
//
// Params:
//   - path: the file path to read.
//
// Returns:
//   - string: the file content as string.
//   - error: nil on success, error on failure.
func readFileAsString(path string) (string, error) {
	content, err := os.ReadFile(path)
	// Check if read succeeded.
	if err != nil {
		// Return empty string and error on failure.
		return "", err
	}
	// Convert bytes to string and return.
	return string(content), nil
}

// mockMultiWriterConfig implements writerConfig interface for testing.
type mockMultiWriterConfig struct {
	filePath        string
	timestampFormat string
	rotation        service.RotationConfig
}

// File returns the file path.
func (m *mockMultiWriterConfig) File() string {
	return m.filePath
}

// TimestampFormat returns the timestamp format.
func (m *mockMultiWriterConfig) TimestampFormat() string {
	return m.timestampFormat
}

// Rotation returns the rotation configuration.
func (m *mockMultiWriterConfig) Rotation() service.RotationConfig {
	return m.rotation
}

// createMultiWriterConfig creates a mockMultiWriterConfig for tests.
//
// Params:
//   - maxSize: the max size string.
//   - maxFiles: the max files count.
//
// Returns:
//   - *mockMultiWriterConfig: the configured mock config.
func createMultiWriterConfig(maxSize string, maxFiles int) *mockMultiWriterConfig {
	return &mockMultiWriterConfig{
		rotation: service.RotationConfig{
			MaxSize:  maxSize,
			MaxFiles: maxFiles,
		},
	}
}

// TestNewMultiWriter tests the NewMultiWriter constructor.
//
// Params:
//   - t: the testing context.
func TestNewMultiWriter(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// writerCount is the number of writers to create.
		writerCount int
	}{
		{
			name:        "single_writer",
			writerCount: 1,
		},
		{
			name:        "two_writers",
			writerCount: 2,
		},
		{
			name:        "multiple_writers",
			writerCount: 5,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := createMultiWriterConfig("1MB", 3)

			// Create the specified number of writers.
			writers := make([]*logging.Writer, 0, tt.writerCount)
			for i := range tt.writerCount {
				logFileName := "log" + strconv.Itoa(i) + ".log"
				path := filepath.Join(tmpDir, logFileName)
				w, err := logging.NewWriter(path, cfg)
				require.NoError(t, err)
				writers = append(writers, w)
			}

			// Create multiwriter using first writer.
			mw := logging.NewMultiWriter(writers[0])
			assert.NotNil(t, mw)
			defer func() { _ = mw.Close() }()
		})
	}
}

// TestMultiWriter_Write tests the Write method.
//
// Params:
//   - t: the testing context.
func TestMultiWriter_Write(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// message is the message to write.
		message string
	}{
		{
			name:    "simple_message",
			message: "test message\n",
		},
		{
			name:    "empty_message",
			message: "",
		},
		{
			name:    "multiline_message",
			message: "line1\nline2\nline3\n",
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
			cfg := createMultiWriterConfig("1MB", 3)

			path1 := filepath.Join(tmpDir, "log1.log")
			path2 := filepath.Join(tmpDir, "log2.log")

			w1, err := logging.NewWriter(path1, cfg)
			require.NoError(t, err)

			w2, err := logging.NewWriter(path2, cfg)
			require.NoError(t, err)

			mw := logging.NewMultiWriter(w1, w2)

			n, err := mw.Write(msgBytes)
			require.NoError(t, err)
			assert.Equal(t, msgLen, n)

			err = mw.Close()
			require.NoError(t, err)

			// Verify both files have the content.
			content1Str, err := readFileAsString(path1)
			require.NoError(t, err)
			content2Str, err := readFileAsString(path2)
			require.NoError(t, err)
			assert.Equal(t, tt.message, content1Str)
			assert.Equal(t, tt.message, content2Str)
		})
	}
}

// TestMultiWriter_Close tests the Close method.
//
// Params:
//   - t: the testing context.
func TestMultiWriter_Close(t *testing.T) {
	tests := []struct {
		// name is the test case name.
		name string
		// writerCount is the number of writers to create.
		writerCount int
	}{
		{
			name:        "close_single_writer",
			writerCount: 1,
		},
		{
			name:        "close_multiple_writers",
			writerCount: 3,
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		// Run each test case as a subtest.
		t.Run(tt.name, func(t *testing.T) {
			tmpDir := t.TempDir()
			cfg := createMultiWriterConfig("1MB", 3)

			// Create the specified number of writers.
			writers := make([]*logging.Writer, 0, tt.writerCount)
			for i := range tt.writerCount {
				logFileName := "log" + strconv.Itoa(i) + ".log"
				path := filepath.Join(tmpDir, logFileName)
				w, err := logging.NewWriter(path, cfg)
				require.NoError(t, err)
				writers = append(writers, w)
			}

			// Create multiwriter with first writer.
			mw := logging.NewMultiWriter(writers[0])

			err := mw.Close()
			assert.NoError(t, err)
		})
	}
}
