// Package daemon provides daemon event logging infrastructure.
package daemon

import (
	"fmt"
	"maps"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/kodflow/daemon/internal/domain/logging"
)

// Buffer size and format constants.
const (
	// typicalLogLineLength is the initial capacity for log line building.
	typicalLogLineLength int = 128
	// decimalBase is the base for decimal number formatting.
	decimalBase int = 10
	// floatPrecision64 is the bit size for 64-bit float formatting.
	floatPrecision64 int = 64
)

var (
	// builderPool provides reusable strings.Builder instances to reduce allocations.
	// This is safe for concurrent use as sync.Pool handles synchronization.
	builderPool sync.Pool = sync.Pool{
		New: func() any {
			return &strings.Builder{}
		},
	}

	// Ensure TextFormatter implements Formatter.
	_ Formatter = (*TextFormatter)(nil)
)

// getBuilder retrieves a strings.Builder from the pool.
//
// Returns:
//   - *strings.Builder: a builder from the pool.
func getBuilder() *strings.Builder {
	// Get builder from pool (safe cast from pool).
	sb, ok := builderPool.Get().(*strings.Builder)
	// Create new builder if pool returned invalid type.
	if !ok {
		sb = &strings.Builder{}
	}
	// Return builder for string concatenation.
	return sb
}

// putBuilder returns a strings.Builder to the pool after resetting it.
//
// Params:
//   - sb: the builder to return to the pool.
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
// It uses a sync.Pool for strings.Builder instances to minimize allocations.
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
	// Use RFC3339 if no format specified.
	if timestampFormat == "" {
		timestampFormat = "2006-01-02T15:04:05Z07:00"
	}
	// Return formatter with timestamp format.
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
	sb.Grow(typicalLogLineLength)

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
		// Use event type as fallback.
		sb.WriteString(event.EventType)
	}

	// Metadata.
	if len(event.Metadata) > 0 {
		sb.WriteByte(' ')
		formatMetadataToBuilder(sb, event.Metadata)
	}

	// Return formatted log line.
	return sb.String()
}

// formatMetadataToBuilder formats metadata directly to a builder.
// Avoids intermediate string allocation by writing directly.
//
// Params:
//   - sb: the string builder to write to.
//   - meta: the metadata map to format.
func formatMetadataToBuilder(sb *strings.Builder, meta map[string]any) {
	// Sort keys for consistent output.
	keys := slices.Collect(maps.Keys(meta))
	sort.Strings(keys)

	// Format each key-value pair.
	for i, k := range keys {
		// Add space separator between pairs.
		if i > 0 {
			sb.WriteByte(' ')
		}
		sb.WriteString(k)
		sb.WriteByte('=')
		formatValue(sb, meta[k])
	}
}

// formatValue formats a value to the builder using type switch for efficiency.
//
// Params:
//   - sb: the string builder to write to.
//   - v: the value to format (any type for log metadata flexibility).
//
// Note: any type required - log metadata values are runtime-determined and type-heterogeneous
// (primitives, errors, custom types). Type switch handles common cases efficiently.
//
//nolint:ktn-interface-anyuse // any required: log metadata is runtime-determined, type switch for efficient handling
func formatValue(sb *strings.Builder, v any) {
	// Type switch for efficient formatting.
	switch val := v.(type) {
	// String values are written directly.
	case string:
		sb.WriteString(val)
	// Integer types use optimized formatting.
	case int:
		sb.WriteString(strconv.Itoa(val))
	// Int64 types use base 10 formatting.
	case int64:
		sb.WriteString(strconv.FormatInt(val, decimalBase))
	// Uint64 types use base 10 formatting.
	case uint64:
		sb.WriteString(strconv.FormatUint(val, decimalBase))
	// Float64 types use 64-bit precision.
	case float64:
		sb.WriteString(strconv.FormatFloat(val, 'f', -1, floatPrecision64))
	// Boolean values use Go formatting.
	case bool:
		sb.WriteString(strconv.FormatBool(val))
	// Complex types fall back to fmt.
	default:
		// Fallback for complex types.
		fmt.Fprintf(sb, "%v", val)
	}
}
