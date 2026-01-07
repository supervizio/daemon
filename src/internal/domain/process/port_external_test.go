// Package process_test provides black-box tests for the process domain entities.
// These tests validate the public API behavior without accessing internal state.
package process_test

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/process"
)

// TestNewSpec validates Spec creation with various configurations.
// It ensures all parameters are correctly assigned to the resulting Spec.
//
// Params:
//   - t: the testing context
func TestNewSpec(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params process.SpecParams
	}{
		{
			name: "minimal spec",
			params: process.SpecParams{
				Command: "/bin/echo",
			},
		},
		{
			name: "full spec",
			params: process.SpecParams{
				Command: "/usr/bin/python",
				Args:    []string{"-c", "print('hello')"},
				Dir:     "/tmp",
				Env:     map[string]string{"PATH": "/usr/bin", "HOME": "/root"},
				User:    "nobody",
				Group:   "nogroup",
			},
		},
	}

	// Iterate through all test cases.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create the spec with provided parameters.
			spec := process.NewSpec(tt.params)

			// Verify all fields are correctly assigned.
			assert.Equal(t, tt.params.Command, spec.Command, "command should match")
			assert.Equal(t, tt.params.Args, spec.Args, "args should match")
			assert.Equal(t, tt.params.Dir, spec.Dir, "dir should match")
			assert.Equal(t, tt.params.Env, spec.Env, "env should match")
			assert.Equal(t, tt.params.User, spec.User, "user should match")
			assert.Equal(t, tt.params.Group, spec.Group, "group should match")
		})
	}
}

// TestSpecWithOutput validates output writer attachment to a Spec.
// It ensures stdout and stderr writers are correctly set.
//
// Params:
//   - t: the testing context
func TestSpecWithOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		command    string
		hasWriters bool
	}{
		{
			name:       "attach writers to echo spec",
			command:    "/bin/echo",
			hasWriters: true,
		},
		{
			name:       "attach writers to sleep spec",
			command:    "/bin/sleep",
			hasWriters: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a base spec.
			spec := process.NewSpec(process.SpecParams{Command: tt.command})

			// Create mock writers.
			stdout := &bytes.Buffer{}
			stderr := &bytes.Buffer{}

			// Attach output writers.
			specWithOutput := spec.WithOutput(stdout, stderr)

			// Verify writers are attached.
			assert.Equal(t, stdout, specWithOutput.Stdout, "stdout should be set")
			assert.Equal(t, stderr, specWithOutput.Stderr, "stderr should be set")

			// Verify original spec is unchanged (immutability).
			assert.Nil(t, spec.Stdout, "original stdout should be nil")
			assert.Nil(t, spec.Stderr, "original stderr should be nil")
		})
	}
}

// TestExitResultFields validates ExitResult struct field access.
// It ensures the struct correctly holds exit code and error information.
//
// Params:
//   - t: the testing context
func TestExitResultFields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		result   process.ExitResult
		wantCode int
		wantErr  bool
	}{
		{
			name: "successful exit",
			result: process.ExitResult{
				Code:  0,
				Error: nil,
			},
			wantCode: 0,
			wantErr:  false,
		},
		{
			name: "failed exit with code 1",
			result: process.ExitResult{
				Code:  1,
				Error: nil,
			},
			wantCode: 1,
			wantErr:  false,
		},
		{
			name: "failed exit with code 127",
			result: process.ExitResult{
				Code:  127,
				Error: nil,
			},
			wantCode: 127,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			assert.Equal(t, tt.wantCode, tt.result.Code, "exit code should match")
			if tt.wantErr {
				assert.NotNil(t, tt.result.Error, "error should not be nil")
			} else {
				assert.Nil(t, tt.result.Error, "error should be nil")
			}
		})
	}
}
