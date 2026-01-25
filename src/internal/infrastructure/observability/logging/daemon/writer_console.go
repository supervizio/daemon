package daemon

import (
	"io"
	"os"
	"sync"

	"github.com/kodflow/daemon/internal/domain/logging"
	"golang.org/x/term"
)

// ANSI color codes for log levels.
const (
	colorReset  = "\033[0m"
	colorDebug  = "\033[36m" // Cyan
	colorInfo   = "\033[32m" // Green
	colorWarn   = "\033[33m" // Yellow
	colorError  = "\033[31m" // Red
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
	if event.Level >= logging.LevelWarn {
		out = w.stderr
	} else {
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
	return err
}

// colorize adds ANSI color codes to the log line.
func (w *ConsoleWriter) colorize(level logging.Level, line string) string {
	var color string
	switch level {
	case logging.LevelDebug:
		color = colorDebug
	case logging.LevelInfo:
		color = colorInfo
	case logging.LevelWarn:
		color = colorWarn
	case logging.LevelError:
		color = colorError
	default:
		return line
	}
	return color + line + colorReset
}

// Close is a no-op for console writer since we don't own stdout/stderr.
//
// Returns:
//   - error: always nil.
func (w *ConsoleWriter) Close() error {
	return nil
}

// isTerminal checks if the given writer is a terminal.
func isTerminal(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// Ensure ConsoleWriter implements logging.Writer.
var _ logging.Writer = (*ConsoleWriter)(nil)
