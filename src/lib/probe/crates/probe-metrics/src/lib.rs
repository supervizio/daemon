//! probe-metrics - Abstract metrics collection traits
//!
//! This crate defines the interfaces for system metrics collection
//! that are implemented by platform-specific code.

use thiserror::Error;

/// Error types for metrics collection.
#[derive(Error, Debug)]
pub enum Error {
    /// Operation not supported on this platform.
    #[error("operation not supported on this platform")]
    NotSupported,

    /// Permission denied.
    #[error("permission denied: {0}")]
    Permission(String),

    /// Resource not found.
    #[error("resource not found: {0}")]
    NotFound(String),

    /// I/O error.
    #[error("I/O error: {0}")]
    Io(#[from] std::io::Error),

    /// Platform-specific error.
    #[error("platform error: {0}")]
    Platform(String),
}

/// Result type alias for metrics operations.
pub type Result<T> = std::result::Result<T, Error>;

// ============================================================================
// CPU METRICS
// ============================================================================

/// System CPU metrics.
#[derive(Debug, Clone, Default)]
pub struct SystemCPU {
    /// User CPU percentage (0-100).
    pub user_percent: f64,
    /// System CPU percentage (0-100).
    pub system_percent: f64,
    /// Idle CPU percentage (0-100).
    pub idle_percent: f64,
    /// I/O wait percentage (Linux only, 0 on other platforms).
    pub iowait_percent: f64,
    /// Steal percentage (VMs only, 0 otherwise).
    pub steal_percent: f64,
    /// Number of CPU cores.
    pub cores: u32,
    /// CPU frequency in MHz.
    pub frequency_mhz: u64,
}

/// Load average (Unix systems).
#[derive(Debug, Clone, Default)]
pub struct LoadAverage {
    /// 1-minute load average.
    pub load_1min: f64,
    /// 5-minute load average.
    pub load_5min: f64,
    /// 15-minute load average.
    pub load_15min: f64,
}

/// CPU pressure metrics (PSI - Pressure Stall Information).
/// Available on Linux 4.20+ via /proc/pressure/cpu.
#[derive(Debug, Clone, Default)]
pub struct CPUPressure {
    /// Percentage of time some tasks were stalled (10s average).
    pub some_avg10: f64,
    /// Percentage of time some tasks were stalled (60s average).
    pub some_avg60: f64,
    /// Percentage of time some tasks were stalled (300s average).
    pub some_avg300: f64,
    /// Total microseconds some tasks were stalled.
    pub some_total_us: u64,
}

// ============================================================================
// MEMORY METRICS
// ============================================================================

/// System memory metrics.
#[derive(Debug, Clone, Default)]
pub struct SystemMemory {
    /// Total physical memory in bytes.
    pub total_bytes: u64,
    /// Available memory in bytes.
    pub available_bytes: u64,
    /// Used memory in bytes.
    pub used_bytes: u64,
    /// Cached memory in bytes.
    pub cached_bytes: u64,
    /// Buffer memory in bytes (Linux only, 0 on other platforms).
    pub buffers_bytes: u64,
    /// Total swap in bytes.
    pub swap_total_bytes: u64,
    /// Used swap in bytes.
    pub swap_used_bytes: u64,
}

/// Memory pressure metrics (PSI).
/// Available on Linux 4.20+ via /proc/pressure/memory.
#[derive(Debug, Clone, Default)]
pub struct MemoryPressure {
    /// Percentage of time some tasks were stalled (10s average).
    pub some_avg10: f64,
    /// Percentage of time some tasks were stalled (60s average).
    pub some_avg60: f64,
    /// Percentage of time some tasks were stalled (300s average).
    pub some_avg300: f64,
    /// Total microseconds some tasks were stalled.
    pub some_total_us: u64,
    /// Percentage of time all tasks were stalled (10s average).
    pub full_avg10: f64,
    /// Percentage of time all tasks were stalled (60s average).
    pub full_avg60: f64,
    /// Percentage of time all tasks were stalled (300s average).
    pub full_avg300: f64,
    /// Total microseconds all tasks were stalled.
    pub full_total_us: u64,
}

// ============================================================================
// PROCESS METRICS
// ============================================================================

/// Process state.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
#[repr(u8)]
pub enum ProcessState {
    /// Process is running.
    Running = 0,
    /// Process is sleeping.
    Sleeping = 1,
    /// Process is waiting.
    Waiting = 2,
    /// Process is a zombie.
    Zombie = 3,
    /// Process is stopped.
    Stopped = 4,
    /// Unknown state.
    #[default]
    Unknown = 255,
}

/// Process metrics.
#[derive(Debug, Clone, Default)]
pub struct ProcessMetrics {
    /// Process ID.
    pub pid: i32,
    /// CPU usage percentage (0-100 per core).
    pub cpu_percent: f64,
    /// Resident set size in bytes.
    pub memory_rss_bytes: u64,
    /// Virtual memory size in bytes.
    pub memory_vms_bytes: u64,
    /// Memory usage percentage.
    pub memory_percent: f64,
    /// Number of threads.
    pub num_threads: u32,
    /// Number of file descriptors.
    pub num_fds: u32,
    /// Read bytes per second.
    pub read_bytes_per_sec: u64,
    /// Write bytes per second.
    pub write_bytes_per_sec: u64,
    /// Process state.
    pub state: ProcessState,
}

// ============================================================================
// DISK METRICS
// ============================================================================

/// Mounted partition information.
#[derive(Debug, Clone, Default)]
pub struct Partition {
    /// Device path (e.g., /dev/sda1).
    pub device: String,
    /// Mount point (e.g., /).
    pub mount_point: String,
    /// Filesystem type (e.g., ext4, xfs).
    pub fs_type: String,
    /// Mount options.
    pub options: String,
}

/// Disk usage for a mount point.
#[derive(Debug, Clone, Default)]
pub struct DiskUsage {
    /// Mount point path.
    pub path: String,
    /// Total space in bytes.
    pub total_bytes: u64,
    /// Used space in bytes.
    pub used_bytes: u64,
    /// Free space in bytes.
    pub free_bytes: u64,
    /// Usage percentage (0-100).
    pub used_percent: f64,
    /// Total inodes.
    pub inodes_total: u64,
    /// Used inodes.
    pub inodes_used: u64,
    /// Free inodes.
    pub inodes_free: u64,
}

/// Block device I/O statistics.
#[derive(Debug, Clone, Default)]
pub struct DiskIOStats {
    /// Device name (e.g., sda).
    pub device: String,
    /// Read operations completed.
    pub reads_completed: u64,
    /// Sectors read.
    pub sectors_read: u64,
    /// Time spent reading (milliseconds).
    pub read_time_ms: u64,
    /// Write operations completed.
    pub writes_completed: u64,
    /// Sectors written.
    pub sectors_written: u64,
    /// Time spent writing (milliseconds).
    pub write_time_ms: u64,
    /// I/O operations currently in progress.
    pub io_in_progress: u64,
    /// Time spent doing I/O (milliseconds).
    pub io_time_ms: u64,
    /// Weighted time spent doing I/O (milliseconds).
    pub weighted_io_time_ms: u64,
}

// ============================================================================
// NETWORK METRICS
// ============================================================================

/// Network interface information.
#[derive(Debug, Clone, Default)]
pub struct NetInterface {
    /// Interface name (e.g., eth0).
    pub name: String,
    /// MAC address.
    pub mac_address: String,
    /// IPv4 addresses.
    pub ipv4_addresses: Vec<String>,
    /// IPv6 addresses.
    pub ipv6_addresses: Vec<String>,
    /// MTU (Maximum Transmission Unit).
    pub mtu: u32,
    /// Whether interface is up.
    pub is_up: bool,
    /// Whether interface is loopback.
    pub is_loopback: bool,
}

/// Network interface statistics.
#[derive(Debug, Clone, Default)]
pub struct NetStats {
    /// Interface name.
    pub interface: String,
    /// Bytes received.
    pub rx_bytes: u64,
    /// Packets received.
    pub rx_packets: u64,
    /// Receive errors.
    pub rx_errors: u64,
    /// Receive drops.
    pub rx_drops: u64,
    /// Bytes transmitted.
    pub tx_bytes: u64,
    /// Packets transmitted.
    pub tx_packets: u64,
    /// Transmit errors.
    pub tx_errors: u64,
    /// Transmit drops.
    pub tx_drops: u64,
}

// ============================================================================
// I/O METRICS
// ============================================================================

/// System-wide I/O statistics.
#[derive(Debug, Clone, Default)]
pub struct IOStats {
    /// Total read operations.
    pub read_ops: u64,
    /// Total bytes read.
    pub read_bytes: u64,
    /// Total write operations.
    pub write_ops: u64,
    /// Total bytes written.
    pub write_bytes: u64,
}

/// Context switch statistics.
///
/// Includes both per-process and system-wide context switches.
#[derive(Debug, Clone, Default)]
pub struct ContextSwitches {
    /// Voluntary context switches (process yielded CPU).
    pub voluntary: u64,
    /// Involuntary context switches (preempted by scheduler).
    pub involuntary: u64,
    /// System-wide total context switches.
    pub system_total: u64,
}

/// I/O pressure metrics (PSI).
/// Available on Linux 4.20+ via /proc/pressure/io.
#[derive(Debug, Clone, Default)]
pub struct IOPressure {
    /// Percentage of time some tasks were stalled (10s average).
    pub some_avg10: f64,
    /// Percentage of time some tasks were stalled (60s average).
    pub some_avg60: f64,
    /// Percentage of time some tasks were stalled (300s average).
    pub some_avg300: f64,
    /// Total microseconds some tasks were stalled.
    pub some_total_us: u64,
    /// Percentage of time all tasks were stalled (10s average).
    pub full_avg10: f64,
    /// Percentage of time all tasks were stalled (60s average).
    pub full_avg60: f64,
    /// Percentage of time all tasks were stalled (300s average).
    pub full_avg300: f64,
    /// Total microseconds all tasks were stalled.
    pub full_total_us: u64,
}

// ============================================================================
// COLLECTOR TRAITS
// ============================================================================

/// Trait for CPU metrics collection.
pub trait CPUCollector: Send + Sync {
    /// Collect system-wide CPU metrics.
    fn collect_system(&self) -> Result<SystemCPU>;
    /// Collect CPU pressure metrics (PSI).
    fn collect_pressure(&self) -> Result<CPUPressure>;
}

/// Trait for memory metrics collection.
pub trait MemoryCollector: Send + Sync {
    /// Collect system-wide memory metrics.
    fn collect_system(&self) -> Result<SystemMemory>;
    /// Collect memory pressure metrics (PSI).
    fn collect_pressure(&self) -> Result<MemoryPressure>;
}

/// Trait for load average collection.
pub trait LoadCollector: Send + Sync {
    /// Collect system load average.
    fn collect(&self) -> Result<LoadAverage>;
}

/// Trait for process metrics collection.
pub trait ProcessCollector: Send + Sync {
    /// Collect metrics for a specific process.
    fn collect(&self, pid: i32) -> Result<ProcessMetrics>;
    /// Collect metrics for all processes.
    fn collect_all(&self) -> Result<Vec<ProcessMetrics>>;
}

/// Trait for disk metrics collection.
pub trait DiskCollector: Send + Sync {
    /// List all mounted partitions.
    fn list_partitions(&self) -> Result<Vec<Partition>>;
    /// Collect disk usage for a specific path.
    fn collect_usage(&self, path: &str) -> Result<DiskUsage>;
    /// Collect disk usage for all partitions.
    fn collect_all_usage(&self) -> Result<Vec<DiskUsage>>;
    /// Collect I/O statistics for all block devices.
    fn collect_io(&self) -> Result<Vec<DiskIOStats>>;
    /// Collect I/O statistics for a specific device.
    fn collect_device_io(&self, device: &str) -> Result<DiskIOStats>;
}

/// Trait for network metrics collection.
pub trait NetworkCollector: Send + Sync {
    /// List all network interfaces.
    fn list_interfaces(&self) -> Result<Vec<NetInterface>>;
    /// Collect statistics for a specific interface.
    fn collect_stats(&self, interface: &str) -> Result<NetStats>;
    /// Collect statistics for all interfaces.
    fn collect_all_stats(&self) -> Result<Vec<NetStats>>;
}

/// Trait for I/O metrics collection.
pub trait IOCollector: Send + Sync {
    /// Collect system-wide I/O statistics.
    fn collect_stats(&self) -> Result<IOStats>;
    /// Collect I/O pressure metrics (PSI).
    fn collect_pressure(&self) -> Result<IOPressure>;
}

// ============================================================================
// THERMAL METRICS
// ============================================================================

/// Thermal zone information and temperature reading.
#[derive(Debug, Clone, Default)]
pub struct ThermalZone {
    /// Device name (e.g., "coretemp", "acpitz", "nvme").
    pub name: String,
    /// Zone label (e.g., "Core 0", "Package id 0").
    pub label: String,
    /// Current temperature in Celsius.
    pub temp_celsius: f64,
    /// Maximum safe temperature in Celsius (if available).
    pub temp_max: Option<f64>,
    /// Critical temperature in Celsius (if available).
    pub temp_crit: Option<f64>,
}

/// Trait for thermal metrics collection.
pub trait ThermalCollector: Send + Sync {
    /// Check if thermal monitoring is supported.
    fn is_supported(&self) -> bool;
    /// List all thermal zones.
    fn list_zones(&self) -> Result<Vec<ThermalZone>>;
    /// Collect current temperatures for all zones.
    fn collect_temperatures(&self) -> Result<Vec<ThermalZone>>;
}

// ============================================================================
// NETWORK CONNECTIONS
// ============================================================================

/// Socket state (matching Linux TCP states).
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
#[repr(u8)]
pub enum SocketState {
    /// Established connection.
    Established = 1,
    /// Sent SYN.
    SynSent = 2,
    /// Received SYN.
    SynRecv = 3,
    /// FIN-WAIT-1.
    FinWait1 = 4,
    /// FIN-WAIT-2.
    FinWait2 = 5,
    /// TIME-WAIT.
    TimeWait = 6,
    /// Closed.
    Close = 7,
    /// CLOSE-WAIT.
    CloseWait = 8,
    /// LAST-ACK.
    LastAck = 9,
    /// Listening for connections.
    Listen = 10,
    /// CLOSING.
    Closing = 11,
    /// Unknown state.
    #[default]
    Unknown = 0,
}

impl SocketState {
    /// Create from Linux TCP state code.
    pub fn from_linux_state(state: u8) -> Self {
        match state {
            1 => SocketState::Established,
            2 => SocketState::SynSent,
            3 => SocketState::SynRecv,
            4 => SocketState::FinWait1,
            5 => SocketState::FinWait2,
            6 => SocketState::TimeWait,
            7 => SocketState::Close,
            8 => SocketState::CloseWait,
            9 => SocketState::LastAck,
            10 => SocketState::Listen,
            11 => SocketState::Closing,
            _ => SocketState::Unknown,
        }
    }
}

/// Address family for network connections.
#[derive(Debug, Clone, Copy, PartialEq, Eq, Default)]
#[repr(u8)]
pub enum AddressFamily {
    /// IPv4 address.
    #[default]
    IPv4 = 4,
    /// IPv6 address.
    IPv6 = 6,
}

/// TCP connection information.
#[derive(Debug, Clone, Default)]
pub struct TcpConnection {
    /// Address family (IPv4 or IPv6).
    pub family: AddressFamily,
    /// Local IP address.
    pub local_addr: String,
    /// Local port.
    pub local_port: u16,
    /// Remote IP address.
    pub remote_addr: String,
    /// Remote port.
    pub remote_port: u16,
    /// Connection state.
    pub state: SocketState,
    /// Process ID owning this connection (-1 if unknown).
    pub pid: i32,
    /// Process name (empty if unknown).
    pub process_name: String,
    /// Socket inode number.
    pub inode: u64,
    /// Receive queue size.
    pub rx_queue: u32,
    /// Transmit queue size.
    pub tx_queue: u32,
}

/// UDP socket information.
#[derive(Debug, Clone, Default)]
pub struct UdpConnection {
    /// Address family (IPv4 or IPv6).
    pub family: AddressFamily,
    /// Local IP address.
    pub local_addr: String,
    /// Local port.
    pub local_port: u16,
    /// Remote IP address (may be 0.0.0.0 for unconnected).
    pub remote_addr: String,
    /// Remote port (may be 0 for unconnected).
    pub remote_port: u16,
    /// Connection state.
    pub state: SocketState,
    /// Process ID owning this socket (-1 if unknown).
    pub pid: i32,
    /// Process name (empty if unknown).
    pub process_name: String,
    /// Socket inode number.
    pub inode: u64,
    /// Receive queue size.
    pub rx_queue: u32,
    /// Transmit queue size.
    pub tx_queue: u32,
}

/// Unix domain socket information.
#[derive(Debug, Clone, Default)]
pub struct UnixSocket {
    /// Socket path (may be empty for abstract sockets).
    pub path: String,
    /// Socket type (stream, dgram, seqpacket).
    pub socket_type: String,
    /// Connection state.
    pub state: SocketState,
    /// Process ID owning this socket (-1 if unknown).
    pub pid: i32,
    /// Process name (empty if unknown).
    pub process_name: String,
    /// Socket inode number.
    pub inode: u64,
}

/// Aggregated TCP connection statistics.
#[derive(Debug, Clone, Default)]
pub struct TcpStats {
    /// Number of established connections.
    pub established: u32,
    /// Number of SYN_SENT connections.
    pub syn_sent: u32,
    /// Number of SYN_RECV connections.
    pub syn_recv: u32,
    /// Number of FIN_WAIT1 connections.
    pub fin_wait1: u32,
    /// Number of FIN_WAIT2 connections.
    pub fin_wait2: u32,
    /// Number of TIME_WAIT connections.
    pub time_wait: u32,
    /// Number of CLOSE connections.
    pub close: u32,
    /// Number of CLOSE_WAIT connections.
    pub close_wait: u32,
    /// Number of LAST_ACK connections.
    pub last_ack: u32,
    /// Number of LISTEN sockets.
    pub listen: u32,
    /// Number of CLOSING connections.
    pub closing: u32,
}

/// Trait for network connection collection.
pub trait ConnectionCollector: Send + Sync {
    /// Collect all TCP connections.
    fn collect_tcp(&self) -> Result<Vec<TcpConnection>>;

    /// Collect all UDP sockets.
    fn collect_udp(&self) -> Result<Vec<UdpConnection>>;

    /// Collect all Unix domain sockets.
    fn collect_unix(&self) -> Result<Vec<UnixSocket>>;

    /// Collect aggregated TCP statistics.
    fn collect_tcp_stats(&self) -> Result<TcpStats>;

    /// Collect connections for a specific process.
    fn collect_process_connections(
        &self,
        pid: i32,
    ) -> Result<(Vec<TcpConnection>, Vec<UdpConnection>)>;

    /// Find which process owns a specific port.
    fn find_process_by_port(&self, port: u16, tcp: bool) -> Result<Option<i32>>;
}

// ============================================================================
// AGGREGATED METRICS
// ============================================================================

/// All pressure metrics combined (Linux PSI).
#[derive(Debug, Clone, Default)]
pub struct AllPressure {
    /// CPU pressure metrics.
    pub cpu: CPUPressure,
    /// Memory pressure metrics.
    pub memory: MemoryPressure,
    /// I/O pressure metrics.
    pub io: IOPressure,
}

/// All system metrics collected in one call.
///
/// This structure contains all the metrics that can be collected
/// by the system collector in a single aggregated call.
#[derive(Debug, Clone, Default)]
pub struct AllMetrics {
    /// System CPU metrics.
    pub cpu: SystemCPU,
    /// System memory metrics.
    pub memory: SystemMemory,
    /// System load average.
    pub load: LoadAverage,
    /// System I/O statistics.
    pub io_stats: IOStats,
    /// Disk partition list.
    pub partitions: Vec<Partition>,
    /// Disk usage for all partitions.
    pub disk_usage: Vec<DiskUsage>,
    /// Disk I/O statistics.
    pub disk_io: Vec<DiskIOStats>,
    /// Network interface list.
    pub net_interfaces: Vec<NetInterface>,
    /// Network statistics.
    pub net_stats: Vec<NetStats>,
    /// Pressure metrics (Linux only, None on other platforms).
    pub pressure: Option<AllPressure>,
    /// Timestamp when metrics were collected (nanoseconds since epoch).
    pub timestamp_ns: u64,
}

/// Combined system collector interface.
pub trait SystemCollector: Send + Sync {
    /// Get CPU collector.
    fn cpu(&self) -> &dyn CPUCollector;
    /// Get memory collector.
    fn memory(&self) -> &dyn MemoryCollector;
    /// Get load collector.
    fn load(&self) -> &dyn LoadCollector;
    /// Get process collector.
    fn process(&self) -> &dyn ProcessCollector;
    /// Get disk collector.
    fn disk(&self) -> &dyn DiskCollector;
    /// Get network collector.
    fn network(&self) -> &dyn NetworkCollector;
    /// Get I/O collector.
    fn io(&self) -> &dyn IOCollector;

    /// Collect all metrics in one call.
    ///
    /// This is more efficient than calling each collector individually
    /// as it reduces the number of system calls and provides a consistent
    /// snapshot of all metrics at approximately the same point in time.
    fn collect_all(&self) -> Result<AllMetrics> {
        use std::time::{SystemTime, UNIX_EPOCH};

        let timestamp_ns = SystemTime::now()
            .duration_since(UNIX_EPOCH)
            .map(|d| d.as_nanos() as u64)
            .unwrap_or(0);

        // Collect all metrics, using defaults for any that fail
        let cpu = self.cpu().collect_system().unwrap_or_default();
        let memory = self.memory().collect_system().unwrap_or_default();
        let load = self.load().collect().unwrap_or_default();
        let io_stats = self.io().collect_stats().unwrap_or_default();

        let partitions = self.disk().list_partitions().unwrap_or_default();
        let disk_usage = self.disk().collect_all_usage().unwrap_or_default();
        let disk_io = self.disk().collect_io().unwrap_or_default();

        let net_interfaces = self.network().list_interfaces().unwrap_or_default();
        let net_stats = self.network().collect_all_stats().unwrap_or_default();

        // Try to collect pressure metrics (Linux only)
        let pressure = match (
            self.cpu().collect_pressure(),
            self.memory().collect_pressure(),
            self.io().collect_pressure(),
        ) {
            (Ok(cpu_p), Ok(mem_p), Ok(io_p)) => Some(AllPressure {
                cpu: cpu_p,
                memory: mem_p,
                io: io_p,
            }),
            _ => None,
        };

        Ok(AllMetrics {
            cpu,
            memory,
            load,
            io_stats,
            partitions,
            disk_usage,
            disk_io,
            net_interfaces,
            net_stats,
            pressure,
            timestamp_ns,
        })
    }
}
