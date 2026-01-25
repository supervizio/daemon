package logging_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/stretchr/testify/assert"
)

func TestNewLogEvent(t *testing.T) {
	t.Parallel()

	before := time.Now()
	event := logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started")
	after := time.Now()

	assert.Equal(t, logging.LevelInfo, event.Level)
	assert.Equal(t, "nginx", event.Service)
	assert.Equal(t, "started", event.EventType)
	assert.Equal(t, "Service started", event.Message)
	assert.NotNil(t, event.Metadata)
	assert.Empty(t, event.Metadata)
	assert.True(t, event.Timestamp.After(before) || event.Timestamp.Equal(before))
	assert.True(t, event.Timestamp.Before(after) || event.Timestamp.Equal(after))
}

func TestLogEvent_WithMeta(t *testing.T) {
	t.Parallel()

	original := logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started")
	modified := original.WithMeta("pid", 1234)

	// Original should be unchanged.
	assert.Empty(t, original.Metadata)

	// Modified should have the new metadata.
	assert.Equal(t, 1234, modified.Metadata["pid"])
	assert.Len(t, modified.Metadata, 1)

	// Adding more metadata should not affect previous copies.
	modified2 := modified.WithMeta("exit_code", 0)
	assert.Len(t, modified.Metadata, 1)
	assert.Len(t, modified2.Metadata, 2)
	assert.Equal(t, 1234, modified2.Metadata["pid"])
	assert.Equal(t, 0, modified2.Metadata["exit_code"])
}

func TestLogEvent_WithMetadata(t *testing.T) {
	t.Parallel()

	original := logging.NewLogEvent(logging.LevelError, "nginx", "failed", "Service failed")
	modified := original.WithMetadata(map[string]any{
		"pid":       1234,
		"exit_code": 1,
		"error":     "exit code 1",
	})

	// Original should be unchanged.
	assert.Empty(t, original.Metadata)

	// Modified should have all the metadata.
	assert.Len(t, modified.Metadata, 3)
	assert.Equal(t, 1234, modified.Metadata["pid"])
	assert.Equal(t, 1, modified.Metadata["exit_code"])
	assert.Equal(t, "exit code 1", modified.Metadata["error"])
}

func TestLogEvent_WithMetadata_Nil(t *testing.T) {
	t.Parallel()

	original := logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started")
	modified := original.WithMetadata(nil)

	// Should return the same event when nil is passed.
	assert.Equal(t, original.Timestamp, modified.Timestamp)
	assert.Equal(t, original.Level, modified.Level)
	assert.Equal(t, original.Service, modified.Service)
}

func TestLogEvent_WithMetadata_Merge(t *testing.T) {
	t.Parallel()

	original := logging.NewLogEvent(logging.LevelInfo, "nginx", "started", "Service started").
		WithMeta("existing", "value")

	modified := original.WithMetadata(map[string]any{
		"new_key": "new_value",
	})

	// Should have both existing and new metadata.
	assert.Len(t, modified.Metadata, 2)
	assert.Equal(t, "value", modified.Metadata["existing"])
	assert.Equal(t, "new_value", modified.Metadata["new_key"])

	// Original should still have only one key.
	assert.Len(t, original.Metadata, 1)
}
