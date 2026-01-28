package daemon

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/stretchr/testify/assert"
)

func TestLevelFilter_Internal(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		minLevel    logging.Level
		eventLevel  logging.Level
		shouldWrite bool
	}{
		{
			name:        "filter debug when minLevel is info",
			minLevel:    logging.LevelInfo,
			eventLevel:  logging.LevelDebug,
			shouldWrite: false,
		},
		{
			name:        "pass error when minLevel is info",
			minLevel:    logging.LevelInfo,
			eventLevel:  logging.LevelError,
			shouldWrite: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			mock := &mockInternalWriter{}
			filter := &LevelFilter{
				writer:   mock,
				minLevel: tt.minLevel,
			}

			event := logging.NewLogEvent(tt.eventLevel, "test", "event", "message")
			err := filter.Write(event)
			assert.NoError(t, err)

			if tt.shouldWrite {
				assert.Equal(t, 1, mock.writeCount)
			} else {
				assert.Equal(t, 0, mock.writeCount)
			}
		})
	}
}

type mockInternalWriter struct {
	writeCount int
}

func (m *mockInternalWriter) Write(event logging.LogEvent) error {
	m.writeCount++
	return nil
}

func (m *mockInternalWriter) Close() error {
	return nil
}
