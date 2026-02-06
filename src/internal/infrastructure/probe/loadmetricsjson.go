//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// LoadMetricsJSON contains load average metrics for JSON output.
// It provides 1, 5, and 15 minute system load averages.
type LoadMetricsJSON struct {
	Load1Min  float64 `json:"load_1min"`
	Load5Min  float64 `json:"load_5min"`
	Load15Min float64 `json:"load_15min"`
}
