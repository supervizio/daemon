// Package logging provides nopcloser.go implementing a no-op closer for writers.
// It wraps an io.Writer and provides a no-op Close method.
package logging

import (
	"io"
)

// nopCloser wraps an io.Writer and provides a no-op Close.
// It is used when output should go to os.Stdout or os.Stderr without closing them.
type nopCloser struct {
	// Writer is the embedded writer that receives all write operations.
	io.Writer
}

// Close implements io.Closer with a no-op operation.
// It always returns nil since the underlying writer should not be closed.
//
// Returns:
//   - error: always returns nil.
func (n *nopCloser) Close() error {
	return nil
}
