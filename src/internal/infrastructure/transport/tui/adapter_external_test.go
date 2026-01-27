// Package tui_test provides external (black-box) tests for the TUI adapter.
// External tests validate the public API without accessing internal state.
package tui_test

import (
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestNewLogBuffer verifies LogBuffer creation with various sizes.
func TestNewLogBuffer(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		maxSize  int
		wantSize int
	}{
		{name: "positive size", maxSize: 50, wantSize: 50},
		{name: "zero size uses default", maxSize: 0, wantSize: 100},
		{name: "negative size uses default", maxSize: -1, wantSize: 100},
	}

	// Run test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			buf := tui.NewLogBuffer(tt.maxSize)
			require.NotNil(t, buf)
		})
	}
}

// TestLogBufferAddAndEntries verifies adding entries and retrieving them.
func TestLogBufferAddAndEntries(t *testing.T) {
	t.Parallel()

	buf := tui.NewLogBuffer(3)

	// Add entries.
	for i := range 5 {
		entry := model.LogEntry{
			Timestamp: time.Now(),
			Level:     "INFO",
			Message:   "test message",
		}
		_ = i
		buf.Add(entry)
	}

	// Verify ring buffer behavior.
	entries := buf.Entries()
	assert.Len(t, entries, 3)
}

// TestLogBufferSummary verifies log summary counts.
func TestLogBufferSummary(t *testing.T) {
	t.Parallel()

	buf := tui.NewLogBuffer(10)

	// Add mixed level entries.
	buf.Add(model.LogEntry{Level: "INFO"})
	buf.Add(model.LogEntry{Level: "WARN"})
	buf.Add(model.LogEntry{Level: "ERROR"})
	buf.Add(model.LogEntry{Level: "INFO"})

	summary := buf.Summary()
	assert.Equal(t, 2, summary.InfoCount)
	assert.Equal(t, 1, summary.WarnCount)
	assert.Equal(t, 1, summary.ErrorCount)
	assert.True(t, summary.HasAlerts)
}

// TestLogAdapterBasic verifies LogAdapter basic operations.
func TestLogAdapterBasic(t *testing.T) {
	t.Parallel()

	adapter := tui.NewLogAdapter()
	require.NotNil(t, adapter)
	require.NotNil(t, adapter.Buffer())

	// Add entry via adapter.
	adapter.AddLog(model.LogEntry{Level: "INFO", Message: "test"})

	summary := adapter.LogSummary()
	assert.Equal(t, 1, summary.InfoCount)
}

// TestNewDynamicServiceProvider verifies provider creation.
func TestNewDynamicServiceProvider(t *testing.T) {
	t.Parallel()

	provider := tui.NewDynamicServiceProvider(nil, nil)
	require.NotNil(t, provider)

	// Nil provider returns nil services.
	services := provider.Services()
	assert.Nil(t, services)
}

// mockSnapshotsProvider implements TUISnapshotser for testing.
type mockSnapshotsProvider struct {
	snapshots []tui.TUISnapshotData
}

// TUISnapshots returns mock snapshots.
func (m *mockSnapshotsProvider) TUISnapshots() []tui.TUISnapshotData {
	// Return configured snapshots.
	return m.snapshots
}

// TestDynamicServiceProviderWithSnapshots verifies service conversion.
func TestDynamicServiceProviderWithSnapshots(t *testing.T) {
	t.Parallel()

	mock := &mockSnapshotsProvider{
		snapshots: []tui.TUISnapshotData{
			{Name: "test-service", State: process.StateRunning, PID: 1234, Uptime: 60},
		},
	}

	provider := tui.NewDynamicServiceProvider(mock, nil)
	services := provider.Services()

	require.Len(t, services, 1)
	assert.Equal(t, "test-service", services[0].Name)
	assert.Equal(t, process.StateRunning, services[0].State)
	assert.Equal(t, 1234, services[0].PID)
}

// TestSystemMetricsAdapter verifies system metrics adapter.
func TestSystemMetricsAdapter(t *testing.T) {
	t.Parallel()

	adapter := tui.NewSystemMetricsAdapter()
	require.NotNil(t, adapter)

	// Returns empty metrics.
	metrics := adapter.SystemMetrics()
	assert.Equal(t, model.SystemMetrics{}, metrics)
}

// TestTUILogWriter verifies log writer operations.
func TestTUILogWriter(t *testing.T) {
	t.Parallel()

	adapter := tui.NewLogAdapter()
	writer := tui.NewTUILogWriter(adapter)
	require.NotNil(t, writer)

	// Close should not error.
	err := writer.Close()
	assert.NoError(t, err)
}

// TestLogBufferClear verifies buffer clearing.
func TestLogBufferClear(t *testing.T) {
	t.Parallel()

	buf := tui.NewLogBuffer(10)

	// Add entries.
	buf.Add(model.LogEntry{Level: "INFO"})
	buf.Add(model.LogEntry{Level: "ERROR"})

	// Clear buffer.
	buf.Clear()

	// Verify empty.
	entries := buf.Entries()
	assert.Nil(t, entries)

	summary := buf.Summary()
	assert.Equal(t, 0, summary.InfoCount)
	assert.Equal(t, 0, summary.ErrorCount)
}
