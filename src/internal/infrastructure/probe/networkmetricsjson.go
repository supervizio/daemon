//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// NetworkMetricsJSON contains network-related metrics for JSON output.
// Uses types from collect_all.go: NetInterfaceInfo, NetStatsInfo.
type NetworkMetricsJSON struct {
	Interfaces []NetInterfaceJSON `json:"interfaces,omitempty"`
	Stats      []NetStatsJSON     `json:"stats,omitempty"`
}
