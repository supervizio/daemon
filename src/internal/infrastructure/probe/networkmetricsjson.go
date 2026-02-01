//go:build cgo

package probe

// NetworkMetricsJSON contains network-related metrics for JSON output.
// Uses types from collect_all.go: NetInterfaceInfo, NetStatsInfo.
type NetworkMetricsJSON struct {
	Interfaces []NetInterfaceJSON `json:"interfaces,omitempty"`
	Stats      []NetStatsJSON     `json:"stats,omitempty"`
}
