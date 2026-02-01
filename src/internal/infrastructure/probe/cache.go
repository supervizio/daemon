//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

import (
	"time"
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
