package daemon_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockWriter struct {
	events []logging.LogEvent
}

func (m *mockWriter) Write(event logging.LogEvent) error {
	m.events = append(m.events, event)
	return nil
}

func (m *mockWriter) Close() error {
	return nil
}

func TestWithLevelFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		minLevel logging.Level
	}{
		{
			name:     "info level filter",
			minLevel: logging.LevelInfo,
		},
		{
			name:     "error level filter",
			minLevel: logging.LevelError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockWriter{}
			filter := daemon.WithLevelFilter(mock, tt.minLevel)
			assert.NotNil(t, filter)
		})
	}
}

func TestNewLevelFilter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		minLevel logging.Level
	}{
		{
			name:     "create with constructor",
			minLevel: logging.LevelWarn,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockWriter{}
			filter := daemon.NewLevelFilter(mock, tt.minLevel)
			assert.NotNil(t, filter)
		})
	}
}

func TestLevelFilter_Write(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		minLevel   logging.Level
		eventLevel logging.Level
		shouldPass bool
	}{
		{
			name:       "debug below info threshold",
			minLevel:   logging.LevelInfo,
			eventLevel: logging.LevelDebug,
			shouldPass: false,
		},
		{
			name:       "info at info threshold",
			minLevel:   logging.LevelInfo,
			eventLevel: logging.LevelInfo,
			shouldPass: true,
		},
		{
			name:       "error above info threshold",
			minLevel:   logging.LevelInfo,
			eventLevel: logging.LevelError,
			shouldPass: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockWriter{}
			filter := daemon.WithLevelFilter(mock, tt.minLevel)

			event := logging.NewLogEvent(tt.eventLevel, "test", "event", "message")
			err := filter.Write(event)
			require.NoError(t, err)

			if tt.shouldPass {
				assert.Len(t, mock.events, 1)
			} else {
				assert.Empty(t, mock.events)
			}
		})
	}
}

func TestLevelFilter_Close(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
	}{
		{name: "close level filter"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockWriter{}
			filter := daemon.WithLevelFilter(mock, logging.LevelInfo)
			err := filter.Close()
			assert.NoError(t, err)
		})
	}
}
