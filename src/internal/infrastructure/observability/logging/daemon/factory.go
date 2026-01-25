package daemon

import (
	"fmt"
	"path/filepath"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/logging"
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
	case "console":
		return NewConsoleWriter(), nil

	case "file":
		path := wcfg.File.Path
		if !filepath.IsAbs(path) && baseDir != "" {
			path = filepath.Join(baseDir, path)
		}
		return NewFileWriter(path, wcfg.File.Rotation)

	case "json":
		path := wcfg.JSON.Path
		if !filepath.IsAbs(path) && baseDir != "" {
			path = filepath.Join(baseDir, path)
		}
		return NewJSONWriter(path, wcfg.JSON.Rotation)

	default:
		return nil, fmt.Errorf("unknown writer type: %s", wcfg.Type)
	}
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
