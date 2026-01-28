package daemon_test

import (
	"bytes"
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewBufferedWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "create buffered writer"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cw := daemon.NewConsoleWriter()
			bw := daemon.NewBufferedWriter(cw)
			assert.NotNil(t, bw)
			_ = bw.Close()
		})
	}
}

func TestBufferedWriter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		events []logging.LogEvent
	}{
		{
			name: "write single event",
			events: []logging.LogEvent{
				logging.NewLogEvent(logging.LevelInfo, "test", "event", "Test message"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			cw := daemon.NewConsoleWriterWithOptions(buf, buf, false)
			bw := daemon.NewBufferedWriter(cw)

			for _, event := range tt.events {
				err := bw.Write(event)
				assert.NoError(t, err)
			}

			_ = bw.Close()
		})
	}
}

func TestBufferedWriter_Flush(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "flush buffered events"},
		{name: "flush when already flushed"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			cw := daemon.NewConsoleWriterWithOptions(buf, buf, false)
			bw := daemon.NewBufferedWriter(cw)

			event := logging.NewLogEvent(logging.LevelInfo, "test", "event", "Test message")
			err := bw.Write(event)
			require.NoError(t, err)

			err = bw.Flush()
			assert.NoError(t, err)

			// Second flush should be idempotent
			if tt.name == "flush when already flushed" {
				err = bw.Flush()
				assert.NoError(t, err)
			}

			_ = bw.Close()
		})
	}
}

func TestBufferedWriter_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "close buffered writer"},
		{name: "close twice"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			buf := &bytes.Buffer{}
			cw := daemon.NewConsoleWriterWithOptions(buf, buf, false)
			bw := daemon.NewBufferedWriter(cw)

			err := bw.Close()
			assert.NoError(t, err)

			// Second close should be idempotent
			if tt.name == "close twice" {
				err = bw.Close()
				assert.NoError(t, err)
			}
		})
	}
}
