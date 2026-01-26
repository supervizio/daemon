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

// JSONLogEntry is the JSON structure for log events.
type JSONLogEntry struct {
	Timestamp string         `json:"ts"`
	Level     string         `json:"level"`
	Service   string         `json:"service,omitempty"`
	Event     string         `json:"event"`
	Message   string         `json:"message,omitempty"`
	Metadata  map[string]any `json:",inline"`
}

// JSONWriter writes log events as JSON lines to a file.
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
func NewJSONWriter(path string, rotation config.RotationConfig) (*JSONWriter, error) {
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

	// Build the JSON entry with metadata flattened.
	// Pre-allocate for base fields (5) + metadata.
	entry := make(map[string]any, 5+len(event.Metadata))
	entry["ts"] = event.Timestamp.Format("2006-01-02T15:04:05Z07:00")
	entry["level"] = event.Level.String()
	if event.Service != "" {
		entry["service"] = event.Service
	}
	entry["event"] = event.EventType
	if event.Message != "" {
		entry["message"] = event.Message
	}

	// Flatten metadata into the entry.
	for k, v := range event.Metadata {
		entry[k] = v
	}

	return w.encoder.Encode(entry)
}

// Close closes the file.
//
// Returns:
//   - error: nil on success, error on failure.
func (w *JSONWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	return w.file.Close()
}

// Ensure JSONWriter implements logging.Writer.
var _ logging.Writer = (*JSONWriter)(nil)
