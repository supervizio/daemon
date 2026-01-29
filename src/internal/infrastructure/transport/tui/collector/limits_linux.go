//go:build linux

// Package collector provides Linux-specific cgroup limit collection.
package collector

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/kodflow/daemon/internal/infrastructure/transport/tui/model"
)

const (
	// cgroupMax is the value used in cgroups to indicate unlimited.
	cgroupMax string = "max"

	// minCPUPartsCount is the minimum number of parts expected in cpu.max format.
	minCPUPartsCount int = 2

	// decimalBase is the base for decimal number parsing.
	decimalBase int = 10

	// bitSize64 is the bit size for 64-bit integers.
	bitSize64 int = 64

	// memoryUnlimitedThreshold is the threshold to detect unlimited memory (< 4 EiB).
	memoryUnlimitedThreshold uint64 = 1 << 62

	// cgroupPathParts is the number of parts in cgroup path format (hierarchy:controller:path).
	cgroupPathParts int = 3

	// cgroupPathIndex is the index of the cgroup path in split parts.
	cgroupPathIndex int = 2

	// cgroupHierarchyIndex is the index of the hierarchy ID in split parts.
	cgroupHierarchyIndex int = 0

	// cgroupControllerIndex is the index of the controller list in split parts.
	cgroupControllerIndex int = 1

	// defaultCgroupV1Capacity is the initial capacity for cgroup v1 controller map.
	defaultCgroupV1Capacity int = 8
)

// Cached cgroup paths - computed once per process lifetime.
var (
	cachedCgroupV2Path  func() string            = sync.OnceValue(getCgroupV2Path)
	cachedCgroupV1Paths func() map[string]string = sync.OnceValue(getCgroupV1Paths)
)

// collectCgroupLimits gathers cgroup v1/v2 limits.
//
// Params:
//   - limits: target resource limits struct to populate
func collectCgroupLimits(limits *model.ResourceLimits) {
	// Try cgroup v2 first (preferred on modern systems).
	if collectCgroupV2(limits) {
		// Successfully collected from cgroup v2.
		return
	}

	// Fall back to cgroup v1.
	collectCgroupV1(limits)
}

// collectCgroupV2 reads cgroup v2 limits.
// Uses cached cgroup path for efficiency.
//
// Params:
//   - limits: target resource limits struct to populate
//
// Returns:
//   - bool: true if cgroup v2 was available and read successfully
func collectCgroupV2(limits *model.ResourceLimits) bool {
	// Find cgroup path (cached).
	cgroupPath := cachedCgroupV2Path()
	// Check if cgroup v2 is available.
	if cgroupPath == "" {
		// Cgroup v2 not available.
		return false
	}

	// Collect CPU limits.
	collectCgroupV2CPU(cgroupPath, limits)

	// Collect memory limits.
	collectCgroupV2Memory(cgroupPath, limits)

	// Collect PIDs limits.
	collectCgroupV2PIDs(cgroupPath, limits)

	// Successfully collected cgroup v2 limits.
	return true
}

// collectCgroupV2CPU reads CPU limits from cgroup v2.
//
// Params:
//   - cgroupPath: path to the cgroup v2 directory
//   - limits: target resource limits struct to populate
func collectCgroupV2CPU(cgroupPath string, limits *model.ResourceLimits) {
	// CPU limits (cpu.max).
	parseCgroupV2CPUMax(cgroupPath, limits)

	// CPU set (cpuset.cpus.effective or cpuset.cpus).
	parseCgroupV2CPUSet(cgroupPath, limits)
}

// parseCgroupV2CPUMax parses cpu.max file for quota/period.
//
// Params:
//   - cgroupPath: path to the cgroup v2 directory
//   - limits: target resource limits struct to populate
func parseCgroupV2CPUMax(cgroupPath string, limits *model.ResourceLimits) {
	content, err := os.ReadFile(filepath.Join(cgroupPath, "cpu.max"))
	// Handle read error.
	if err != nil {
		// Cannot read cpu.max.
		return
	}

	contentStr := string(content)
	parts := strings.Fields(contentStr)
	// Check if both parts are present.
	if len(parts) < minCPUPartsCount {
		// Insufficient parts.
		return
	}
	// Check if quota is unlimited.
	if parts[0] == cgroupMax {
		// Quota is unlimited.
		return
	}

	// Parse CPU quota.
	quota, err := strconv.ParseInt(parts[0], decimalBase, bitSize64)
	// Handle parse error.
	if err != nil {
		// Cannot parse quota.
		return
	}

	// Parse CPU period.
	period, err := strconv.ParseInt(parts[1], decimalBase, bitSize64)
	// Handle parse error.
	if err != nil {
		// Cannot parse period.
		return
	}

	limits.CPUQuotaRaw = quota
	limits.CPUPeriod = period
	limits.CPUQuota = float64(quota) / float64(period)
}

// parseCgroupV2CPUSet reads cpuset from cgroup v2.
//
// Params:
//   - cgroupPath: path to the cgroup v2 directory
//   - limits: target resource limits struct to populate
func parseCgroupV2CPUSet(cgroupPath string, limits *model.ResourceLimits) {
	// Try effective file first, fall back to regular file.
	content, err := os.ReadFile(filepath.Join(cgroupPath, "cpuset.cpus.effective"))
	// handle non-nil condition.
	if err != nil {
		// Fall back to regular cpuset file.
		content, err = os.ReadFile(filepath.Join(cgroupPath, "cpuset.cpus"))
		// handle non-nil condition.
		if err != nil {
			// Neither file readable.
			return
		}
	}
	// Convert and trim the content.
	cpusetStr := string(content)
	cpuset := strings.TrimSpace(cpusetStr)
	// Only set if we got a valid value.
	if cpuset != "" {
		limits.CPUSet = cpuset
	}
}

// collectCgroupV2Memory reads memory limits from cgroup v2.
//
// Params:
//   - cgroupPath: path to the cgroup v2 directory
//   - limits: target resource limits struct to populate
func collectCgroupV2Memory(cgroupPath string, limits *model.ResourceLimits) {
	// Memory max.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "memory.max")); err == nil {
		memMaxStr := string(content)
		s := strings.TrimSpace(memMaxStr)
		// Check if memory is not unlimited.
		if s != cgroupMax {
			// Parse memory limit.
			if val, err := strconv.ParseUint(s, decimalBase, bitSize64); err == nil {
				limits.MemoryMax = val
			}
		}
	}

	// Memory current.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "memory.current")); err == nil {
		memCurrentStr := string(content)
		// Parse memory current value.
		if val, err := strconv.ParseUint(strings.TrimSpace(memCurrentStr), decimalBase, bitSize64); err == nil {
			limits.MemoryCurrent = val
		}
	}
}

// collectCgroupV2PIDs reads PIDs limits from cgroup v2.
//
// Params:
//   - cgroupPath: path to the cgroup v2 directory
//   - limits: target resource limits struct to populate
func collectCgroupV2PIDs(cgroupPath string, limits *model.ResourceLimits) {
	// PIDs max.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "pids.max")); err == nil {
		pidsMaxStr := string(content)
		s := strings.TrimSpace(pidsMaxStr)
		// Check if PIDs is not unlimited.
		if s != cgroupMax {
			// Parse PIDs limit.
			if val, err := strconv.ParseInt(s, decimalBase, bitSize64); err == nil {
				limits.PIDsMax = val
			}
		}
	}

	// PIDs current.
	if content, err := os.ReadFile(filepath.Join(cgroupPath, "pids.current")); err == nil {
		pidsCurrentStr := string(content)
		// Parse PIDs current value.
		if val, err := strconv.ParseInt(strings.TrimSpace(pidsCurrentStr), decimalBase, bitSize64); err == nil {
			limits.PIDsCurrent = val
		}
	}
}

// getCgroupV2Path returns the cgroup v2 path for the current process.
//
// Returns:
//   - string: cgroup v2 path or empty string if not available
func getCgroupV2Path() string {
	// Check if cgroup v2 is mounted.
	if _, err := os.Stat("/sys/fs/cgroup/cgroup.controllers"); err != nil {
		// Cgroup v2 not available.
		return ""
	}

	// Read our cgroup path.
	content, err := os.ReadFile("/proc/self/cgroup")
	// Handle read error.
	if err != nil {
		// Cannot read cgroup info.
		return ""
	}

	cgroupContentStr := string(content)
	// Format: 0::/path
	// Parse each line to find cgroup v2 entry.
	for line := range strings.SplitSeq(cgroupContentStr, "\n") {
		parts := strings.SplitN(line, ":", cgroupPathParts)
		// Check for valid cgroup v2 entry (hierarchy 0).
		if len(parts) == cgroupPathParts && parts[cgroupHierarchyIndex] == "0" {
			path := strings.TrimSpace(parts[cgroupPathIndex])
			// Handle root cgroup.
			if path == "" || path == "/" {
				// Use base cgroup path.
				return "/sys/fs/cgroup"
			}
			// Return full cgroup path.
			return filepath.Join("/sys/fs/cgroup", path)
		}
	}

	// Default to base cgroup path.
	return "/sys/fs/cgroup"
}

// collectCgroupV1 reads cgroup v1 limits.
// Uses cached cgroup paths for efficiency.
//
// Params:
//   - limits: target resource limits struct to populate
func collectCgroupV1(limits *model.ResourceLimits) {
	// Find cgroup paths (cached).
	cgroupPaths := cachedCgroupV1Paths()

	// Collect CPU limits.
	collectCgroupV1CPU(cgroupPaths, limits)

	// Collect memory limits.
	collectCgroupV1Memory(cgroupPaths, limits)

	// Collect PIDs limits.
	collectCgroupV1PIDs(cgroupPaths, limits)
}

// collectCgroupV1CPU reads CPU limits from cgroup v1.
//
// Params:
//   - cgroupPaths: map of controller name to path
//   - limits: target resource limits struct to populate
func collectCgroupV1CPU(cgroupPaths map[string]string, limits *model.ResourceLimits) {
	// CPU quota (cpu.cfs_quota_us / cpu.cfs_period_us).
	parseCgroupV1CPUQuota(cgroupPaths, limits)

	// CPU set.
	parseCgroupV1CPUSet(cgroupPaths, limits)
}

// parseCgroupV1CPUQuota parses CPU quota/period from cgroup v1.
//
// Params:
//   - cgroupPaths: map of controller name to path
//   - limits: target resource limits struct to populate
func parseCgroupV1CPUQuota(cgroupPaths map[string]string, limits *model.ResourceLimits) {
	// Check if CPU controller path exists.
	cpuPath, ok := cgroupPaths["cpu"]
	// Skip if CPU controller not found.
	if !ok {
		// CPU controller not available.
		return
	}

	// Read CPU quota file.
	content, err := os.ReadFile(filepath.Join(cpuPath, "cpu.cfs_quota_us"))
	// Handle read error.
	if err != nil {
		// Cannot read quota.
		return
	}

	quotaStr := string(content)
	// Parse and validate quota value.
	quota, err := strconv.ParseInt(strings.TrimSpace(quotaStr), decimalBase, bitSize64)
	// Skip if parsing failed or quota disabled.
	if err != nil || quota <= 0 {
		// Invalid quota.
		return
	}

	limits.CPUQuotaRaw = quota

	// Read CPU period file.
	content, err = os.ReadFile(filepath.Join(cpuPath, "cpu.cfs_period_us"))
	// Handle read error.
	if err != nil {
		// Cannot read period.
		return
	}

	periodStr := string(content)
	// Parse and validate period value.
	period, err := strconv.ParseInt(strings.TrimSpace(periodStr), decimalBase, bitSize64)
	// Skip if parsing failed or period invalid.
	if err != nil || period <= 0 {
		// Invalid period.
		return
	}

	limits.CPUPeriod = period
	limits.CPUQuota = float64(quota) / float64(period)
}

// parseCgroupV1CPUSet parses cpuset from cgroup v1.
//
// Params:
//   - cgroupPaths: map of controller name to path
//   - limits: target resource limits struct to populate
func parseCgroupV1CPUSet(cgroupPaths map[string]string, limits *model.ResourceLimits) {
	// Check if cpuset controller path exists.
	cpusetPath, ok := cgroupPaths["cpuset"]
	// Skip if cpuset controller not found.
	if !ok {
		// Cpuset controller not available.
		return
	}

	// Read cpuset.cpus file.
	content, err := os.ReadFile(filepath.Join(cpusetPath, "cpuset.cpus"))
	// Handle read error.
	if err != nil {
		// Cannot read cpuset.
		return
	}

	cpusetStr := string(content)
	limits.CPUSet = strings.TrimSpace(cpusetStr)
}

// collectCgroupV1Memory reads memory limits from cgroup v1.
//
// Params:
//   - cgroupPaths: map of controller name to path
//   - limits: target resource limits struct to populate
func collectCgroupV1Memory(cgroupPaths map[string]string, limits *model.ResourceLimits) {
	// Check if memory controller path exists.
	memPath, ok := cgroupPaths["memory"]
	// Skip if memory controller not found.
	if !ok {
		// Memory controller not available.
		return
	}

	// Read memory limit file.
	if content, err := os.ReadFile(filepath.Join(memPath, "memory.limit_in_bytes")); err == nil {
		memLimitStr := string(content)
		// Parse memory limit value.
		if val, err := strconv.ParseUint(strings.TrimSpace(memLimitStr), decimalBase, bitSize64); err == nil {
			// Check for "unlimited" (very large value).
			if val < memoryUnlimitedThreshold {
				limits.MemoryMax = val
			}
		}
	}

	// Read memory usage file.
	if content, err := os.ReadFile(filepath.Join(memPath, "memory.usage_in_bytes")); err == nil {
		memUsageStr := string(content)
		// Parse memory usage value.
		if val, err := strconv.ParseUint(strings.TrimSpace(memUsageStr), decimalBase, bitSize64); err == nil {
			limits.MemoryCurrent = val
		}
	}
}

// collectCgroupV1PIDs reads PIDs limits from cgroup v1.
//
// Params:
//   - cgroupPaths: map of controller name to path
//   - limits: target resource limits struct to populate
func collectCgroupV1PIDs(cgroupPaths map[string]string, limits *model.ResourceLimits) {
	// Check if pids controller path exists.
	pidsPath, ok := cgroupPaths["pids"]
	// Skip if pids controller not found.
	if !ok {
		// PIDs controller not available.
		return
	}

	// Read PIDs max file.
	if content, err := os.ReadFile(filepath.Join(pidsPath, "pids.max")); err == nil {
		pidsMaxStr := string(content)
		s := strings.TrimSpace(pidsMaxStr)
		// Check if not unlimited.
		if s != cgroupMax {
			// Parse PIDs max value.
			if val, err := strconv.ParseInt(s, decimalBase, bitSize64); err == nil {
				limits.PIDsMax = val
			}
		}
	}

	// Read PIDs current file.
	if content, err := os.ReadFile(filepath.Join(pidsPath, "pids.current")); err == nil {
		pidsCurrentStr := string(content)
		// Parse PIDs current value.
		if val, err := strconv.ParseInt(strings.TrimSpace(pidsCurrentStr), decimalBase, bitSize64); err == nil {
			limits.PIDsCurrent = val
		}
	}
}

// getCgroupV1Paths returns controller -> path mapping for cgroup v1.
//
// Returns:
//   - map[string]string: controller name to filesystem path mapping
func getCgroupV1Paths() map[string]string {
	// Pre-allocate for typical cgroup v1 controllers.
	paths := make(map[string]string, defaultCgroupV1Capacity)

	content, err := os.ReadFile("/proc/self/cgroup")
	// Handle read error.
	if err != nil {
		// Return empty map on error.
		return paths
	}

	cgroupContentStr := string(content)
	// Format: hierarchy-ID:controller-list:cgroup-path
	// Parse each line of cgroup file.
	for line := range strings.SplitSeq(cgroupContentStr, "\n") {
		parts := strings.SplitN(line, ":", cgroupPathParts)
		// Validate part count.
		if len(parts) != cgroupPathParts {
			// Skip malformed lines.
			continue
		}

		controllers := strings.Split(parts[cgroupControllerIndex], ",")
		cgroupPath := strings.TrimSpace(parts[cgroupPathIndex])

		// Process each controller.
		for _, controller := range controllers {
			// Skip empty controller names.
			if controller == "" {
				// Ignore empty entries.
				continue
			}

			// Build full path.
			fullPath := filepath.Join("/sys/fs/cgroup", controller, cgroupPath)
			// Verify path exists.
			if _, err := os.Stat(fullPath); err == nil {
				paths[controller] = fullPath
			}
		}
	}

	// Return controller to path mapping.
	return paths
}
