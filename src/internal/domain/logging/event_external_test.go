package logging_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewLogEvent(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		level     logging.Level
		service   string
		eventType string
		message   string
	}{
		{
			name:      "info level event",
			level:     logging.LevelInfo,
			service:   "nginx",
			eventType: "started",
			message:   "Service started",
		},
		{
			name:      "error level event",
			level:     logging.LevelError,
			service:   "postgres",
			eventType: "failed",
			message:   "Service failed",
		},
		{
			name:      "daemon level event",
			level:     logging.LevelDebug,
			service:   "",
			eventType: "init",
			message:   "Daemon initializing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			before := time.Now()
			event := logging.NewLogEvent(tt.level, tt.service, tt.eventType, tt.message)
			after := time.Now()

			assert.Equal(t, tt.level, event.Level)
			assert.Equal(t, tt.service, event.Service)
			assert.Equal(t, tt.eventType, event.EventType)
			assert.Equal(t, tt.message, event.Message)
			assert.NotNil(t, event.Metadata)
			assert.Empty(t, event.Metadata)
			assert.True(t, event.Timestamp.After(before) || event.Timestamp.Equal(before))
			assert.True(t, event.Timestamp.Before(after) || event.Timestamp.Equal(after))
		})
	}
}

func TestLogEvent_WithMeta(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name          string
		initialMeta   map[string]any
		addKey        string
		addValue      any
		expectedCount int
	}{
		{
			name:          "add to empty metadata",
			initialMeta:   nil,
			addKey:        "pid",
			addValue:      1234,
			expectedCount: 1,
		},
		{
			name:          "add to existing metadata",
			initialMeta:   map[string]any{"existing": "value"},
			addKey:        "new_key",
			addValue:      "new_value",
			expectedCount: 2,
		},
		{
			name:          "override existing key",
			initialMeta:   map[string]any{"key": "old"},
			addKey:        "key",
			addValue:      "new",
			expectedCount: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started")
			if tt.initialMeta != nil {
				original = original.WithMetadata(tt.initialMeta)
			}

			modified := original.WithMeta(tt.addKey, tt.addValue)

			assert.Equal(t, tt.expectedCount, len(modified.Metadata))
			assert.Equal(t, tt.addValue, modified.Metadata[tt.addKey])
		})
	}
}

func TestLogEvent_WithMetadata(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		metadata map[string]any
		expected int
	}{
		{
			name: "add multiple metadata",
			metadata: map[string]any{
				"pid":       1234,
				"exit_code": 1,
				"error":     "exit code 1",
			},
			expected: 3,
		},
		{
			name:     "nil metadata returns unchanged",
			metadata: nil,
			expected: 0,
		},
		{
			name:     "empty metadata map",
			metadata: map[string]any{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := logging.NewLogEvent(logging.LevelError, "nginx", "failed", "Service failed")
			modified := original.WithMetadata(tt.metadata)

			assert.Empty(t, original.Metadata)
			assert.Equal(t, tt.expected, len(modified.Metadata))

			if tt.metadata != nil {
				for k, v := range tt.metadata {
					assert.Equal(t, v, modified.Metadata[k])
				}
			}
		})
	}
}

func TestLogEvent_WithMetadata_Merge(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		initialMeta    map[string]any
		additionalMeta map[string]any
		expectedKeys   []string
	}{
		{
			name:           "merge with existing metadata",
			initialMeta:    map[string]any{"existing": "value"},
			additionalMeta: map[string]any{"new_key": "new_value"},
			expectedKeys:   []string{"existing", "new_key"},
		},
		{
			name:           "override existing key",
			initialMeta:    map[string]any{"key": "old"},
			additionalMeta: map[string]any{"key": "new"},
			expectedKeys:   []string{"key"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			original := logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started").
				WithMetadata(tt.initialMeta)

			modified := original.WithMetadata(tt.additionalMeta)

			assert.Equal(t, len(tt.expectedKeys), len(modified.Metadata))
			for _, key := range tt.expectedKeys {
				assert.Contains(t, modified.Metadata, key)
			}
		})
	}
}
