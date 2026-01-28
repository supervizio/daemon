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
	if len(cfg.Writers) == 0 {
		cfg = config.DefaultDaemonLogging()
	}

	var writers []logging.Writer

	for _, wcfg := range cfg.Writers {
		w, err := buildWriter(wcfg, baseDir)
		if err != nil {
			for _, created := range writers {
				_ = created.Close()
			}

			return nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
		}

		level, err := logging.ParseLevel(wcfg.Level)
		if err != nil {
			level = logging.LevelInfo
		}

		writers = append(writers, WithLevelFilter(w, level))
	}

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
	switch wcfg.Type {
	case writerTypeConsole:
		return NewConsoleWriter(), nil

	case writerTypeFile:
		path := wcfg.File.Path
		if path == "" {
			return nil, ErrFilePathRequired
		}
		resolvedPath, err := resolvePath(path, baseDir)
		if err != nil {
			return nil, fmt.Errorf("file writer: %w", err)
		}

		return NewFileWriter(resolvedPath, wcfg.File.Rotation)

	case writerTypeJSON:
		path := wcfg.JSON.Path
		if path == "" {
			return nil, ErrJSONPathRequired
		}
		resolvedPath, err := resolvePath(path, baseDir)
		if err != nil {
			return nil, fmt.Errorf("json writer: %w", err)
		}

		return NewJSONWriter(resolvedPath)

	default:
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
	if filepath.IsAbs(path) {
		return path, nil
	}
	if baseDir == "" {
		return path, nil
	}

	resolvedPath := filepath.Clean(filepath.Join(baseDir, path))
	cleanBase := filepath.Clean(baseDir)

	// Security check: ensure path doesn't escape base directory.
	if !strings.HasPrefix(resolvedPath, cleanBase+string(os.PathSeparator)) && resolvedPath != cleanBase {
		return "", fmt.Errorf("path %q escapes base directory %q: %w", path, baseDir, ErrPathEscapesBase)
	}

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

	for _, wcfg := range cfg.Writers {
		// Skip console writers in interactive mode.
		if wcfg.Type == writerTypeConsole {
			continue
		}

		w, err := buildWriter(wcfg, baseDir)
		if err != nil {
			for _, created := range writers {
				_ = created.Close()
			}

			return nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
		}

		level, err := logging.ParseLevel(wcfg.Level)
		if err != nil {
			level = logging.LevelInfo
		}

		writers = append(writers, WithLevelFilter(w, level))
	}

	return New(writers...), nil
}

// DefaultLogger creates a logger with default console output.
// This is a convenience function for when no configuration is available.
//
// Returns:
//   - logging.Logger: the default console logger.
func DefaultLogger() logging.Logger {
	writer := WithLevelFilter(NewConsoleWriter(), logging.LevelInfo)

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
	if len(cfg.Writers) == 0 {
		cfg = config.DefaultDaemonLogging()
	}

	var writers []logging.Writer
	var bufferedConsole *BufferedWriter

	for _, wcfg := range cfg.Writers {
		var w logging.Writer
		var err error

		// Handle console writer specially for buffering.
		if wcfg.Type == writerTypeConsole {
			if bufferedConsole == nil {
				consoleWriter := NewConsoleWriter()
				bufferedConsole = NewBufferedWriter(consoleWriter)
			}
			w = bufferedConsole
		} else {
			w, err = buildWriter(wcfg, baseDir)
			if err != nil {
				for _, created := range writers {
					_ = created.Close()
				}

				return nil, nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
			}
		}

		level, err := logging.ParseLevel(wcfg.Level)
		if err != nil {
			level = logging.LevelInfo
		}

		writers = append(writers, WithLevelFilter(w, level))
	}

	return New(writers...), bufferedConsole, nil
}
