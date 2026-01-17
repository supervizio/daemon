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

// NewInvalidFormatError creates a new InvalidFormatError.
//
// Params:
//   - file: the name of the file that failed validation.
//   - content: the content that failed to parse.
//   - expected: the expected number of fields.
//   - got: the actual number of fields found.
//
// Returns:
//   - *InvalidFormatError: a new format validation error.
func NewInvalidFormatError(file, content string, expected, got int) *InvalidFormatError {
	// Create and return new error with provided values.
	return &InvalidFormatError{
		File:     file,
		Content:  content,
		Expected: expected,
		Got:      got,
	}
}

// Error implements the error interface.
//
// Returns:
//   - string: formatted error message
func (e *InvalidFormatError) Error() string {
	// Return detailed error message
	return fmt.Sprintf("invalid %s format %q: expected %d fields, got %d", e.File, e.Content, e.Expected, e.Got)
}

// Unwrap returns the base error for errors.Is/As support.
//
// Returns:
//   - error: the base error
func (e *InvalidFormatError) Unwrap() error {
	// Return base error for error chain
	return ErrInvalidFormat
}
