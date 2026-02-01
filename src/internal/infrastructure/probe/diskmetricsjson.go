//go:build cgo

package probe

// DiskMetricsJSON contains disk-related metrics for JSON output.
// Uses types from collect_all.go: PartitionInfo, DiskUsageInfo, DiskIOInfo.
type DiskMetricsJSON struct {
	Partitions []PartitionInfo `json:"partitions,omitempty"`
	Usage      []DiskUsageInfo `json:"usage,omitempty"`
	IO         []DiskIOInfo    `json:"io,omitempty"`
}
