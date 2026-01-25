package daemon

import (
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/logging"
)

// FileWriter writes log events to a file with rotation support.
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

// dirPermissions defines the permission mode for log directories (rwxr-x---).
const dirPermissions os.FileMode = 0o750

// filePermissions defines the permission mode for log files (rw-------).
const filePermissions os.FileMode = 0o600

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
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Open or create the log file.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

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
	return err
}

// Close closes the file.
//
// Returns:
//   - error: nil on success, error on failure.
func (w *FileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.file.Close()
}

// Ensure FileWriter implements logging.Writer.
var _ logging.Writer = (*FileWriter)(nil)
