//go:build cgo

package probe

// ContextSwitchMetricsJSON contains context switch statistics.
// It tracks system-wide and per-process context switch counts.
type ContextSwitchMetricsJSON struct {
	SystemTotal uint64                 `json:"system_total"`
	Self        *ContextSwitchInfoJSON `json:"self,omitempty"`
}
