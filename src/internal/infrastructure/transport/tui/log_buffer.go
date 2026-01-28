// Package tui provides terminal user interface rendering for superviz.io.
package tui

import (
	"sync"
	"time"

	domainlogging "github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// LogBuffer constants.
const (
	// defaultLogBufferSize is the default log buffer capacity.
	defaultLogBufferSize int = 100

	// logPeriod is the time period for log summaries.
	logPeriod time.Duration = 5 * time.Minute
)

// LogBuffer is a thread-safe ring buffer for log entries.
// Uses a proper ring buffer to avoid memory leaks from slice shifting.
type LogBuffer struct {
	mu         sync.RWMutex
	entries    []model.LogEntry
	head       int // Index of oldest entry (read position).
	tail       int // Index for next write.
	count      int // Number of entries in buffer.
	maxSize    int
	infoCount  int
	warnCount  int
	errorCount int
}

// NewLogBuffer creates a new log buffer with the specified capacity.
// Pre-allocates the full buffer to avoid allocations during Add.
//
// Params:
//   - maxSize: the maximum buffer capacity (default 100 if <= 0).
//
// Returns:
//   - *LogBuffer: the created buffer.
func NewLogBuffer(maxSize int) *LogBuffer {
	if maxSize <= 0 {
		maxSize = defaultLogBufferSize
	}
	// Pre-populate entries using append pattern for ring buffer.
	entries := make([]model.LogEntry, 0, maxSize)
	for range maxSize {
		entries = append(entries, model.LogEntry{})
	}
	return &LogBuffer{
		entries: entries,
		maxSize: maxSize,
		head:    0,
		tail:    0,
		count:   0,
	}
}

// Add adds a log entry to the buffer using ring buffer semantics.
// This avoids memory leaks from slice shifting.
//
// Params:
//   - entry: the log entry to add.
func (b *LogBuffer) Add(entry model.LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	switch entry.Level {
	case "INFO":
		b.infoCount++
	case "WARN", "WARNING":
		b.warnCount++
	case "ERROR", "ERR":
		b.errorCount++
	}

	b.entries[b.tail] = entry
	b.tail = (b.tail + 1) % b.maxSize

	if b.count < b.maxSize {
		b.count++
	} else {
		b.head = (b.head + 1) % b.maxSize
	}
}

// AddFromDomainEvent adds a domain LogEvent to the buffer.
//
// Params:
//   - event: the domain log event.
func (b *LogBuffer) AddFromDomainEvent(event domainlogging.LogEvent) {
	entry := model.LogEntry{
		Timestamp: event.Timestamp,
		Level:     event.Level.String(),
		Service:   event.Service,
		EventType: event.EventType,
		Message:   event.Message,
		Metadata:  event.Metadata,
	}
	b.Add(entry)
}

// Entries returns a copy of all entries in chronological order.
//
// Returns:
//   - []model.LogEntry: the log entries.
func (b *LogBuffer) Entries() []model.LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return b.entriesLocked()
}

// entriesLocked returns entries without acquiring lock (caller must hold lock).
//
// Returns:
//   - []model.LogEntry: the log entries in chronological order.
func (b *LogBuffer) entriesLocked() []model.LogEntry {
	if b.count == 0 {
		return nil
	}

	result := make([]model.LogEntry, 0, b.count)
	for i := range b.count {
		idx := (b.head + i) % b.maxSize
		result = append(result, b.entries[idx])
	}
	return result
}

// Summary returns the log summary.
// Uses entriesLocked() to avoid deadlock from re-acquiring RLock.
//
// Returns:
//   - model.LogSummary: the log summary.
func (b *LogBuffer) Summary() model.LogSummary {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return model.LogSummary{
		Period:        logPeriod,
		InfoCount:     b.infoCount,
		WarnCount:     b.warnCount,
		ErrorCount:    b.errorCount,
		RecentEntries: b.entriesLocked(), // Use internal method to avoid deadlock.
		HasAlerts:     b.errorCount > 0,
	}
}

// Clear resets the buffer without deallocating memory.
func (b *LogBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Reset ring buffer state (entries array stays allocated).
	b.head = 0
	b.tail = 0
	b.count = 0
	b.infoCount = 0
	b.warnCount = 0
	b.errorCount = 0

	// Clear entry references to allow GC of log data.
	for i := range b.entries {
		b.entries[i] = model.LogEntry{}
	}
}
