package daemon_test

import (
	"bytes"
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConsoleWriter_SplitByLevel(t *testing.T) {
	t.Parallel()

	tests := []struct {
		level        logging.Level
		expectStdout bool
		expectStderr bool
	}{
		{logging.LevelDebug, true, false},
		{logging.LevelInfo, true, false},
		{logging.LevelWarn, false, true},
		{logging.LevelError, false, true},
	}

	for _, tt := range tests {
		t.Run(tt.level.String(), func(t *testing.T) {
			t.Parallel()
			// Create separate buffers for each subtest.
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			writer := daemon.NewConsoleWriterWithOptions(stdout, stderr, false)

			event := logging.NewLogEvent(tt.level, "test", "event", "message")
			err := writer.Write(event)
			require.NoError(t, err)

			if tt.expectStdout {
				assert.NotEmpty(t, stdout.String(), "expected output on stdout")
				assert.Empty(t, stderr.String(), "expected no output on stderr")
			}
			if tt.expectStderr {
				assert.Empty(t, stdout.String(), "expected no output on stdout")
				assert.NotEmpty(t, stderr.String(), "expected output on stderr")
			}
		})
	}
}

func TestConsoleWriter_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		level    logging.Level
		service  string
		event    string
		message  string
		metadata map[string]any
		contains []string
	}{
		{
			name:     "info with metadata",
			level:    logging.LevelInfo,
			service:  "nginx",
			event:    "started",
			message:  "Service started",
			metadata: map[string]any{"pid": 1234},
			contains: []string{"[INFO]", "nginx", "started", "pid=1234"},
		},
		{
			name:     "error without metadata",
			level:    logging.LevelError,
			service:  "postgres",
			event:    "failed",
			message:  "Database connection failed",
			metadata: nil,
			contains: []string{"[ERROR]", "postgres", "failed"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			writer := daemon.NewConsoleWriterWithOptions(stdout, stderr, false)

			event := logging.NewLogEvent(tt.level, tt.service, tt.event, tt.message).
				WithMetadata(tt.metadata)
			err := writer.Write(event)
			require.NoError(t, err)

			output := stdout.String() + stderr.String()
			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}

func TestConsoleWriter_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "close console writer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := daemon.NewConsoleWriter()
			err := writer.Close()
			assert.NoError(t, err)
		})
	}
}

func TestNewConsoleWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "create console writer with defaults"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := daemon.NewConsoleWriter()
			assert.NotNil(t, writer)
			_ = writer.Close()
		})
	}
}

func TestNewConsoleWriterWithOptions(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		color bool
	}{
		{
			name:  "with color",
			color: true,
		},
		{
			name:  "without color",
			color: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			writer := daemon.NewConsoleWriterWithOptions(stdout, stderr, tt.color)
			assert.NotNil(t, writer)
			_ = writer.Close()
		})
	}
}

func TestConsoleWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level logging.Level
	}{
		{
			name:  "write debug",
			level: logging.LevelDebug,
		},
		{
			name:  "write info",
			level: logging.LevelInfo,
		},
		{
			name:  "write warn",
			level: logging.LevelWarn,
		},
		{
			name:  "write error",
			level: logging.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			writer := daemon.NewConsoleWriterWithOptions(stdout, stderr, false)

			event := logging.NewLogEvent(tt.level, "test", "event", "message")
			err := writer.Write(event)
			assert.NoError(t, err)

			_ = writer.Close()
		})
	}
}
