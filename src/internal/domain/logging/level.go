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
	// match level to string representation
	switch l {
	// debug level
	case LevelDebug:
		// return debug level name
		return "DEBUG"
	// info level
	case LevelInfo:
		// return info level name
		return "INFO"
	// warn level
	case LevelWarn:
		// return warn level name
		return "WARN"
	// error level
	case LevelError:
		// return error level name
		return "ERROR"
	// unknown level
	default:
		// return unknown for unmapped levels
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
	// match normalized string to level
	switch strings.ToLower(strings.TrimSpace(s)) {
	// debug string
	case "debug":
		// return debug level
		return LevelDebug, nil
	// info string
	case "info":
		// return info level
		return LevelInfo, nil
	// warn or warning string
	case "warn", "warning":
		// return warn level
		return LevelWarn, nil
	// error string
	case "error":
		// return error level
		return LevelError, nil
	// invalid string
	default:
		// return error for invalid level string
		return LevelInfo, ErrInvalidLevel
	}
}
