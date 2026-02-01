// Package process provides domain entities and value objects for process lifecycle management.
package process

// Spec contains process execution parameters.
// This is a value object passed to the Executor.
// Note: I/O configuration (stdout/stderr) is handled at the infrastructure layer,
// not in the domain, following hexagonal architecture principles.
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
	// convert params to spec
	return Spec(params)
}
