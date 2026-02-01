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

	// reject empty input
	if s == "" {
		// return empty size error
		return 0, ErrEmptySize
	}

	multiplier, numStr := extractSizeComponents(s)
	numStr = strings.TrimSpace(numStr)
	num, err := strconv.ParseInt(numStr, Base10, BitSize64)

	// validate numeric parsing
	if err != nil {
		// return parsing error
		return 0, fmt.Errorf("invalid size number %q: %w", numStr, err)
	}

	// reject negative values
	if num < 0 {
		// return negative size error
		return 0, fmt.Errorf("%w: %d", ErrNegativeSize, num)
	}

	// calculate final size in bytes
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
	// match suffix to unit multiplier
	switch {
	// gigabyte suffix
	case strings.HasSuffix(s, "GB"):
		// return gigabyte multiplier
		return Gigabyte, strings.TrimSuffix(s, "GB")
	// megabyte suffix
	case strings.HasSuffix(s, "MB"):
		// return megabyte multiplier
		return Megabyte, strings.TrimSuffix(s, "MB")
	// kilobyte suffix
	case strings.HasSuffix(s, "KB"):
		// return kilobyte multiplier
		return Kilobyte, strings.TrimSuffix(s, "KB")
	// byte suffix
	case strings.HasSuffix(s, "B"):
		// return byte multiplier
		return Byte, strings.TrimSuffix(s, "B")
	// no suffix
	default:
		// return byte multiplier
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
	// select appropriate unit for size
	switch {
	// format as gigabytes
	case bytes >= Gigabyte:
		// return gigabyte format
		return strconv.FormatInt(bytes/Gigabyte, Base10) + "GB"
	// format as megabytes
	case bytes >= Megabyte:
		// return megabyte format
		return strconv.FormatInt(bytes/Megabyte, Base10) + "MB"
	// format as kilobytes
	case bytes >= Kilobyte:
		// return kilobyte format
		return strconv.FormatInt(bytes/Kilobyte, Base10) + "KB"
	// format as bytes
	default:
		// return byte format
		return strconv.FormatInt(bytes, Base10) + "B"
	}
}
