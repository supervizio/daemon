//go:build cgo

package probe_test

import (
	"context"
	"testing"
	"time"

	"github.com/kodflow/daemon/internal/infrastructure/probe"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCacheIsEnabled verifies cache status check.
func TestCacheIsEnabled(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "ReturnsBoolean"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			// Initially disabled
			assert.False(t, probe.CacheIsEnabled())

			// Enable cache
			err = probe.CacheEnable()
			require.NoError(t, err)

			// Should be enabled now
			assert.True(t, probe.CacheIsEnabled())

			// Disable cache
			err = probe.CacheDisable()
			require.NoError(t, err)

			// Should be disabled again
			assert.False(t, probe.CacheIsEnabled())
		})
	}
}

// TestCacheEnable verifies cache enabling.
func TestCacheEnable(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "enables cache successfully"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			err = probe.CacheEnable()
			require.NoError(t, err)

			assert.True(t, probe.CacheIsEnabled())

			err = probe.CacheDisable()
			require.NoError(t, err)
		})
	}
}

// TestCacheEnableWithPolicy verifies cache enabling with policy.
func TestCacheEnableWithPolicy(t *testing.T) {
	tests := []struct {
		name   string
		policy probe.CachePolicy
	}{
		{name: "CachePolicyDefault", policy: probe.CachePolicyDefault},
		{name: "CachePolicyHighFreq", policy: probe.CachePolicyHighFreq},
		{name: "CachePolicyLowFreq", policy: probe.CachePolicyLowFreq},
		{name: "CachePolicyNoCache", policy: probe.CachePolicyNoCache},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			err = probe.CacheEnableWithPolicy(tt.policy)
			require.NoError(t, err)

			err = probe.CacheDisable()
			require.NoError(t, err)
		})
	}
}

// TestCacheDisable verifies cache disabling.
func TestCacheDisable(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "disables cache successfully"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			err = probe.CacheEnable()
			require.NoError(t, err)

			err = probe.CacheDisable()
			require.NoError(t, err)

			assert.False(t, probe.CacheIsEnabled())
		})
	}
}

// TestCacheSetTTL verifies TTL setting for metric types.
func TestCacheSetTTL(t *testing.T) {
	tests := []struct {
		name       string
		metricType probe.MetricType
		ttl        time.Duration
	}{
		{name: "MetricCPUSystem", metricType: probe.MetricCPUSystem, ttl: 5 * time.Second},
		{name: "MetricCPUPressure", metricType: probe.MetricCPUPressure, ttl: 5 * time.Second},
		{name: "MetricMemorySystem", metricType: probe.MetricMemorySystem, ttl: 5 * time.Second},
		{name: "MetricMemoryPressure", metricType: probe.MetricMemoryPressure, ttl: 5 * time.Second},
		{name: "MetricLoad", metricType: probe.MetricLoad, ttl: 5 * time.Second},
		{name: "MetricDiskPartitions", metricType: probe.MetricDiskPartitions, ttl: 5 * time.Second},
		{name: "MetricDiskUsage", metricType: probe.MetricDiskUsage, ttl: 5 * time.Second},
		{name: "MetricDiskIO", metricType: probe.MetricDiskIO, ttl: 5 * time.Second},
		{name: "MetricNetInterfaces", metricType: probe.MetricNetInterfaces, ttl: 5 * time.Second},
		{name: "MetricNetStats", metricType: probe.MetricNetStats, ttl: 5 * time.Second},
		{name: "MetricIOStats", metricType: probe.MetricIOStats, ttl: 5 * time.Second},
		{name: "MetricIOPressure", metricType: probe.MetricIOPressure, ttl: 5 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			err = probe.CacheEnable()
			require.NoError(t, err)
			defer func() { _ = probe.CacheDisable() }()

			err = probe.CacheSetTTL(tt.metricType, tt.ttl)
			require.NoError(t, err)
		})
	}
}

// TestCacheInvalidateAll verifies cache invalidation.
func TestCacheInvalidateAll(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "invalidates all cache successfully"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			err = probe.CacheEnable()
			require.NoError(t, err)
			defer func() { _ = probe.CacheDisable() }()

			err = probe.CacheInvalidateAll()
			require.NoError(t, err)
		})
	}
}

// TestCacheInvalidate verifies specific metric type cache invalidation.
func TestCacheInvalidate(t *testing.T) {
	tests := []struct {
		name       string
		metricType probe.MetricType
	}{
		{name: "invalidates MetricCPUSystem", metricType: probe.MetricCPUSystem},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			err = probe.CacheEnable()
			require.NoError(t, err)
			defer func() { _ = probe.CacheDisable() }()

			err = probe.CacheInvalidate(tt.metricType)
			require.NoError(t, err)
		})
	}
}

// TestCachedCPUCollector verifies cached CPU collector.
func TestCachedCPUCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects CPU metrics with cache"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			err = probe.CacheEnable()
			require.NoError(t, err)
			defer func() { _ = probe.CacheDisable() }()

			collector := probe.NewCachedCPUCollector()
			require.NotNil(t, collector)

			ctx := context.Background()
			cpu, err := collector.CollectSystem(ctx)
			require.NoError(t, err)
			assert.GreaterOrEqual(t, cpu.UsagePercent, 0.0)
		})
	}
}

// TestCachedMemoryCollector verifies cached memory collector.
func TestCachedMemoryCollector(t *testing.T) {
	tests := []struct {
		name string
	}{
		{name: "collects memory metrics with cache"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := probe.Init()
			require.NoError(t, err)
			defer probe.Shutdown()

			err = probe.CacheEnable()
			require.NoError(t, err)
			defer func() { _ = probe.CacheDisable() }()

			collector := probe.NewCachedMemoryCollector()
			require.NotNil(t, collector)

			ctx := context.Background()
			mem, err := collector.CollectSystem(ctx)
			require.NoError(t, err)
			assert.Greater(t, mem.Total, uint64(0))
		})
	}
}
