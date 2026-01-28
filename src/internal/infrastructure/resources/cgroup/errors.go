//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

import (
	"errors"
	"fmt"
)

// Package errors.
var (
	// ErrUnknownVersion is returned when the cgroup version cannot be detected.
	ErrUnknownVersion error = errors.New("unknown cgroup version")

	// ErrPathNotFound is returned when the cgroup path cannot be detected.
	ErrPathNotFound error = errors.New("cgroup path not found in /proc/self/cgroup")

	// ErrInvalidFormat is the base error for format validation failures.
	ErrInvalidFormat error = errors.New("invalid format")
)

// InvalidFormatError represents a format validation error.
// It contains details about the parsing failure including the file, content, and field mismatch.
type InvalidFormatError struct {
	File     string
	Content  string
	Expected int
	Got      int
}

// NewInvalidFormatError wraps parsing failures with context.
//
// Params:
//   - file: cgroup file that failed to parse
//   - content: raw content that could not be parsed
//   - expected: number of fields expected
//   - got: actual number of fields found
//
// Returns:
//   - *InvalidFormatError: structured error with parsing context
func NewInvalidFormatError(file, content string, expected, got int) *InvalidFormatError {
	return &InvalidFormatError{File: file, Content: content, Expected: expected, Got: got}
}

// Error formats the validation failure with file and field count details.
//
// Returns:
//   - string: formatted error message with file, content, and field counts
func (e *InvalidFormatError) Error() string {
	return fmt.Sprintf("invalid %s format %q: expected %d fields, got %d", e.File, e.Content, e.Expected, e.Got)
}

// Unwrap enables errors.Is/As to match ErrInvalidFormat.
//
// Returns:
//   - error: base ErrInvalidFormat for error chain matching
func (e *InvalidFormatError) Unwrap() error { return ErrInvalidFormat }
