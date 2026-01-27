// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/kodflow/daemon/internal/domain/config"
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

// JSONLogEntry is the JSON structure for log events.
// Metadata fields are inlined into the root of the JSON object.
type JSONLogEntry struct {
	Timestamp string         `json:"ts"`
	Level     string         `json:"level"`
	Service   string         `json:"service,omitempty"`
	Event     string         `json:"event"`
	Message   string         `json:"message,omitempty"`
	Metadata  map[string]any `json:",inline"`
}

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
//   - rotation: the rotation configuration (unused for now, future support).
//
// Returns:
//   - *JSONWriter: the created JSON writer.
//   - error: nil on success, error on failure.
func NewJSONWriter(path string, _ config.RotationConfig) (*JSONWriter, error) {
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

	// Return JSON writer with encoder.
	return &JSONWriter{
		file:    file,
		path:    path,
		encoder: json.NewEncoder(file),
	}, nil
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
	entry := jsonMapPool.Get().(map[string]any)

	// Build the JSON entry with metadata flattened.
	entry["ts"] = event.Timestamp.Format("2006-01-02T15:04:05Z07:00")
	entry["level"] = event.Level.String()
	// Add service if present.
	if event.Service != "" {
		entry["service"] = event.Service
	}
	entry["event"] = event.EventType
	// Add message if present.
	if event.Message != "" {
		entry["message"] = event.Message
	}

	// Flatten metadata into the entry.
	for k, v := range event.Metadata {
		entry[k] = v
	}

	err := w.encoder.Encode(entry)

	// Clear and return map to pool.
	for k := range entry {
		delete(entry, k)
	}
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
