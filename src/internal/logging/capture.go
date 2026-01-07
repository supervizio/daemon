package logging

import (
	"io"
	"os"
	"sync"

	"github.com/kodflow/daemon/internal/config"
)

// Capture captures stdout and stderr for a service.
type Capture struct {
	mu     sync.Mutex
	stdout io.WriteCloser
	stderr io.WriteCloser
	config *config.ServiceLogging
	closed bool
}

// NewCapture creates a new output capture for a service.
func NewCapture(serviceName string, cfg *config.Config, svcCfg *config.ServiceLogging) (*Capture, error) {
	c := &Capture{
		config: svcCfg,
	}

	// Setup stdout
	if svcCfg.Stdout.File != "" {
		path := cfg.GetServiceLogPath(serviceName, svcCfg.Stdout.File)
		writer, err := NewWriter(path, &svcCfg.Stdout)
		if err != nil {
			return nil, err
		}
		c.stdout = writer
	} else {
		c.stdout = &nopCloser{os.Stdout}
	}

	// Setup stderr
	if svcCfg.Stderr.File != "" {
		path := cfg.GetServiceLogPath(serviceName, svcCfg.Stderr.File)
		writer, err := NewWriter(path, &svcCfg.Stderr)
		if err != nil {
			c.stdout.Close()
			return nil, err
		}
		c.stderr = writer
	} else {
		c.stderr = &nopCloser{os.Stderr}
	}

	return c, nil
}

// Stdout returns the stdout writer.
func (c *Capture) Stdout() io.Writer {
	return c.stdout
}

// Stderr returns the stderr writer.
func (c *Capture) Stderr() io.Writer {
	return c.stderr
}

// Close closes both output streams.
func (c *Capture) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.closed {
		return nil
	}
	c.closed = true

	var firstErr error
	if err := c.stdout.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	if err := c.stderr.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	return firstErr
}

// nopCloser wraps an io.Writer and provides a no-op Close.
type nopCloser struct {
	io.Writer
}

func (n *nopCloser) Close() error {
	return nil
}

// LineWriter writes lines with optional prefix.
type LineWriter struct {
	writer io.Writer
	prefix string
	buf    []byte
}

// NewLineWriter creates a writer that prefixes each line.
func NewLineWriter(w io.Writer, prefix string) *LineWriter {
	return &LineWriter{
		writer: w,
		prefix: prefix,
	}
}

// Write implements io.Writer with line buffering.
func (lw *LineWriter) Write(p []byte) (n int, err error) {
	lw.buf = append(lw.buf, p...)

	for {
		idx := -1
		for i, b := range lw.buf {
			if b == '\n' {
				idx = i
				break
			}
		}

		if idx < 0 {
			break
		}

		line := lw.buf[:idx+1]
		lw.buf = lw.buf[idx+1:]

		if lw.prefix != "" {
			if _, err := lw.writer.Write([]byte(lw.prefix)); err != nil {
				return 0, err
			}
		}
		if _, err := lw.writer.Write(line); err != nil {
			return 0, err
		}
	}

	return len(p), nil
}

// Flush writes any remaining buffered data.
func (lw *LineWriter) Flush() error {
	if len(lw.buf) > 0 {
		if lw.prefix != "" {
			if _, err := lw.writer.Write([]byte(lw.prefix)); err != nil {
				return err
			}
		}
		if _, err := lw.writer.Write(lw.buf); err != nil {
			return err
		}
		if _, err := lw.writer.Write([]byte{'\n'}); err != nil {
			return err
		}
		lw.buf = nil
	}
	return nil
}
