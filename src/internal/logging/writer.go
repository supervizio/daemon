// Package logging provides log writing with rotation for daemon services.
package logging

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/kodflow/daemon/internal/config"
)

// Writer is a log writer with optional rotation.
type Writer struct {
	mu       sync.Mutex
	file     *os.File
	writer   *bufio.Writer
	path     string
	maxSize  int64
	maxFiles int
	compress bool
	size     int64

	timestampFormat string
	addTimestamp    bool
}

// NewWriter creates a new log writer.
func NewWriter(path string, cfg *config.LogStreamConfig) (*Writer, error) {
	// Ensure directory exists
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	// Get current size
	info, err := file.Stat()
	if err != nil {
		file.Close()
		return nil, fmt.Errorf("getting file info: %w", err)
	}

	maxSize, err := config.ParseSize(cfg.Rotation.MaxSize)
	if err != nil {
		maxSize = 100 * 1024 * 1024 // Default 100MB
	}

	w := &Writer{
		file:            file,
		writer:          bufio.NewWriter(file),
		path:            path,
		maxSize:         maxSize,
		maxFiles:        cfg.Rotation.MaxFiles,
		compress:        cfg.Rotation.Compress,
		size:            info.Size(),
		timestampFormat: cfg.TimestampFormat,
		addTimestamp:    cfg.TimestampFormat != "",
	}

	return w, nil
}

// Write implements io.Writer.
func (w *Writer) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Check if rotation needed
	if w.maxSize > 0 && w.size+int64(len(p)) > w.maxSize {
		if err := w.rotate(); err != nil {
			return 0, fmt.Errorf("rotating log: %w", err)
		}
	}

	// Add timestamp if configured
	if w.addTimestamp {
		ts := FormatTimestamp(time.Now(), w.timestampFormat)
		if _, err := w.writer.WriteString(ts + " "); err != nil {
			return 0, err
		}
		w.size += int64(len(ts) + 1)
	}

	n, err = w.writer.Write(p)
	if err != nil {
		return n, err
	}
	w.size += int64(n)

	// Flush for each write to ensure logs are persisted
	if err := w.writer.Flush(); err != nil {
		return n, err
	}

	return n, nil
}

// rotate rotates the log file.
func (w *Writer) rotate() error {
	// Flush and close current file
	if err := w.writer.Flush(); err != nil {
		return err
	}
	if err := w.file.Close(); err != nil {
		return err
	}

	// Rotate existing files
	if err := w.rotateFiles(); err != nil {
		return err
	}

	// Create new file
	file, err := os.OpenFile(w.path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err
	}

	w.file = file
	w.writer = bufio.NewWriter(file)
	w.size = 0

	return nil
}

// rotateFiles rotates the backup files.
func (w *Writer) rotateFiles() error {
	// Remove oldest file if at max
	oldest := fmt.Sprintf("%s.%d", w.path, w.maxFiles)
	os.Remove(oldest)

	// Rotate existing files
	for i := w.maxFiles - 1; i >= 1; i-- {
		oldPath := fmt.Sprintf("%s.%d", w.path, i)
		newPath := fmt.Sprintf("%s.%d", w.path, i+1)
		os.Rename(oldPath, newPath)
	}

	// Rename current to .1
	if err := os.Rename(w.path, w.path+".1"); err != nil && !os.IsNotExist(err) {
		return err
	}

	return nil
}

// Close closes the log writer.
func (w *Writer) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Close()
}

// Sync flushes the buffer to disk.
func (w *Writer) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if err := w.writer.Flush(); err != nil {
		return err
	}
	return w.file.Sync()
}

// Path returns the log file path.
func (w *Writer) Path() string {
	return w.path
}

// Size returns the current log file size.
func (w *Writer) Size() int64 {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.size
}

// MultiWriter writes to multiple writers.
type MultiWriter struct {
	writers []io.WriteCloser
}

// NewMultiWriter creates a writer that duplicates output to multiple writers.
func NewMultiWriter(writers ...io.WriteCloser) *MultiWriter {
	return &MultiWriter{writers: writers}
}

// Write writes to all writers.
func (mw *MultiWriter) Write(p []byte) (n int, err error) {
	for _, w := range mw.writers {
		n, err = w.Write(p)
		if err != nil {
			return n, err
		}
	}
	return len(p), nil
}

// Close closes all writers.
func (mw *MultiWriter) Close() error {
	var firstErr error
	for _, w := range mw.writers {
		if err := w.Close(); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}
