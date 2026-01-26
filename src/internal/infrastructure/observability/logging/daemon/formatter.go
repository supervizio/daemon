package daemon

import (
	"fmt"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/kodflow/daemon/internal/domain/logging"
)

// builderPool provides reusable strings.Builder instances to reduce allocations.
// This is safe for concurrent use as sync.Pool handles synchronization.
var builderPool = sync.Pool{
	New: func() any {
		return &strings.Builder{}
	},
}

// getBuilder retrieves a strings.Builder from the pool.
func getBuilder() *strings.Builder {
	return builderPool.Get().(*strings.Builder)
}

// putBuilder returns a strings.Builder to the pool after resetting it.
func putBuilder(sb *strings.Builder) {
	sb.Reset()
	builderPool.Put(sb)
}

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
// Uses sync.Pool for strings.Builder to minimize allocations in hot paths.
//
// Params:
//   - event: the log event to format.
//
// Returns:
//   - string: the formatted log line.
func (f *TextFormatter) Format(event logging.LogEvent) string {
	sb := getBuilder()
	defer putBuilder(sb)

	// Pre-grow for typical log line length.
	sb.Grow(128)

	// Timestamp.
	sb.WriteString(event.Timestamp.Format(f.timestampFormat))
	sb.WriteByte(' ')

	// Level.
	sb.WriteByte('[')
	sb.WriteString(event.Level.String())
	sb.WriteString("] ")

	// Service name.
	if event.Service != "" {
		sb.WriteString(event.Service)
		sb.WriteByte(' ')
	}

	// Use message if set, otherwise fall back to event type.
	if event.Message != "" {
		sb.WriteString(event.Message)
	} else {
		sb.WriteString(event.EventType)
	}

	// Metadata.
	if len(event.Metadata) > 0 {
		sb.WriteByte(' ')
		formatMetadataToBuilder(sb, event.Metadata)
	}

	return sb.String()
}

// formatMetadataToBuilder formats metadata directly to a builder.
// Avoids intermediate string allocation by writing directly.
func formatMetadataToBuilder(sb *strings.Builder, meta map[string]any) {
	// Sort keys for consistent output.
	keys := make([]string, 0, len(meta))
	for k := range meta {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		formatValue(sb, meta[k])
	}
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
		fmt.Fprintf(sb, "%v", val)
	}
}

// Ensure TextFormatter implements Formatter.
var _ Formatter = (*TextFormatter)(nil)
