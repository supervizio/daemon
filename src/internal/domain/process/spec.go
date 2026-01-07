// Package process provides domain entities and value objects for process lifecycle management.
package process

import "io"

// Spec contains process execution parameters.
// This is a value object passed to the Executor.
type Spec struct {
	// Command is the executable path or command to run.
	Command string
	// Args contains command-line arguments.
	Args []string
	// Dir is the working directory.
	Dir string
	// Env contains environment variables as key=value pairs.
	Env map[string]string
	// User specifies the username to run as.
	User string
	// Group specifies the group to run as.
	Group string
	// Stdout is the writer for standard output.
	Stdout io.Writer
	// Stderr is the writer for standard error.
	Stderr io.Writer
}

// NewSpec creates a new process specification from configuration parameters.
// It initializes a Spec with the provided execution parameters.
//
// Params:
//   - params: the configuration parameters for the process
//
// Returns:
//   - Spec: a configured process specification ready for execution
func NewSpec(params SpecParams) Spec {
	// Build and return the specification with all provided parameters.
	return Spec{
		Command: params.Command,
		Args:    params.Args,
		Dir:     params.Dir,
		Env:     params.Env,
		User:    params.User,
		Group:   params.Group,
	}
}

// WithOutput returns a copy of the spec with stdout and stderr set.
// This method creates a new Spec with the output writers configured.
//
// Params:
//   - stdout: writer for standard output
//   - stderr: writer for standard error
//
// Returns:
//   - Spec: a new specification with output writers set
func (s Spec) WithOutput(stdout, stderr io.Writer) Spec {
	s.Stdout = stdout
	s.Stderr = stderr

	// Return the modified copy with output writers attached.
	return s
}
