//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// QuotaMetricsJSON contains resource quota information.
// It indicates support status and provides limit and usage data.
type QuotaMetricsJSON struct {
	Supported bool           `json:"supported"`
	Limits    *QuotaInfoJSON `json:"limits,omitempty"`
	Usage     *UsageInfoJSON `json:"usage,omitempty"`
}
