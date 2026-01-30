//go:build cgo

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

// CachePolicy represents a cache policy preset.
type CachePolicy uint32

// Cache policy presets.
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
type MetricType uint8

// Metric types for cache TTL configuration.
const (
	MetricCPUSystem      MetricType = 0
	MetricCPUPressure    MetricType = 1
	MetricMemorySystem   MetricType = 2
	MetricMemoryPressure MetricType = 3
	MetricLoad           MetricType = 4
	MetricDiskPartitions MetricType = 5
	MetricDiskUsage      MetricType = 6
	MetricDiskIO         MetricType = 7
	MetricNetInterfaces  MetricType = 8
	MetricNetStats       MetricType = 9
	MetricIOStats        MetricType = 10
	MetricIOPressure     MetricType = 11
)

// CacheEnable enables caching with default policies.
//
// After calling this, all metric collection calls will use caching.
// Call CacheDisable to disable caching.
func CacheEnable() error {
	result := C.probe_cache_enable()
	return resultToError(result)
}

// CacheEnableWithPolicy enables caching with a specific policy preset.
func CacheEnableWithPolicy(policy CachePolicy) error {
	result := C.probe_cache_enable_with_policy(C.uint32_t(policy))
	return resultToError(result)
}

// CacheDisable disables caching and reverts to direct collection.
func CacheDisable() error {
	result := C.probe_cache_disable()
	return resultToError(result)
}

// CacheIsEnabled returns whether caching is currently enabled.
func CacheIsEnabled() bool {
	return bool(C.probe_cache_is_enabled())
}

// CacheSetTTL sets the TTL for a specific metric type.
func CacheSetTTL(metricType MetricType, ttl time.Duration) error {
	ttlMs := uint64(ttl.Milliseconds())
	result := C.probe_cache_set_ttl(C.uint8_t(metricType), C.uint64_t(ttlMs))
	return resultToError(result)
}

// CacheInvalidateAll invalidates all cached metrics.
func CacheInvalidateAll() error {
	result := C.probe_cache_invalidate_all()
	return resultToError(result)
}

// CacheInvalidate invalidates a specific metric type from the cache.
func CacheInvalidate(metricType MetricType) error {
	result := C.probe_cache_invalidate(C.uint8_t(metricType))
	return resultToError(result)
}

// CachedCPUCollector provides CPU metrics collection with caching support.
// It implements the metrics.CPUCollector interface.
type CachedCPUCollector struct{}

// NewCachedCPUCollector creates a new cached CPU collector.
func NewCachedCPUCollector() *CachedCPUCollector {
	return &CachedCPUCollector{}
}

// CollectSystem collects system-wide CPU metrics with caching.
func (c *CachedCPUCollector) CollectSystem(_ context.Context) (metrics.SystemCPU, error) {
	if err := checkInitialized(); err != nil {
		return metrics.SystemCPU{}, err
	}

	var cCPU C.SystemCPU
	result := C.probe_collect_cpu_cached(&cCPU)
	if err := resultToError(result); err != nil {
		return metrics.SystemCPU{}, err
	}

	usagePercent := 100.0 - float64(cCPU.idle_percent)
	if usagePercent < 0 {
		usagePercent = 0
	}

	return metrics.SystemCPU{
		UsagePercent: usagePercent,
		Timestamp:    time.Now(),
	}, nil
}

// CollectProcess collects CPU metrics for a specific process.
// Process metrics are not cached.
func (c *CachedCPUCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessCPU, error) {
	// Process metrics are not cached, delegate to regular collector
	return (&CPUCollector{}).CollectProcess(ctx, pid)
}

// CollectAllProcesses is not implemented.
func (c *CachedCPUCollector) CollectAllProcesses(_ context.Context) ([]metrics.ProcessCPU, error) {
	return nil, ErrNotSupported
}

// CollectLoadAverage collects system load average with caching.
func (c *CachedCPUCollector) CollectLoadAverage(_ context.Context) (metrics.LoadAverage, error) {
	if err := checkInitialized(); err != nil {
		return metrics.LoadAverage{}, err
	}

	var cLoad C.LoadAverage
	result := C.probe_collect_load_cached(&cLoad)
	if err := resultToError(result); err != nil {
		return metrics.LoadAverage{}, err
	}

	return metrics.LoadAverage{
		Load1:     float64(cLoad.load_1min),
		Load5:     float64(cLoad.load_5min),
		Load15:    float64(cLoad.load_15min),
		Timestamp: time.Now(),
	}, nil
}

// CollectPressure collects CPU pressure metrics.
// Pressure metrics are not cached.
func (c *CachedCPUCollector) CollectPressure(ctx context.Context) (metrics.CPUPressure, error) {
	return (&CPUCollector{}).CollectPressure(ctx)
}

// CachedMemoryCollector provides memory metrics collection with caching support.
// It implements the metrics.MemoryCollector interface.
type CachedMemoryCollector struct{}

// NewCachedMemoryCollector creates a new cached memory collector.
func NewCachedMemoryCollector() *CachedMemoryCollector {
	return &CachedMemoryCollector{}
}

// CollectSystem collects system-wide memory metrics with caching.
func (c *CachedMemoryCollector) CollectSystem(_ context.Context) (metrics.SystemMemory, error) {
	if err := checkInitialized(); err != nil {
		return metrics.SystemMemory{}, err
	}

	var cMem C.SystemMemory
	result := C.probe_collect_memory_cached(&cMem)
	if err := resultToError(result); err != nil {
		return metrics.SystemMemory{}, err
	}

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
func (c *CachedMemoryCollector) CollectProcess(ctx context.Context, pid int) (metrics.ProcessMemory, error) {
	return (&MemoryCollector{}).CollectProcess(ctx, pid)
}

// CollectPressure collects memory pressure metrics.
// Pressure metrics are not cached.
func (c *CachedMemoryCollector) CollectPressure(ctx context.Context) (metrics.MemoryPressure, error) {
	return (&MemoryCollector{}).CollectPressure(ctx)
}
