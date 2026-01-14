// Package probe provides domain abstractions for service probing.
package healthcheck

// Target represents the target of a probe.
// It contains all information needed to probe different types of services
// including network addresses, paths, and commands.
type Target struct {
	// Network specifies the network protocol.
	// Supported values: "tcp", "tcp4", "tcp6", "udp", "udp4", "udp6".
	Network string

	// Address is the target address in host:port format.
	// Examples: "localhost:8080", "192.168.1.1:50051".
	Address string

	// Path is the HTTP endpoint path for HTTP probes.
	// Example: "/health", "/ready".
	Path string

	// Service is the gRPC service name for gRPC health probes.
	// Empty string means check the server's overall health.
	// Example: "myapp.v1.UserService".
	Service string

	// Command is the command to execute for exec probes.
	// Example: "/app/health.sh".
	Command string

	// Args contains the command arguments for exec probes.
	Args []string

	// Method is the HTTP method for HTTP probes.
	// Example: "GET", "HEAD".
	Method string

	// StatusCode is the expected HTTP status code for HTTP probes.
	// Default is 200 if not specified.
	StatusCode int
}

// NewTarget creates a new probe target with the specified network and address.
//
// Params:
//   - network: the network protocol (e.g., "tcp", "udp", "icmp").
//   - address: the target address in host:port format.
//
// Returns:
//   - Target: a target configured with the specified network and address.
func NewTarget(network, address string) Target {
	// Return target with basic configuration.
	return Target{
		Network: network,
		Address: address,
	}
}

// NewTCPTarget creates a target for TCP probes.
//
// Params:
//   - address: the target address in host:port format.
//
// Returns:
//   - Target: a target configured for TCP probing.
func NewTCPTarget(address string) Target {
	// Return target with TCP network.
	return Target{
		Network: "tcp",
		Address: address,
	}
}

// NewUDPTarget creates a target for UDP probes.
//
// Params:
//   - address: the target address in host:port format.
//
// Returns:
//   - Target: a target configured for UDP probing.
func NewUDPTarget(address string) Target {
	// Return target with UDP network.
	return Target{
		Network: "udp",
		Address: address,
	}
}

// NewHTTPTarget creates a target for HTTP probes.
//
// Params:
//   - address: the full URL to probe.
//   - method: the HTTP method to use.
//   - statusCode: the expected HTTP status code.
//
// Returns:
//   - Target: a target configured for HTTP probing.
func NewHTTPTarget(address, method string, statusCode int) Target {
	// Return target with HTTP configuration.
	return Target{
		Network:    "tcp",
		Address:    address,
		Method:     method,
		StatusCode: statusCode,
	}
}

// NewGRPCTarget creates a target for gRPC health probes.
//
// Params:
//   - address: the gRPC server address in host:port format.
//   - service: the service name to check (empty for overall health).
//
// Returns:
//   - Target: a target configured for gRPC health probing.
func NewGRPCTarget(address, service string) Target {
	// Return target with gRPC configuration.
	return Target{
		Network: "tcp",
		Address: address,
		Service: service,
	}
}

// NewExecTarget creates a target for command execution probes.
//
// Params:
//   - command: the command to execute.
//   - args: the command arguments.
//
// Returns:
//   - Target: a target configured for exec probing.
func NewExecTarget(command string, args ...string) Target {
	// Return target with exec configuration.
	return Target{
		Command: command,
		Args:    args,
	}
}

// NewICMPTarget creates a target for ICMP ping probes.
//
// Params:
//   - address: the target IP address or hostname.
//
// Returns:
//   - Target: a target configured for ICMP probing.
func NewICMPTarget(address string) Target {
	// Return target with ICMP configuration.
	return Target{
		Network: "icmp",
		Address: address,
	}
}
