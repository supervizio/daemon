//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

import "time"

// AllSystemMetrics contains all metrics that can be collected on the current platform.
// This is used for the --probe CLI command to output comprehensive system information.
type AllSystemMetrics struct {
	// Metadata about the collection
	Timestamp     time.Time `json:"timestamp"`
	Platform      string    `json:"platform"`
	Hostname      string    `json:"hostname,omitempty"`
	OSVersion     string    `json:"os_version"`
	KernelVersion string    `json:"kernel_version"`
	Arch          string    `json:"arch"`
	CollectedAt   int64     `json:"collected_at_us"`

	// Basic system metrics
	CPU    *CPUMetricsJSON    `json:"cpu"`
	Memory *MemoryMetricsJSON `json:"memory"`
	Load   *LoadMetricsJSON   `json:"load"`

	// Disk metrics
	Disk *DiskMetricsJSON `json:"disk"`

	// Network metrics
	Network *NetworkMetricsJSON `json:"network"`

	// I/O metrics
	IO *IOMetricsJSON `json:"io"`

	// Process metrics
	Process *ProcessMetricsJSON `json:"process"`

	// Thermal metrics (Linux only)
	Thermal *ThermalMetricsJSON `json:"thermal"`

	// Context switches (Linux only)
	ContextSwitches *ContextSwitchMetricsJSON `json:"context_switches"`

	// Network connections (Linux only)
	Connections *ConnectionMetricsJSON `json:"connections"`

	// Resource quotas and container info
	Quota     *QuotaMetricsJSON     `json:"quota"`
	Container *ContainerMetricsJSON `json:"container"`
	Runtime   *RuntimeMetricsJSON   `json:"runtime"`
}
