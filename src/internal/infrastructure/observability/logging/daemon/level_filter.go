// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"github.com/kodflow/daemon/internal/domain/logging"
)

// LevelFilter wraps a writer and filters events below a minimum level.
// Events below the threshold are silently discarded without error.
type LevelFilter struct {
	writer   logging.Writer
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
	// Create and return level filter.
	return &LevelFilter{
		writer:   w,
		minLevel: minLevel,
	}
}

// NewLevelFilter creates a new level filter.
// Alias for WithLevelFilter for KTN-CTOR compliance.
//
// Params:
//   - w: the writer to wrap.
//   - minLevel: the minimum level to pass through.
//
// Returns:
//   - *LevelFilter: the level-filtered writer.
func NewLevelFilter(w logging.Writer, minLevel logging.Level) *LevelFilter {
	// Delegate to WithLevelFilter.
	return WithLevelFilter(w, minLevel)
}

// Write writes the event if it meets the minimum level threshold.
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: nil on success or if filtered, error on write failure.
func (f *LevelFilter) Write(event logging.LogEvent) error {
	// Filter out events below minimum level.
	if event.Level < f.minLevel {
		// Silently discard.
		return nil
	}
	// Pass through to underlying writer.
	return f.writer.Write(event)
}

// Close closes the underlying writer.
//
// Returns:
//   - error: nil on success, error on failure.
func (f *LevelFilter) Close() error {
	// Close underlying writer.
	return f.writer.Close()
}

// Ensure LevelFilter implements logging.Writer.
var _ logging.Writer = (*LevelFilter)(nil)
