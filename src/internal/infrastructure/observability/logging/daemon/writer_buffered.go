// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"cmp"
	"sync"

	"github.com/kodflow/daemon/internal/domain/logging"
)

// Buffer size limits to prevent OOM.
const (
	defaultBufferCap int = 64
	maxBufferSize    int = 1024
)

// BufferedWriter buffers log events until Flush is called.
// This is useful for delaying console output until after the MOTD banner is displayed.
// Buffer size is limited to maxBufferSize to prevent unbounded memory growth.
type BufferedWriter struct {
	// mu protects concurrent access.
	mu sync.Mutex
	// inner is the underlying writer to delegate to after flush.
	inner logging.Writer
	// buffer holds events until Flush is called.
	buffer []logging.LogEvent
	// flushed indicates whether Flush has been called.
	flushed bool
	// closed indicates whether Close has been called (prevents write after close).
	closed bool
	// dropped counts events dropped due to buffer overflow.
	dropped int
}

// NewBufferedWriter creates a new buffered writer wrapping the given writer.
//
// Params:
//   - inner: the underlying writer to delegate to after flush.
//
// Returns:
//   - *BufferedWriter: the created buffered writer.
func NewBufferedWriter(inner logging.Writer) *BufferedWriter {
	return &BufferedWriter{
		inner:  inner,
		buffer: make([]logging.LogEvent, 0, defaultBufferCap),
	}
}

// Write buffers or writes an event depending on whether Flush has been called.
// If buffer is full (maxBufferSize reached), events are dropped to prevent OOM.
// Returns nil silently if writer is closed (no error, events discarded).
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: nil on success, error on failure (only after flush).
func (w *BufferedWriter) Write(event logging.LogEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Closed writer discards events (no error to avoid cascading failures).
	if w.closed {
		return nil
	}

	if w.flushed {
		// After flush, write directly to inner writer.
		return w.inner.Write(event)
	}

	// Before flush, buffer the event (with size limit to prevent OOM).
	if len(w.buffer) >= maxBufferSize {
		// Buffer full, drop oldest event to make room.
		w.dropped++
		w.buffer = w.buffer[1:]
	}
	w.buffer = append(w.buffer, event)
	return nil
}

// Flush writes all buffered events to the inner writer and enables direct writes.
// After calling Flush, all subsequent Write calls go directly to the inner writer.
//
// Returns:
//   - error: the first error encountered during flush, or nil if all succeeded.
func (w *BufferedWriter) Flush() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.flushed {
		// Already flushed.
		return nil
	}

	// Write all buffered events.
	var firstErr error
	for _, event := range w.buffer {
		if err := w.inner.Write(event); err != nil && firstErr == nil {
			firstErr = err
		}
	}

	// Clear buffer and mark as flushed.
	w.buffer = nil
	w.flushed = true

	return firstErr
}

// Close flushes any remaining buffered events and closes the inner writer.
// Thread-safe: sets closed flag under lock to prevent write-after-close races.
//
// Returns:
//   - error: error from flushing or closing, or nil.
func (w *BufferedWriter) Close() error {
	w.mu.Lock()
	if w.closed {
		w.mu.Unlock()
		return nil
	}
	w.closed = true
	w.mu.Unlock()

	flushErr := w.Flush()
	closeErr := w.inner.Close()
	// Return first non-nil error (flush takes priority).
	return cmp.Or(flushErr, closeErr)
}

// Ensure BufferedWriter implements logging.Writer.
var _ logging.Writer = (*BufferedWriter)(nil)
