package daemon

import (
	"fmt"
	"sort"
	"strconv"
	"strings"

	"github.com/kodflow/daemon/internal/domain/logging"
)

// Formatter formats log events into strings.
type Formatter interface {
	// Format formats a log event into a string.
	//
	// Params:
	//   - event: the log event to format.
	//
	// Returns:
	//   - string: the formatted log line.
	Format(event logging.LogEvent) string
}

// TextFormatter formats log events as human-readable text.
type TextFormatter struct {
	// timestampFormat is the Go time format string.
	timestampFormat string
}

// NewTextFormatter creates a new text formatter.
//
// Params:
//   - timestampFormat: the Go time format string (default: RFC3339).
//
// Returns:
//   - *TextFormatter: the created formatter.
func NewTextFormatter(timestampFormat string) *TextFormatter {
	if timestampFormat == "" {
		timestampFormat = "2006-01-02T15:04:05Z07:00"
	}
	return &TextFormatter{timestampFormat: timestampFormat}
}

// Format formats a log event as text.
//
// Params:
//   - event: the log event to format.
//
// Returns:
//   - string: the formatted log line.
func (f *TextFormatter) Format(event logging.LogEvent) string {
	var sb strings.Builder

	// Timestamp.
	sb.WriteString(event.Timestamp.Format(f.timestampFormat))
	sb.WriteString(" ")

	// Level.
	sb.WriteString("[")
	sb.WriteString(event.Level.String())
	sb.WriteString("] ")

	// Service name.
	if event.Service != "" {
		sb.WriteString(event.Service)
		sb.WriteString(" ")
	}

	// Use message if set, otherwise fall back to event type.
	if event.Message != "" {
		sb.WriteString(event.Message)
	} else {
		sb.WriteString(event.EventType)
	}

	// Metadata.
	if len(event.Metadata) > 0 {
		sb.WriteString(" ")
		sb.WriteString(formatMetadata(event.Metadata))
	}

	return sb.String()
}

// formatMetadata formats metadata as key=value pairs.
// Uses type switch for common types to avoid fmt.Sprintf allocations.
func formatMetadata(meta map[string]any) string {
	// Sort keys for consistent output.
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// Use strings.Builder to avoid intermediate allocations.
	var sb strings.Builder
	sb.Grow(len(keys) * 16) // Pre-allocate for typical key=value pairs.

	for i, k := range keys {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		formatValue(&sb, meta[k])
	}
	return sb.String()
}

// formatValue formats a value to the builder using type switch for efficiency.
func formatValue(sb *strings.Builder, v any) {
	switch val := v.(type) {
	case string:
		sb.WriteString(val)
	case int:
		sb.WriteString(strconv.Itoa(val))
	case int64:
		sb.WriteString(strconv.FormatInt(val, 10))
	case uint64:
		sb.WriteString(strconv.FormatUint(val, 10))
	case float64:
		sb.WriteString(strconv.FormatFloat(val, 'f', -1, 64))
	case bool:
		sb.WriteString(strconv.FormatBool(val))
	default:
		// Fallback for complex types.
		sb.WriteString(fmt.Sprintf("%v", val))
	}
}

// Ensure TextFormatter implements Formatter.
var _ Formatter = (*TextFormatter)(nil)
