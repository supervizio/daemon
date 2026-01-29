// Package tui provides terminal user interface rendering for superviz.io.
package tui

import domainlogging "github.com/kodflow/daemon/internal/domain/logging"

// TUILogWriter implements domain/logging.Writer to capture logs for TUI.
// It forwards log events to a LogAdapter for display.
type TUILogWriter struct {
	adapter *LogAdapter
}

// NewTUILogWriter creates a writer that sends logs to the TUI.
//
// Params:
//   - adapter: the log adapter to write to.
//
// Returns:
//   - *TUILogWriter: the created writer.
func NewTUILogWriter(adapter *LogAdapter) *TUILogWriter {
	// return computed result.
	return &TUILogWriter{
		adapter: adapter,
	}
}

// Write implements domain/logging.Writer.
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: always nil (errors are ignored).
func (w *TUILogWriter) Write(event domainlogging.LogEvent) error {
	// handle non-nil condition.
	if w.adapter != nil {
		w.adapter.AddDomainEvent(event)
	}
	// return nil to indicate no error.
	return nil
}

// Close implements domain/logging.Writer.
//
// Returns:
//   - error: always nil (no cleanup needed).
func (w *TUILogWriter) Close() error {
	// return nil to indicate no error.
	return nil
}
