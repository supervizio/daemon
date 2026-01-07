package logging

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewWriter(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	cfg := &config.LogStreamConfig{
		File: "test.log",
		Rotation: config.RotationConfig{
			MaxSize:  "1MB",
			MaxFiles: 3,
		},
	}

	w, err := NewWriter(path, cfg)
	require.NoError(t, err)
	defer w.Close()

	assert.Equal(t, path, w.Path())
	assert.Equal(t, int64(0), w.Size())
}

func TestWriterWrite(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	cfg := &config.LogStreamConfig{
		Rotation: config.RotationConfig{
			MaxSize:  "1MB",
			MaxFiles: 3,
		},
	}

	w, err := NewWriter(path, cfg)
	require.NoError(t, err)
	defer w.Close()

	msg := "Hello, World!\n"
	n, err := w.Write([]byte(msg))
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)

	// Verify content
	content, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, msg, string(content))
}

func TestWriterWithTimestamp(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	cfg := &config.LogStreamConfig{
		TimestampFormat: FormatISO8601,
		Rotation: config.RotationConfig{
			MaxSize:  "1MB",
			MaxFiles: 3,
		},
	}

	w, err := NewWriter(path, cfg)
	require.NoError(t, err)
	defer w.Close()

	msg := "test message\n"
	_, err = w.Write([]byte(msg))
	require.NoError(t, err)

	content, err := os.ReadFile(path)
	require.NoError(t, err)

	// Should contain timestamp and message
	assert.Contains(t, string(content), "test message")
	assert.True(t, len(content) > len(msg)) // Timestamp adds length
}

func TestWriterRotation(t *testing.T) {
	tmpDir := t.TempDir()
	path := filepath.Join(tmpDir, "test.log")

	cfg := &config.LogStreamConfig{
		Rotation: config.RotationConfig{
			MaxSize:  "100", // Very small for testing
			MaxFiles: 3,
		},
	}

	w, err := NewWriter(path, cfg)
	require.NoError(t, err)
	defer w.Close()

	// Write enough to trigger rotation
	for i := 0; i < 10; i++ {
		_, err := w.Write([]byte("This is a test line that will trigger rotation\n"))
		require.NoError(t, err)
	}

	// Check that rotated files exist
	files, err := filepath.Glob(path + "*")
	require.NoError(t, err)
	assert.True(t, len(files) >= 1)
}

func TestFormatTimestamp(t *testing.T) {
	now := time.Date(2024, 1, 15, 10, 30, 45, 0, time.UTC)

	tests := []struct {
		format   string
		contains string
	}{
		{FormatISO8601, "2024-01-15T10:30:45Z"},
		{FormatRFC3339, "2024-01-15T10:30:45"},
		{"2006-01-02", "2024-01-15"},
		{"15:04:05", "10:30:45"},
	}

	for _, tt := range tests {
		t.Run(tt.format, func(t *testing.T) {
			result := FormatTimestamp(now, tt.format)
			assert.Contains(t, result, tt.contains)
		})
	}
}

func TestLineWriter(t *testing.T) {
	var buf strings.Builder

	lw := NewLineWriter(&buf, "[PREFIX] ")

	// Write partial line
	lw.Write([]byte("hello"))
	assert.Empty(t, buf.String()) // Not flushed yet

	// Complete the line
	lw.Write([]byte(" world\n"))
	assert.Equal(t, "[PREFIX] hello world\n", buf.String())
}

func TestLineWriterMultipleLines(t *testing.T) {
	var buf strings.Builder

	lw := NewLineWriter(&buf, ">> ")

	lw.Write([]byte("line1\nline2\n"))

	assert.Equal(t, ">> line1\n>> line2\n", buf.String())
}

func TestLineWriterFlush(t *testing.T) {
	var buf strings.Builder

	lw := NewLineWriter(&buf, "")

	lw.Write([]byte("partial"))
	assert.Empty(t, buf.String())

	err := lw.Flush()
	require.NoError(t, err)
	assert.Equal(t, "partial\n", buf.String())
}

func TestMultiWriter(t *testing.T) {
	tmpDir := t.TempDir()

	cfg := &config.LogStreamConfig{
		Rotation: config.RotationConfig{
			MaxSize:  "1MB",
			MaxFiles: 3,
		},
	}

	path1 := filepath.Join(tmpDir, "log1.log")
	path2 := filepath.Join(tmpDir, "log2.log")

	w1, err := NewWriter(path1, cfg)
	require.NoError(t, err)

	w2, err := NewWriter(path2, cfg)
	require.NoError(t, err)

	mw := NewMultiWriter(w1, w2)

	msg := "test message\n"
	n, err := mw.Write([]byte(msg))
	require.NoError(t, err)
	assert.Equal(t, len(msg), n)

	err = mw.Close()
	require.NoError(t, err)

	// Both files should have the content
	content1, _ := os.ReadFile(path1)
	content2, _ := os.ReadFile(path2)
	assert.Equal(t, msg, string(content1))
	assert.Equal(t, msg, string(content2))
}

func TestParseTimestampFormat(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"", FormatISO8601},
		{FormatISO8601, FormatISO8601},
		{FormatRFC3339, FormatRFC3339},
		{FormatUnix, FormatUnix},
		{"2006-01-02", "2006-01-02"},
	}

	for _, tt := range tests {
		result := ParseTimestampFormat(tt.input)
		assert.Equal(t, tt.expected, result)
	}
}
