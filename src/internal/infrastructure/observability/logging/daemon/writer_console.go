// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"io"
	"os"
	"sync"

	"golang.org/x/term"

	"github.com/kodflow/daemon/internal/domain/logging"
)

// ANSI color codes for log levels.
const (
	colorReset string = "\033[0m"
	colorDebug string = "\033[36m" // Cyan
	colorInfo  string = "\033[32m" // Green
	colorWarn  string = "\033[33m" // Yellow
	colorError string = "\033[31m" // Red
)

// ConsoleWriter writes log events to stdout/stderr based on level.
// DEBUG and INFO go to stdout, WARN and ERROR go to stderr.
// This is the DEFAULT writer when no configuration is provided.
type ConsoleWriter struct {
	// mu protects concurrent writes.
	mu sync.Mutex
	// stdout is the writer for INFO and below.
	stdout io.Writer
	// stderr is the writer for WARN and above.
	stderr io.Writer
	// format is the text formatter.
	format Formatter
	// color indicates whether to use ANSI colors.
	color bool
}

// NewConsoleWriter creates a new console writer with auto-detected color support.
//
// Returns:
//   - *ConsoleWriter: the created console writer.
func NewConsoleWriter() *ConsoleWriter {
	// Create console writer with OS defaults and auto-detected color.
	return NewConsoleWriterWithOptions(os.Stdout, os.Stderr, isTerminal(os.Stdout))
}

// NewConsoleWriterWithOptions creates a console writer with explicit options.
//
// Params:
//   - stdout: the writer for INFO and below.
//   - stderr: the writer for WARN and above.
//   - color: whether to use ANSI colors.
//
// Returns:
//   - *ConsoleWriter: the created console writer.
func NewConsoleWriterWithOptions(stdout, stderr io.Writer, color bool) *ConsoleWriter {
	// Create console writer with custom options.
	return &ConsoleWriter{
		stdout: stdout,
		stderr: stderr,
		format: NewTextFormatter(""),
		color:  color,
	}
}

// Write writes a log event to the appropriate output stream.
// DEBUG and INFO go to stdout, WARN and ERROR go to stderr.
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: nil on success, error on failure.
func (w *ConsoleWriter) Write(event logging.LogEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Choose output based on level.
	var out io.Writer
	// WARN and ERROR go to stderr.
	if event.Level >= logging.LevelWarn {
		out = w.stderr
	} else {
		// DEBUG and INFO go to stdout.
		out = w.stdout
	}

	// Format the event.
	line := w.format.Format(event)

	// Add color if enabled.
	if w.color {
		line = w.colorize(event.Level, line)
	}

	// Write with newline.
	_, err := out.Write([]byte(line + "\n"))
	// Return write result.
	return err
}

// colorize adds ANSI color codes to the log line.
//
// Params:
//   - level: the log level to colorize.
//   - line: the log line to colorize.
//
// Returns:
//   - string: the colorized log line.
func (w *ConsoleWriter) colorize(level logging.Level, line string) string {
	var color string
	// Select color based on log level.
	switch level {
	// Debug level uses cyan.
	case logging.LevelDebug:
		color = colorDebug
	// Info level uses green.
	case logging.LevelInfo:
		color = colorInfo
	// Warn level uses yellow.
	case logging.LevelWarn:
		color = colorWarn
	// Error level uses red.
	case logging.LevelError:
		color = colorError
	// Unknown level, no color.
	default:
		// No color for unknown level.
		return line
	}
	// Wrap line with color codes.
	return color + line + colorReset
}

// Close is a no-op for console writer since we don't own stdout/stderr.
//
// Returns:
//   - error: always nil.
func (w *ConsoleWriter) Close() error {
	// No-op, we don't own stdout/stderr.
	return nil
}

// isTerminal checks if the given writer is a terminal.
//
// Params:
//   - w: the writer to check.
//
// Returns:
//   - bool: true if the writer is a terminal, false otherwise.
func isTerminal(w io.Writer) bool {
	// Check if writer is a file and supports terminal detection.
	if f, ok := w.(*os.File); ok {
		// Check if file descriptor is a terminal.
		return term.IsTerminal(int(f.Fd()))
	}
	// Not a file, not a terminal.
	return false
}

// Ensure ConsoleWriter implements logging.Writer.
var _ logging.Writer = (*ConsoleWriter)(nil)
