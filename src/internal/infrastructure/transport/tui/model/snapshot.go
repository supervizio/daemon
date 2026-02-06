// Package model provides data types for the TUI.
package model

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
)

// Default capacities for pre-allocation to reduce allocations.
const (
	// defaultNetworkCap is the typical number of network interfaces (2-5).
	defaultNetworkCap int = 6
	// defaultSandboxCap is the number of sandbox types (docker, podman, k8s, lxc, systemd).
	defaultSandboxCap int = 5
	// defaultLogEntryCap is the default log buffer size.
	defaultLogEntryCap int = 100
	// defaultServicesCap is the typical number of services (5-15).
	defaultServicesCap int = 16
)

// RuntimeMode represents the execution environment.
type RuntimeMode int

// Runtime mode constants.
const (
	// ModeUnknown indicates the runtime mode could not be determined.
	ModeUnknown RuntimeMode = iota
	// ModeHost indicates running directly on the host.
	ModeHost
	// ModeContainer indicates running inside a container.
	ModeContainer
	// ModeVM indicates running inside a virtual machine.
	ModeVM
)

// PortStatus represents the display status of a port.
type PortStatus int

// Port status constants for coloring.
const (
	// PortStatusOK indicates port state matches configuration (green).
	PortStatusOK PortStatus = iota
	// PortStatusWarning indicates mismatch between config and reality (yellow).
	PortStatusWarning
	// PortStatusError indicates expected port but nothing listening (red).
	PortStatusError
	// PortStatusUnknown indicates status cannot be determined.
	PortStatusUnknown
)

// Snapshot represents a complete state capture for TUI display.
// All fields support "unknown" state for graceful degradation.
type Snapshot struct {
	// Timestamp of this snapshot.
	Timestamp time.Time `json:"timestamp"`

	// Context contains runtime environment information.
	Context RuntimeContext `json:"context"`

	// Limits contains resource limits (cgroups, ulimits).
	Limits ResourceLimits `json:"limits"`

	// Services contains supervised service states.
	Services []ServiceSnapshot `json:"services"`

	// System contains host-level metrics.
	System SystemMetrics `json:"system"`

	// Network contains network interface information.
	Network []NetworkInterface `json:"network"`

	// Logs contains log summary (raw mode) or badge (interactive).
	Logs LogSummary `json:"logs"`

	// Sandboxes contains detected container runtimes.
	Sandboxes []SandboxInfo `json:"sandboxes"`
}

// RuntimeContext provides environment detection.
// It contains information about the host, OS, kernel, runtime mode, and daemon state.
type RuntimeContext struct {
	// Hostname of the machine.
	Hostname string `json:"hostname" dto:"out,priv,pub"`
	// OS name (linux, darwin, freebsd).
	OS string `json:"os"`
	// Arch (amd64, arm64).
	Arch string `json:"arch"`
	// Kernel version string.
	Kernel string `json:"kernel"`
	// Mode: host, container, vm.
	Mode RuntimeMode `json:"mode"`
	// ContainerRuntime if Mode is container (docker, podman, lxc).
	ContainerRuntime string `json:"container_runtime"`
	// PID of the daemon process.
	DaemonPID int `json:"daemon_pid"`
	// Version of superviz.io.
	Version string `json:"version"`
	// StartTime of the daemon.
	StartTime time.Time `json:"start_time"`
	// Uptime calculated from StartTime.
	Uptime time.Duration `json:"uptime"`
	// PrimaryIP is the main IP address.
	PrimaryIP string `json:"primary_ip"`
	// DNS nameservers.
	DNSServers []string `json:"dns_servers"`
	// DNS search domains.
	DNSSearch []string `json:"dns_search"`
	// ConfigPath is the path to the configuration file.
	ConfigPath string `json:"config_path"`
}

// String returns the string representation of RuntimeMode.
//
// Returns:
//   - string: human-readable name of the runtime mode
func (m RuntimeMode) String() string {
	// Match the mode to its string representation.
	switch m {
	// Unknown mode returns "unknown".
	case ModeUnknown:
		// Return unknown string.
		return "unknown"
	// Host mode returns "host".
	case ModeHost:
		// Return host string.
		return "host"
	// Container mode returns "container".
	case ModeContainer:
		// Return container string.
		return "container"
	// VM mode returns "vm".
	case ModeVM:
		// Return vm string.
		return "vm"
	}

	// Default fallback for any unrecognized mode.
	return "unknown"
}

// ResourceLimits contains cgroup and ulimit information.
// It tracks CPU quotas, memory limits, and process limits from cgroups v2.
type ResourceLimits struct {
	// CPUQuota as a fraction (e.g., 2.0 = 2 cores).
	CPUQuota float64 `json:"cpu_quota" dto:"out,priv,pub"`
	// CPUQuotaRaw is the raw quota value (e.g., 200000).
	CPUQuotaRaw int64 `json:"cpu_quota_raw"`
	// CPUPeriod is the cgroup period (e.g., 100000).
	CPUPeriod int64 `json:"cpu_period"`
	// CPUSet is the allowed CPU cores (e.g., "0-3").
	CPUSet string `json:"cpu_set"`
	// MemoryMax in bytes (0 = unlimited).
	MemoryMax uint64 `json:"memory_max"`
	// MemoryCurrent in bytes.
	MemoryCurrent uint64 `json:"memory_current"`
	// PIDsMax is the maximum number of processes.
	PIDsMax int64 `json:"pids_max"`
	// PIDsCurrent is the current process count.
	PIDsCurrent int64 `json:"pids_current"`
	// HasLimits indicates if any limits are set.
	HasLimits bool `json:"has_limits"`
}

// ServiceSnapshot contains per-service state for display.
// It includes process state, health status, resource usage, and listener information.
type ServiceSnapshot struct {
	// Name of the service.
	Name string `json:"name" dto:"out,priv,pub"`
	// State of the process.
	State process.State `json:"state"`
	// PID if running (0 otherwise).
	PID int `json:"pid"`
	// Uptime since last start.
	Uptime time.Duration `json:"uptime"`
	// RestartCount since daemon start.
	RestartCount int `json:"restart_count"`
	// LastExitCode if failed.
	LastExitCode int `json:"last_exit_code"`
	// LastError message if failed.
	LastError string `json:"last_error"`
	// Health status.
	Health health.Status `json:"health"`
	// HasHealthChecks indicates whether health probes are configured.
	HasHealthChecks bool `json:"has_health_checks"`
	// HealthLatency of last probe.
	HealthLatency time.Duration `json:"health_latency"`
	// Listeners with their states.
	Listeners []ListenerSnapshot `json:"listeners"`
	// Ports are the TCP/UDP ports the service is listening on.
	Ports []int `json:"ports"`
	// CPUPercent usage (0-100).
	CPUPercent float64 `json:"cpu_percent"`
	// MemoryRSS in bytes.
	MemoryRSS uint64 `json:"memory_rss"`
	// MemoryPercent of total system memory.
	MemoryPercent float64 `json:"memory_percent"`
}

// ListenerSnapshot contains listener state for display.
// It tracks whether ports are listening, reachable, and match their configuration.
type ListenerSnapshot struct {
	// Name of the listener.
	Name string `json:"name" dto:"out,priv,pub"`
	// Protocol (tcp, udp).
	Protocol string `json:"protocol"`
	// Port number.
	Port int `json:"port"`
	// Address (bind address).
	Address string `json:"address"`
	// Exposed indicates if this port should be publicly accessible.
	Exposed bool `json:"exposed"`
	// Listening indicates if the port is actually listening.
	Listening bool `json:"listening"`
	// Reachable indicates if the port is reachable externally (for exposed ports).
	Reachable bool `json:"reachable"`
	// Status for display coloring.
	Status PortStatus `json:"status"`
	// State of the listener (deprecated, use Status).
	State string `json:"state"`
	// ProbeType (tcp, http, grpc, exec).
	ProbeType string `json:"probe_type"`
	// Latency of last probe.
	Latency time.Duration `json:"latency"`
}

// SystemMetrics contains host-level resource usage.
// It includes CPU, memory, swap, disk, and load average metrics.
type SystemMetrics struct {
	// CPUPercent usage (0-100).
	CPUPercent float64 `json:"cpu_percent" dto:"out,priv,pub"`
	// CPUUser percent.
	CPUUser float64 `json:"cpu_user"`
	// CPUSystem percent.
	CPUSystem float64 `json:"cpu_system"`
	// CPUIdle percent.
	CPUIdle float64 `json:"cpu_idle"`
	// CPUIOWait percent.
	CPUIOWait float64 `json:"cpu_io_wait"`
	// LoadAvg1 is 1-minute load average.
	LoadAvg1 float64 `json:"load_avg_1"`
	// LoadAvg5 is 5-minute load average.
	LoadAvg5 float64 `json:"load_avg_5"`
	// LoadAvg15 is 15-minute load average.
	LoadAvg15 float64 `json:"load_avg_15"`
	// MemoryTotal in bytes.
	MemoryTotal uint64 `json:"memory_total"`
	// MemoryUsed in bytes.
	MemoryUsed uint64 `json:"memory_used"`
	// MemoryAvailable in bytes.
	MemoryAvailable uint64 `json:"memory_available"`
	// MemoryPercent used (0-100).
	MemoryPercent float64 `json:"memory_percent"`
	// MemoryCached in bytes.
	MemoryCached uint64 `json:"memory_cached"`
	// MemoryBuffers in bytes.
	MemoryBuffers uint64 `json:"memory_buffers"`
	// SwapTotal in bytes.
	SwapTotal uint64 `json:"swap_total"`
	// SwapUsed in bytes.
	SwapUsed uint64 `json:"swap_used"`
	// SwapPercent used (0-100).
	SwapPercent float64 `json:"swap_percent"`
	// DiskTotal in bytes (root filesystem).
	DiskTotal uint64 `json:"disk_total"`
	// DiskUsed in bytes.
	DiskUsed uint64 `json:"disk_used"`
	// DiskAvailable in bytes.
	DiskAvailable uint64 `json:"disk_available"`
	// DiskPercent used (0-100).
	DiskPercent float64 `json:"disk_percent"`
	// DiskPath is the mount point measured (usually "/").
	DiskPath string `json:"disk_path"`
}

// NetworkInterface contains per-interface statistics.
// It tracks IP address, bandwidth (Rx/Tx bytes per second), and interface status.
type NetworkInterface struct {
	// Name of the interface (eth0, lo).
	Name string `json:"name" dto:"out,priv,pub"`
	// IP address (first IPv4).
	IP string `json:"ip"`
	// RxBytesPerSec received bytes per second.
	RxBytesPerSec uint64 `json:"rx_bytes_per_sec"`
	// TxBytesPerSec transmitted bytes per second.
	TxBytesPerSec uint64 `json:"tx_bytes_per_sec"`
	// Speed in bits per second (0 = unknown).
	Speed uint64 `json:"speed"`
	// IsUp indicates if interface is up.
	IsUp bool `json:"is_up"`
	// IsLoopback indicates if this is lo.
	IsLoopback bool `json:"is_loopback"`
}

// LogSummary contains aggregated log information.
// It provides log counts by level and recent entries for the TUI display.
type LogSummary struct {
	// Period over which logs are summarized.
	Period time.Duration `json:"period" dto:"out,priv,pub"`
	// InfoCount of INFO level logs.
	InfoCount int `json:"info_count"`
	// WarnCount of WARN level logs.
	WarnCount int `json:"warn_count"`
	// ErrorCount of ERROR level logs.
	ErrorCount int `json:"error_count"`
	// RecentEntries are the last N log lines (raw mode).
	RecentEntries []LogEntry `json:"recent_entries"`
	// HasAlerts indicates active alerts.
	HasAlerts bool `json:"has_alerts"`
}

// LogEntry represents a single log line.
// It contains timestamp, level, service, event type, message, and optional metadata.
type LogEntry struct {
	// Timestamp of the log.
	Timestamp time.Time `json:"timestamp" dto:"out,priv,pub"`
	// Level (INFO, WARN, ERROR).
	Level string `json:"level"`
	// Service name.
	Service string `json:"service"`
	// EventType categorizes the event (started, stopped, failed, etc.).
	EventType string `json:"event_type"`
	// Message content.
	Message string `json:"message"`
	// Metadata contains additional key-value data.
	Metadata map[string]any `json:"metadata"`
}

// SandboxInfo contains detected container runtime information.
// It indicates whether a runtime (Docker, Podman, Kubernetes, etc.) was detected.
type SandboxInfo struct {
	// Name of the runtime (docker, podman, kubernetes, lxc).
	Name string `json:"name" dto:"out,priv,pub"`
	// Detected indicates if the runtime was found.
	Detected bool `json:"detected"`
	// Endpoint is the socket/API path.
	Endpoint string `json:"endpoint"`
	// Version if detectable.
	Version string `json:"version"`
}

// NewSnapshot creates an empty snapshot with current timestamp.
// Uses pre-allocated capacities for typical workloads.
//
// Returns:
//   - *Snapshot: new snapshot instance with pre-allocated slices
func NewSnapshot() *Snapshot {
	// Return a new snapshot with default capacities.
	return &Snapshot{
		Timestamp: time.Now(),
		Services:  make([]ServiceSnapshot, 0, defaultServicesCap),
		Network:   make([]NetworkInterface, 0, defaultNetworkCap),
		Sandboxes: make([]SandboxInfo, 0, defaultSandboxCap),
		Logs: LogSummary{
			RecentEntries: make([]LogEntry, 0, defaultLogEntryCap),
		},
	}
}

// NewSnapshotWithCapacity creates a snapshot with specified service capacity.
// Use this when the number of services is known ahead of time.
//
// Params:
//   - serviceCount: number of services to pre-allocate capacity for
//
// Returns:
//   - *Snapshot: new snapshot instance with custom service capacity
func NewSnapshotWithCapacity(serviceCount int) *Snapshot {
	// Return a new snapshot with the specified service capacity.
	return &Snapshot{
		Timestamp: time.Now(),
		Services:  make([]ServiceSnapshot, 0, serviceCount),
		Network:   make([]NetworkInterface, 0, defaultNetworkCap),
		Sandboxes: make([]SandboxInfo, 0, defaultSandboxCap),
		Logs: LogSummary{
			RecentEntries: make([]LogEntry, 0, defaultLogEntryCap),
		},
	}
}

// ServiceCount returns total number of services.
//
// Returns:
//   - int: total count of services in the snapshot
func (s *Snapshot) ServiceCount() int {
	// Return the length of the services slice.
	return len(s.Services)
}

// RunningCount returns number of running services.
//
// Returns:
//   - int: count of services in running state
func (s *Snapshot) RunningCount() int {
	count := 0

	// Iterate through all services to count running ones.
	for i := range s.Services {
		// Check if the service is in running state.
		if s.Services[i].State == process.StateRunning {
			count++
		}
	}

	// Return the final running count.
	return count
}

// FailedCount returns number of failed services.
//
// Returns:
//   - int: count of services in failed state
func (s *Snapshot) FailedCount() int {
	count := 0

	// Iterate through all services to count failed ones.
	for i := range s.Services {
		// Check if the service is in failed state.
		if s.Services[i].State == process.StateFailed {
			count++
		}
	}

	// Return the final failed count.
	return count
}

// HealthyCount returns number of healthy services.
//
// Returns:
//   - int: count of services with healthy status
func (s *Snapshot) HealthyCount() int {
	count := 0

	// Iterate through all services to count healthy ones.
	for i := range s.Services {
		// Check if the service has healthy status.
		if s.Services[i].Health == health.StatusHealthy {
			count++
		}
	}

	// Return the final healthy count.
	return count
}
