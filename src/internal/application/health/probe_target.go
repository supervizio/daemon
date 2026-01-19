// Package health provides health monitoring for services.
package health

// ProbeTarget defines the target for a health probe.
// It contains all necessary information to execute different types of probes (HTTP, gRPC, exec, etc.).
type ProbeTarget struct {
	// Address is the target address (host:port).
	Address string
	// Path is the HTTP path (for HTTP probes).
	Path string
	// Service is the gRPC service name (for gRPC probes).
	Service string
	// Method is the HTTP method (for HTTP probes).
	Method string
	// StatusCode is the expected HTTP status code.
	StatusCode int
	// Command is the command to execute (for exec probes).
	Command string
	// Args are the command arguments (for exec probes).
	Args []string
}
