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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	writer := daemon.NewConsoleWriterWithOptions(stdout, stderr, false)

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
			// Reset buffers.
			stdout.Reset()
			stderr.Reset()

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

	stdout := &bytes.Buffer{}
	stderr := &bytes.Buffer{}
	writer := daemon.NewConsoleWriterWithOptions(stdout, stderr, false)

	event := logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started").
		WithMeta("pid", 1234)
	err := writer.Write(event)
	require.NoError(t, err)

	output := stdout.String()
	assert.Contains(t, output, "[INFO]")
	assert.Contains(t, output, "nginx")
	assert.Contains(t, output, "started")
	assert.Contains(t, output, "pid=1234")
}

func TestConsoleWriter_Close(t *testing.T) {
	t.Parallel()

	writer := daemon.NewConsoleWriter()
	err := writer.Close()
	assert.NoError(t, err)
}
