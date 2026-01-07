// Package process_test provides black-box tests for the spec_params.go file.
// These tests validate the SpecParams struct behavior.
package process_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kodflow/daemon/internal/domain/process"
)

// TestSpecParams_Fields validates direct field access on SpecParams struct.
// It ensures all struct fields can be accessed and contain expected values.
//
// Params:
//   - t: the testing context
func TestSpecParams_Fields(t *testing.T) {
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
			name:    "minimal params",
			command: "/bin/echo",
			args:    nil,
			dir:     "",
			env:     nil,
			user:    "",
			group:   "",
		},
		{
			name:    "full params",
			command: "/usr/bin/python",
			args:    []string{"-c", "print('hello')"},
			dir:     "/tmp",
			env:     map[string]string{"PATH": "/usr/bin"},
			user:    "nobody",
			group:   "nogroup",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			params := process.SpecParams{
				Command: tt.command,
				Args:    tt.args,
				Dir:     tt.dir,
				Env:     tt.env,
				User:    tt.user,
				Group:   tt.group,
			}

			assert.Equal(t, tt.command, params.Command)
			assert.Equal(t, tt.args, params.Args)
			assert.Equal(t, tt.dir, params.Dir)
			assert.Equal(t, tt.env, params.Env)
			assert.Equal(t, tt.user, params.User)
			assert.Equal(t, tt.group, params.Group)
		})
	}
}
