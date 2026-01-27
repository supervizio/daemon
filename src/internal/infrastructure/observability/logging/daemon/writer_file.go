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
	// dirPermissions defines the permission mode for log directories (rwxr-x---).
	dirPermissions os.FileMode = 0o750
	// filePermissions defines the permission mode for log files (rw-------).
	filePermissions os.FileMode = 0o600
)

// FileWriter writes log events to a file with rotation support.
// Writes are protected by a mutex for concurrent access safety.
type FileWriter struct {
	// mu protects concurrent writes.
	mu sync.Mutex
	// file is the underlying file handle.
	file *os.File
	// path is the file path.
	path string
	// format is the text formatter.
	format Formatter
	// rotation is the rotation configuration.
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
func NewFileWriter(path string, rotation config.RotationConfig) (*FileWriter, error) {
	// Create directory if it doesn't exist.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		// Failed to create directory.
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Open or create the log file.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	// Failed to open file.
	if err != nil {
		// Failed to open log file.
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	// Return file writer with rotation config.
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

	// Format the event.
	line := w.format.Format(event)

	// Write with newline.
	_, err := w.file.WriteString(line + "\n")
	// Return write result.
	return err
}

// Close closes the file.
//
// Returns:
//   - error: nil on success, error on failure.
func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Close the file handle.
	return w.file.Close()
}

// Ensure FileWriter implements logging.Writer.
var _ logging.Writer = (*FileWriter)(nil)
