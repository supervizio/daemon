//! Cache policy configuration for different metric types.

use std::time::Duration;

/// Types of metrics that can be cached.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Hash)]
#[repr(u8)]
pub enum MetricType {
    /// System CPU metrics.
    CpuSystem = 0,
    /// CPU pressure metrics (PSI).
    CpuPressure = 1,
    /// System memory metrics.
    MemorySystem = 2,
    /// Memory pressure metrics (PSI).
    MemoryPressure = 3,
    /// System load average.
    Load = 4,
    /// Disk partition list.
    DiskPartitions = 5,
    /// Disk usage metrics.
    DiskUsage = 6,
    /// Disk I/O statistics.
    DiskIo = 7,
    /// Network interface list.
    NetInterfaces = 8,
    /// Network interface statistics.
    NetStats = 9,
    /// System I/O statistics.
    IoStats = 10,
    /// I/O pressure metrics (PSI).
    IoPressure = 11,
}

impl MetricType {
    /// Convert from u8 value.
    pub fn from_u8(value: u8) -> Option<Self> {
        match value {
            0 => Some(Self::CpuSystem),
            1 => Some(Self::CpuPressure),
            2 => Some(Self::MemorySystem),
            3 => Some(Self::MemoryPressure),
            4 => Some(Self::Load),
            5 => Some(Self::DiskPartitions),
            6 => Some(Self::DiskUsage),
            7 => Some(Self::DiskIo),
            8 => Some(Self::NetInterfaces),
            9 => Some(Self::NetStats),
            10 => Some(Self::IoStats),
            11 => Some(Self::IoPressure),
            _ => None,
        }
    }
}

/// Cache TTL policies for different metric types.
///
/// Different metrics have different volatility levels:
/// - CPU metrics change rapidly -> short TTL (100ms)
/// - Memory metrics are relatively stable -> medium TTL (500ms)
/// - Disk partition list rarely changes -> long TTL (30s)
/// - Network stats change frequently -> short TTL (500ms)
#[derive(Debug, Clone)]
pub struct CachePolicies {
    cpu_system_ttl: Duration,
    cpu_pressure_ttl: Duration,
    memory_system_ttl: Duration,
    memory_pressure_ttl: Duration,
    load_ttl: Duration,
    disk_partitions_ttl: Duration,
    disk_usage_ttl: Duration,
    disk_io_ttl: Duration,
    net_interfaces_ttl: Duration,
    net_stats_ttl: Duration,
    io_stats_ttl: Duration,
    io_pressure_ttl: Duration,
}

impl Default for CachePolicies {
    fn default() -> Self {
        Self {
            // CPU metrics - high volatility
            cpu_system_ttl: Duration::from_millis(100),
            cpu_pressure_ttl: Duration::from_millis(500),

            // Memory metrics - medium volatility
            memory_system_ttl: Duration::from_millis(500),
            memory_pressure_ttl: Duration::from_millis(500),

            // Load average - low volatility (already averaged)
            load_ttl: Duration::from_secs(1),

            // Disk metrics - varying volatility
            disk_partitions_ttl: Duration::from_secs(30), // Rarely changes
            disk_usage_ttl: Duration::from_secs(5),       // Moderate changes
            disk_io_ttl: Duration::from_secs(1),          // Frequent changes

            // Network metrics - medium volatility
            net_interfaces_ttl: Duration::from_secs(30), // Rarely changes
            net_stats_ttl: Duration::from_millis(500),   // Frequent changes

            // I/O metrics - high volatility
            io_stats_ttl: Duration::from_millis(500),
            io_pressure_ttl: Duration::from_millis(500),
        }
    }
}

impl CachePolicies {
    /// Create policies with no caching (TTL = 0).
    pub fn no_cache() -> Self {
        Self {
            cpu_system_ttl: Duration::ZERO,
            cpu_pressure_ttl: Duration::ZERO,
            memory_system_ttl: Duration::ZERO,
            memory_pressure_ttl: Duration::ZERO,
            load_ttl: Duration::ZERO,
            disk_partitions_ttl: Duration::ZERO,
            disk_usage_ttl: Duration::ZERO,
            disk_io_ttl: Duration::ZERO,
            net_interfaces_ttl: Duration::ZERO,
            net_stats_ttl: Duration::ZERO,
            io_stats_ttl: Duration::ZERO,
            io_pressure_ttl: Duration::ZERO,
        }
    }

    /// Create policies with a uniform TTL for all metrics.
    pub fn uniform(ttl: Duration) -> Self {
        Self {
            cpu_system_ttl: ttl,
            cpu_pressure_ttl: ttl,
            memory_system_ttl: ttl,
            memory_pressure_ttl: ttl,
            load_ttl: ttl,
            disk_partitions_ttl: ttl,
            disk_usage_ttl: ttl,
            disk_io_ttl: ttl,
            net_interfaces_ttl: ttl,
            net_stats_ttl: ttl,
            io_stats_ttl: ttl,
            io_pressure_ttl: ttl,
        }
    }

    /// Create policies optimized for high-frequency collection (short TTLs).
    pub fn high_frequency() -> Self {
        Self {
            cpu_system_ttl: Duration::from_millis(50),
            cpu_pressure_ttl: Duration::from_millis(100),
            memory_system_ttl: Duration::from_millis(100),
            memory_pressure_ttl: Duration::from_millis(100),
            load_ttl: Duration::from_millis(500),
            disk_partitions_ttl: Duration::from_secs(10),
            disk_usage_ttl: Duration::from_secs(1),
            disk_io_ttl: Duration::from_millis(500),
            net_interfaces_ttl: Duration::from_secs(10),
            net_stats_ttl: Duration::from_millis(100),
            io_stats_ttl: Duration::from_millis(100),
            io_pressure_ttl: Duration::from_millis(100),
        }
    }

    /// Create policies optimized for low-frequency collection (longer TTLs).
    pub fn low_frequency() -> Self {
        Self {
            cpu_system_ttl: Duration::from_secs(1),
            cpu_pressure_ttl: Duration::from_secs(5),
            memory_system_ttl: Duration::from_secs(5),
            memory_pressure_ttl: Duration::from_secs(5),
            load_ttl: Duration::from_secs(10),
            disk_partitions_ttl: Duration::from_secs(60),
            disk_usage_ttl: Duration::from_secs(30),
            disk_io_ttl: Duration::from_secs(10),
            net_interfaces_ttl: Duration::from_secs(60),
            net_stats_ttl: Duration::from_secs(5),
            io_stats_ttl: Duration::from_secs(5),
            io_pressure_ttl: Duration::from_secs(5),
        }
    }

    /// Get the TTL for a specific metric type.
    pub fn get_ttl(&self, metric: MetricType) -> Duration {
        match metric {
            MetricType::CpuSystem => self.cpu_system_ttl,
            MetricType::CpuPressure => self.cpu_pressure_ttl,
            MetricType::MemorySystem => self.memory_system_ttl,
            MetricType::MemoryPressure => self.memory_pressure_ttl,
            MetricType::Load => self.load_ttl,
            MetricType::DiskPartitions => self.disk_partitions_ttl,
            MetricType::DiskUsage => self.disk_usage_ttl,
            MetricType::DiskIo => self.disk_io_ttl,
            MetricType::NetInterfaces => self.net_interfaces_ttl,
            MetricType::NetStats => self.net_stats_ttl,
            MetricType::IoStats => self.io_stats_ttl,
            MetricType::IoPressure => self.io_pressure_ttl,
        }
    }

    /// Set the TTL for a specific metric type.
    pub fn set_ttl(&mut self, metric: MetricType, ttl: Duration) {
        match metric {
            MetricType::CpuSystem => self.cpu_system_ttl = ttl,
            MetricType::CpuPressure => self.cpu_pressure_ttl = ttl,
            MetricType::MemorySystem => self.memory_system_ttl = ttl,
            MetricType::MemoryPressure => self.memory_pressure_ttl = ttl,
            MetricType::Load => self.load_ttl = ttl,
            MetricType::DiskPartitions => self.disk_partitions_ttl = ttl,
            MetricType::DiskUsage => self.disk_usage_ttl = ttl,
            MetricType::DiskIo => self.disk_io_ttl = ttl,
            MetricType::NetInterfaces => self.net_interfaces_ttl = ttl,
            MetricType::NetStats => self.net_stats_ttl = ttl,
            MetricType::IoStats => self.io_stats_ttl = ttl,
            MetricType::IoPressure => self.io_pressure_ttl = ttl,
        }
    }

    /// Set TTL for CPU-related metrics.
    pub fn with_cpu_ttl(mut self, ttl: Duration) -> Self {
        self.cpu_system_ttl = ttl;
        self.cpu_pressure_ttl = ttl;
        self
    }

    /// Set TTL for memory-related metrics.
    pub fn with_memory_ttl(mut self, ttl: Duration) -> Self {
        self.memory_system_ttl = ttl;
        self.memory_pressure_ttl = ttl;
        self
    }

    /// Set TTL for disk-related metrics.
    pub fn with_disk_ttl(mut self, ttl: Duration) -> Self {
        self.disk_partitions_ttl = ttl;
        self.disk_usage_ttl = ttl;
        self.disk_io_ttl = ttl;
        self
    }

    /// Set TTL for network-related metrics.
    pub fn with_network_ttl(mut self, ttl: Duration) -> Self {
        self.net_interfaces_ttl = ttl;
        self.net_stats_ttl = ttl;
        self
    }

    /// Set TTL for I/O-related metrics.
    pub fn with_io_ttl(mut self, ttl: Duration) -> Self {
        self.io_stats_ttl = ttl;
        self.io_pressure_ttl = ttl;
        self
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_default_policies() {
        let policies = CachePolicies::default();
        assert_eq!(
            policies.get_ttl(MetricType::CpuSystem),
            Duration::from_millis(100)
        );
        assert_eq!(
            policies.get_ttl(MetricType::MemorySystem),
            Duration::from_millis(500)
        );
        assert_eq!(
            policies.get_ttl(MetricType::DiskPartitions),
            Duration::from_secs(30)
        );
    }

    #[test]
    fn test_no_cache_policies() {
        let policies = CachePolicies::no_cache();
        assert_eq!(policies.get_ttl(MetricType::CpuSystem), Duration::ZERO);
        assert_eq!(policies.get_ttl(MetricType::MemorySystem), Duration::ZERO);
    }

    #[test]
    fn test_uniform_policies() {
        let ttl = Duration::from_secs(5);
        let policies = CachePolicies::uniform(ttl);
        assert_eq!(policies.get_ttl(MetricType::CpuSystem), ttl);
        assert_eq!(policies.get_ttl(MetricType::MemorySystem), ttl);
        assert_eq!(policies.get_ttl(MetricType::DiskPartitions), ttl);
    }

    #[test]
    fn test_set_ttl() {
        let mut policies = CachePolicies::default();
        let new_ttl = Duration::from_secs(10);
        policies.set_ttl(MetricType::CpuSystem, new_ttl);
        assert_eq!(policies.get_ttl(MetricType::CpuSystem), new_ttl);
    }

    #[test]
    fn test_builder_pattern() {
        let policies = CachePolicies::default()
            .with_cpu_ttl(Duration::from_millis(200))
            .with_memory_ttl(Duration::from_secs(1));

        assert_eq!(
            policies.get_ttl(MetricType::CpuSystem),
            Duration::from_millis(200)
        );
        assert_eq!(
            policies.get_ttl(MetricType::CpuPressure),
            Duration::from_millis(200)
        );
        assert_eq!(
            policies.get_ttl(MetricType::MemorySystem),
            Duration::from_secs(1)
        );
    }

    #[test]
    fn test_metric_type_from_u8() {
        assert_eq!(MetricType::from_u8(0), Some(MetricType::CpuSystem));
        assert_eq!(MetricType::from_u8(5), Some(MetricType::DiskPartitions));
        assert_eq!(MetricType::from_u8(255), None);
    }
}
