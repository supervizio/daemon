//go:build cgo

package probe

// MemoryMetricsJSON contains memory-related metrics for JSON output.
// It includes total, available, used, cached, and swap memory statistics.
type MemoryMetricsJSON struct {
	TotalBytes     uint64              `json:"total_bytes"`
	AvailableBytes uint64              `json:"available_bytes"`
	UsedBytes      uint64              `json:"used_bytes"`
	CachedBytes    uint64              `json:"cached_bytes"`
	BuffersBytes   uint64              `json:"buffers_bytes"`
	SwapTotalBytes uint64              `json:"swap_total_bytes"`
	SwapUsedBytes  uint64              `json:"swap_used_bytes"`
	UsedPercent    float64             `json:"used_percent"`
	Pressure       *MemoryPressureJSON `json:"pressure,omitempty"`
}
