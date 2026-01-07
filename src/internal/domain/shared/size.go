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

// Parsing constants for strconv.ParseInt.
const (
	// base10 is the numeric base for decimal parsing.
	base10 int = 10
	// bitSize64 is the bit size for int64 parsing.
	bitSize64 int = 64
)

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

	// Check for empty input string.
	if s == "" {
		// Return error for empty size string.
		return 0, ErrEmptySize
	}

	multiplier, numStr := extractSizeComponents(s)
	numStr = strings.TrimSpace(numStr)
	num, err := strconv.ParseInt(numStr, base10, bitSize64)

	// Check for parsing errors.
	if err != nil {
		// Return wrapped error for invalid size number.
		return 0, fmt.Errorf("invalid size number %q: %w", numStr, err)
	}

	// Check for negative size values.
	if num < 0 {
		// Return error for negative size.
		return 0, fmt.Errorf("%w: %d", ErrNegativeSize, num)
	}

	// Return the calculated size in bytes.
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
	// Determine the size unit suffix and extract the numeric part.
	switch {
	// Handle gigabyte suffix.
	case strings.HasSuffix(s, "GB"):
		// Return gigabyte multiplier and numeric part.
		return Gigabyte, strings.TrimSuffix(s, "GB")
	// Handle megabyte suffix.
	case strings.HasSuffix(s, "MB"):
		// Return megabyte multiplier and numeric part.
		return Megabyte, strings.TrimSuffix(s, "MB")
	// Handle kilobyte suffix.
	case strings.HasSuffix(s, "KB"):
		// Return kilobyte multiplier and numeric part.
		return Kilobyte, strings.TrimSuffix(s, "KB")
	// Handle byte suffix.
	case strings.HasSuffix(s, "B"):
		// Return byte multiplier and numeric part.
		return Byte, strings.TrimSuffix(s, "B")
	// Handle no suffix (default to bytes).
	default:
		// Return byte multiplier and original string.
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
	// Determine the appropriate unit based on size magnitude.
	switch {
	// Format as gigabytes for sizes >= 1GB.
	case bytes >= Gigabyte:
		// Return gigabyte formatted string.
		return fmt.Sprintf("%dGB", bytes/Gigabyte)
	// Format as megabytes for sizes >= 1MB.
	case bytes >= Megabyte:
		// Return megabyte formatted string.
		return fmt.Sprintf("%dMB", bytes/Megabyte)
	// Format as kilobytes for sizes >= 1KB.
	case bytes >= Kilobyte:
		// Return kilobyte formatted string.
		return fmt.Sprintf("%dKB", bytes/Kilobyte)
	// Format as bytes for smaller sizes.
	default:
		// Return byte formatted string.
		return fmt.Sprintf("%dB", bytes)
	}
}
