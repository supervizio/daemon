package daemon_test

import (
	"sync"
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockWriter is a test double for logging.Writer.
type mockWriter struct {
	mu     sync.Mutex
	events []logging.LogEvent
	closed bool
}

func (m *mockWriter) Write(event logging.LogEvent) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.events = append(m.events, event)
	return nil
}

func (m *mockWriter) Close() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.closed = true
	return nil
}

func (m *mockWriter) Events() []logging.LogEvent {
	m.mu.Lock()
	defer m.mu.Unlock()
	return m.events
}

func TestMultiLogger_Log(t *testing.T) {
	t.Parallel()

	mock1 := &mockWriter{}
	mock2 := &mockWriter{}
	logger := daemon.New(mock1, mock2)

	event := logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started").
		WithMeta("pid", 1234)

	logger.Log(event)

	// Both writers should receive the event.
	assert.Len(t, mock1.Events(), 1)
	assert.Len(t, mock2.Events(), 1)
	assert.Equal(t, "nginx", mock1.Events()[0].Service)
	assert.Equal(t, "started", mock2.Events()[0].EventType)
}

func TestMultiLogger_ConvenienceMethods(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		logFunc func(logger logging.Logger)
		level   logging.Level
		service string
		event   string
		metaKey string
		metaVal any
	}{
		{
			name: "Debug",
			logFunc: func(l logging.Logger) {
				l.Debug("svc", "debug_event", "debug message", map[string]any{"key": "value"})
			},
			level:   logging.LevelDebug,
			service: "svc",
			event:   "debug_event",
			metaKey: "key",
			metaVal: "value",
		},
		{
			name: "Info",
			logFunc: func(l logging.Logger) {
				l.Info("svc", "info_event", "info message", map[string]any{"pid": 123})
			},
			level:   logging.LevelInfo,
			service: "svc",
			event:   "info_event",
			metaKey: "pid",
			metaVal: 123,
		},
		{
			name: "Warn",
			logFunc: func(l logging.Logger) {
				l.Warn("svc", "warn_event", "warn message", map[string]any{"code": 1})
			},
			level:   logging.LevelWarn,
			service: "svc",
			event:   "warn_event",
			metaKey: "code",
			metaVal: 1,
		},
		{
			name: "Error",
			logFunc: func(l logging.Logger) {
				l.Error("svc", "error_event", "error message", map[string]any{"err": "failed"})
			},
			level:   logging.LevelError,
			service: "svc",
			event:   "error_event",
			metaKey: "err",
			metaVal: "failed",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockWriter{}
			logger := daemon.New(mock)

			tt.logFunc(logger)

			require.Len(t, mock.Events(), 1)
			event := mock.Events()[0]
			assert.Equal(t, tt.level, event.Level)
			assert.Equal(t, tt.service, event.Service)
			assert.Equal(t, tt.event, event.EventType)
			assert.Equal(t, tt.metaVal, event.Metadata[tt.metaKey])
		})
	}
}

func TestMultiLogger_Close(t *testing.T) {
	t.Parallel()

	mock1 := &mockWriter{}
	mock2 := &mockWriter{}
	logger := daemon.New(mock1, mock2)

	err := logger.Close()
	require.NoError(t, err)

	assert.True(t, mock1.closed)
	assert.True(t, mock2.closed)
}
