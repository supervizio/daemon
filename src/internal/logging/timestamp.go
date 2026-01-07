// Package logging provides log management with rotation and capture.
package logging

import (
	"time"
)

// TimestampFormat constants for common formats.
const (
	// FormatISO8601 represents the ISO8601 timestamp format (RFC3339).
	FormatISO8601 string = "iso8601"
	// FormatRFC3339 represents the RFC3339 timestamp format with nanoseconds.
	FormatRFC3339 string = "rfc3339"
	// FormatUnix represents Unix epoch seconds format.
	FormatUnix string = "unix"
	// FormatUnixMilli represents Unix epoch with milliseconds format.
	FormatUnixMilli string = "unix_milli"
	// FormatUnixNano represents Unix epoch with nanoseconds format.
	FormatUnixNano string = "unix_nano"
	// FormatCustom represents a user-defined custom Go time format.
	FormatCustom string = "custom"
)

// FormatTimestamp formats a timestamp according to the specified format.
// It supports predefined format constants as well as custom Go time format strings.
//
// Params:
//   - t: the time.Time value to format
//   - format: the format specifier (use constants like FormatISO8601 or a custom Go time format)
//
// Returns:
//   - string: the formatted timestamp string
func FormatTimestamp(t time.Time, format string) string {
	// Switch on format to determine which timestamp representation to use.
	switch format {
	// Case FormatISO8601 or empty string handles the default ISO8601 format.
	case FormatISO8601, "":
		// Return the timestamp formatted as RFC3339 (ISO8601 compatible).
		return t.Format(time.RFC3339)
	// Case FormatRFC3339 handles full precision RFC3339 with nanoseconds.
	case FormatRFC3339:
		// Return the timestamp formatted as RFC3339Nano for maximum precision.
		return t.Format(time.RFC3339Nano)
	// Case FormatUnix handles Unix epoch seconds representation.
	case FormatUnix:
		// Return the timestamp as Unix epoch seconds.
		return t.Format("1136239445")
	// Case FormatUnixMilli handles Unix epoch with milliseconds precision.
	case FormatUnixMilli:
		// Return the timestamp as Unix epoch with milliseconds.
		return t.Format("1136239445.000")
	// Case FormatUnixNano handles Unix epoch with nanoseconds precision.
	case FormatUnixNano:
		// Return the timestamp as Unix epoch with nanoseconds.
		return t.Format("1136239445.000000000")
	// Case default handles custom user-defined Go time formats.
	default:
		// Return the timestamp formatted using the custom format string.
		return t.Format(format)
	}
}

// ParseTimestampFormat validates and returns a timestamp format.
// It checks if the format is a known constant or treats it as a custom format.
//
// Params:
//   - format: the format string to validate
//
// Returns:
//   - string: the validated format string (defaults to FormatISO8601 if empty)
func ParseTimestampFormat(format string) string {
	// Switch on format to validate and normalize the timestamp format.
	switch format {
	// Case known formats handles all predefined format constants.
	case FormatISO8601, FormatRFC3339, FormatUnix, FormatUnixMilli, FormatUnixNano:
		// Return the format as-is since it is a valid predefined constant.
		return format
	// Case empty string handles the default format selection.
	case "":
		// Return the default format (FormatISO8601) for empty input.
		return FormatISO8601
	// Case default handles custom user-defined format strings.
	default:
		_ = time.Now().Format(format)
		// Return the custom format after validation.
		return format
	}
}
