//go:build linux

// Package cgroup_test provides external tests for the cgroup package.
package cgroup_test

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/infrastructure/resources/cgroup"
)

// TestInvalidFormatError_Error tests the Error method.
func TestInvalidFormatError_Error(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		err      *cgroup.InvalidFormatError
		expected string
	}{
		{
			name: "cpu_stat_format_error",
			err: &cgroup.InvalidFormatError{
				File:     "cpu.stat",
				Content:  "invalid content",
				Expected: 2,
				Got:      1,
			},
			expected: `invalid cpu.stat format "invalid content": expected 2 fields, got 1`,
		},
		{
			name: "memory_max_format_error",
			err: &cgroup.InvalidFormatError{
				File:     "memory.max",
				Content:  "bad",
				Expected: 1,
				Got:      3,
			},
			expected: `invalid memory.max format "bad": expected 1 fields, got 3`,
		},
		{
			name: "empty_content",
			err: &cgroup.InvalidFormatError{
				File:     "cpu.max",
				Content:  "",
				Expected: 2,
				Got:      0,
			},
			expected: `invalid cpu.max format "": expected 2 fields, got 0`,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify error message.
			assert.Equal(t, tt.expected, tt.err.Error())
		})
	}
}

// TestInvalidFormatError_Unwrap tests the Unwrap method.
func TestInvalidFormatError_Unwrap(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name         string
		err          *cgroup.InvalidFormatError
		expectedBase error
	}{
		{
			name: "cpu_stat_unwrap",
			err: &cgroup.InvalidFormatError{
				File:     "cpu.stat",
				Content:  "bad",
				Expected: 2,
				Got:      1,
			},
			expectedBase: cgroup.ErrInvalidFormat,
		},
		{
			name: "memory_max_unwrap",
			err: &cgroup.InvalidFormatError{
				File:     "memory.max",
				Content:  "invalid",
				Expected: 1,
				Got:      2,
			},
			expectedBase: cgroup.ErrInvalidFormat,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify Unwrap returns ErrInvalidFormat.
			assert.ErrorIs(t, tt.err, tt.expectedBase)

			// Verify the unwrapped error.
			unwrapped := tt.err.Unwrap()
			assert.Equal(t, tt.expectedBase, unwrapped)
		})
	}
}

// TestNewInvalidFormatError tests the constructor.
func TestNewInvalidFormatError(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		file     string
		content  string
		expected int
		got      int
	}{
		{
			name:     "cpu_stat_error",
			file:     "cpu.stat",
			content:  "invalid",
			expected: 2,
			got:      1,
		},
		{
			name:     "memory_max_error",
			file:     "memory.max",
			content:  "max 100",
			expected: 1,
			got:      2,
		},
		{
			name:     "empty_content_error",
			file:     "cpu.max",
			content:  "",
			expected: 2,
			got:      0,
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create error using constructor.
			err := cgroup.NewInvalidFormatError(tt.file, tt.content, tt.expected, tt.got)

			// Verify fields.
			assert.Equal(t, tt.file, err.File)
			assert.Equal(t, tt.content, err.Content)
			assert.Equal(t, tt.expected, err.Expected)
			assert.Equal(t, tt.got, err.Got)

			// Verify it implements error interface.
			var _ error = err

			// Verify it wraps ErrInvalidFormat.
			assert.True(t, errors.Is(err, cgroup.ErrInvalidFormat))
		})
	}
}

// TestErrUnknownVersion tests the ErrUnknownVersion sentinel error.
func TestErrUnknownVersion(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		err              error
		expectedContains []string
	}{
		{
			name:             "contains_unknown",
			err:              cgroup.ErrUnknownVersion,
			expectedContains: []string{"unknown"},
		},
		{
			name:             "contains_cgroup",
			err:              cgroup.ErrUnknownVersion,
			expectedContains: []string{"cgroup"},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify error is not nil.
			assert.NotNil(t, tt.err)

			// Verify error message contains expected strings.
			for _, substr := range tt.expectedContains {
				assert.Contains(t, tt.err.Error(), substr)
			}
		})
	}
}

// TestErrPathNotFound tests the ErrPathNotFound sentinel error.
func TestErrPathNotFound(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		err              error
		expectedContains []string
	}{
		{
			name:             "contains_not_found",
			err:              cgroup.ErrPathNotFound,
			expectedContains: []string{"not found"},
		},
		{
			name:             "contains_cgroup",
			err:              cgroup.ErrPathNotFound,
			expectedContains: []string{"cgroup"},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify error is not nil.
			assert.NotNil(t, tt.err)

			// Verify error message contains expected strings.
			for _, substr := range tt.expectedContains {
				assert.Contains(t, tt.err.Error(), substr)
			}
		})
	}
}

// TestErrInvalidFormat tests the ErrInvalidFormat sentinel error.
func TestErrInvalidFormat(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name             string
		err              error
		expectedContains []string
	}{
		{
			name:             "contains_invalid",
			err:              cgroup.ErrInvalidFormat,
			expectedContains: []string{"invalid"},
		},
		{
			name:             "contains_format",
			err:              cgroup.ErrInvalidFormat,
			expectedContains: []string{"format"},
		},
	}

	// Iterate through test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Verify error is not nil.
			assert.NotNil(t, tt.err)

			// Verify error message contains expected strings.
			for _, substr := range tt.expectedContains {
				assert.Contains(t, tt.err.Error(), substr)
			}
		})
	}
}
