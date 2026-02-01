//go:build cgo

package probe

// QuotaMetricsJSON contains resource quota information.
// It indicates support status and provides limit and usage data.
type QuotaMetricsJSON struct {
	Supported bool           `json:"supported"`
	Limits    *QuotaInfoJSON `json:"limits,omitempty"`
	Usage     *UsageInfoJSON `json:"usage,omitempty"`
}
