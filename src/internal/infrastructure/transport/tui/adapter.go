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

// TUISnapshotData contains service data for TUI display.
type TUISnapshotData struct {
	Name   string
	State  process.State
	PID    int
	Uptime int64
}

// TUISnapshotsProvider provides service snapshots for TUI display.
type TUISnapshotsProvider interface {
	// TUISnapshots returns service data for TUI display.
	TUISnapshots() []TUISnapshotData
}

// SupervisorMetrics provides process metrics from the metrics tracker.
type SupervisorMetrics interface {
	// Get returns metrics for a specific service.
	Get(serviceName string) (domainmetrics.ProcessMetrics, bool)
}

// DynamicServiceProvider queries the supervisor on each call.
type DynamicServiceProvider struct {
	provider TUISnapshotsProvider
	metrics  SupervisorMetrics
}

// NewDynamicServiceProvider creates a new dynamic service provider.
func NewDynamicServiceProvider(provider TUISnapshotsProvider, metrics SupervisorMetrics) *DynamicServiceProvider {
	return &DynamicServiceProvider{
		provider: provider,
		metrics:  metrics,
	}
}

// Services implements ServiceProvider.
func (p *DynamicServiceProvider) Services() []model.ServiceSnapshot {
	if p.provider == nil {
		return nil
	}

	snapshots := p.provider.TUISnapshots()
	result := make([]model.ServiceSnapshot, 0, len(snapshots))

	for _, snap := range snapshots {
		ss := model.ServiceSnapshot{
			Name:   snap.Name,
			State:  snap.State,
			PID:    snap.PID,
			Uptime: time.Duration(snap.Uptime) * time.Second,
		}

		// Add metrics if available.
		if p.metrics != nil {
			if m, ok := p.metrics.Get(snap.Name); ok {
				ss.CPUPercent = m.CPU.UsagePercent
				ss.MemoryRSS = m.Memory.RSS
			}
		}

		result = append(result, ss)
	}

	return result
}

// SystemMetricsAdapter provides system metrics.
type SystemMetricsAdapter struct {
	// System metrics will be collected via collectors.
	// This is a placeholder that can be extended.
}

// NewSystemMetricsAdapter creates a new system metrics adapter.
func NewSystemMetricsAdapter() *SystemMetricsAdapter {
	return &SystemMetricsAdapter{}
}

// SystemMetrics implements MetricsProvider.
func (a *SystemMetricsAdapter) SystemMetrics() model.SystemMetrics {
	// System metrics are collected by the TUI collectors.
	// This returns empty metrics; the TUI will use collectors instead.
	return model.SystemMetrics{}
}

// LogBuffer is a thread-safe ring buffer for log entries.
type LogBuffer struct {
	mu         sync.RWMutex
	entries    []model.LogEntry
	maxSize    int
	infoCount  int
	warnCount  int
	errorCount int
}

// NewLogBuffer creates a new log buffer with the specified capacity.
func NewLogBuffer(maxSize int) *LogBuffer {
	if maxSize <= 0 {
		maxSize = 100
	}
	return &LogBuffer{
		entries: make([]model.LogEntry, 0, maxSize),
		maxSize: maxSize,
	}
}

// Add adds a log entry to the buffer.
func (b *LogBuffer) Add(entry model.LogEntry) {
	b.mu.Lock()
	defer b.mu.Unlock()

	// Update counts.
	switch entry.Level {
	case "INFO":
		b.infoCount++
	case "WARN", "WARNING":
		b.warnCount++
	case "ERROR", "ERR":
		b.errorCount++
	}

	// Add entry.
	if len(b.entries) >= b.maxSize {
		// Remove oldest entry and shift.
		b.entries = b.entries[1:]
	}
	b.entries = append(b.entries, entry)
}

// AddFromDomainEvent adds a domain LogEvent to the buffer.
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

// Entries returns a copy of all entries.
func (b *LogBuffer) Entries() []model.LogEntry {
	b.mu.RLock()
	defer b.mu.RUnlock()

	result := make([]model.LogEntry, len(b.entries))
	copy(result, b.entries)
	return result
}

// Summary returns the log summary.
func (b *LogBuffer) Summary() model.LogSummary {
	b.mu.RLock()
	defer b.mu.RUnlock()

	return model.LogSummary{
		Period:        5 * time.Minute,
		InfoCount:     b.infoCount,
		WarnCount:     b.warnCount,
		ErrorCount:    b.errorCount,
		RecentEntries: b.Entries(),
		HasAlerts:     b.errorCount > 0,
	}
}

// Clear resets the buffer.
func (b *LogBuffer) Clear() {
	b.mu.Lock()
	defer b.mu.Unlock()

	b.entries = b.entries[:0]
	b.infoCount = 0
	b.warnCount = 0
	b.errorCount = 0
}

// LogAdapter provides log summary from a log buffer.
type LogAdapter struct {
	buffer *LogBuffer
}

// NewLogAdapter creates a new log adapter with a default buffer size.
func NewLogAdapter() *LogAdapter {
	return &LogAdapter{
		buffer: NewLogBuffer(100),
	}
}

// NewLogAdapterWithBuffer creates a new log adapter with a custom buffer.
func NewLogAdapterWithBuffer(buffer *LogBuffer) *LogAdapter {
	return &LogAdapter{
		buffer: buffer,
	}
}

// LogSummary implements HealthProvider.
func (a *LogAdapter) LogSummary() model.LogSummary {
	if a.buffer == nil {
		return model.LogSummary{}
	}
	return a.buffer.Summary()
}

// AddLog adds a log entry to the adapter.
func (a *LogAdapter) AddLog(entry model.LogEntry) {
	if a.buffer != nil {
		a.buffer.Add(entry)
	}
}

// AddDomainEvent adds a domain log event to the adapter.
func (a *LogAdapter) AddDomainEvent(event domainlogging.LogEvent) {
	if a.buffer != nil {
		a.buffer.AddFromDomainEvent(event)
	}
}

// Buffer returns the underlying buffer.
func (a *LogAdapter) Buffer() *LogBuffer {
	return a.buffer
}

// TUILogWriter implements domain/logging.Writer to capture logs for TUI.
type TUILogWriter struct {
	adapter *LogAdapter
}

// NewTUILogWriter creates a writer that sends logs to the TUI.
func NewTUILogWriter(adapter *LogAdapter) *TUILogWriter {
	return &TUILogWriter{
		adapter: adapter,
	}
}

// Write implements domain/logging.Writer.
func (w *TUILogWriter) Write(event domainlogging.LogEvent) error {
	if w.adapter != nil {
		w.adapter.AddDomainEvent(event)
	}
	return nil
}

// Close implements domain/logging.Writer.
func (w *TUILogWriter) Close() error {
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
	if a.buffer == nil || path == "" {
		return nil
	}

	if maxLines <= 0 {
		maxLines = 100
	}

	// Open file.
	file, err := os.Open(path)
	if err != nil {
		// File not found is not an error - daemon may be starting fresh.
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	defer func() { _ = file.Close() }()

	// Read all lines (simple approach for small files).
	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}
	if err := scanner.Err(); err != nil {
		return err
	}

	// Take only the last maxLines.
	startIdx := 0
	if len(lines) > maxLines {
		startIdx = len(lines) - maxLines
	}

	// Parse and add entries.
	for _, line := range lines[startIdx:] {
		if entry, ok := parseLogLine(line); ok {
			a.buffer.Add(entry)
		}
	}

	return nil
}

// logLineRegex parses log lines in format:
// 2006-01-02T15:04:05Z07:00 [LEVEL] service Message key=value
var logLineRegex = regexp.MustCompile(
	`^(\d{4}-\d{2}-\d{2}T\d{2}:\d{2}:\d{2}[^\s]*)\s+\[([A-Z]+)\]\s+(.*)$`,
)

// parseLogLine parses a log line into a LogEntry.
func parseLogLine(line string) (model.LogEntry, bool) {
	matches := logLineRegex.FindStringSubmatch(line)
	if len(matches) < 4 {
		return model.LogEntry{}, false
	}

	// Parse timestamp.
	ts, err := time.Parse(time.RFC3339, matches[1])
	if err != nil {
		// Try alternative format without timezone.
		ts, err = time.Parse("2006-01-02T15:04:05Z", matches[1])
		if err != nil {
			return model.LogEntry{}, false
		}
	}

	level := matches[2]
	remainder := matches[3]

	// Parse remainder: "service Message key=value key2=value2"
	// or just "Message key=value" if no service.
	entry := model.LogEntry{
		Timestamp: ts,
		Level:     level,
		Metadata:  make(map[string]any),
	}

	// Split remainder into parts.
	parts := strings.Fields(remainder)
	if len(parts) == 0 {
		return entry, true
	}

	// First part could be service name or start of message.
	// Service names don't contain "=" and are typically short identifiers.
	msgStartIdx := 0
	if len(parts) > 1 && !strings.Contains(parts[0], "=") && isServiceName(parts[0]) {
		entry.Service = parts[0]
		msgStartIdx = 1
	}

	// Find where metadata starts (first key=value pair).
	metaStartIdx := len(parts)
	for i := msgStartIdx; i < len(parts); i++ {
		if strings.Contains(parts[i], "=") {
			metaStartIdx = i
			break
		}
	}

	// Message is between msgStartIdx and metaStartIdx.
	if metaStartIdx > msgStartIdx {
		entry.Message = strings.Join(parts[msgStartIdx:metaStartIdx], " ")
	}

	// Parse metadata.
	for i := metaStartIdx; i < len(parts); i++ {
		if idx := strings.Index(parts[i], "="); idx > 0 {
			key := parts[i][:idx]
			value := parts[i][idx+1:]
			entry.Metadata[key] = value
		}
	}

	return entry, true
}

// isServiceName checks if a string looks like a service name.
func isServiceName(s string) bool {
	// Service names are typically alphanumeric with dashes/underscores.
	// They don't start with uppercase words like "Service", "Daemon", etc.
	if len(s) == 0 {
		return false
	}
	// Common message starters are not service names.
	commonStarters := []string{"Service", "Daemon", "Supervisor", "Failed", "Started", "Stopped"}
	for _, starter := range commonStarters {
		if s == starter {
			return false
		}
	}
	return true
}
