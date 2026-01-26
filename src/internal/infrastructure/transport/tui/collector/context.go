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

// Cached static system data - computed once per process lifetime.
var (
	cachedHostname    = sync.OnceValue(getHostnameOnce)
	cachedKernel      = sync.OnceValue(getKernelVersion)
	cachedRuntimeMode = sync.OnceValue(detectRuntimeModeOnce)
	cachedDNSConfig   = sync.OnceValue(getDNSConfigOnce)
	cachedPrimaryIP   = sync.OnceValue(getPrimaryIP)
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
func getHostnameOnce() string {
	if hostname, err := os.Hostname(); err == nil {
		return hostname
	}
	return unknownValue
}

// detectRuntimeModeOnce wraps detectRuntimeMode for sync.OnceValue.
func detectRuntimeModeOnce() runtimeModeResult {
	mode, rt := detectRuntimeMode()
	return runtimeModeResult{mode: mode, runtime: rt}
}

// getDNSConfigOnce wraps getDNSConfig for sync.OnceValue.
func getDNSConfigOnce() dnsConfigResult {
	servers, search := getDNSConfig()
	return dnsConfigResult{servers: servers, search: search}
}

// unknownValue is the default string for unknown fields.
const unknownValue = "unknown"

// ContextCollector gathers runtime context information.
type ContextCollector struct {
	version    string
	startTime  time.Time
	configPath string
}

// NewContextCollector creates a context collector.
func NewContextCollector(version string) *ContextCollector {
	return &ContextCollector{
		version:   version,
		startTime: time.Now(),
	}
}

// SetConfigPath sets the configuration file path.
func (c *ContextCollector) SetConfigPath(path string) {
	c.configPath = path
}

// CollectInto populates context information.
// Uses cached values for static data (hostname, kernel, runtime mode, DNS, IP).
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

	return nil
}

// detectRuntimeMode determines if we're in a container, VM, or host.
// Reads /proc/1/cgroup only once to avoid duplicate I/O.
func detectRuntimeMode() (model.RuntimeMode, string) {
	// Check for container indicators.

	// Docker/containerd.
	if fileExists("/.dockerenv") {
		return model.ModeContainer, "docker"
	}

	// Kubernetes (check for service account).
	if fileExists("/var/run/secrets/kubernetes.io/serviceaccount/token") {
		return model.ModeContainer, "kubernetes"
	}

	// LXC.
	if content, err := os.ReadFile("/proc/1/environ"); err == nil {
		if strings.Contains(string(content), "container=lxc") {
			return model.ModeContainer, "lxc"
		}
	}

	// Read /proc/1/cgroup ONCE and check all patterns.
	if content, err := os.ReadFile("/proc/1/cgroup"); err == nil {
		cgroupContent := string(content)

		// Podman detection.
		if strings.Contains(cgroupContent, "libpod") || strings.Contains(cgroupContent, "podman") {
			return model.ModeContainer, "podman"
		}

		// Docker/containerd detection.
		if strings.Contains(cgroupContent, "docker") || strings.Contains(cgroupContent, "containerd") {
			return model.ModeContainer, "docker"
		}

		// Generic container detection: non-root cgroup suggests container.
		if !strings.Contains(cgroupContent, ":/\n") {
			return model.ModeContainer, unknownValue
		}
	}

	// Check for VM indicators.
	if isVM() {
		return model.ModeVM, ""
	}

	return model.ModeHost, ""
}

// isVM checks for virtual machine indicators.
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

	for _, path := range vmIndicators {
		if content, err := os.ReadFile(path); err == nil {
			lower := strings.ToLower(string(content))
			for _, vendor := range vmVendors {
				if strings.Contains(lower, vendor) {
					return true
				}
			}
		}
	}

	// Check cpuinfo for hypervisor flag.
	if content, err := os.ReadFile("/proc/cpuinfo"); err == nil {
		if strings.Contains(string(content), "hypervisor") {
			return true
		}
	}

	return false
}

// getDNSConfig reads DNS configuration.
func getDNSConfig() (servers []string, search []string) {
	content, err := os.ReadFile("/etc/resolv.conf")
	if err != nil {
		return nil, nil
	}

	lines := strings.Split(string(content), "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#") {
			continue
		}

		fields := strings.Fields(line)
		if len(fields) < 2 {
			continue
		}

		switch fields[0] {
		case "nameserver":
			servers = append(servers, fields[1])
		case "search", "domain":
			search = append(search, fields[1:]...)
		}
	}

	return servers, search
}

// fileExists checks if a file exists.
func fileExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
