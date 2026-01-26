package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/logging"
)

// Writer type constants.
const (
	writerTypeConsole = "console"
	writerTypeFile    = "file"
	writerTypeJSON    = "json"
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
	// If no config provided, use default (console only, INFO level).
	if len(cfg.Writers) == 0 {
		cfg = config.DefaultDaemonLogging()
	}

	var writers []logging.Writer

	for _, wcfg := range cfg.Writers {
		w, err := buildWriter(wcfg, baseDir)
		if err != nil {
			// Close any already created writers.
			for _, created := range writers {
				_ = created.Close()
			}
			return nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
		}

		// Parse level for filtering.
		level, err := logging.ParseLevel(wcfg.Level)
		if err != nil {
			// Default to INFO if level parsing fails.
			level = logging.LevelInfo
		}

		// Wrap with level filter.
		writers = append(writers, WithLevelFilter(w, level))
	}

	return New(writers...), nil
}

// buildWriter creates a writer based on configuration type.
func buildWriter(wcfg config.WriterConfig, baseDir string) (logging.Writer, error) {
	switch wcfg.Type {
	case writerTypeConsole:
		return NewConsoleWriter(), nil

	case writerTypeFile:
		path := wcfg.File.Path
		if path == "" {
			return nil, fmt.Errorf("file writer requires File.Path configuration")
		}
		resolvedPath, err := resolvePath(path, baseDir)
		if err != nil {
			return nil, fmt.Errorf("file writer: %w", err)
		}
		return NewFileWriter(resolvedPath, wcfg.File.Rotation)

	case writerTypeJSON:
		path := wcfg.JSON.Path
		if path == "" {
			return nil, fmt.Errorf("json writer requires JSON.Path configuration")
		}
		resolvedPath, err := resolvePath(path, baseDir)
		if err != nil {
			return nil, fmt.Errorf("json writer: %w", err)
		}
		return NewJSONWriter(resolvedPath, wcfg.JSON.Rotation)

	default:
		return nil, fmt.Errorf("unknown writer type: %s", wcfg.Type)
	}
}

// resolvePath resolves a path relative to baseDir and validates it doesn't escape.
func resolvePath(path, baseDir string) (string, error) {
	if filepath.IsAbs(path) {
		return path, nil
	}
	if baseDir == "" {
		return path, nil
	}

	// Resolve the path and ensure it doesn't escape the base directory.
	resolvedPath := filepath.Clean(filepath.Join(baseDir, path))
	cleanBase := filepath.Clean(baseDir)

	// Check that resolved path is within base directory.
	if !strings.HasPrefix(resolvedPath, cleanBase+string(os.PathSeparator)) && resolvedPath != cleanBase {
		return "", fmt.Errorf("path %q escapes base directory %q", path, baseDir)
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

	// Return empty MultiLogger if no writers (TUI writer will be added later).
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
	// If no config provided, use default (console only, INFO level).
	if len(cfg.Writers) == 0 {
		cfg = config.DefaultDaemonLogging()
	}

	var writers []logging.Writer
	var bufferedConsole *BufferedWriter

	for _, wcfg := range cfg.Writers {
		var w logging.Writer
		var err error

		if wcfg.Type == writerTypeConsole {
			// Reuse buffered console if already created (handles multiple console entries).
			if bufferedConsole == nil {
				consoleWriter := NewConsoleWriter()
				bufferedConsole = NewBufferedWriter(consoleWriter)
			}
			w = bufferedConsole
		} else {
			w, err = buildWriter(wcfg, baseDir)
			if err != nil {
				// Close any already created writers.
				for _, created := range writers {
					_ = created.Close()
				}
				return nil, nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
			}
		}

		// Parse level for filtering.
		level, err := logging.ParseLevel(wcfg.Level)
		if err != nil {
			// Default to INFO if level parsing fails.
			level = logging.LevelInfo
		}

		// Wrap with level filter.
		writers = append(writers, WithLevelFilter(w, level))
	}

	return New(writers...), bufferedConsole, nil
}
