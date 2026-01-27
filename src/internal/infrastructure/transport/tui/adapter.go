// Package tui provides terminal user interface rendering for superviz.io.
package tui

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"

	domainlogging "github.com/kodflow/daemon/internal/domain/logging"
	domainmetrics "github.com/kodflow/daemon/internal/domain/metrics"
	"github.com/kodflow/daemon/internal/domain/process"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// Buffer size constants.
const (
	// defaultLogBufferSize is the default log buffer capacity.
	defaultLogBufferSize int = 100
	// defaultMaxLines is the default number of lines to load from log files.
	defaultMaxLines int = 100
	// logPeriod is the time period for log summaries.
	logPeriod time.Duration = 5 * time.Minute
)

// Log level comparison constants.
const (
	// minRegexGroups is the minimum number of groups for a valid log line regex match.
	minRegexGroups int = 4

	// regexGroupLevel is the regex capture group index for log level.
	regexGroupLevel int = 2

	// regexGroupRemainder is the regex capture group index for log remainder.
	regexGroupRemainder int = 3

	// initialMetadataCapacity is the initial capacity for log entry metadata maps.
	initialMetadataCapacity int = 4
)

// TUISnapshotData contains service data for TUI display.
// It provides a minimal view of service state optimized for terminal rendering.
type TUISnapshotData struct {
	Name   string
	State  process.State
	PID    int
	Uptime int64
}

// TUISnapshotser provides service snapshots for TUI display.
type TUISnapshotser interface {
	// TUISnapshots returns service data for TUI display.
	TUISnapshots() []TUISnapshotData
}

// ProcessMetricsProvider provides process metrics from the metrics tracker.
// It wraps a metrics tracker to provide TUI-compatible process metrics.
type ProcessMetricsProvider interface {
	// Get returns metrics for a specific service.
	Get(serviceName string) (domainmetrics.ProcessMetrics, bool)
	// Has checks if metrics exist for a service.
	Has(serviceName string) bool
}

// DynamicServiceProvider queries the supervisor on each call.
// It bridges the supervisor's metric tracking with the TUI's display model.
type DynamicServiceProvider struct {
	provider TUISnapshotser
	metrics  ProcessMetricsProvider
}

// NewDynamicServiceProvider creates a new dynamic service provider.
//
// Params:
//   - provider: the TUI snapshots provider.
//   - metrics: the supervisor metrics tracker.
//
// Returns:
//   - *DynamicServiceProvider: the created provider.
func NewDynamicServiceProvider(provider TUISnapshotser, metrics ProcessMetricsProvider) *DynamicServiceProvider {
	// Create and return provider with injected dependencies.
	return &DynamicServiceProvider{
		provider: provider,
		metrics:  metrics,
	}
}

// Services implements ServiceProvider.
//
// Returns:
//   - []model.ServiceSnapshot: the service snapshots.
func (p *DynamicServiceProvider) Services() []model.ServiceSnapshot {
	// Return nil if no provider is available.
	if p.provider == nil {
		// No provider available.
		return nil
	}

	snapshots := p.provider.TUISnapshots()
	result := make([]model.ServiceSnapshot, 0, len(snapshots))

	// Convert each snapshot to TUI model.
	for _, snap := range snapshots {
		ss := model.ServiceSnapshot{
			Name:   snap.Name,
			State:  snap.State,
			PID:    snap.PID,
			Uptime: time.Duration(snap.Uptime) * time.Second,
		}

		// Add metrics if available.
		if p.metrics != nil {
			// Try to get metrics for this service.
			if m, ok := p.metrics.Get(snap.Name); ok {
				ss.CPUPercent = m.CPU.UsagePercent
				ss.MemoryRSS = m.Memory.RSS
			}
		}

		result = append(result, ss)
	}

	// Return collected snapshots.
	return result
}

// SystemMetricsAdapter provides system metrics.
// This is a placeholder that delegates to TUI collectors.
type SystemMetricsAdapter struct {
	// System metrics will be collected via collectors.
	// This is a placeholder that can be extended.
}

// NewSystemMetricsAdapter creates a new system metrics adapter.
//
// Returns:
//   - *SystemMetricsAdapter: the created adapter.
func NewSystemMetricsAdapter() *SystemMetricsAdapter {
	// Return new adapter instance.
	return &SystemMetricsAdapter{}
}

// SystemMetrics implements MetricsProvider.
//
// Returns:
//   - model.SystemMetrics: empty metrics (TUI collectors handle this).
func (a *SystemMetricsAdapter) SystemMetrics() model.SystemMetrics {
	// System metrics are collected by the TUI collectors.
	// This returns empty metrics; the TUI will use collectors instead.
	return model.SystemMetrics{}
}

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
	// Use default size if invalid.
	if maxSize <= 0 {
		maxSize = defaultLogBufferSize
	}
	// Return pre-allocated buffer.
	return &LogBuffer{
		entries: make([]model.LogEntry, maxSize), // Pre-allocate full capacity.
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

	// Update counts based on log level.
	switch entry.Level {
	// Handle INFO level entries.
	case "INFO":
		b.infoCount++
	// Handle WARN and WARNING level entries.
	case "WARN", "WARNING":
		b.warnCount++
	// Handle ERROR and ERR level entries.
	case "ERROR", "ERR":
		b.errorCount++
	}

	// Write to current tail position (overwrites oldest if full).
	b.entries[b.tail] = entry
	b.tail = (b.tail + 1) % b.maxSize

	// Update count and head position.
	if b.count < b.maxSize {
		// Buffer not yet full.
		b.count++
	} else {
		// Buffer is full, advance head (oldest entry overwritten).
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

	// Return entries with lock held.
	return b.entriesLocked()
}

// entriesLocked returns entries without acquiring lock (caller must hold lock).
//
// Returns:
//   - []model.LogEntry: the log entries in chronological order.
func (b *LogBuffer) entriesLocked() []model.LogEntry {
	// Return nil if buffer is empty.
	if b.count == 0 {
		// Empty buffer.
		return nil
	}

	result := make([]model.LogEntry, b.count)
	// Copy entries in chronological order.
	for i := range b.count {
		idx := (b.head + i) % b.maxSize
		result[i] = b.entries[idx]
	}
	// Return chronologically sorted entries.
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

	// Return summary with current state.
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

// LogAdapter provides log summary from a log buffer.
// It bridges the domain logging system with the TUI display.
type LogAdapter struct {
	buffer *LogBuffer
}

// NewLogAdapter creates a new log adapter with a default buffer size.
//
// Returns:
//   - *LogAdapter: the created adapter.
func NewLogAdapter() *LogAdapter {
	// Return adapter with default buffer.
	return &LogAdapter{
		buffer: NewLogBuffer(defaultLogBufferSize),
	}
}

// NewLogAdapterWithBuffer creates a new log adapter with a custom buffer.
//
// Params:
//   - buffer: the log buffer to use.
//
// Returns:
//   - *LogAdapter: the created adapter.
func NewLogAdapterWithBuffer(buffer *LogBuffer) *LogAdapter {
	// Return adapter with provided buffer.
	return &LogAdapter{
		buffer: buffer,
	}
}

// LogSummary implements HealthProvider.
//
// Returns:
//   - model.LogSummary: the log summary.
func (a *LogAdapter) LogSummary() model.LogSummary {
	// Return empty summary if no buffer is available.
	if a.buffer == nil {
		// No buffer available.
		return model.LogSummary{}
	}
	// Return buffer summary.
	return a.buffer.Summary()
}

// AddLog adds a log entry to the adapter.
//
// Params:
//   - entry: the log entry to add.
func (a *LogAdapter) AddLog(entry model.LogEntry) {
	// Add to buffer if available.
	if a.buffer != nil {
		a.buffer.Add(entry)
	}
}

// AddDomainEvent adds a domain log event to the adapter.
//
// Params:
//   - event: the domain log event.
func (a *LogAdapter) AddDomainEvent(event domainlogging.LogEvent) {
	// Add to buffer if available.
	if a.buffer != nil {
		a.buffer.AddFromDomainEvent(event)
	}
}

// Buffer returns the underlying buffer.
//
// Returns:
//   - *LogBuffer: the log buffer.
func (a *LogAdapter) Buffer() *LogBuffer {
	// Return buffer reference.
	return a.buffer
}

// TUILogWriter implements domain/logging.Writer to capture logs for TUI.
// It forwards log events to a LogAdapter for display.
type TUILogWriter struct {
	adapter *LogAdapter
}

// NewTUILogWriter creates a writer that sends logs to the TUI.
//
// Params:
//   - adapter: the log adapter to write to.
//
// Returns:
//   - *TUILogWriter: the created writer.
func NewTUILogWriter(adapter *LogAdapter) *TUILogWriter {
	// Return writer with adapter.
	return &TUILogWriter{
		adapter: adapter,
	}
}

// Write implements domain/logging.Writer.
//
// Params:
//   - event: the log event to write.
//
// Returns:
//   - error: always nil (errors are ignored).
func (w *TUILogWriter) Write(event domainlogging.LogEvent) error {
	// Write to adapter if available.
	if w.adapter != nil {
		w.adapter.AddDomainEvent(event)
	}
	// Always return success.
	return nil
}

// Close implements domain/logging.Writer.
//
// Returns:
//   - error: always nil (no cleanup needed).
func (w *TUILogWriter) Close() error {
	// No cleanup needed.
	return nil
}

// LoadLogHistory loads recent log entries from a log file into the adapter.
// It reads the last maxLines from the file and parses them into LogEntry structs.
//
// Params:
//   - path: the path to the log file.
//   - maxLines: maximum number of lines to load (default 100 if <= 0).
//
// Returns:
//   - error: nil on success, error on failure (file not found is not an error).
func (a *LogAdapter) LoadLogHistory(path string, maxLines int) error {
	// Return early if no buffer or path.
	if a.buffer == nil || path == "" {
		// Nothing to load.
		return nil
	}

	// Use default if invalid.
	if maxLines <= 0 {
		maxLines = defaultMaxLines
	}

	// Read last lines using helper function.
	lines, err := readLastLines(path, maxLines)
	// Check for file read errors.
	if err != nil {
		// Failed to read file.
		return err
	}

	// Parse and add entries.
	for _, line := range lines {
		// Try to parse line.
		if entry, ok := parseLogLine(line); ok {
			a.buffer.Add(entry)
		}
	}

	// Success.
	return nil
}

// logLineRegex parses log lines in format:
// 2006-01-02T15:04:05Z07:00 [LEVEL] service Message key=value
var logLineRegex *regexp.Regexp = regexp.MustCompile(
	`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[^\s]*)\s+\[([A-Z]+)\]\s+(.*)$`,
)

// parseLogLine parses a log line into a LogEntry.
//
// Params:
//   - line: the log line to parse.
//
// Returns:
//   - model.LogEntry: the parsed log entry.
//   - bool: true if parsing succeeded, false otherwise.
func parseLogLine(line string) (model.LogEntry, bool) {
	matches := logLineRegex.FindStringSubmatch(line)
	// Return early if line doesn't match format.
	if len(matches) < minRegexGroups {
		// Invalid format.
		return model.LogEntry{}, false
	}

	// Parse timestamp using helper function.
	ts, ok := parseLogTimestamp(matches[1])
	// Validate timestamp parsing result.
	if !ok {
		// Invalid timestamp.
		return model.LogEntry{}, false
	}

	level := matches[regexGroupLevel]
	remainder := matches[regexGroupRemainder]

	// Parse remainder: "service Message key=value key2=value2"
	// or just "Message key=value" if no service.
	entry := model.LogEntry{
		Timestamp: ts,
		Level:     level,
		Metadata:  make(map[string]any, initialMetadataCapacity),
	}

	// Parse remainder using helper function.
	parseLogRemainder(&entry, remainder)

	// Return parsed entry.
	return entry, true
}

// isServiceName checks if a string looks like a service name.
//
// Params:
//   - s: the string to check.
//
// Returns:
//   - bool: true if the string looks like a service name.
func isServiceName(s string) bool {
	// Service names are typically alphanumeric with dashes/underscores.
	// They don't start with uppercase words like "Service", "Daemon", etc.
	if len(s) == 0 {
		// Empty string is not a service name.
		return false
	}
	// Common message starters are not service names.
	commonStarters := []string{"Service", "Daemon", "Supervisor", "Failed", "Started", "Stopped"}
	// Check against known non-service names.
	for _, starter := range commonStarters {
		// Compare against each common message starter.
		if s == starter {
			// Matched a common starter.
			return false
		}
	}
	// Looks like a service name.
	return true
}

// parseLogTimestamp parses a timestamp string into a time.Time.
// Supports RFC3339 format with or without timezone.
//
// Params:
//   - s: the timestamp string to parse.
//
// Returns:
//   - time.Time: the parsed timestamp.
//   - bool: true if parsing succeeded, false otherwise.
func parseLogTimestamp(s string) (time.Time, bool) {
	ts, err := time.Parse(time.RFC3339, s)
	// Check for RFC3339 parse error.
	if err != nil {
		// Try alternative format without timezone.
		ts, err = time.Parse("2006-01-02T15:04:05Z", s)
		// Check for alternative format parse error.
		if err != nil {
			// Parse failed.
			return time.Time{}, false
		}
	}
	// Parse succeeded.
	return ts, true
}

// parseLogRemainder parses the remainder of a log line into a LogEntry.
// It extracts service name, message, and metadata key=value pairs.
//
// Params:
//   - entry: the log entry to populate.
//   - remainder: the remainder of the log line after timestamp and level.
func parseLogRemainder(entry *model.LogEntry, remainder string) {
	parts := strings.Fields(remainder)
	// Return early if no parts.
	if len(parts) == 0 {
		// Nothing to parse.
		return
	}

	// First part could be service name or start of message.
	msgStartIdx := 0
	// Check if first part is a service name.
	if len(parts) > 1 && !strings.Contains(parts[0], "=") && isServiceName(parts[0]) {
		entry.Service = parts[0]
		msgStartIdx = 1
	}

	// Find where metadata starts (first key=value pair).
	metaStartIdx := len(parts)
	// Search for first metadata key=value.
	for i := msgStartIdx; i < len(parts); i++ {
		// Check if this part contains a key=value pair.
		if strings.Contains(parts[i], "=") {
			metaStartIdx = i
			// Found metadata start.
			break
		}
	}

	// Message is between msgStartIdx and metaStartIdx.
	if metaStartIdx > msgStartIdx {
		entry.Message = strings.Join(parts[msgStartIdx:metaStartIdx], " ")
	}

	// Parse metadata key=value pairs.
	for i := metaStartIdx; i < len(parts); i++ {
		// Find equals sign.
		if idx := strings.Index(parts[i], "="); idx > 0 {
			key := parts[i][:idx]
			value := parts[i][idx+1:]
			entry.Metadata[key] = value
		}
	}
}

// readLastLines reads the last n lines from a file.
// Returns nil, nil if the file does not exist.
//
// Params:
//   - path: the path to the file.
//   - maxLines: the maximum number of lines to read.
//
// Returns:
//   - []string: the last lines from the file.
//   - error: nil on success, error on failure.
func readLastLines(path string, maxLines int) ([]string, error) {
	file, err := os.Open(path)
	// Check for file open errors.
	if err != nil {
		// File not found is not an error.
		if os.IsNotExist(err) {
			// File does not exist.
			return nil, nil
		}
		// Other error.
		return nil, err
	}
	defer func() { _ = file.Close() }()

	lines := make([]string, 0, maxLines)
	scanner := bufio.NewScanner(file)
	// Read all lines, keeping only the last maxLines.
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		// Trim to max size if exceeded.
		if len(lines) > maxLines {
			copy(lines, lines[1:])
			lines = lines[:maxLines]
		}
	}
	// Check for scanner errors.
	if err := scanner.Err(); err != nil {
		// Scanner error.
		return nil, err
	}

	// Return collected lines.
	return lines, nil
}
