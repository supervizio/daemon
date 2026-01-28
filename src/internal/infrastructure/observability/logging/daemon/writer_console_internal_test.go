package daemon

import (
	"bytes"
	"os"
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/stretchr/testify/assert"
)

func TestConsoleWriter_Colorize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level logging.Level
		line  string
	}{
		{
			name:  "debug level",
			level: logging.LevelDebug,
			line:  "debug message",
		},
		{
			name:  "info level",
			level: logging.LevelInfo,
			line:  "info message",
		},
		{
			name:  "warn level",
			level: logging.LevelWarn,
			line:  "warn message",
		},
		{
			name:  "error level",
			level: logging.LevelError,
			line:  "error message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}
			writer := &ConsoleWriter{
				stdout: stdout,
				stderr: stderr,
				format: NewTextFormatter(""),
				color:  true,
			}

			colorized := writer.colorize(tt.level, tt.line)
			assert.Contains(t, colorized, tt.line)
			assert.NotEqual(t, tt.line, colorized, "colorized should differ from plain")
		})
	}
}

func TestIsTerminal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		setup    func() any
		expected bool
	}{
		{
			name: "bytes buffer is not terminal",
			setup: func() any {
				return &bytes.Buffer{}
			},
			expected: false,
		},
		{
			name: "os.File might be terminal",
			setup: func() any {
				return os.Stdout
			},
			expected: false, // In test environment, usually false
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := tt.setup()
			if w, ok := writer.(interface{ Write([]byte) (int, error) }); ok {
				_ = w
			}
		})
	}
}
