package daemon_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
)

func TestNewTextFormatter(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		timestampFormat string
	}{
		{
			name:            "default format",
			timestampFormat: "",
		},
		{
			name:            "custom format",
			timestampFormat: "2006-01-02",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatter := daemon.NewTextFormatter(tt.timestampFormat)
			assert.NotNil(t, formatter)
		})
	}
}

func TestTextFormatter_Format(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		event    logging.LogEvent
		contains []string
	}{
		{
			name: "info event with service",
			event: logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started").
				WithMeta("pid", 1234),
			contains: []string{"[INFO]", "nginx", "Service started", "pid=1234"},
		},
		{
			name: "error event without service",
			event: logging.NewLogEvent(logging.LevelError, "", "failure", "System error").
				WithMeta("code", 500),
			contains: []string{"[ERROR]", "System error", "code=500"},
		},
		{
			name:     "event without metadata",
			event:    logging.NewLogEvent(logging.LevelDebug, "test", "event", "Test message"),
			contains: []string{"[DEBUG]", "test", "Test message"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			formatter := daemon.NewTextFormatter("")
			output := formatter.Format(tt.event)

			for _, expected := range tt.contains {
				assert.Contains(t, output, expected)
			}
		})
	}
}
