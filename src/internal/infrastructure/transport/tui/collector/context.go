// Package collector provides data collectors for TUI snapshot.
package collector

import (
	"os"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

const (
	// minDNSFields is the minimum number of fields in resolv.conf lines.
	minDNSFields int = 2
	// unknownValue is the default string for unknown fields.
	unknownValue string = "unknown"
)

// Cached static system data - computed once per process lifetime.
var (
	cachedHostname    func() string            = sync.OnceValue(getHostnameOnce)
	cachedKernel      func() string            = sync.OnceValue(getKernelVersion)
	cachedRuntimeMode func() runtimeModeResult = sync.OnceValue(detectRuntimeModeOnce)
	cachedDNSConfig   func() dnsConfigResult   = sync.OnceValue(getDNSConfigOnce)
	cachedPrimaryIP   func() string            = sync.OnceValue(getPrimaryIP)
)

// runtimeModeResult holds the cached runtime mode detection result.
type runtimeModeResult struct {
	mode    model.RuntimeMode
	runtime string
}

// dnsConfigResult holds the cached DNS configuration.
type dnsConfigResult struct {
	servers []string
	search  []string
}

// getHostnameOnce returns the hostname (called once).
//
// Returns:
//   - string: hostname or "unknown"
func getHostnameOnce() string {
	// Try to get hostname from OS.
	if hostname, err := os.Hostname(); err == nil {
		// Successfully retrieved hostname.
		return hostname
	}
	// Failed to get hostname.
	return unknownValue
}

// detectRuntimeModeOnce wraps detectRuntimeMode for sync.OnceValue.
//
// Returns:
//   - runtimeModeResult: detected runtime mode and runtime name
func detectRuntimeModeOnce() runtimeModeResult {
	mode, rt := detectRuntimeMode()
	// Wrap in result struct.
	return runtimeModeResult{mode: mode, runtime: rt}
}

// getDNSConfigOnce wraps getDNSConfig for sync.OnceValue.
//
// Returns:
//   - dnsConfigResult: DNS servers and search domains
func getDNSConfigOnce() dnsConfigResult {
	servers, search := getDNSConfig()
	// Wrap in result struct.
	return dnsConfigResult{servers: servers, search: search}
}

// ContextCollector gathers runtime context information.
// It collects both static (cached) and dynamic system data.
type ContextCollector struct {
	version    string
	startTime  time.Time
	configPath string
}

// NewContextCollector creates a context collector.
//
// Params:
//   - version: daemon version string
//
// Returns:
//   - *ContextCollector: configured context collector
func NewContextCollector(version string) *ContextCollector {
	// Initialize with current time as start time.
	return &ContextCollector{
		version:   version,
		startTime: time.Now(),
	}
}

// SetConfigPath sets the configuration file path.
//
// Params:
//   - path: configuration file path
func (c *ContextCollector) SetConfigPath(path string) {
	c.configPath = path
}

// CollectInto populates context information.
// Uses cached values for static data (hostname, kernel, runtime mode, DNS, IP).
//
// Params:
//   - snap: target snapshot to populate
//
// Returns:
//   - error: always returns nil
func (c *ContextCollector) CollectInto(snap *model.Snapshot) error {
	ctx := &snap.Context

	// Basic info (dynamic).
	ctx.Version = c.version
	ctx.StartTime = c.startTime
	ctx.Uptime = time.Since(c.startTime)
	ctx.DaemonPID = os.Getpid()
	ctx.OS = runtime.GOOS
	ctx.Arch = runtime.GOARCH

	// Static info (cached - computed once per process lifetime).
	ctx.Hostname = cachedHostname()
	ctx.Kernel = cachedKernel()

	// Runtime mode (cached).
	rtResult := cachedRuntimeMode()
	ctx.Mode = rtResult.mode
	ctx.ContainerRuntime = rtResult.runtime

	// DNS info (cached).
	dnsResult := cachedDNSConfig()
	ctx.DNSServers = dnsResult.servers
	ctx.DNSSearch = dnsResult.search

	// Primary IP (cached).
	ctx.PrimaryIP = cachedPrimaryIP()

	// Config path.
	ctx.ConfigPath = c.configPath

	// Always return nil for graceful operation.
	return nil
}

// detectRuntimeMode determines if we're in a container, VM, or host.
// Reads /proc/1/cgroup only once to avoid duplicate I/O.
//
// Returns:
//   - model.RuntimeMode: detected runtime mode (container, VM, or host)
//   - string: container runtime name (docker, kubernetes, etc.) or empty
func detectRuntimeMode() (model.RuntimeMode, string) {
	// Check for container indicators.

	// Docker/containerd.
	if fileExists("/.dockerenv") {
		// Docker environment file found.
		return model.ModeContainer, "docker"
	}

	// Kubernetes (check for service account).
	if fileExists("/var/run/secrets/kubernetes.io/serviceaccount/token") {
		// Kubernetes service account token found.
		return model.ModeContainer, "kubernetes"
	}

	// LXC.
	if content, err := os.ReadFile("/proc/1/environ"); err == nil {
		// Check for LXC container marker.
		if strings.Contains(string(content), "container=lxc") {
			// LXC container detected.
			return model.ModeContainer, "lxc"
		}
	}

	// Read /proc/1/cgroup ONCE and check all patterns.
	if content, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		cgroupContent := string(content)

		// Podman detection.
		if strings.Contains(cgroupContent, "libpod") || strings.Contains(cgroupContent, "podman") {
			// Podman cgroup markers found.
			return model.ModeContainer, "podman"
		}

		// Docker/containerd detection.
		if strings.Contains(cgroupContent, "docker") || strings.Contains(cgroupContent, "containerd") {
			// Docker/containerd cgroup markers found.
			return model.ModeContainer, "docker"
		}

		// Generic container detection: non-root cgroup suggests container.
		if !strings.Contains(cgroupContent, ":/\n") {
			// Non-root cgroup detected.
			return model.ModeContainer, unknownValue
		}
	}

	// Check for VM indicators.
	if isVM() {
		// Virtual machine detected.
		return model.ModeVM, ""
	}

	// Default to host mode.
	return model.ModeHost, ""
}

// isVM checks for virtual machine indicators.
//
// Returns:
//   - bool: true if running in a virtual machine
func isVM() bool {
	// Check DMI/SMBIOS for VM vendors.
	vmIndicators := []string{
		"/sys/class/dmi/id/product_name",
		"/sys/class/dmi/id/sys_vendor",
		"/sys/class/dmi/id/board_vendor",
	}

	vmVendors := []string{
		"vmware", "virtualbox", "kvm", "qemu",
		"xen", "hyper-v", "microsoft", "parallels",
		"virtual", "bochs", "bhyve",
	}

	// Iterate through DMI paths.
	for _, path := range vmIndicators {
		// Try to read each indicator file.
		if content, err := os.ReadFile(path); err == nil {
			lower := strings.ToLower(string(content))
			// Check against known VM vendors.
			for _, vendor := range vmVendors {
				// Match vendor string.
				if strings.Contains(lower, vendor) {
					// VM vendor detected.
					return true
				}
			}
		}
	}

	// Check cpuinfo for hypervisor flag.
	if content, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		// Look for hypervisor CPU flag.
		if strings.Contains(string(content), "hypervisor") {
			// Hypervisor flag found.
			return true
		}
	}

	// No VM indicators found.
	return false
}

// getDNSConfig reads DNS configuration.
//
// Returns:
//   - servers: DNS nameservers from resolv.conf
//   - search: DNS search domains from resolv.conf
func getDNSConfig() (servers []string, search []string) {
	content, err := os.ReadFile("/etc/resolv.conf")
	// Handle read error.
	if err != nil {
		// File not readable.
		return nil, nil
	}

	lines := strings.Split(string(content), "\n")
	// Parse each line of resolv.conf.
	for _, line := range lines {
		line = strings.TrimSpace(line)
		// Skip comment lines.
		if strings.HasPrefix(line, "#") {
			// Ignore comments.
			continue
		}

		fields := strings.Fields(line)
		// Skip lines with insufficient fields.
		if len(fields) < minDNSFields {
			// Not enough fields.
			continue
		}

		// Handle DNS directive type.
		switch fields[0] {
		// Nameserver directive.
		case "nameserver":
			servers = append(servers, fields[1])
		// Search or domain directive.
		case "search", "domain":
			search = append(search, fields[1:]...)
		}
	}

	// Return parsed DNS configuration.
	return servers, search
}

// fileExists checks if a file exists.
//
// Params:
//   - path: filesystem path to check
//
// Returns:
//   - bool: true if file exists and is accessible
func fileExists(path string) bool {
	_, err := os.Stat(path)
	// Return true if stat succeeded.
	return err == nil
}
