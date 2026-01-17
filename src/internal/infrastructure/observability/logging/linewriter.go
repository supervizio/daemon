// Package logging provides linewriter.go implementing a line-buffered writer with prefix.
// It buffers input and writes complete lines with the configured prefix.
package logging

import (
	"io"
)

// LineWriter constants for line buffering and processing.
const (
	// newlineChar is the newline character used to detect line endings.
	newlineChar byte = '\n'
	// indexNotFound is the sentinel value indicating no newline was found in buffer.
	indexNotFound int = -1
	// zeroBytes is the return value when no bytes were written due to an error.
	zeroBytes int = 0
)

// LineWriter writes lines with optional prefix.
// It buffers input and writes complete lines with the configured prefix.
type LineWriter struct {
	// writer is the underlying writer that receives formatted output.
	writer io.Writer
	// prefix is the string prepended to each line.
	prefix string
	// buf holds incomplete line data until a newline is received.
	buf []byte
}

// NewLineWriter creates a writer that prefixes each line.
// It wraps an existing writer and adds prefix support with line buffering.
//
// Params:
//   - w: the underlying writer to write to.
//   - prefix: the string to prepend to each line.
//
// Returns:
//   - *LineWriter: the initialized line writer instance.
func NewLineWriter(w io.Writer, prefix string) *LineWriter {
	// Return a new LineWriter initialized with the provided writer and prefix.
	return &LineWriter{
		writer: w,
		prefix: prefix,
	}
}

// Write implements io.Writer with line buffering.
// It buffers data until complete lines are available, then writes them with prefix.
//
// Params:
//   - p: the byte slice to write.
//
// Returns:
//   - int: the number of bytes from p that were processed.
//   - error: an error if writing to the underlying writer fails.
func (lw *LineWriter) Write(p []byte) (n int, err error) {
	lw.buf = append(lw.buf, p...)

	// Iterate through the buffer to find and process complete lines.
	for {
		idx := indexNotFound
		// Iterate through buffer bytes to find the next newline character.
		for i, b := range lw.buf {
			// Check if current byte is a newline character.
			if b == newlineChar {
				idx = i
				break
			}
		}

		// Check if no newline was found, meaning no complete line is available.
		if idx == indexNotFound {
			break
		}

		line := lw.buf[:idx+1]
		lw.buf = lw.buf[idx+1:]

		// Check if a prefix is configured and should be written.
		if lw.prefix != "" {
			// Check if prefix write operation failed.
			if _, err := lw.writer.Write([]byte(lw.prefix)); err != nil {
				// Return zero bytes written and propagate the prefix write error.
				return zeroBytes, err
			}
		}
		// Check if line write operation failed.
		if _, err := lw.writer.Write(line); err != nil {
			// Return zero bytes written and propagate the line write error.
			return zeroBytes, err
		}
	}

	// Return the total number of bytes processed from the input.
	return len(p), nil
}

// Flush writes any remaining buffered data.
// It ensures all data is written even if no trailing newline was received.
//
// Returns:
//   - error: an error if writing to the underlying writer fails.
func (lw *LineWriter) Flush() error {
	// Check if there is any buffered data remaining to be flushed.
	if len(lw.buf) > zeroBytes {
		// Check if a prefix is configured and should be written.
		if lw.prefix != "" {
			// Check if prefix write operation failed.
			if _, err := lw.writer.Write([]byte(lw.prefix)); err != nil {
				// Return the prefix write error.
				return err
			}
		}
		// Check if buffer content write operation failed.
		if _, err := lw.writer.Write(lw.buf); err != nil {
			// Return the buffer content write error.
			return err
		}
		// Check if trailing newline write operation failed.
		if _, err := lw.writer.Write([]byte{newlineChar}); err != nil {
			// Return the newline write error.
			return err
		}
		lw.buf = nil
	}
	// Return nil indicating successful flush or no data to flush.
	return nil
}
