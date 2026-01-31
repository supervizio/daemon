//! probe-cache - TTL caching layer for metrics collection
//!
//! This crate provides a caching wrapper around system collectors to reduce
//! syscall overhead by caching metrics within configurable time windows.
//!
//! # Example
//!
//! ```ignore
//! use probe_cache::{CachedCollector, CachePolicies};
//! use probe_platform::new_collector;
//!
//! let collector = new_collector();
//! let policies = CachePolicies::default();
//! let cached = CachedCollector::new(collector, policies);
//!
//! // First call reads from system
//! let cpu1 = cached.cpu().collect_system();
//!
//! // Second call within TTL returns cached value
//! let cpu2 = cached.cpu().collect_system();
//! ```

mod policy;
mod ttl;

pub use policy::{CachePolicies, MetricType};
pub use ttl::{CacheEntry, TtlCache};

use parking_lot::RwLock;
use probe_metrics::{
    CPUCollector, CPUPressure, DiskCollector, DiskIOStats, DiskUsage, IOCollector, IOPressure,
    IOStats, LoadAverage, LoadCollector, MemoryCollector, MemoryPressure, NetInterface, NetStats,
    NetworkCollector, Partition, ProcessCollector, Result, SystemCPU, SystemCollector,
    SystemMemory,
};
use std::sync::Arc;

/// Cached metrics storage.
#[derive(Default)]
struct MetricsCache {
    cpu_system: Option<CacheEntry<SystemCPU>>,
    cpu_pressure: Option<CacheEntry<CPUPressure>>,
    memory_system: Option<CacheEntry<SystemMemory>>,
    memory_pressure: Option<CacheEntry<MemoryPressure>>,
    load: Option<CacheEntry<LoadAverage>>,
    partitions: Option<CacheEntry<Vec<Partition>>>,
    disk_usage: Option<CacheEntry<Vec<DiskUsage>>>,
    disk_io: Option<CacheEntry<Vec<DiskIOStats>>>,
    net_interfaces: Option<CacheEntry<Vec<NetInterface>>>,
    net_stats: Option<CacheEntry<Vec<NetStats>>>,
    io_stats: Option<CacheEntry<IOStats>>,
    io_pressure: Option<CacheEntry<IOPressure>>,
}

/// A caching wrapper around a SystemCollector.
///
/// Caches metric results for configurable TTL periods to reduce
/// the overhead of repeated system calls.
pub struct CachedCollector<T: SystemCollector> {
    inner: Arc<T>,
    cache: RwLock<MetricsCache>,
    policies: CachePolicies,
}

impl<T: SystemCollector> CachedCollector<T> {
    /// Create a new cached collector with the given policies.
    pub fn new(inner: T, policies: CachePolicies) -> Self {
        Self {
            inner: Arc::new(inner),
            cache: RwLock::new(MetricsCache::default()),
            policies,
        }
    }

    /// Create a new cached collector with default policies.
    pub fn with_defaults(inner: T) -> Self {
        Self::new(inner, CachePolicies::default())
    }

    /// Invalidate all cached metrics.
    pub fn invalidate_all(&self) {
        let mut cache = self.cache.write();
        *cache = MetricsCache::default();
    }

    /// Invalidate a specific metric type.
    pub fn invalidate(&self, metric: MetricType) {
        let mut cache = self.cache.write();
        match metric {
            MetricType::CpuSystem => cache.cpu_system = None,
            MetricType::CpuPressure => cache.cpu_pressure = None,
            MetricType::MemorySystem => cache.memory_system = None,
            MetricType::MemoryPressure => cache.memory_pressure = None,
            MetricType::Load => cache.load = None,
            MetricType::DiskPartitions => cache.partitions = None,
            MetricType::DiskUsage => cache.disk_usage = None,
            MetricType::DiskIo => cache.disk_io = None,
            MetricType::NetInterfaces => cache.net_interfaces = None,
            MetricType::NetStats => cache.net_stats = None,
            MetricType::IoStats => cache.io_stats = None,
            MetricType::IoPressure => cache.io_pressure = None,
        }
    }

    /// Update the TTL for a specific metric type.
    pub fn set_ttl(&mut self, metric: MetricType, ttl: std::time::Duration) {
        self.policies.set_ttl(metric, ttl);
    }

    /// Get the inner collector reference.
    pub fn inner(&self) -> &T {
        &self.inner
    }
}

// Implement SystemCollector for CachedCollector
impl<T: SystemCollector + 'static> SystemCollector for CachedCollector<T> {
    fn cpu(&self) -> &dyn CPUCollector {
        self
    }

    fn memory(&self) -> &dyn MemoryCollector {
        self
    }

    fn load(&self) -> &dyn LoadCollector {
        self
    }

    fn process(&self) -> &dyn ProcessCollector {
        // Process metrics are not cached (too dynamic)
        self.inner.process()
    }

    fn disk(&self) -> &dyn DiskCollector {
        self
    }

    fn network(&self) -> &dyn NetworkCollector {
        self
    }

    fn io(&self) -> &dyn IOCollector {
        self
    }
}

// Implement CPUCollector with caching
impl<T: SystemCollector + 'static> CPUCollector for CachedCollector<T> {
    fn collect_system(&self) -> Result<SystemCPU> {
        let ttl = self.policies.get_ttl(MetricType::CpuSystem);

        // Check cache first (read lock)
        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.cpu_system
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        // Cache miss - collect and store (write lock)
        let value = self.inner.cpu().collect_system()?;
        let mut cache = self.cache.write();
        cache.cpu_system = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }

    fn collect_pressure(&self) -> Result<CPUPressure> {
        let ttl = self.policies.get_ttl(MetricType::CpuPressure);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.cpu_pressure
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.cpu().collect_pressure()?;
        let mut cache = self.cache.write();
        cache.cpu_pressure = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }
}

// Implement MemoryCollector with caching
impl<T: SystemCollector + 'static> MemoryCollector for CachedCollector<T> {
    fn collect_system(&self) -> Result<SystemMemory> {
        let ttl = self.policies.get_ttl(MetricType::MemorySystem);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.memory_system
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.memory().collect_system()?;
        let mut cache = self.cache.write();
        cache.memory_system = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }

    fn collect_pressure(&self) -> Result<MemoryPressure> {
        let ttl = self.policies.get_ttl(MetricType::MemoryPressure);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.memory_pressure
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.memory().collect_pressure()?;
        let mut cache = self.cache.write();
        cache.memory_pressure = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }
}

// Implement LoadCollector with caching
impl<T: SystemCollector + 'static> LoadCollector for CachedCollector<T> {
    fn collect(&self) -> Result<LoadAverage> {
        let ttl = self.policies.get_ttl(MetricType::Load);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.load
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.load().collect()?;
        let mut cache = self.cache.write();
        cache.load = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }
}

// Implement DiskCollector with caching
impl<T: SystemCollector + 'static> DiskCollector for CachedCollector<T> {
    fn list_partitions(&self) -> Result<Vec<Partition>> {
        let ttl = self.policies.get_ttl(MetricType::DiskPartitions);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.partitions
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.disk().list_partitions()?;
        let mut cache = self.cache.write();
        cache.partitions = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }

    fn collect_usage(&self, path: &str) -> Result<DiskUsage> {
        // Individual path lookups are not cached
        self.inner.disk().collect_usage(path)
    }

    fn collect_all_usage(&self) -> Result<Vec<DiskUsage>> {
        let ttl = self.policies.get_ttl(MetricType::DiskUsage);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.disk_usage
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.disk().collect_all_usage()?;
        let mut cache = self.cache.write();
        cache.disk_usage = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }

    fn collect_io(&self) -> Result<Vec<DiskIOStats>> {
        let ttl = self.policies.get_ttl(MetricType::DiskIo);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.disk_io
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.disk().collect_io()?;
        let mut cache = self.cache.write();
        cache.disk_io = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }

    fn collect_device_io(&self, device: &str) -> Result<DiskIOStats> {
        // Individual device lookups are not cached
        self.inner.disk().collect_device_io(device)
    }
}

// Implement NetworkCollector with caching
impl<T: SystemCollector + 'static> NetworkCollector for CachedCollector<T> {
    fn list_interfaces(&self) -> Result<Vec<NetInterface>> {
        let ttl = self.policies.get_ttl(MetricType::NetInterfaces);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.net_interfaces
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.network().list_interfaces()?;
        let mut cache = self.cache.write();
        cache.net_interfaces = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }

    fn collect_stats(&self, interface: &str) -> Result<NetStats> {
        // Individual interface lookups are not cached
        self.inner.network().collect_stats(interface)
    }

    fn collect_all_stats(&self) -> Result<Vec<NetStats>> {
        let ttl = self.policies.get_ttl(MetricType::NetStats);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.net_stats
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.network().collect_all_stats()?;
        let mut cache = self.cache.write();
        cache.net_stats = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }
}

// Implement IOCollector with caching
impl<T: SystemCollector + 'static> IOCollector for CachedCollector<T> {
    fn collect_stats(&self) -> Result<IOStats> {
        let ttl = self.policies.get_ttl(MetricType::IoStats);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.io_stats
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.io().collect_stats()?;
        let mut cache = self.cache.write();
        cache.io_stats = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }

    fn collect_pressure(&self) -> Result<IOPressure> {
        let ttl = self.policies.get_ttl(MetricType::IoPressure);

        {
            let cache = self.cache.read();
            if let Some(entry) = &cache.io_pressure
                && entry.is_valid(ttl)
            {
                return Ok(entry.value.clone());
            }
        }

        let value = self.inner.io().collect_pressure()?;
        let mut cache = self.cache.write();
        cache.io_pressure = Some(CacheEntry::new(value.clone()));
        Ok(value)
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_cache_policies_default() {
        let policies = CachePolicies::default();
        assert!(policies.get_ttl(MetricType::CpuSystem).as_millis() > 0);
    }
}
