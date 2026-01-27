// Package daemon provides daemon event logging infrastructure.
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
	writerTypeConsole string = "console"
	writerTypeFile    string = "file"
	writerTypeJSON    string = "json"
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

	// Create and configure each writer from config.
	for _, wcfg := range cfg.Writers {
		// Build the writer based on type.
		w, err := buildWriter(wcfg, baseDir)
		// Handle writer creation errors.
		if err != nil {
			// Close any already created writers.
			for _, created := range writers {
				_ = created.Close()
			}
			// Return error with writer type context.
			return nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
		}

		// Parse level for filtering.
		level, err := logging.ParseLevel(wcfg.Level)
		// Use default level if parsing fails.
		if err != nil {
			// Default to INFO if level parsing fails.
			level = logging.LevelInfo
		}

		// Wrap with level filter.
		writers = append(writers, WithLevelFilter(w, level))
	}

	// Return logger with all configured writers.
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
	// Route to appropriate writer constructor based on type.
	switch wcfg.Type {
	// Console writer for stdout/stderr.
	case writerTypeConsole:
		// Create console writer with defaults.
		return NewConsoleWriter(), nil

	// File writer with rotation support.
	case writerTypeFile:
		path := wcfg.File.Path
		// Validate path is configured.
		if path == "" {
			// Missing required path configuration.
			return nil, fmt.Errorf("file writer requires File.Path configuration")
		}
		// Resolve path relative to base directory.
		resolvedPath, err := resolvePath(path, baseDir)
		// Handle path resolution errors.
		if err != nil {
			// Path resolution failed.
			return nil, fmt.Errorf("file writer: %w", err)
		}
		// Create file writer with rotation.
		return NewFileWriter(resolvedPath, wcfg.File.Rotation)

	// JSON writer for structured logging.
	case writerTypeJSON:
		path := wcfg.JSON.Path
		// Validate path is configured.
		if path == "" {
			// Missing required path configuration.
			return nil, fmt.Errorf("json writer requires JSON.Path configuration")
		}
		// Resolve path relative to base directory.
		resolvedPath, err := resolvePath(path, baseDir)
		// Handle path resolution errors.
		if err != nil {
			// Path resolution failed.
			return nil, fmt.Errorf("json writer: %w", err)
		}
		// Create JSON writer with rotation.
		return NewJSONWriter(resolvedPath, wcfg.JSON.Rotation)

	// Unknown writer type.
	default:
		// Invalid writer type in configuration.
		return nil, fmt.Errorf("unknown writer type: %s", wcfg.Type)
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
	// Absolute paths are used as-is.
	if filepath.IsAbs(path) {
		// Return absolute path unchanged.
		return path, nil
	}
	// No base directory means relative paths are used as-is.
	if baseDir == "" {
		// Return relative path unchanged.
		return path, nil
	}

	// Resolve the path and ensure it doesn't escape the base directory.
	resolvedPath := filepath.Clean(filepath.Join(baseDir, path))
	cleanBase := filepath.Clean(baseDir)

	// Check that resolved path is within base directory.
	if !strings.HasPrefix(resolvedPath, cleanBase+string(os.PathSeparator)) && resolvedPath != cleanBase {
		// Path escapes base directory (security violation).
		return "", fmt.Errorf("path %q escapes base directory %q", path, baseDir)
	}

	// Return validated path.
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

	// Create writers excluding console type.
	for _, wcfg := range cfg.Writers {
		// Skip console writers in interactive mode.
		if wcfg.Type == writerTypeConsole {
			continue
		}

		// Build non-console writer.
		w, err := buildWriter(wcfg, baseDir)
		// Handle writer creation errors.
		if err != nil {
			// Clean up on error.
			for _, created := range writers {
				_ = created.Close()
			}
			// Return error with context.
			return nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
		}

		// Parse level for filtering.
		level, err := logging.ParseLevel(wcfg.Level)
		// Use default level if parsing fails.
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
	// Return logger with default console writer.
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

	// Create writers with buffered console.
	for _, wcfg := range cfg.Writers {
		var w logging.Writer
		var err error

		// Handle console writer specially.
		if wcfg.Type == writerTypeConsole {
			// Reuse buffered console if already created (handles multiple console entries).
			if bufferedConsole == nil {
				consoleWriter := NewConsoleWriter()
				bufferedConsole = NewBufferedWriter(consoleWriter)
			}
			w = bufferedConsole
		} else {
			// Create non-console writer.
			w, err = buildWriter(wcfg, baseDir)
			// Handle writer creation errors.
			if err != nil {
				// Close any already created writers.
				for _, created := range writers {
					_ = created.Close()
				}
				// Return error with context.
				return nil, nil, fmt.Errorf("building writer %s: %w", wcfg.Type, err)
			}
		}

		// Parse level for filtering.
		level, err := logging.ParseLevel(wcfg.Level)
		// Use default level if parsing fails.
		if err != nil {
			// Default to INFO if level parsing fails.
			level = logging.LevelInfo
		}

		// Wrap with level filter.
		writers = append(writers, WithLevelFilter(w, level))
	}

	// Return logger with buffered console.
	return New(writers...), bufferedConsole, nil
}
