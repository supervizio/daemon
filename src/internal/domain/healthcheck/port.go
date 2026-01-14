// Package probe provides domain abstractions for service probing.
// It defines the Prober interface and related types for health checking
// across multiple protocols (TCP, UDP, HTTP, gRPC, exec, ICMP).
package healthcheck

import "context"

// Prober executes probes against targets.
// This is the domain port that infrastructure adapters implement
// to provide specific probing mechanisms for different protocols.
type Prober interface {
	// Probe executes a probe against the target.
	//
	// Params:
	//   - ctx: the context for cancellation and timeout control.
	//   - target: the target to probe.
	//
	// Returns:
	//   - Result: the result of the probe including latency and output.
	Probe(ctx context.Context, target Target) Result

	// Type returns the probe type (tcp, udp, http, grpc, exec, icmp).
	//
	// Returns:
	//   - string: the type identifier for this prober.
	Type() string
}
