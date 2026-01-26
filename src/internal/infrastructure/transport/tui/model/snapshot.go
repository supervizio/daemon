// Package model provides data types for the TUI.
package model

import (
	"time"

	"github.com/kodflow/daemon/internal/domain/health"
	"github.com/kodflow/daemon/internal/domain/process"
)

// Snapshot represents a complete state capture for TUI display.
// All fields support "unknown" state for graceful degradation.
type Snapshot struct {
	// Timestamp of this snapshot.
	Timestamp time.Time

	// Context contains runtime environment information.
	Context RuntimeContext

	// Limits contains resource limits (cgroups, ulimits).
	Limits ResourceLimits

	// Services contains supervised service states.
	Services []ServiceSnapshot

	// System contains host-level metrics.
	System SystemMetrics

	// Network contains network interface information.
	Network []NetworkInterface

	// Logs contains log summary (raw mode) or badge (interactive).
	Logs LogSummary

	// Sandboxes contains detected container runtimes.
	Sandboxes []SandboxInfo
}

// RuntimeContext provides environment detection.
type RuntimeContext struct {
	// Hostname of the machine.
	Hostname string
	// OS name (linux, darwin, freebsd).
	OS string
	// Arch (amd64, arm64).
	Arch string
	// Kernel version string.
	Kernel string
	// Mode: host, container, vm.
	Mode RuntimeMode
	// ContainerRuntime if Mode is container (docker, podman, lxc).
	ContainerRuntime string
	// PID of the daemon process.
	DaemonPID int
	// Version of superviz.io.
	Version string
	// StartTime of the daemon.
	StartTime time.Time
	// Uptime calculated from StartTime.
	Uptime time.Duration
	// PrimaryIP is the main IP address.
	PrimaryIP string
	// DNS nameservers.
	DNSServers []string
	// DNS search domains.
	DNSSearch []string
	// ConfigPath is the path to the configuration file.
	ConfigPath string
}

// RuntimeMode represents the execution environment.
type RuntimeMode int

// Runtime mode constants.
const (
	ModeUnknown RuntimeMode = iota
	ModeHost
	ModeContainer
	ModeVM
)

// String returns the string representation of RuntimeMode.
func (m RuntimeMode) String() string {
	switch m {
	case ModeUnknown:
		return "unknown"
	case ModeHost:
		return "host"
	case ModeContainer:
		return "container"
	case ModeVM:
		return "vm"
	}
	return "unknown"
}

// ResourceLimits contains cgroup and ulimit information.
type ResourceLimits struct {
	// CPUQuota as a fraction (e.g., 2.0 = 2 cores).
	CPUQuota float64
	// CPUQuotaRaw is the raw quota value (e.g., 200000).
	CPUQuotaRaw int64
	// CPUPeriod is the cgroup period (e.g., 100000).
	CPUPeriod int64
	// CPUSet is the allowed CPU cores (e.g., "0-3").
	CPUSet string
	// MemoryMax in bytes (0 = unlimited).
	MemoryMax uint64
	// MemoryCurrent in bytes.
	MemoryCurrent uint64
	// PIDsMax is the maximum number of processes.
	PIDsMax int64
	// PIDsCurrent is the current process count.
	PIDsCurrent int64
	// HasLimits indicates if any limits are set.
	HasLimits bool
}

// ServiceSnapshot contains per-service state for display.
type ServiceSnapshot struct {
	// Name of the service.
	Name string
	// State of the process.
	State process.State
	// PID if running (0 otherwise).
	PID int
	// Uptime since last start.
	Uptime time.Duration
	// RestartCount since daemon start.
	RestartCount int
	// LastExitCode if failed.
	LastExitCode int
	// LastError message if failed.
	LastError string
	// Health status.
	Health health.Status
	// HasHealthChecks indicates whether health probes are configured.
	HasHealthChecks bool
	// HealthLatency of last probe.
	HealthLatency time.Duration
	// Listeners with their states.
	Listeners []ListenerSnapshot
	// Ports are the TCP/UDP ports the service is listening on.
	Ports []int
	// CPUPercent usage (0-100).
	CPUPercent float64
	// MemoryRSS in bytes.
	MemoryRSS uint64
	// MemoryPercent of total system memory.
	MemoryPercent float64
}

// PortStatus represents the display status of a port.
type PortStatus int

// Port status constants for coloring.
const (
	// PortStatusOK: port state matches configuration (green).
	PortStatusOK PortStatus = iota
	// PortStatusWarning: mismatch between config and reality (yellow).
	PortStatusWarning
	// PortStatusError: expected port but nothing listening (red).
	PortStatusError
	// PortStatusUnknown: status cannot be determined.
	PortStatusUnknown
)

// ListenerSnapshot contains listener state for display.
type ListenerSnapshot struct {
	// Name of the listener.
	Name string
	// Protocol (tcp, udp).
	Protocol string
	// Port number.
	Port int
	// Address (bind address).
	Address string
	// Exposed indicates if this port should be publicly accessible.
	Exposed bool
	// Listening indicates if the port is actually listening.
	Listening bool
	// Reachable indicates if the port is reachable externally (for exposed ports).
	Reachable bool
	// Status for display coloring.
	Status PortStatus
	// State of the listener (deprecated, use Status).
	State string
	// ProbeType (tcp, http, grpc, exec).
	ProbeType string
	// Latency of last probe.
	Latency time.Duration
}

// SystemMetrics contains host-level resource usage.
type SystemMetrics struct {
	// CPUPercent usage (0-100).
	CPUPercent float64
	// CPUUser percent.
	CPUUser float64
	// CPUSystem percent.
	CPUSystem float64
	// CPUIdle percent.
	CPUIdle float64
	// CPUIOWait percent.
	CPUIOWait float64
	// LoadAvg1 is 1-minute load average.
	LoadAvg1 float64
	// LoadAvg5 is 5-minute load average.
	LoadAvg5 float64
	// LoadAvg15 is 15-minute load average.
	LoadAvg15 float64
	// MemoryTotal in bytes.
	MemoryTotal uint64
	// MemoryUsed in bytes.
	MemoryUsed uint64
	// MemoryAvailable in bytes.
	MemoryAvailable uint64
	// MemoryPercent used (0-100).
	MemoryPercent float64
	// MemoryCached in bytes.
	MemoryCached uint64
	// MemoryBuffers in bytes.
	MemoryBuffers uint64
	// SwapTotal in bytes.
	SwapTotal uint64
	// SwapUsed in bytes.
	SwapUsed uint64
	// SwapPercent used (0-100).
	SwapPercent float64
	// DiskTotal in bytes (root filesystem).
	DiskTotal uint64
	// DiskUsed in bytes.
	DiskUsed uint64
	// DiskAvailable in bytes.
	DiskAvailable uint64
	// DiskPercent used (0-100).
	DiskPercent float64
	// DiskPath is the mount point measured (usually "/").
	DiskPath string
}

// NetworkInterface contains per-interface statistics.
type NetworkInterface struct {
	// Name of the interface (eth0, lo).
	Name string
	// IP address (first IPv4).
	IP string
	// RxBytesPerSec received bytes per second.
	RxBytesPerSec uint64
	// TxBytesPerSec transmitted bytes per second.
	TxBytesPerSec uint64
	// Speed in bits per second (0 = unknown).
	Speed uint64
	// IsUp indicates if interface is up.
	IsUp bool
	// IsLoopback indicates if this is lo.
	IsLoopback bool
}

// LogSummary contains aggregated log information.
type LogSummary struct {
	// Period over which logs are summarized.
	Period time.Duration
	// InfoCount of INFO level logs.
	InfoCount int
	// WarnCount of WARN level logs.
	WarnCount int
	// ErrorCount of ERROR level logs.
	ErrorCount int
	// RecentEntries are the last N log lines (raw mode).
	RecentEntries []LogEntry
	// HasAlerts indicates active alerts.
	HasAlerts bool
}

// LogEntry represents a single log line.
type LogEntry struct {
	// Timestamp of the log.
	Timestamp time.Time
	// Level (INFO, WARN, ERROR).
	Level string
	// Service name.
	Service string
	// EventType categorizes the event (started, stopped, failed, etc.).
	EventType string
	// Message content.
	Message string
	// Metadata contains additional key-value data.
	Metadata map[string]any
}

// SandboxInfo contains detected container runtime information.
type SandboxInfo struct {
	// Name of the runtime (docker, podman, kubernetes, lxc).
	Name string
	// Detected indicates if the runtime was found.
	Detected bool
	// Endpoint is the socket/API path.
	Endpoint string
	// Version if detectable.
	Version string
}

// Default capacities for pre-allocation to reduce allocations.
const (
	defaultNetworkCap   = 6   // Typical system has 2-5 interfaces.
	defaultSandboxCap   = 5   // 4-5 sandbox types (docker, podman, k8s, lxc, systemd).
	defaultLogEntryCap  = 100 // Default log buffer size.
	defaultServicesCap  = 16  // Typical setup has 5-15 services.
	defaultListenersCap = 4   // Typical service has 1-3 listeners.
)

// NewSnapshot creates an empty snapshot with current timestamp.
// Uses pre-allocated capacities for typical workloads.
func NewSnapshot() *Snapshot {
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
func NewSnapshotWithCapacity(serviceCount int) *Snapshot {
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
func (s *Snapshot) ServiceCount() int {
	return len(s.Services)
}

// RunningCount returns number of running services.
func (s *Snapshot) RunningCount() int {
	count := 0
	for _, svc := range s.Services {
		if svc.State == process.StateRunning {
			count++
		}
	}
	return count
}

// FailedCount returns number of failed services.
func (s *Snapshot) FailedCount() int {
	count := 0
	for _, svc := range s.Services {
		if svc.State == process.StateFailed {
			count++
		}
	}
	return count
}

// HealthyCount returns number of healthy services.
func (s *Snapshot) HealthyCount() int {
	count := 0
	for _, svc := range s.Services {
		if svc.Health == health.StatusHealthy {
			count++
		}
	}
	return count
}
