// Package collector provides data collectors for TUI snapshot.
package collector

// dnsConfigResult is a private result struct for sync.OnceValue.
type dnsConfigResult struct {
	servers []string
	search  []string
}
