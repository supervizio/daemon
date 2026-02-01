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

	// vmVendorPatterns contains patterns identifying VM environments.
	vmVendorPatterns []string = []string{
		"vmware", "virtualbox", "kvm", "qemu", "xen", "hyper-v",
		"microsoft", "parallels", "virtual", "bochs", "bhyve",
	}
)

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

// Gather populates context information.
// Uses cached values for static data (hostname, kernel, runtime mode, DNS, IP).
//
// Params:
//   - snap: target snapshot to populate
//
// Returns:
//   - error: always returns nil
func (c *ContextCollector) Gather(snap *model.Snapshot) error {
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
	if mode, runtime, found := detectContainerByFiles(); found {
		// Container detected via files.
		return mode, runtime
	}

	// Check cgroup-based container detection.
	if mode, runtime, found := detectContainerByCgroup(); found {
		// Container detected via cgroup.
		return mode, runtime
	}

	// Check for VM indicators.
	if isVM() {
		// Virtual machine detected.
		return model.ModeVM, ""
	}

	// Default to host mode.
	return model.ModeHost, ""
}

// detectContainerByFiles checks for container markers via file existence.
//
// Returns:
//   - model.RuntimeMode: detected runtime mode
//   - string: container runtime name
//   - bool: true if container detected
func detectContainerByFiles() (model.RuntimeMode, string, bool) {
	// Docker/containerd.
	if fileExists("/.dockerenv") {
		// Docker environment file found.
		return model.ModeContainer, "docker", true
	}

	// Kubernetes (check for service account).
	if fileExists("/var/run/secrets/kubernetes.io/serviceaccount/token") {
		// Kubernetes service account token found.
		return model.ModeContainer, "kubernetes", true
	}

	// LXC.
	if content, err := os.ReadFile("/proc/1/environ"); err == nil {
		contentStr := string(content)
		// Check for LXC container marker.
		if strings.Contains(contentStr, "container=lxc") {
			// LXC container detected.
			return model.ModeContainer, "lxc", true
		}
	}

	// No container detected by files.
	return model.ModeHost, "", false
}

// detectContainerByCgroup checks for container markers via cgroup content.
//
// Returns:
//   - model.RuntimeMode: detected runtime mode
//   - string: container runtime name
//   - bool: true if container detected
func detectContainerByCgroup() (model.RuntimeMode, string, bool) {
	// Read /proc/1/cgroup ONCE and check all patterns.
	content, err := os.ReadFile("/proc/1/cgroup")
	// Handle read error.
	if err != nil {
		// Cannot read cgroup file.
		return model.ModeHost, "", false
	}

	cgroupContent := string(content)

	// Podman detection.
	if strings.Contains(cgroupContent, "libpod") || strings.Contains(cgroupContent, "podman") {
		// Podman cgroup markers found.
		return model.ModeContainer, "podman", true
	}

	// Docker/containerd detection.
	if strings.Contains(cgroupContent, "docker") || strings.Contains(cgroupContent, "containerd") {
		// Docker/containerd cgroup markers found.
		return model.ModeContainer, "docker", true
	}

	// Generic container detection: non-root cgroup suggests container.
	if !strings.Contains(cgroupContent, ":/\n") {
		// Non-root cgroup detected.
		return model.ModeContainer, unknownValue, true
	}

	// No container detected via cgroup.
	return model.ModeHost, "", false
}

// isVM checks for virtual machine indicators.
//
// Returns:
//   - bool: true if running in a virtual machine
func isVM() bool {
	// Check DMI/SMBIOS paths for VM vendors.
	if checkDMIForVM("/sys/class/dmi/id/product_name") {
		// return true for success.
		return true
	}
	// evaluate condition.
	if checkDMIForVM("/sys/class/dmi/id/sys_vendor") {
		// return true for success.
		return true
	}
	// evaluate condition.
	if checkDMIForVM("/sys/class/dmi/id/board_vendor") {
		// return true for success.
		return true
	}

	// Check cpuinfo for hypervisor flag.
	if content, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		contentStr := string(content)
		// Look for hypervisor CPU flag.
		if strings.Contains(contentStr, "hypervisor") {
			// Hypervisor flag found.
			return true
		}
	}

	// No VM indicators found.
	return false
}

// checkDMIForVM checks a DMI path for known VM vendor strings.
//
// Params:
//   - path: path to DMI file to check
//
// Returns:
//   - bool: true if VM vendor detected
func checkDMIForVM(path string) bool {
	content, err := os.ReadFile(path)
	// handle non-nil condition.
	if err != nil {
		// return false for failure.
		return false
	}
	contentStr := string(content)
	lower := strings.ToLower(contentStr)
	// Check against known VM vendors using loop to reduce complexity.
	return containsAnyVMVendor(lower)
}

// containsAnyVMVendor checks if content contains any known VM vendor pattern.
//
// Params:
//   - content: lowercased content to check
//
// Returns:
//   - bool: true if any VM vendor pattern found
func containsAnyVMVendor(content string) bool {
	// iterate over collection.
	for _, vendor := range vmVendorPatterns {
		// evaluate condition.
		if strings.Contains(content, vendor) {
			// return true for success.
			return true
		}
	}
	// return false for failure.
	return false
}

// getDNSConfig reads DNS configuration.
//
// Returns:
//   - servers: DNS nameservers from resolv.conf
//   - search: DNS search domains from resolv.conf
func getDNSConfig() (servers, search []string) {
	content, err := os.ReadFile("/etc/resolv.conf")
	// Handle read error.
	if err != nil {
		// File not readable.
		return nil, nil
	}

	contentStr := string(content)
	// Parse each line of resolv.conf.
	for line := range strings.SplitSeq(contentStr, "\n") {
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
