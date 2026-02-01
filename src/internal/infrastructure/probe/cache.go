//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
//
//nolint:ktn-struct-onefile // Cached collectors are logically grouped with cache management
package probe

/*
#include "probe.h"
*/
import "C"

import (
	"context"
	"time"

	"github.com/kodflow/daemon/internal/domain/metrics"
)

// fullPercentCache represents 100% for cache percentage calculations.
const fullPercentCache float64 = 100.0

// CachePolicy represents a cache policy preset.
// It defines the TTL behavior for cached metrics.
type CachePolicy uint32

// Cache policy presets for different collection frequencies.
const (
	// CachePolicyDefault provides balanced TTLs for general use.
	CachePolicyDefault CachePolicy = 0
	// CachePolicyHighFreq provides shorter TTLs for frequent collection.
	CachePolicyHighFreq CachePolicy = 1
	// CachePolicyLowFreq provides longer TTLs for infrequent collection.
	CachePolicyLowFreq CachePolicy = 2
	// CachePolicyNoCache disables caching (TTL=0).
	CachePolicyNoCache CachePolicy = 3
)

// MetricType represents a type of metric for cache configuration.
// It identifies specific metric categories for TTL settings.
type MetricType uint8

// Metric types for cache TTL configuration.
const (
	// MetricCPUSystem is the system CPU metrics type.
	MetricCPUSystem MetricType = 0
	// MetricCPUPressure is the CPU pressure metrics type.
	MetricCPUPressure MetricType = 1
	// MetricMemorySystem is the system memory metrics type.
	MetricMemorySystem MetricType = 2
	// MetricMemoryPressure is the memory pressure metrics type.
	MetricMemoryPressure MetricType = 3
	// MetricLoad is the load average metrics type.
	MetricLoad MetricType = 4
	// MetricDiskPartitions is the disk partitions metrics type.
	MetricDiskPartitions MetricType = 5
	// MetricDiskUsage is the disk usage metrics type.
	MetricDiskUsage MetricType = 6
	// MetricDiskIO is the disk I/O metrics type.
	MetricDiskIO MetricType = 7
	// MetricNetInterfaces is the network interfaces metrics type.
	MetricNetInterfaces MetricType = 8
	// MetricNetStats is the network statistics metrics type.
	MetricNetStats MetricType = 9
	// MetricIOStats is the I/O statistics metrics type.
	MetricIOStats MetricType = 10
	// MetricIOPressure is the I/O pressure metrics type.
	MetricIOPressure MetricType = 11
)

// CacheEnable enables caching for probe metrics.
// After calling this, all metric collection calls will use caching.
//
// Returns:
//   - error: nil on success, error if operation fails
func CacheEnable() error {
	result := C.probe_cache_enable()
	// Convert result to Go error.
	return resultToError(result)
}

// CacheEnableWithPolicy enables caching with a specific policy preset.
//
// Params:
//   - policy: cache policy preset to use
//
// Returns:
//   - error: nil on success, error if operation fails
func CacheEnableWithPolicy(policy CachePolicy) error {
	result := C.probe_cache_enable_with_policy(C.uint32_t(policy))
	// Convert result to Go error.
	return resultToError(result)
}

// CacheDisable disables caching and reverts to direct collection.
//
// Returns:
//   - error: nil on success, error if operation fails
func CacheDisable() error {
	result := C.probe_cache_disable()
	// Convert result to Go error.
	return resultToError(result)
}

// CacheIsEnabled returns whether caching is currently enabled.
//
// Returns:
//   - bool: true if caching is enabled
func CacheIsEnabled() bool {
	// Delegate to Rust library for cache status.
	return bool(C.probe_cache_is_enabled())
}

// CacheSetTTL sets the TTL for a specific metric type.
//
// Params:
//   - metricType: type of metric to configure
//   - ttl: time-to-live duration for cached values
//
// Returns:
//   - error: nil on success, error if operation fails
func CacheSetTTL(metricType MetricType, ttl time.Duration) error {
	ttlMs := uint64(ttl.Milliseconds())
	result := C.probe_cache_set_ttl(C.uint8_t(metricType), C.uint64_t(ttlMs))
	// Convert result to Go error.
	return resultToError(result)
}

// CacheInvalidateAll invalidates all cached metrics.
//
// Returns:
//   - error: nil on success, error if operation fails
func CacheInvalidateAll() error {
	result := C.probe_cache_invalidate_all()
	// Convert result to Go error.
	return resultToError(result)
}

// CacheInvalidate invalidates a specific metric type from the cache.
//
// Params:
//   - metricType: type of metric to invalidate
//
// Returns:
//   - error: nil on success, error if operation fails
func CacheInvalidate(metricType MetricType) error {
	result := C.probe_cache_invalidate(C.uint8_t(metricType))
	// Convert result to Go error.
	return resultToError(result)
}

// CachedCPUCollector provides CPU metrics collection with caching support.
// It implements the metrics.CPUCollector interface.
type CachedCPUCollector struct{}

// NewCachedCPUCollector creates a new cached CPU collector.
//
// Returns:
//   - *CachedCPUCollector: new cached CPU collector instance
func NewCachedCPUCollector() *CachedCPUCollector {
	// Return a new empty collector instance.
	return &CachedCPUCollector{}
}

// CollectSystem collects system-wide CPU metrics with caching.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.SystemCPU: cached system-wide CPU statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CachedCPUCollector) CollectSystem(_ context.Context) (metrics.SystemCPU, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.SystemCPU{}, err
	}

	var cCPU C.SystemCPU
	result := C.probe_collect_cpu_cached(&cCPU)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.SystemCPU{}, err
	}

	usagePercent := fullPercentCache - float64(cCPU.idle_percent)
	// Clamp negative values to zero.
	if usagePercent < 0 {
		usagePercent = 0
	}

	// Return cached CPU metrics with current timestamp.
	return metrics.SystemCPU{
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
	}, nil
}

// CollectProcess collects CPU metrics for a specific process.
// Process metrics are not cached.
//
// Params:
//   - ctx: context for cancellation
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessCPU: CPU metrics for the process
//   - error: nil on success, error if collection fails
func (c *CachedCPUCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	// Process metrics are not cached, delegate to regular collector.
	return (&CPUCollector{}).CollectProcess(ctx, pid)
}

// CollectAllProcesses is not implemented.
//
// Params:
//   - ctx: context for cancellation (unused)
//
// Returns:
//   - []metrics.ProcessCPU: always nil
//   - error: always ErrNotSupported
func (c *CachedCPUCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessCPU, error) {
	// Return not supported error.
	return nil, ErrNotSupported
}

// CollectLoadAverage collects system load average with caching.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.LoadAverage: cached system load averages
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CachedCPUCollector) CollectLoadAverage(_ context.Context) (metrics.LoadAverage, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.LoadAverage{}, err
	}

	var cLoad C.LoadAverage
	result := C.probe_collect_load_cached(&cLoad)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.LoadAverage{}, err
	}

	// Return cached load average with current timestamp.
	return metrics.LoadAverage{
		Load1:     float64(cLoad.load_1min),
		Load5:     float64(cLoad.load_5min),
		Load15:    float64(cLoad.load_15min),
		Timestamp: time.Now(),
	}, nil
}

// CollectPressure collects CPU pressure metrics.
// Pressure metrics are not cached.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.CPUPressure: CPU pressure statistics
//   - error: nil on success, error if collection fails
func (c *CachedCPUCollector) CollectPressure(ctx context.Context) (metrics.CPUPressure, error) {
	// Delegate to regular collector as pressure is not cached.
	return (&CPUCollector{}).CollectPressure(ctx)
}

// CachedMemoryCollector provides memory metrics collection with caching support.
// It implements the metrics.MemoryCollector interface.
type CachedMemoryCollector struct{}

// NewCachedMemoryCollector creates a new cached memory collector.
//
// Returns:
//   - *CachedMemoryCollector: new cached memory collector instance
func NewCachedMemoryCollector() *CachedMemoryCollector {
	// Return a new empty collector instance.
	return &CachedMemoryCollector{}
}

// CollectSystem collects system-wide memory metrics with caching.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - metrics.SystemMemory: cached system-wide memory statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (c *CachedMemoryCollector) CollectSystem(_ context.Context) (metrics.SystemMemory, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return empty metrics with initialization error.
		return metrics.SystemMemory{}, err
	}

	var cMem C.SystemMemory
	result := C.probe_collect_memory_cached(&cMem)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return empty metrics with collection error.
		return metrics.SystemMemory{}, err
	}

	// Return cached memory metrics with current timestamp.
	return metrics.SystemMemory{
		Total:     uint64(cMem.total_bytes),
		Available: uint64(cMem.available_bytes),
		Used:      uint64(cMem.used_bytes),
		Cached:    uint64(cMem.cached_bytes),
		Buffers:   uint64(cMem.buffers_bytes),
		SwapTotal: uint64(cMem.swap_total_bytes),
		SwapUsed:  uint64(cMem.swap_used_bytes),
		SwapFree:  uint64(cMem.swap_total_bytes) - uint64(cMem.swap_used_bytes),
		Timestamp: time.Now(),
	}, nil
}

// CollectProcess collects memory metrics for a specific process.
// Process metrics are not cached.
//
// Params:
//   - ctx: context for cancellation
//   - pid: process ID to collect metrics for
//
// Returns:
//   - metrics.ProcessMemory: memory metrics for the process
//   - error: nil on success, error if collection fails
func (c *CachedMemoryCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	// Delegate to regular collector as process metrics are not cached.
	return (&MemoryCollector{}).CollectProcess(ctx, pid)
}

// CollectAllProcesses is not implemented.
//
// Params:
//   - ctx: context for cancellation (unused)
//
// Returns:
//   - []metrics.ProcessMemory: always nil
//   - error: always ErrNotSupported
func (c *CachedMemoryCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessMemory, error) {
	// Return not supported error.
	return nil, ErrNotSupported
}

// CollectPressure collects memory pressure metrics.
// Pressure metrics are not cached.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - metrics.MemoryPressure: memory pressure statistics
//   - error: nil on success, error if collection fails
func (c *CachedMemoryCollector) CollectPressure(ctx context.Context) (metrics.MemoryPressure, error) {
	// Delegate to regular collector as pressure is not cached.
	return (&MemoryCollector{}).CollectPressure(ctx)
}
