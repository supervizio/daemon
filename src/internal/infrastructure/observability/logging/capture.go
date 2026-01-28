// Package logging provides capture.go implementing stdout and stderr capture for services.
// It provides writers that can redirect process output to files or passthrough to standard streams.
package logging

import (
	"io"
	"os"
	"sync"

	"github.com/kodflow/daemon/internal/domain/config"
)

// GetServiceLogPather defines the interface for configuration access.
// It provides the method needed to get service log paths.
type GetServiceLogPather interface {
	// GetServiceLogPath returns the full path for a service log file.
	GetServiceLogPath(serviceName, logFile string) string
}

// serviceLogging defines the interface for service logging configuration.
// It provides access to stdout and stderr stream configurations.
type serviceLogging interface {
	// StdoutConfig returns a mutable pointer to the stdout configuration.
	StdoutConfig() *config.LogStreamConfig
	// StderrConfig returns a mutable pointer to the stderr configuration.
	StderrConfig() *config.LogStreamConfig
}

// Capture captures stdout and stderr for a service.
// It wraps output streams and provides thread-safe close operations.
type Capture struct {
	// mu protects concurrent access to the capture state.
	mu sync.Mutex
	// stdout is the writer for standard output.
	stdout io.WriteCloser
	// stderr is the writer for standard error.
	stderr io.WriteCloser
	// closed indicates whether the capture has been closed.
	closed bool
}

// NewCapture creates a new output capture for a service.
// It initializes stdout and stderr writers based on the service configuration.
//
// Params:
//   - serviceName: the name of the service being captured.
//   - cfg: the global configuration containing log path information.
//   - svcCfg: the service-specific logging configuration.
//
// Returns:
//   - *Capture: the initialized capture instance.
//   - error: an error if writer creation fails.
func NewCapture(serviceName string, cfg GetServiceLogPather, svcCfg serviceLogging) (*Capture, error) {
	c := &Capture{}

	if svcCfg.StdoutConfig().File() != "" {
		path := cfg.GetServiceLogPath(serviceName, svcCfg.StdoutConfig().File())
		writer, err := NewWriter(path, svcCfg.StdoutConfig())
		if err != nil {
			return nil, err
		}
		c.stdout = writer
	} else {
		c.stdout = &nopCloser{os.Stdout}
	}

	if svcCfg.StderrConfig().File() != "" {
		path := cfg.GetServiceLogPath(serviceName, svcCfg.StderrConfig().File())
		writer, err := NewWriter(path, svcCfg.StderrConfig())
		if err != nil {
			_ = c.stdout.Close()

			return nil, err
		}
		c.stderr = writer
	} else {
		c.stderr = &nopCloser{os.Stderr}
	}

	return c, nil
}

// Stdout returns the stdout writer.
// It provides access to the configured standard output stream.
//
// Returns:
//   - io.Writer: the stdout writer instance.
func (c *Capture) Stdout() io.Writer {
	return c.stdout
}

// Stderr returns the stderr writer.
// It provides access to the configured standard error stream.
//
// Returns:
//   - io.Writer: the stderr writer instance.
func (c *Capture) Stderr() io.Writer {
	return c.stderr
}

// Close closes both output streams.
// It is thread-safe and can be called multiple times safely.
//
// Returns:
//   - error: the first error encountered during close operations, if any.
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
