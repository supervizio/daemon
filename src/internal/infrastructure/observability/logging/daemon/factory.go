// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/logging"
)

// Writer type constants.
const (
	writerTypeConsole string = "console"
	writerTypeFile    string = "file"
	writerTypeJSON    string = "json"
)

// Sentinel errors for factory operations.
var (
	// ErrPathEscapesBase indicates the resolved path escapes the base directory.
	ErrPathEscapesBase error = errors.New("path escapes base directory")
	// ErrFilePathRequired indicates a file writer requires a path.
	ErrFilePathRequired error = errors.New("file writer requires path")
	// ErrJSONPathRequired indicates a JSON writer requires a path.
	ErrJSONPathRequired error = errors.New("json writer requires path")
	// ErrUnknownWriterType indicates an unknown writer type was specified.
	ErrUnknownWriterType error = errors.New("unknown writer type")
)

// BuildLogger creates a MultiLogger from configuration.
// If no writers are configured, it returns a default console logger.
//
// Params:
//   - cfg: the daemon logging configuration.
//   - baseDir: the base directory for log files.
//
// Returns:
//   - logging.Logger: the created logger.
//   - error: nil on success, error on failure.
func BuildLogger(cfg config.DaemonLogging, baseDir string) (logging.Logger, error) {
	// use default config if no writers specified
	if len(cfg.Writers) == 0 {
		cfg = config.DefaultDaemonLogging()
	}

	var writers []logging.Writer

	// create writer for each configuration
	for i := range cfg.Writers {
		wcfg := &cfg.Writers[i]
		w, err := buildWriter(*wcfg, baseDir)
		// handle writer creation failure
		if err != nil {
			// clean up already created writers
			for _, created := range writers {
				_ = created.Close()
			}
			// wrap error with context
			return nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
		}

		level, err := logging.ParseLevel(wcfg.Level)
		// use default level if parsing fails
		if err != nil {
			level = logging.LevelInfo
		}

		writers = append(writers, WithLevelFilter(w, level))
	}
	// create logger with configured writers
	return New(writers...), nil
}

// buildWriter creates a writer based on configuration type.
//
// Params:
//   - wcfg: the writer configuration.
//   - baseDir: the base directory for log files.
//
// Returns:
//   - logging.Writer: the created writer.
//   - error: nil on success, error on failure.
func buildWriter(wcfg config.WriterConfig, baseDir string) (logging.Writer, error) {
	// dispatch to appropriate writer type
	switch wcfg.Type {
	// console writer outputs to stdout/stderr
	case writerTypeConsole:
		// create console writer for terminal output
		return NewConsoleWriter(), nil
	// file writer outputs to log file
	case writerTypeFile:
		path := wcfg.File.Path
		// validate file path is provided
		if path == "" {
			// return error for missing path
			return nil, ErrFilePathRequired
		}
		resolvedPath, err := resolvePath(path, baseDir)
		// handle path resolution failure
		if err != nil {
			// wrap error with context
			return nil, fmt.Errorf("file writer: %w", err)
		}
		// create file writer with resolved path
		return NewFileWriter(resolvedPath, wcfg.File.Rotation)
	// json writer outputs structured logs
	case writerTypeJSON:
		path := wcfg.JSON.Path
		// validate JSON path is provided
		if path == "" {
			// return error for missing path
			return nil, ErrJSONPathRequired
		}
		resolvedPath, err := resolvePath(path, baseDir)
		// handle path resolution failure
		if err != nil {
			// wrap error with context
			return nil, fmt.Errorf("json writer: %w", err)
		}
		// create JSON writer with resolved path
		return NewJSONWriter(resolvedPath)
	// unknown writer type
	default:
		// return error for unknown type
		return nil, fmt.Errorf("%w: %s", ErrUnknownWriterType, wcfg.Type)
	}
}

// resolvePath resolves a path relative to baseDir and validates it doesn't escape.
//
// Params:
//   - path: the path to resolve.
//   - baseDir: the base directory.
//
// Returns:
//   - string: the resolved path.
//   - error: nil on success, error if path escapes baseDir.
func resolvePath(path, baseDir string) (string, error) {
	// use absolute path as-is
	if filepath.IsAbs(path) {
		// return absolute path unchanged
		return path, nil
	}
	// use relative path as-is if no base dir
	if baseDir == "" {
		// return relative path unchanged
		return path, nil
	}

	resolvedPath := filepath.Clean(filepath.Join(baseDir, path))
	cleanBase := filepath.Clean(baseDir)

	// Security check: ensure path doesn't escape base directory.
	// validate path is within base directory
	if !strings.HasPrefix(resolvedPath, cleanBase+string(os.PathSeparator)) && resolvedPath != cleanBase {
		// return error for path escape attempt
		return "", fmt.Errorf("path %q escapes base directory %q: %w", path, baseDir, ErrPathEscapesBase)
	}
	// return validated resolved path
	return resolvedPath, nil
}

// BuildLoggerWithoutConsole creates a MultiLogger from configuration
// but excludes console writers. Used for interactive TUI mode where
// console output would pollute the display.
//
// Params:
//   - cfg: the daemon logging configuration.
//   - baseDir: the base directory for log files.
//
// Returns:
//   - logging.Logger: the created logger (without console writers).
//   - error: nil on success, error on failure.
func BuildLoggerWithoutConsole(cfg config.DaemonLogging, baseDir string) (logging.Logger, error) {
	var writers []logging.Writer

	// create writers excluding console type
	for i := range cfg.Writers {
		wcfg := &cfg.Writers[i]
		// Skip console writers in interactive mode.
		// skip console writers for TUI mode
		if wcfg.Type == writerTypeConsole {
			continue
		}

		w, err := buildWriter(*wcfg, baseDir)
		// handle writer creation failure
		if err != nil {
			// clean up already created writers
			for _, created := range writers {
				_ = created.Close()
			}
			// wrap error with context
			return nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
		}

		level, err := logging.ParseLevel(wcfg.Level)
		// use default level if parsing fails
		if err != nil {
			level = logging.LevelInfo
		}

		writers = append(writers, WithLevelFilter(w, level))
	}
	// create logger with non-console writers
	return New(writers...), nil
}

// DefaultLogger creates a logger with default console output.
// This is a convenience function for when no configuration is available.
//
// Returns:
//   - logging.Logger: the default console logger.
func DefaultLogger() logging.Logger {
	writer := WithLevelFilter(NewConsoleWriter(), logging.LevelInfo)
	// create logger with default console writer
	return New(writer)
}

// NewSilentLogger creates a logger with no output.
// Used for interactive mode when file logging is not configured.
//
// Returns:
//   - logging.Logger: a logger that discards all output.
func NewSilentLogger() logging.Logger {
	return New() // Empty MultiLogger with no writers.
}

// BuildLoggerWithBufferedConsole creates a MultiLogger with a buffered console writer.
// The buffered console holds logs until Flush() is called, allowing the MOTD banner
// to be displayed first before any log output.
//
// Params:
//   - cfg: the daemon logging configuration.
//   - baseDir: the base directory for log files.
//
// Returns:
//   - logging.Logger: the created logger.
//   - *BufferedWriter: the buffered console writer (nil if no console configured).
//   - error: nil on success, error on failure.
func BuildLoggerWithBufferedConsole(cfg config.DaemonLogging, baseDir string) (logging.Logger, *BufferedWriter, error) {
	// use default config if no writers specified
	if len(cfg.Writers) == 0 {
		cfg = config.DefaultDaemonLogging()
	}

	var writers []logging.Writer
	var bufferedConsole *BufferedWriter

	// create writer for each configuration
	for i := range cfg.Writers {
		wcfg := &cfg.Writers[i]
		var w logging.Writer
		var err error

		// Handle console writer specially for buffering.
		// create buffered console writer if console type
		if wcfg.Type == writerTypeConsole {
			// create buffered console writer once
			if bufferedConsole == nil {
				consoleWriter := NewConsoleWriter()
				bufferedConsole = NewBufferedWriter(consoleWriter)
			}
			w = bufferedConsole
		} else {
			// create non-console writer
			w, err = buildWriter(*wcfg, baseDir)
			// handle writer creation failure
			if err != nil {
				// clean up already created writers
				for _, created := range writers {
					_ = created.Close()
				}
				// wrap error with context
				return nil, nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
			}
		}

		level, err := logging.ParseLevel(wcfg.Level)
		// use default level if parsing fails
		if err != nil {
			level = logging.LevelInfo
		}

		writers = append(writers, WithLevelFilter(w, level))
	}
	// create logger with buffered console
	return New(writers...), bufferedConsole, nil
}
