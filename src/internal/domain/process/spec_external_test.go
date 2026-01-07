// Package process_test provides black-box tests for the spec.go file.
// These tests validate the public API behavior of Spec without accessing internal state.
package process_test

import (
	"bytes"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/process"
)

// TestNewSpec_TableDriven validates Spec creation with various configurations.
// It ensures all parameters are correctly assigned to the resulting Spec.
//
// Params:
//   - t: the testing context
func TestNewSpec_TableDriven(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		params process.SpecParams
	}{
		{
			name: "minimal spec with command only",
			params: process.SpecParams{
				Command: "/bin/echo",
			},
		},
		{
			name: "spec with command and args",
			params: process.SpecParams{
				Command: "/usr/bin/python",
				Args:    []string{"-c", "print('hello')"},
			},
		},
		{
			name: "full spec with all fields",
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

// TestSpec_WithOutput validates output writer attachment to a Spec.
// It ensures stdout and stderr writers are correctly set and the original spec is unchanged.
//
// Params:
//   - t: the testing context
func TestSpec_WithOutput(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		stdout  io.Writer
		stderr  io.Writer
	}{
		{
			name:    "attach buffer writers",
			command: "/bin/echo",
			stdout:  &bytes.Buffer{},
			stderr:  &bytes.Buffer{},
		},
		{
			name:    "attach nil writers",
			command: "/bin/cat",
			stdout:  nil,
			stderr:  nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// Create a base spec.
			spec := process.NewSpec(process.SpecParams{Command: tt.command})

			// Attach output writers.
			specWithOutput := spec.WithOutput(tt.stdout, tt.stderr)

			// Verify writers are attached.
			assert.Equal(t, tt.stdout, specWithOutput.Stdout, "stdout should be set")
			assert.Equal(t, tt.stderr, specWithOutput.Stderr, "stderr should be set")

			// Verify original spec is unchanged (immutability).
			assert.Nil(t, spec.Stdout, "original stdout should be nil")
			assert.Nil(t, spec.Stderr, "original stderr should be nil")
		})
	}
}

// TestSpec_Fields validates direct field access on Spec struct.
// It ensures all struct fields can be accessed directly.
//
// Params:
//   - t: the testing context
func TestSpec_Fields(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		command string
		args    []string
		dir     string
		env     map[string]string
		user    string
		group   string
	}{
		{
			name:    "echo command",
			command: "/bin/echo",
			args:    []string{"hello"},
			dir:     "/tmp",
			env:     map[string]string{"LANG": "en_US.UTF-8"},
			user:    "root",
			group:   "root",
		},
		{
			name:    "sleep command",
			command: "/bin/sleep",
			args:    []string{"1"},
			dir:     "/var",
			env:     nil,
			user:    "nobody",
			group:   "nogroup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			spec := process.Spec{
				Command: tt.command,
				Args:    tt.args,
				Dir:     tt.dir,
				Env:     tt.env,
				User:    tt.user,
				Group:   tt.group,
			}

			assert.Equal(t, tt.command, spec.Command)
			assert.Equal(t, tt.args, spec.Args)
			assert.Equal(t, tt.dir, spec.Dir)
			assert.Equal(t, tt.env, spec.Env)
			assert.Equal(t, tt.user, spec.User)
			assert.Equal(t, tt.group, spec.Group)
		})
	}
}
