// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/logging"
)

// File permissions constants.
const (
	// Permission mode for log directories (rwxr-x---).
	dirPermissions os.FileMode = 0o750
	// Permission mode for log files (rw-------).
	filePermissions os.FileMode = 0o600
)

// FileWriter writes log events to a file with rotation support.
// Writes are protected by a mutex for concurrent access safety.
type FileWriter struct {
	mu       sync.Mutex
	file     *os.File
	path     string
	format   Formatter
	rotation config.RotationConfig
}

// NewFileWriter creates a new file writer with rotation support.
//
// Params:
//   - path: the file path.
//   - rotation: the rotation configuration.
//
// Returns:
//   - *FileWriter: the created file writer.
//   - error: nil on success, error on failure.
func NewFileWriter(path string, rotation config.RotationConfig) (fw *FileWriter, err error) {
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	// Create log directory with restricted permissions.
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		// Failed to create directory.
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	// Check for file open error.
	if err != nil {
		// Failed to open file.
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	// Ensure cleanup on panic or error.
	defer func() {
		// Close file if error occurred.
		if err != nil && file != nil {
			_ = file.Close()
		}
	}()

	// Return initialized file writer.
	return &FileWriter{
		file:     file,
		path:     path,
		format:   NewTextFormatter(""),
		rotation: rotation,
	}, nil
}

// Write writes a log event to the file.
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: nil on success, error on failure.
func (w *FileWriter) Write(event logging.LogEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	line := w.format.Format(event)
	_, err := w.file.WriteString(line + "\n")
	// Return write error.
	return err
}

// Close closes the file.
//
// Returns:
//   - error: nil on success, error on failure.
func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Close underlying file.
	return w.file.Close()
}

// Ensure FileWriter implements logging.Writer.
var _ logging.Writer = (*FileWriter)(nil)
