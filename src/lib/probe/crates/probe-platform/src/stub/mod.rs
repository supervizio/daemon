//! Stub platform implementation for unsupported platforms
//!
//! Returns NotSupported errors for all operations.

use crate::{
    CPUCollector, CPUPressure, DiskCollector, DiskIOStats, DiskUsage, Error, IOCollector,
    IOPressure, IOStats, LoadAverage, LoadCollector, MemoryCollector, MemoryPressure, NetInterface,
    NetStats, NetworkCollector, Partition, ProcessCollector, ProcessMetrics, Result, SystemCPU,
    SystemCollector, SystemMemory,
};

/// Stub system collector for unsupported platforms.
pub struct StubCollector {
    cpu: StubCPUCollector,
    memory: StubMemoryCollector,
    load: StubLoadCollector,
    process: StubProcessCollector,
    disk: StubDiskCollector,
    network: StubNetworkCollector,
    io: StubIOCollector,
}

impl StubCollector {
    /// Create a new stub collector.
    pub fn new() -> Self {
        Self {
            cpu: StubCPUCollector,
            memory: StubMemoryCollector,
            load: StubLoadCollector,
            process: StubProcessCollector,
            disk: StubDiskCollector,
            network: StubNetworkCollector,
            io: StubIOCollector,
        }
    }
}

impl Default for StubCollector {
    fn default() -> Self {
        Self::new()
    }
}

impl SystemCollector for StubCollector {
    fn cpu(&self) -> &dyn CPUCollector {
        &self.cpu
    }

    fn memory(&self) -> &dyn MemoryCollector {
        &self.memory
    }

    fn load(&self) -> &dyn LoadCollector {
        &self.load
    }

    fn process(&self) -> &dyn ProcessCollector {
        &self.process
    }

    fn disk(&self) -> &dyn DiskCollector {
        &self.disk
    }

    fn network(&self) -> &dyn NetworkCollector {
        &self.network
    }

    fn io(&self) -> &dyn IOCollector {
        &self.io
    }
}

// ============================================================================
// CPU COLLECTOR
// ============================================================================

struct StubCPUCollector;

impl CPUCollector for StubCPUCollector {
    fn collect_system(&self) -> Result<SystemCPU> {
        Err(Error::NotSupported)
    }

    fn collect_pressure(&self) -> Result<CPUPressure> {
        Err(Error::NotSupported)
    }
}

// ============================================================================
// MEMORY COLLECTOR
// ============================================================================

struct StubMemoryCollector;

impl MemoryCollector for StubMemoryCollector {
    fn collect_system(&self) -> Result<SystemMemory> {
        Err(Error::NotSupported)
    }

    fn collect_pressure(&self) -> Result<MemoryPressure> {
        Err(Error::NotSupported)
    }
}

// ============================================================================
// LOAD COLLECTOR
// ============================================================================

struct StubLoadCollector;

impl LoadCollector for StubLoadCollector {
    fn collect(&self) -> Result<LoadAverage> {
        Err(Error::NotSupported)
    }
}

// ============================================================================
// PROCESS COLLECTOR
// ============================================================================

struct StubProcessCollector;

impl ProcessCollector for StubProcessCollector {
    fn collect(&self, _pid: i32) -> Result<ProcessMetrics> {
        Err(Error::NotSupported)
    }

    fn collect_all(&self) -> Result<Vec<ProcessMetrics>> {
        Err(Error::NotSupported)
    }
}

// ============================================================================
// DISK COLLECTOR
// ============================================================================

struct StubDiskCollector;

impl DiskCollector for StubDiskCollector {
    fn list_partitions(&self) -> Result<Vec<Partition>> {
        Err(Error::NotSupported)
    }

    fn collect_usage(&self, _path: &str) -> Result<DiskUsage> {
        Err(Error::NotSupported)
    }

    fn collect_all_usage(&self) -> Result<Vec<DiskUsage>> {
        Err(Error::NotSupported)
    }

    fn collect_io(&self) -> Result<Vec<DiskIOStats>> {
        Err(Error::NotSupported)
    }

    fn collect_device_io(&self, _device: &str) -> Result<DiskIOStats> {
        Err(Error::NotSupported)
    }
}

// ============================================================================
// NETWORK COLLECTOR
// ============================================================================

struct StubNetworkCollector;

impl NetworkCollector for StubNetworkCollector {
    fn list_interfaces(&self) -> Result<Vec<NetInterface>> {
        Err(Error::NotSupported)
    }

    fn collect_stats(&self, _interface: &str) -> Result<NetStats> {
        Err(Error::NotSupported)
    }

    fn collect_all_stats(&self) -> Result<Vec<NetStats>> {
        Err(Error::NotSupported)
    }
}

// ============================================================================
// I/O COLLECTOR
// ============================================================================

struct StubIOCollector;

impl IOCollector for StubIOCollector {
    fn collect_stats(&self) -> Result<IOStats> {
        Err(Error::NotSupported)
    }

    fn collect_pressure(&self) -> Result<IOPressure> {
        Err(Error::NotSupported)
    }
}
