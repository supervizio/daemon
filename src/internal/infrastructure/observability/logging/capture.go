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

	// Check if stdout file path is configured for file-based logging.
	if svcCfg.StdoutConfig().File() != "" {
		path := cfg.GetServiceLogPath(serviceName, svcCfg.StdoutConfig().File())
		writer, err := NewWriter(path, svcCfg.StdoutConfig())
		// Check if writer creation failed.
		if err != nil {
			// Return nil capture and propagate the writer creation error.
			return nil, err
		}
		c.stdout = writer
	} else {
		// Else use a no-op closer wrapping os.Stdout for passthrough mode.
		c.stdout = &nopCloser{os.Stdout}
	}

	// Check if stderr file path is configured for file-based logging.
	if svcCfg.StderrConfig().File() != "" {
		path := cfg.GetServiceLogPath(serviceName, svcCfg.StderrConfig().File())
		writer, err := NewWriter(path, svcCfg.StderrConfig())
		// Check if writer creation failed and cleanup is needed.
		if err != nil {
			// Ignore close error since we're returning a different error
			_ = c.stdout.Close()
			// Return nil capture and propagate the stderr writer creation error.
			return nil, err
		}
		c.stderr = writer
	} else {
		// Else use a no-op closer wrapping os.Stderr for passthrough mode.
		c.stderr = &nopCloser{os.Stderr}
	}

	// Return the fully initialized capture instance with no error.
	return c, nil
}

// Stdout returns the stdout writer.
// It provides access to the configured standard output stream.
//
// Returns:
//   - io.Writer: the stdout writer instance.
func (c *Capture) Stdout() io.Writer {
	// Return the configured stdout writer for the capture.
	return c.stdout
}

// Stderr returns the stderr writer.
// It provides access to the configured standard error stream.
//
// Returns:
//   - io.Writer: the stderr writer instance.
func (c *Capture) Stderr() io.Writer {
	// Return the configured stderr writer for the capture.
	return c.stderr
}

// Close closes both output streams.
// It is thread-safe and can be called multiple times safely.
//
// Returns:
//   - error: the first error encountered during close operations, if any.
func (c *Capture) Close() error {
	c.mu.Lock()
	// Defer unlocking the mutex to ensure it is released on function exit.
	defer c.mu.Unlock()

	// Check if the capture has already been closed to prevent double-close.
	if c.closed {
		// Return nil since already closed successfully.
		return nil
	}
	c.closed = true

	var firstErr error
	// Check if stdout close returns an error and capture it.
	if err := c.stdout.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	// Check if stderr close returns an error and capture it.
	if err := c.stderr.Close(); err != nil && firstErr == nil {
		firstErr = err
	}
	// Return the first error encountered, or nil if both closed successfully.
	return firstErr
}
