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
var ErrInvalidLevel = errors.New("invalid log level")

// String returns the string representation of the level.
func (l Level) String() string {
	switch l {
	case LevelDebug:
		return "DEBUG"
	case LevelInfo:
		return "INFO"
	case LevelWarn:
		return "WARN"
	case LevelError:
		return "ERROR"
	default:
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
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug, nil
	case "info":
		return LevelInfo, nil
	case "warn", "warning":
		return LevelWarn, nil
	case "error":
		return LevelError, nil
	default:
		return LevelInfo, ErrInvalidLevel
	}
}
