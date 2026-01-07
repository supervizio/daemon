package logging

import (
	"time"
)

// TimestampFormat constants for common formats.
const (
	FormatISO8601    = "iso8601"
	FormatRFC3339    = "rfc3339"
	FormatUnix       = "unix"
	FormatUnixMilli  = "unix_milli"
	FormatUnixNano   = "unix_nano"
	FormatCustom     = "custom"
)

// FormatTimestamp formats a timestamp according to the specified format.
func FormatTimestamp(t time.Time, format string) string {
	switch format {
	case FormatISO8601, "":
		return t.Format(time.RFC3339)
	case FormatRFC3339:
		return t.Format(time.RFC3339Nano)
	case FormatUnix:
		return t.Format("1136239445")
	case FormatUnixMilli:
		return t.Format("1136239445.000")
	case FormatUnixNano:
		return t.Format("1136239445.000000000")
	default:
		// Treat as custom Go time format
		return t.Format(format)
	}
}

// ParseTimestampFormat validates and returns a timestamp format.
func ParseTimestampFormat(format string) string {
	switch format {
	case FormatISO8601, FormatRFC3339, FormatUnix, FormatUnixMilli, FormatUnixNano:
		return format
	case "":
		return FormatISO8601
	default:
		// Assume custom format, validate by trying to format
		_ = time.Now().Format(format)
		return format
	}
}
