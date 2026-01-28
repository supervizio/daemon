// Package process_test provides black-box tests for the spec.go file.
// These tests validate the public API behavior of Spec without accessing internal state.
package process_test

import (
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
