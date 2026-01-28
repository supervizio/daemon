// Package shared provides common domain types used across multiple domain packages.
// It contains utility functions and constants for size parsing and formatting.
package shared

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
)

// Size unit multipliers.
const (
	Byte     int64 = 1
	Kilobyte int64 = 1024
	Megabyte int64 = 1024 * Kilobyte
	Gigabyte int64 = 1024 * Megabyte
)

// Use constants from constants.go: Base10, BitSize64.

// Error variables for size parsing.
var (
	// ErrEmptySize indicates an empty size string was provided.
	ErrEmptySize error = errors.New("empty size string")
	// ErrNegativeSize indicates a negative size value was provided.
	ErrNegativeSize error = errors.New("size cannot be negative")
)

// ParseSize parses a human-readable size string into bytes.
// Supported formats: "100", "100B", "100KB", "100MB", "100GB"
// Case-insensitive.
//
// Params:
//   - s: size string to parse
//
// Returns:
//   - int64: size in bytes
//   - error: parsing error if format is invalid
func ParseSize(s string) (int64, error) {
	s = strings.TrimSpace(strings.ToUpper(s))

	if s == "" {
		return 0, ErrEmptySize
	}

	multiplier, numStr := extractSizeComponents(s)
	numStr = strings.TrimSpace(numStr)
	num, err := strconv.ParseInt(numStr, Base10, BitSize64)

	if err != nil {
		return 0, fmt.Errorf("invalid size number %q: %w", numStr, err)
	}

	if num < 0 {
		return 0, fmt.Errorf("%w: %d", ErrNegativeSize, num)
	}

	return num * multiplier, nil
}

// extractSizeComponents extracts the multiplier and numeric string from a size string.
//
// Params:
//   - s: uppercase size string to parse
//
// Returns:
//   - multiplier: the multiplier based on the unit suffix
//   - numericPart: the numeric portion of the string
func extractSizeComponents(s string) (multiplier int64, numericPart string) {
	switch {
	case strings.HasSuffix(s, "GB"):
		return Gigabyte, strings.TrimSuffix(s, "GB")
	case strings.HasSuffix(s, "MB"):
		return Megabyte, strings.TrimSuffix(s, "MB")
	case strings.HasSuffix(s, "KB"):
		return Kilobyte, strings.TrimSuffix(s, "KB")
	case strings.HasSuffix(s, "B"):
		return Byte, strings.TrimSuffix(s, "B")
	default:
		return Byte, s
	}
}

// FormatSize formats a size in bytes to a human-readable string.
//
// Params:
//   - bytes: size in bytes
//
// Returns:
//   - string: human-readable size string
func FormatSize(bytes int64) string {
	switch {
	case bytes >= Gigabyte:
		return strconv.FormatInt(bytes/Gigabyte, Base10) + "GB"
	case bytes >= Megabyte:
		return strconv.FormatInt(bytes/Megabyte, Base10) + "MB"
	case bytes >= Kilobyte:
		return strconv.FormatInt(bytes/Kilobyte, Base10) + "KB"
	default:
		return strconv.FormatInt(bytes, Base10) + "B"
	}
}
