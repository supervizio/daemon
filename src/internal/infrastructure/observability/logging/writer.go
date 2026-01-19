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
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		// Return directory creation error to caller.
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Open and initialize the log file.
	file, size, err := openLogFile(path)

	// Check if file opening succeeded.
	if err != nil {
		// Return file open error to caller.
		return nil, err
	}

	// Get rotation config from interface.
	rotation := cfg.Rotation()

	// Parse max size from config, using default on failure.
	maxSize := parseMaxSize(rotation.MaxSize)

	// Determine max files count from config.
	maxFiles := rotation.MaxFiles

	// Check if max files is not configured (zero), use default.
	if maxFiles == 0 {
		// Use default max files backup count when config is zero.
		maxFiles = defaultMaxFilesBackup
	}

	// Get timestamp format from interface.
	timestampFormat := cfg.TimestampFormat()

	// Return the initialized writer with rotation settings from config.
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
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		// Directory creation failed - return error to caller.
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Open and initialize the log file.
	file, size, err := openLogFile(path)

	// Check if file opening succeeded.
	if err != nil {
		// File open failed - return error to caller.
		return nil, err
	}

	// Get rotation config from struct.
	rotation := cfg.Rotation()

	// Parse max size from config, using default on failure.
	maxSize := parseMaxSize(rotation.MaxSize)

	// Get timestamp format from struct.
	timestampFormat := cfg.TimestampFormat()

	// Return the initialized writer.
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
	// Create file opener for the log file path.
	opener := newFileOpener(path)

	// Open the file using the opener.
	f, err := opener.open()

	// Check if file opening succeeded.
	if err != nil {
		// Return nil file and zero size when file open fails.
		return nil, 0, fmt.Errorf("opening log file: %w", err)
	}

	// Get current file size.
	info, err := f.Stat()

	// Check if file stat operation succeeded.
	if err != nil {
		// Close file on stat error before returning.
		_ = f.Close()
		// Return nil file and zero size when stat fails.
		return nil, 0, fmt.Errorf("getting file info: %w", err)
	}

	// Return opened file and its current size on success.
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
	// Parse max size from config string.
	maxSize, err := shared.ParseSize(sizeStr)

	// Check if size parsing succeeded, use default on failure.
	if err != nil {
		// Return default on parse failure.
		return defaultMaxSize
	}

	// Return parsed size.
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

	// Defer unlocking the mutex to ensure it is released on function exit.
	defer w.mu.Unlock()

	// Check if rotation is needed based on max size and incoming data.
	if w.maxSize > 0 && w.size+int64(len(p)) > w.maxSize {
		// Check if rotation operation succeeds.
		if err := w.rotate(); err != nil {
			// Rotation failed - return error to caller.
			return 0, fmt.Errorf("rotating log: %w", err)
		}
	}

	// Check if timestamp should be added to the log entry.
	if w.addTimestamp {
		ts := FormatTimestamp(time.Now(), w.timestampFormat)

		// Check if timestamp write operation succeeds.
		if _, err := w.writer.WriteString(ts + " "); err != nil {
			// Timestamp write failed - return error to caller.
			return 0, err
		}
		w.size += int64(len(ts) + timestampSeparatorLen)
	}

	// Write data
	n, err = w.writer.Write(p)

	// Check if write operation succeeded.
	if err != nil {
		// Write failed - return bytes written and error.
		return n, err
	}
	w.size += int64(n)

	// Check if flush operation succeeds to ensure logs are persisted.
	if err := w.writer.Flush(); err != nil {
		// Flush failed - return bytes written and error.
		return n, err
	}

	// Return successful write count.
	return n, nil
}

// rotate rotates the log file by closing the current file,
// shifting backup files, and opening a new log file.
//
// Returns:
//   - error: nil on success, error on failure
func (w *Writer) rotate() error {
	// Check if buffer flush succeeds before closing the file.
	if err := w.writer.Flush(); err != nil {
		// Flush failed - return error to caller.
		return err
	}

	// Check if file close operation succeeds.
	if err := w.file.Close(); err != nil {
		// Close failed - return error to caller.
		return err
	}

	// Check if file rotation succeeds.
	if err := w.rotateFiles(); err != nil {
		// Rotation failed - return error to caller.
		return err
	}

	// Create new file using openNewFile helper.
	file, err := w.openNewFile()

	// Check if new file creation succeeded.
	if err != nil {
		// File creation failed - return error to caller.
		return err
	}

	w.file = file
	w.writer = bufio.NewWriter(file)
	w.size = 0

	// Return nil indicating success.
	return nil
}

// openNewFile creates and opens a new log file.
//
// Returns:
//   - *os.File: the opened file handle
//   - error: nil on success, error on failure
func (w *Writer) openNewFile() (*os.File, error) {
	// Create file opener for the writer path.
	opener := newFileOpener(w.path)

	// Open the file using the opener.
	f, err := opener.open()

	// Check if file opening succeeded.
	if err != nil {
		// Return nil file when open fails.
		return nil, err
	}

	// Return the opened file handle on success.
	return f, nil
}

// rotateFiles rotates the backup files by shifting existing backups
// and renaming the current log file to .1 extension.
//
// Returns:
//   - error: nil on success, error on failure
func (w *Writer) rotateFiles() error {
	// Remove oldest file if at max
	oldest := fmt.Sprintf("%s.%d", w.path, w.maxFiles)
	// Ignore error - file may not exist
	_ = os.Remove(oldest)

	// Iterate through existing backup files from oldest to newest to shift them.
	for i := w.maxFiles - firstBackupIndex; i >= firstBackupIndex; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i)
		newPath := fmt.Sprintf("%s.%d", w.path, i+firstBackupIndex)
		// Ignore error - file may not exist
		_ = os.Rename(oldPath, newPath)
	}

	// Check if renaming current file to first backup succeeds (ignore if file doesn't exist).
	if err := os.Rename(w.path, w.path+firstBackupSuffix); err != nil && !os.IsNotExist(err) {
		// Rename failed for existing file - return error to caller.
		return err
	}

	// Return nil indicating success.
	return nil
}

// Close closes the log writer by flushing the buffer
// and closing the underlying file handle.
//
// Returns:
//   - error: nil on success, error on failure
func (w *Writer) Close() error {
	w.mu.Lock()

	// Defer unlocking the mutex to ensure it is released on function exit.
	defer w.mu.Unlock()

	// Check if buffer flush succeeds before closing.
	if err := w.writer.Flush(); err != nil {
		// Flush failed - return error to caller.
		return err
	}

	// Close file and return result.
	return w.file.Close()
}

// Sync flushes the buffer and synchronizes the file to disk
// ensuring all data is persisted to storage.
//
// Returns:
//   - error: nil on success, error on failure
func (w *Writer) Sync() error {
	w.mu.Lock()

	// Defer unlocking the mutex to ensure it is released on function exit.
	defer w.mu.Unlock()

	// Check if buffer flush succeeds before syncing.
	if err := w.writer.Flush(); err != nil {
		// Flush failed - return error to caller.
		return err
	}

	// Sync file and return result.
	return w.file.Sync()
}

// Path returns the absolute path to the log file.
//
// Returns:
//   - string: the file path
func (w *Writer) Path() string {
	// Return the stored path value.
	return w.path
}

// Size returns the current size of the log file in bytes.
//
// Returns:
//   - int64: the current file size in bytes
func (w *Writer) Size() int64 {
	w.mu.Lock()

	// Defer unlocking the mutex to ensure it is released on function exit.
	defer w.mu.Unlock()

	// Return the current size value.
	return w.size
}
