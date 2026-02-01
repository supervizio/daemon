//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// DiskMetricsJSON contains disk-related metrics for JSON output.
// Uses types from collect_all.go: PartitionInfo, DiskUsageInfo, DiskIOInfo.
type DiskMetricsJSON struct {
	Partitions []PartitionInfo `json:"partitions,omitempty"`
	Usage      []DiskUsageInfo `json:"usage,omitempty"`
	IO         []DiskIOInfo    `json:"io,omitempty"`
}
