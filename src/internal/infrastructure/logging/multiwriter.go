// Package logging provides log management with rotation and capture.
// It handles automatic file rotation based on size and file count.
package logging

import (
	"io"
)

// MultiWriter writes to multiple writers.
// It duplicates output to all underlying writers for logging to multiple destinations.
type MultiWriter struct {
	// writers is the slice of writers to duplicate output to.
	writers []io.WriteCloser
}

// NewMultiWriter creates a writer that duplicates output to multiple writers.
//
// Params:
//   - writers: writers to duplicate output to
//
// Returns:
//   - *MultiWriter: new multi-writer instance
func NewMultiWriter(writers ...io.WriteCloser) *MultiWriter {
	// Return new MultiWriter with provided writers.
	return &MultiWriter{writers: writers}
}

// Write writes to all writers.
//
// Params:
//   - p: bytes to write
//
// Returns:
//   - int: number of bytes written
//   - error: nil on success, error on failure
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	// Iterate through each writer to duplicate the output.
	for _, w := range mw.writers {
		n, err = w.Write(p)

		// Check if write operation succeeded for this writer.
		if err != nil {
			// Write failed - return bytes written and error.
			return n, err
		}
	}

	// Return total bytes written.
	return len(p), nil
}

// Close closes all underlying writers and returns the first error encountered.
//
// Returns:
//   - error: nil on success, first error encountered on failure
func (mw *MultiWriter) Close() error {
	var firstErr error

	// Iterate through each writer to close them all.
	for _, w := range mw.writers {
		// Check if close operation succeeded and track the first error.
		if err := w.Close(); err != nil && firstErr == nil {
			// Track first error.
			firstErr = err
		}
	}

	// Return first error encountered (or nil if all succeeded).
	return firstErr
}
