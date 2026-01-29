// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"encoding/json"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"sync"

	"github.com/kodflow/daemon/internal/domain/logging"
)

// JSON entry pool constants.
const (
	// jsonMapInitialCapacity is the pre-allocated capacity for JSON log entries.
	jsonMapInitialCapacity int = 16
)

var (
	// jsonMapPool provides reusable map[string]any instances to reduce allocations.
	// Maps are cleared before returning to pool.
	jsonMapPool sync.Pool = sync.Pool{
		New: func() any {
			return make(map[string]any, jsonMapInitialCapacity) // Pre-allocate for typical log entries
		},
	}

	// Ensure JSONWriter implements logging.Writer.
	_ logging.Writer = (*JSONWriter)(nil)
)

// JSONWriter writes log events as JSON lines to a file.
// Writes are protected by a mutex for concurrent access safety.
type JSONWriter struct {
	// mu protects concurrent writes.
	mu sync.Mutex
	// file is the underlying file handle.
	file *os.File
	// path is the file path.
	path string
	// encoder is the JSON encoder.
	encoder *json.Encoder
}

// NewJSONWriter creates a new JSON writer with rotation support.
//
// Params:
//   - path: the file path.
//
// Returns:
//   - *JSONWriter: the created JSON writer.
//   - error: nil on success, error on failure.
//
// Goroutine lifecycle: File handle is owned by JSONWriter struct.
// Cleanup: Caller must call Close() to release the file handle.
func NewJSONWriter(path string) (*JSONWriter, error) {
	// Create directory if it doesn't exist.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	// create parent directories if needed
	if err := os.MkdirAll(filepath.Dir(path), dirPermissions); err != nil {
		// Failed to create directory.
		return nil, fmt.Errorf("creating log directory: %w", err)
	}

	// Open or create the log file.
	// nosemgrep: go.lang.correctness.permissions.file_permission.incorrect-default-permission
	// File lifecycle: Opened here, ownership transferred to JSONWriter on success.
	// Cleanup via defer: On error, file is closed. On success, defer is disabled via nil assignment.
	file, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, filePermissions)
	// Failed to open file.
	// handle file open failure
	if err != nil {
		// Failed to open log file.
		return nil, fmt.Errorf("opening log file: %w", err)
	}

	// Defer cleanup for error paths - disabled on success by nil assignment.
	defer func() {
		// close file if not transferred to struct
		if file != nil {
			_ = file.Close()
		}
	}()

	// Build JSONWriter - file ownership transfers here.
	writer := &JSONWriter{
		file:    file,
		path:    path,
		encoder: json.NewEncoder(file),
	}

	// Disable deferred close - ownership successfully transferred.
	file = nil

	// Return JSON writer with encoder.
	return writer, nil
}

// Write writes a log event as a JSON line.
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: nil on success, error on failure.
func (w *JSONWriter) Write(event logging.LogEvent) error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Get a pooled map to reduce allocations in hot path.
	pooled := jsonMapPool.Get()
	entry, ok := pooled.(map[string]any)
	// Handle type assertion failure gracefully.
	// handle invalid type from pool
	if !ok {
		// Type assertion failed - should never happen but handle gracefully.
		entry = make(map[string]any, jsonMapInitialCapacity)
	}

	// Build the JSON entry with metadata flattened.
	entry["ts"] = event.Timestamp.Format("2006-01-02T15:04:05Z07:00")
	entry["level"] = event.Level.String()
	// Add service if present.
	// add service field if present
	if event.Service != "" {
		entry["service"] = event.Service
	}
	entry["event"] = event.EventType
	// Add message if present.
	// add message field if present
	if event.Message != "" {
		entry["message"] = event.Message
	}

	// Flatten metadata into the entry.
	maps.Copy(entry, event.Metadata)

	err := w.encoder.Encode(entry)

	// Clear and return map to pool.
	clear(entry)
	jsonMapPool.Put(entry)

	// Return encode result.
	return err
}

// Close closes the file.
//
// Returns:
//   - error: nil on success, error on failure.
func (w *JSONWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	// Close the file handle.
	return w.file.Close()
}
