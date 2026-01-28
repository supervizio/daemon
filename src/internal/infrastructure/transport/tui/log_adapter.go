// Package tui provides terminal user interface rendering for superviz.io.
package tui

import (
	"bufio"
	"os"
	"regexp"
	"strings"
	"time"

	domainlogging "github.com/kodflow/daemon/internal/domain/logging"
	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

// Log parsing constants.
const (
	// defaultMaxLines is the default number of lines to load from log files.
	defaultMaxLines int = 100

	// minRegexGroups is the minimum number of groups for a valid log line regex match.
	minRegexGroups int = 4

	// regexGroupLevel is the regex capture group index for log level.
	regexGroupLevel int = 2

	// regexGroupRemainder is the regex capture group index for log remainder.
	regexGroupRemainder int = 3

	// initialMetadataCapacity is the initial capacity for log entry metadata maps.
	initialMetadataCapacity int = 4
)

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
	return &LogAdapter{
		buffer: buffer,
	}
}

// Summarize implements LogSummarizer.
//
// Returns:
//   - model.LogSummary: the log summary.
func (a *LogAdapter) Summarize() model.LogSummary {
	if a.buffer == nil {
		return model.LogSummary{}
	}
	return a.buffer.Summary()
}

// AddLog adds a log entry to the adapter.
//
// Params:
//   - entry: the log entry to add.
func (a *LogAdapter) AddLog(entry model.LogEntry) {
	if a.buffer != nil {
		a.buffer.Add(entry)
	}
}

// AddDomainEvent adds a domain log event to the adapter.
//
// Params:
//   - event: the domain log event.
func (a *LogAdapter) AddDomainEvent(event domainlogging.LogEvent) {
	if a.buffer != nil {
		a.buffer.AddFromDomainEvent(event)
	}
}

// Buffer returns the underlying buffer.
//
// Returns:
//   - *LogBuffer: the log buffer.
func (a *LogAdapter) Buffer() *LogBuffer {
	return a.buffer
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
		maxLines = defaultMaxLines
	}

	lines, err := readLastLines(path, maxLines)
	if err != nil {
		return err
	}

	for _, line := range lines {
		if entry, ok := parseLogLine(line); ok {
			a.buffer.Add(entry)
		}
	}

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
	if len(matches) < minRegexGroups {
		return model.LogEntry{}, false
	}

	ts, ok := parseLogTimestamp(matches[1])
	if !ok {
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

	parseLogRemainder(&entry, remainder)

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
		return false
	}
	commonStarters := []string{"Service", "Daemon", "Supervisor", "Failed", "Started", "Stopped"}
	for _, starter := range commonStarters {
		if s == starter {
			return false
		}
	}
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
	if err != nil {
		ts, err = time.Parse("2006-01-02T15:04:05Z", s)
		if err != nil {
			return time.Time{}, false
		}
	}
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
	if len(parts) == 0 {
		return
	}

	// First part could be service name or start of message.
	msgStartIdx := extractServiceName(entry, parts)

	// Find where metadata starts (first key=value pair).
	metaStartIdx := findMetadataStart(parts, msgStartIdx)

	// Message is between msgStartIdx and metaStartIdx.
	if metaStartIdx > msgStartIdx {
		entry.Message = strings.Join(parts[msgStartIdx:metaStartIdx], " ")
	}

	// Parse metadata key=value pairs.
	extractMetadata(entry, parts, metaStartIdx)
}

// extractServiceName extracts service name from parts if present.
//
// Params:
//   - entry: the log entry to populate.
//   - parts: parsed parts of the log line.
//
// Returns:
//   - int: index where message starts (0 if no service, 1 if service found).
func extractServiceName(entry *model.LogEntry, parts []string) int {
	hasMultipleParts := len(parts) > 1
	isNotMetadata := !strings.Contains(parts[0], "=")
	isService := isServiceName(parts[0])
	if hasMultipleParts && isNotMetadata && isService {
		entry.Service = parts[0]
		return 1
	}
	return 0
}

// findMetadataStart finds the index where metadata key=value pairs begin.
//
// Params:
//   - parts: parsed parts of the log line.
//   - startIdx: index to start searching from.
//
// Returns:
//   - int: index of first metadata pair, or len(parts) if none found.
func findMetadataStart(parts []string, startIdx int) int {
	for i := startIdx; i < len(parts); i++ {
		if strings.Contains(parts[i], "=") {
			return i
		}
	}
	return len(parts)
}

// extractMetadata extracts key=value pairs into entry metadata.
//
// Params:
//   - entry: the log entry to populate.
//   - parts: parsed parts of the log line.
//   - startIdx: index where metadata begins.
func extractMetadata(entry *model.LogEntry, parts []string, startIdx int) {
	for i := startIdx; i < len(parts); i++ {
		idx := strings.Index(parts[i], "=")
		if idx <= 0 {
			continue
		}
		key := parts[i][:idx]
		value := parts[i][idx+1:]
		entry.Metadata[key] = value
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
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = file.Close() }()

	lines := make([]string, 0, maxLines)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
		if len(lines) > maxLines {
			copy(lines, lines[1:])
			lines = lines[:maxLines]
		}
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return lines, nil
}
