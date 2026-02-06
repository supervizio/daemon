// Package config provides domain value objects for service configuration.
package config

// NomadDiscoveryConfig configures Nomad allocation discovery.
// Nomad discovery monitors allocations in a Nomad cluster.
type NomadDiscoveryConfig struct {
	// Enabled activates Nomad discovery.
	Enabled bool

	// Address is the Nomad API address.
	// Default: "http://localhost:4646".
	Address string

	// Namespace limits discovery to a specific namespace.
	// Empty means default namespace.
	Namespace string

	// JobFilter filters by job name pattern.
	JobFilter string
}
