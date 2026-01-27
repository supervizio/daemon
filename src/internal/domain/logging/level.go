// Package logging provides domain types for daemon event logging.
package logging

import (
	"errors"
	"strings"
)

// Level represents a log severity level.
type Level int

const (
	// LevelDebug is for detailed diagnostic information.
	LevelDebug Level = iota
	// LevelInfo is for general operational information.
	LevelInfo
	// LevelWarn is for warning conditions.
	LevelWarn
	// LevelError is for error conditions.
	LevelError
)

// ErrInvalidLevel is returned when parsing an invalid level string.
var ErrInvalidLevel error = errors.New("invalid log level")

// String returns the string representation of the level.
//
// Returns:
//   - string: the uppercase string representation (DEBUG, INFO, WARN, ERROR, UNKNOWN).
func (l Level) String() string {
	// Map level to uppercase string representation.
	switch l {
	// Debug level.
	case LevelDebug:
		// Return debug string.
		return "DEBUG"
	// Info level.
	case LevelInfo:
		// Return info string.
		return "INFO"
	// Warning level.
	case LevelWarn:
		// Return warn string.
		return "WARN"
	// Error level.
	case LevelError:
		// Return error string.
		return "ERROR"
	// Unknown level.
	default:
		// Return unknown string.
		return "UNKNOWN"
	}
}

// ParseLevel parses a string into a Level.
// Valid values: "debug", "info", "warn", "warning", "error" (case-insensitive).
//
// Params:
//   - s: the string to parse.
//
// Returns:
//   - Level: the parsed level.
//   - error: ErrInvalidLevel if the string is not a valid level.
func ParseLevel(s string) (Level, error) {
	// Normalize input to lowercase and match against valid levels.
	switch strings.ToLower(strings.TrimSpace(s)) {
	// Debug level.
	case "debug":
		// Return debug level.
		return LevelDebug, nil
	// Info level.
	case "info":
		// Return info level.
		return LevelInfo, nil
	// Warning level (both "warn" and "warning" accepted).
	case "warn", "warning":
		// Return warn level.
		return LevelWarn, nil
	// Error level.
	case "error":
		// Return error level.
		return LevelError, nil
	// Invalid level - return default Info and error.
	default:
		// Return default info level with error.
		return LevelInfo, ErrInvalidLevel
	}
}
