// Package logging provides log management with rotation and capture.
// It handles automatic file rotation based on size and file count.
package logging

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kodflow/daemon/internal/domain/config"
	"github.com/kodflow/daemon/internal/domain/shared"
)

// Writer constants.
const (
	// dirPermissions defines the permission mode for log directories (rwxr-x---).
	dirPermissions os.FileMode = 0o750
	// filePermissions defines the permission mode for log files (rw-------).
	filePermissions os.FileMode = 0o600
	// logFileFlags defines the standard flags for opening log files.
	logFileFlags int = os.O_APPEND | os.O_CREATE | os.O_WRONLY
	// defaultMaxSize defines the default maximum log file size (100MB).
	defaultMaxSize int64 = 100 * 1024 * 1024 // 100MB
	// defaultMaxFilesBackup defines the default number of backup files to keep.
	defaultMaxFilesBackup int = 5
	// timestampSeparatorLen defines the length of the space separator after timestamps.
	timestampSeparatorLen int = 1
	// firstBackupIndex defines the starting index for backup file rotation.
	firstBackupIndex int = 1
	// firstBackupSuffix defines the file extension suffix for the first backup file.
	firstBackupSuffix string = ".1"
)

// Writer is a log writer with optional rotation.
// It provides thread-safe writing with automatic file rotation based on size.
type Writer struct {
	// mu protects concurrent access to the writer.
	mu sync.Mutex
	// file is the underlying file handle for the log file.
	file *os.File
	// writer is the buffered writer wrapping the file.
	writer *bufio.Writer
	// path is the absolute path to the log file.
	path string
	// maxSize is the maximum file size before rotation in bytes.
	maxSize int64
	// maxFiles is the maximum number of rotated backup files to keep.
	maxFiles int
	// compress indicates whether to compress rotated files.
	compress bool
	// size is the current size of the log file in bytes.
	size int64

	// timestampFormat is the format string for timestamps.
	timestampFormat string
	// addTimestamp indicates whether to prepend timestamps to log entries.
	addTimestamp bool
}

// writerConfig defines the interface for log stream configuration.
// It provides access to rotation settings and timestamp format.
type writerConfig interface {
	// File returns the file path for the output stream.
	File() string
	// TimestampFormat returns the timestamp format for log entries.
	TimestampFormat() string
	// Rotation returns the rotation configuration.
	Rotation() config.RotationConfig
}

// NewWriterFromConfig creates a new log writer from a generic config interface.
// It is used when creating writers from interface-based configuration.
//
// Params:
//   - path: path to the log file
//   - cfg: output stream configuration interface
//
// Returns:
//   - *Writer: new writer instance
//   - error: nil on success, error on failure
func NewWriterFromConfig(path string, cfg writerConfig) (*Writer, error) {
	// SECURITY: Directory permissions 0o750 (rwxr-x---) are intentional:
	// - Owner (daemon process): full access for log management and rotation
	// - Group (admin/ops): read+execute for log inspection and aggregation
	// - Other: no access for confidentiality of service logs
	// This is more restrictive than typical 0o755 used by syslog/logrotate.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	// create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		// propagate mkdir error to caller
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	file, size, err := openLogFile(path)
	// handle file open failure
	if err != nil {
		// propagate open error to caller
		return nil, err
	}

	rotation := cfg.Rotation()
	maxSize := parseMaxSize(rotation.MaxSize)
	maxFiles := rotation.MaxFiles
	// use default max files if not set
	if maxFiles == 0 {
		maxFiles = defaultMaxFilesBackup
	}

	timestampFormat := cfg.TimestampFormat()

	// return configured writer with rotation
	return &Writer{
		file:            file,
		writer:          bufio.NewWriter(file),
		path:            path,
		maxSize:         maxSize,
		maxFiles:        maxFiles,
		compress:        rotation.Compress,
		size:            size,
		timestampFormat: timestampFormat,
		addTimestamp:    timestampFormat != "",
	}, nil
}

// NewWriter creates a new log writer with optional rotation support.
//
// Params:
//   - path: path to the log file
//   - cfg: log stream configuration
//
// Returns:
//   - *Writer: new writer instance
//   - error: nil on success, error on failure
func NewWriter(path string, cfg writerConfig) (*Writer, error) {
	// SECURITY: Directory permissions 0o750 (rwxr-x---) are intentional:
	// - Owner (daemon process): full access for log management and rotation
	// - Group (admin/ops): read+execute for log inspection and aggregation
	// - Other: no access for confidentiality of service logs
	// This is more restrictive than typical 0o755 used by syslog/logrotate.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	// create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		// propagate mkdir error to caller
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	file, size, err := openLogFile(path)
	// handle file open failure
	if err != nil {
		// propagate open error to caller
		return nil, err
	}

	rotation := cfg.Rotation()
	maxSize := parseMaxSize(rotation.MaxSize)
	timestampFormat := cfg.TimestampFormat()

	// return configured writer with rotation
	return &Writer{
		file:            file,
		writer:          bufio.NewWriter(file),
		path:            path,
		maxSize:         maxSize,
		maxFiles:        rotation.MaxFiles,
		compress:        rotation.Compress,
		size:            size,
		timestampFormat: timestampFormat,
		addTimestamp:    timestampFormat != "",
	}, nil
}

// openLogFile opens or creates a log file and returns its current size.
//
// Params:
//   - path: path to the log file
//
// Returns:
//   - *os.File: the opened file handle
//   - int64: the current file size
//   - error: nil on success, error on failure
func openLogFile(path string) (*os.File, int64, error) {
	opener := newFileOpener(path)
	f, err := opener.open()
	// handle file open failure
	if err != nil {
		// propagate open error to caller
		return nil, 0, fmt.Errorf("opening log file: %w", err)
	}

	info, err := f.Stat()
	// handle stat failure
	if err != nil {
		_ = f.Close()
		// propagate stat error after cleanup
		return nil, 0, fmt.Errorf("getting file info: %w", err)
	}

	// return opened file with current size
	return f, info.Size(), nil
}

// parseMaxSize parses the max size string and returns bytes.
//
// Params:
//   - sizeStr: size string to parse (e.g., "100MB")
//
// Returns:
//   - int64: parsed size in bytes or default on failure
func parseMaxSize(sizeStr string) int64 {
	maxSize, err := shared.ParseSize(sizeStr)
	// use default size if parsing fails
	if err != nil {
		// fallback to default max size
		return defaultMaxSize
	}

	// return parsed size in bytes
	return maxSize
}

// Write implements io.Writer.
//
// Params:
//   - p: bytes to write
//
// Returns:
//   - int: number of bytes written
//   - error: nil on success, error on failure
func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// rotate if size limit exceeded
	if w.maxSize > 0 && w.size+int64(len(p)) > w.maxSize {
		// perform log rotation
		if err := w.rotate(); err != nil {
			// propagate rotation error to caller
			return 0, fmt.Errorf("rotating log: %w", err)
		}
	}

	// add timestamp prefix if configured
	if w.addTimestamp {
		ts := FormatTimestamp(time.Now(), w.timestampFormat)
		// write timestamp prefix
		if _, err := w.writer.WriteString(ts + " "); err != nil {
			// propagate timestamp write error
			return 0, err
		}
		w.size += int64(len(ts) + timestampSeparatorLen)
	}

	n, err = w.writer.Write(p)
	// handle write failure
	if err != nil {
		// propagate write error to caller
		return n, err
	}
	w.size += int64(n)

	// flush buffer to ensure data is written
	if err := w.writer.Flush(); err != nil {
		// propagate flush error to caller
		return n, err
	}

	// return bytes written successfully
	return n, nil
}

// rotate rotates the log file by closing the current file,
// shifting backup files, and opening a new log file.
//
// Returns:
//   - error: nil on success, error on failure
func (w *Writer) rotate() error {
	// flush buffer before closing
	if err := w.writer.Flush(); err != nil {
		// propagate flush error to caller
		return err
	}

	// close current file
	if err := w.file.Close(); err != nil {
		// propagate close error to caller
		return err
	}

	// shift backup files
	if err := w.rotateFiles(); err != nil {
		// propagate rotate error to caller
		return err
	}

	file, err := w.openNewFile()
	// handle new file creation failure
	if err != nil {
		// propagate open error to caller
		return err
	}

	w.file = file
	w.writer = bufio.NewWriter(file)
	w.size = 0

	// return success after rotation
	return nil
}

// openNewFile creates and opens a new log file.
//
// Returns:
//   - *os.File: the opened file handle
//   - error: nil on success, error on failure
func (w *Writer) openNewFile() (*os.File, error) {
	opener := newFileOpener(w.path)
	f, err := opener.open()
	// handle file open failure
	if err != nil {
		// propagate open error to caller
		return nil, err
	}

	// return opened file handle
	return f, nil
}

// rotateFiles rotates the backup files by shifting existing backups
// and renaming the current log file to .1 extension.
//
// Returns:
//   - error: nil on success, error on failure
func (w *Writer) rotateFiles() error {
	oldest := fmt.Sprintf("%s.%d", w.path, w.maxFiles)
	_ = os.Remove(oldest)

	// Shift backup files from oldest to newest.
	// shift numbered backups
	for i := w.maxFiles - firstBackupIndex; i >= firstBackupIndex; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i)
		newPath := fmt.Sprintf("%s.%d", w.path, i+firstBackupIndex)
		_ = os.Rename(oldPath, newPath)
	}

	// rename current log to .1
	if err := os.Rename(w.path, w.path+firstBackupSuffix); err != nil && !os.IsNotExist(err) {
		// propagate rename error to caller
		return err
	}

	// return success after rotation
	return nil
}

// Close closes the log writer by flushing the buffer
// and closing the underlying file handle.
//
// Returns:
//   - error: nil on success, error on failure
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// flush buffer before closing
	if err := w.writer.Flush(); err != nil {
		// propagate flush error to caller
		return err
	}

	// close file and return result
	return w.file.Close()
}

// Sync flushes the buffer and synchronizes the file to disk
// ensuring all data is persisted to storage.
//
// Returns:
//   - error: nil on success, error on failure
func (w *Writer) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// flush buffer to file
	if err := w.writer.Flush(); err != nil {
		// propagate flush error to caller
		return err
	}

	// sync file to disk and return result
	return w.file.Sync()
}

// Path returns the absolute path to the log file.
//
// Returns:
//   - string: the file path
func (w *Writer) Path() string {
	// return log file path
	return w.path
}

// Size returns the current size of the log file in bytes.
//
// Returns:
//   - int64: the current file size in bytes
func (w *Writer) Size() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()

	// return current file size
	return w.size
}
