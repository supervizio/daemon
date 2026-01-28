package daemon_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testWriter struct {
	events []logging.LogEvent
}

func (w *testWriter) Write(event logging.LogEvent) error {
	w.events = append(w.events, event)
	return nil
}

func (w *testWriter) Close() error {
	return nil
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		writerCount int
	}{
		{
			name:        "no writers",
			writerCount: 0,
		},
		{
			name:        "single writer",
			writerCount: 1,
		},
		{
			name:        "multiple writers",
			writerCount: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			var writers []logging.Writer
			for range tt.writerCount {
				writers = append(writers, &testWriter{})
			}

			logger := daemon.New(writers...)
			assert.NotNil(t, logger)
			_ = logger.Close()
		})
	}
}

func TestNewMultiLogger(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "create with alias constructor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := daemon.NewMultiLogger(&testWriter{})
			assert.NotNil(t, logger)
			_ = logger.Close()
		})
	}
}

func TestMultiLogger_Log(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		level logging.Level
	}{
		{
			name:  "log info event",
			level: logging.LevelInfo,
		},
		{
			name:  "log error event",
			level: logging.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &testWriter{}
			logger := daemon.New(writer)

			event := logging.NewLogEvent(tt.level, "test", "event", "message")
			logger.Log(event)

			require.Len(t, writer.events, 1)
			assert.Equal(t, tt.level, writer.events[0].Level)
		})
	}
}

func TestMultiLogger_Debug(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		service  string
		event    string
		message  string
		metadata map[string]any
	}{
		{
			name:     "debug with metadata",
			service:  "nginx",
			event:    "started",
			message:  "Service started",
			metadata: map[string]any{"pid": 1234},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &testWriter{}
			logger := daemon.New(writer)

			logger.Debug(tt.service, tt.event, tt.message, tt.metadata)

			require.Len(t, writer.events, 1)
			assert.Equal(t, logging.LevelDebug, writer.events[0].Level)
			assert.Equal(t, tt.service, writer.events[0].Service)
		})
	}
}

func TestMultiLogger_Info(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		service  string
		event    string
		message  string
		metadata map[string]any
	}{
		{
			name:     "info without metadata",
			service:  "postgres",
			event:    "connected",
			message:  "Database connected",
			metadata: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &testWriter{}
			logger := daemon.New(writer)

			logger.Info(tt.service, tt.event, tt.message, tt.metadata)

			require.Len(t, writer.events, 1)
			assert.Equal(t, logging.LevelInfo, writer.events[0].Level)
		})
	}
}

func TestMultiLogger_Warn(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service string
	}{
		{
			name:    "warn event",
			service: "redis",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &testWriter{}
			logger := daemon.New(writer)

			logger.Warn(tt.service, "slow", "Slow response", nil)

			require.Len(t, writer.events, 1)
			assert.Equal(t, logging.LevelWarn, writer.events[0].Level)
		})
	}
}

func TestMultiLogger_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		service string
	}{
		{
			name:    "error event",
			service: "api",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer := &testWriter{}
			logger := daemon.New(writer)

			logger.Error(tt.service, "failed", "Request failed", nil)

			require.Len(t, writer.events, 1)
			assert.Equal(t, logging.LevelError, writer.events[0].Level)
		})
	}
}

func TestMultiLogger_AddWriter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "add writer at runtime"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			writer1 := &testWriter{}
			logger := daemon.New(writer1)

			logger.Info("test", "event", "message", nil)
			assert.Len(t, writer1.events, 1)

			writer2 := &testWriter{}
			logger.AddWriter(writer2)

			logger.Info("test", "event2", "message2", nil)
			assert.Len(t, writer1.events, 2)
			assert.Len(t, writer2.events, 1)
		})
	}
}

func TestMultiLogger_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "close multi logger"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			logger := daemon.New(&testWriter{}, &testWriter{})
			err := logger.Close()
			assert.NoError(t, err)
		})
	}
}
