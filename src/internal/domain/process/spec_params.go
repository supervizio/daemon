// Package process provides domain entities and value objects for process lifecycle management.
package process

// SpecParams contains the configuration parameters for creating a process Spec.
// This groups related parameters to simplify the NewSpec function signature.
type SpecParams struct {
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
