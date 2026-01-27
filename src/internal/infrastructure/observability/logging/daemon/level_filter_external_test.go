package daemon_test

import (
	"testing"

	"github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/observability/logging/daemon"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLevelFilter_FiltersBelow(t *testing.T) {
	t.Parallel()

	mock := &mockWriter{}
	filtered := daemon.WithLevelFilter(mock, logging.LevelWarn)

	// Debug and Info should be filtered.
	filtered.Write(logging.NewLogEvent(logging.LevelDebug, "svc", "debug", "msg"))
	filtered.Write(logging.NewLogEvent(logging.LevelInfo, "svc", "info", "msg"))
	assert.Empty(t, mock.Events())

	// Warn and Error should pass through.
	filtered.Write(logging.NewLogEvent(logging.LevelWarn, "svc", "warn", "msg"))
	filtered.Write(logging.NewLogEvent(logging.LevelError, "svc", "error", "msg"))
	assert.Len(t, mock.Events(), 2)
}

func TestLevelFilter_PassesAtAndAbove(t *testing.T) {
	t.Parallel()

	mock := &mockWriter{}
	filtered := daemon.WithLevelFilter(mock, logging.LevelInfo)

	// Debug should be filtered.
	filtered.Write(logging.NewLogEvent(logging.LevelDebug, "svc", "debug", "msg"))
	assert.Empty(t, mock.Events())

	// Info and above should pass through.
	filtered.Write(logging.NewLogEvent(logging.LevelInfo, "svc", "info", "msg"))
	filtered.Write(logging.NewLogEvent(logging.LevelWarn, "svc", "warn", "msg"))
	filtered.Write(logging.NewLogEvent(logging.LevelError, "svc", "error", "msg"))
	assert.Len(t, mock.Events(), 3)
}

func TestLevelFilter_Close(t *testing.T) {
	t.Parallel()

	mock := &mockWriter{}
	filtered := daemon.WithLevelFilter(mock, logging.LevelInfo)

	err := filtered.Close()
	require.NoError(t, err)
	assert.True(t, mock.closed)
}
