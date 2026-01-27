// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"github.com/kodflow/daemon/internal/domain/logging"
)

// LevelFilter wraps a writer and filters events below a minimum level.
// Events below the threshold are silently discarded without error.
type LevelFilter struct {
	// writer is the underlying writer.
	writer logging.Writer
	// minLevel is the minimum level to pass through.
	minLevel logging.Level
}

// WithLevelFilter wraps a writer with level filtering.
//
// Params:
//   - w: the writer to wrap.
//   - minLevel: the minimum level to pass through.
//
// Returns:
//   - *LevelFilter: the level-filtered writer.
func WithLevelFilter(w logging.Writer, minLevel logging.Level) *LevelFilter {
	// Create level filter wrapper.
	return &LevelFilter{
		writer:   w,
		minLevel: minLevel,
	}
}

// Write writes the event if it meets the minimum level threshold.
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: nil on success or if filtered, error on write failure.
func (f *LevelFilter) Write(event logging.LogEvent) error {
	// Skip events below the minimum level.
	if event.Level < f.minLevel {
		// Filter out events below threshold.
		return nil
	}
	// Pass through events at or above threshold.
	return f.writer.Write(event)
}

// Close closes the underlying writer.
//
// Returns:
//   - error: nil on success, error on failure.
func (f *LevelFilter) Close() error {
	// Delegate close to underlying writer.
	return f.writer.Close()
}

// Ensure LevelFilter implements logging.Writer.
var _ logging.Writer = (*LevelFilter)(nil)
