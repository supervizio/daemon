//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

import "time"

// AllSystemMetrics contains all metrics that can be collected on the current platform.
// This is used for the --probe CLI command to output comprehensive system information.
type AllSystemMetrics struct {
	// Metadata about the collection
	Timestamp   time.Time `json:"timestamp"`
	Platform    string    `json:"platform"`
	Hostname    string    `json:"hostname,omitempty"`
	CollectedAt int64     `json:"collected_at_ns"`

	// Basic system metrics
	CPU    *CPUMetricsJSON    `json:"cpu,omitempty"`
	Memory *MemoryMetricsJSON `json:"memory,omitempty"`
	Load   *LoadMetricsJSON   `json:"load,omitempty"`

	// Disk metrics
	Disk *DiskMetricsJSON `json:"disk,omitempty"`

	// Network metrics
	Network *NetworkMetricsJSON `json:"network,omitempty"`

	// I/O metrics
	IO *IOMetricsJSON `json:"io,omitempty"`

	// Process metrics
	Process *ProcessMetricsJSON `json:"process,omitempty"`

	// Thermal metrics (Linux only)
	Thermal *ThermalMetricsJSON `json:"thermal,omitempty"`

	// Context switches (Linux only)
	ContextSwitches *ContextSwitchMetricsJSON `json:"context_switches,omitempty"`

	// Network connections (Linux only)
	Connections *ConnectionMetricsJSON `json:"connections,omitempty"`

	// Resource quotas and container info
	Quota     *QuotaMetricsJSON     `json:"quota,omitempty"`
	Container *ContainerMetricsJSON `json:"container,omitempty"`
	Runtime   *RuntimeMetricsJSON   `json:"runtime,omitempty"`
}
