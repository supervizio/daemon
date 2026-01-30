# Plan Complet - Sonde Système Ultime

## Vision: Capturer TOUTES les Métriques Possibles

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│   ╔═══════════════════════════════════════════════════════════════════╗    │
│   ║                    LIBPROBE - SONDE ULTIME                         ║    │
│   ║                                                                    ║    │
│   ║   "Si le kernel le sait, la sonde doit pouvoir le collecter"      ║    │
│   ║                                                                    ║    │
│   ╚═══════════════════════════════════════════════════════════════════╝    │
│                                                                             │
│   Objectif: 100% des métriques système accessibles sur Linux/BSD/macOS     │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 1. PROCESS - Métriques Complètes

### 1.1 Ce qui existe ✅

```rust
pub struct ProcessMetrics {
    pub pid: i32,
    pub state: ProcessState,       // R/S/D/Z/T
    pub memory_rss_bytes: u64,
    pub memory_vms_bytes: u64,
    pub num_threads: u32,
    pub num_fds: u32,
}
```

### 1.2 Ce qu'il faut ajouter ⏳

```rust
/// Process complet avec TOUTES les informations possibles
pub struct ProcessFull {
    // === IDENTIFICATION ===
    pub pid: i32,
    pub ppid: i32,                      // Parent PID
    pub pgid: i32,                      // Process Group ID
    pub sid: i32,                       // Session ID
    pub tgid: i32,                      // Thread Group ID

    // === NOMS ET COMMANDES ===
    pub name: String,                   // /proc/[pid]/comm
    pub exe: String,                    // /proc/[pid]/exe (symlink)
    pub cmdline: Vec<String>,           // /proc/[pid]/cmdline
    pub cwd: String,                    // /proc/[pid]/cwd (symlink)
    pub root: String,                   // /proc/[pid]/root (symlink)

    // === UTILISATEUR/GROUPE ===
    pub uid_real: u32,
    pub uid_effective: u32,
    pub uid_saved: u32,
    pub uid_filesystem: u32,
    pub gid_real: u32,
    pub gid_effective: u32,
    pub gid_saved: u32,
    pub gid_filesystem: u32,
    pub groups: Vec<u32>,               // /proc/[pid]/status (Groups)

    // === ÉTAT ===
    pub state: ProcessState,
    pub state_char: char,               // R, S, D, Z, T, t, W, X, x, K, W, P
    pub nice: i32,                      // -20 to +19
    pub priority: i32,
    pub num_threads: u32,
    pub start_time: u64,                // Ticks since boot
    pub start_time_unix: u64,           // Unix timestamp

    // === CPU ===
    pub cpu_user_ticks: u64,            // utime
    pub cpu_system_ticks: u64,          // stime
    pub cpu_children_user: u64,         // cutime
    pub cpu_children_system: u64,       // cstime
    pub cpu_percent: f64,               // Calculé (delta)
    pub cpu_affinity: Vec<u32>,         // /proc/[pid]/status (Cpus_allowed_list)
    pub last_cpu: i32,                  // Dernier CPU utilisé

    // === MÉMOIRE DÉTAILLÉE ===
    pub memory_rss: u64,                // Resident Set Size
    pub memory_vms: u64,                // Virtual Memory Size
    pub memory_shared: u64,             // Shared memory
    pub memory_text: u64,               // Code segment
    pub memory_data: u64,               // Data + stack
    pub memory_lib: u64,                // Shared libraries
    pub memory_dirty: u64,              // Dirty pages
    pub memory_swap: u64,               // Swapped out
    pub memory_pss: u64,                // Proportional Set Size (from smaps)
    pub memory_uss: u64,                // Unique Set Size (from smaps)
    pub memory_percent: f64,
    pub page_faults_minor: u64,
    pub page_faults_major: u64,

    // === I/O ===
    pub io_read_bytes: u64,             // /proc/[pid]/io
    pub io_write_bytes: u64,
    pub io_read_chars: u64,             // Including cached
    pub io_write_chars: u64,
    pub io_read_syscalls: u64,
    pub io_write_syscalls: u64,
    pub io_cancelled_write: u64,

    // === FICHIERS ET SOCKETS ===
    pub num_fds: u32,
    pub open_files: Vec<OpenFile>,      // /proc/[pid]/fd/*
    pub open_sockets: Vec<OpenSocket>,  // Filtered from fd
    pub maps: Vec<MemoryMap>,           // /proc/[pid]/maps

    // === CONTEXT SWITCHES ===
    pub voluntary_ctx_switches: u64,
    pub nonvoluntary_ctx_switches: u64,

    // === LIMITES ===
    pub limits: ProcessLimits,          // /proc/[pid]/limits

    // === SÉCURITÉ ===
    pub seccomp_mode: u8,               // 0=disabled, 1=strict, 2=filter
    pub capabilities_effective: u64,
    pub capabilities_permitted: u64,
    pub capabilities_inheritable: u64,
    pub capabilities_bounding: u64,
    pub capabilities_ambient: u64,
    pub no_new_privs: bool,

    // === NAMESPACES ===
    pub ns_cgroup: u64,                 // /proc/[pid]/ns/cgroup
    pub ns_ipc: u64,
    pub ns_mnt: u64,
    pub ns_net: u64,
    pub ns_pid: u64,
    pub ns_pid_children: u64,
    pub ns_time: u64,
    pub ns_user: u64,
    pub ns_uts: u64,

    // === CGROUPS ===
    pub cgroup_path: String,            // /proc/[pid]/cgroup
    pub cgroup_controllers: Vec<String>,

    // === ENVIRONNEMENT ===
    pub environ: HashMap<String, String>, // /proc/[pid]/environ

    // === OOM ===
    pub oom_score: i32,                 // /proc/[pid]/oom_score
    pub oom_score_adj: i32,             // /proc/[pid]/oom_score_adj

    // === SCHEDULING ===
    pub sched_policy: u32,              // SCHED_NORMAL, SCHED_FIFO, etc.
    pub sched_priority: u32,
    pub cpu_time_limit: Option<u64>,    // From cgroup

    // === SIGNAUX ===
    pub signals_pending: u64,
    pub signals_blocked: u64,
    pub signals_ignored: u64,
    pub signals_caught: u64,
}

/// Fichier ouvert par un process
pub struct OpenFile {
    pub fd: i32,
    pub path: String,
    pub mode: String,                   // r, w, rw
    pub flags: u32,
    pub position: u64,
    pub inode: u64,
    pub file_type: FileType,            // Regular, Directory, Socket, Pipe, etc.
}

/// Socket ouvert par un process
pub struct OpenSocket {
    pub fd: i32,
    pub socket_type: SocketType,        // TCP, UDP, Unix, Netlink, etc.
    pub local_addr: Option<SocketAddr>,
    pub remote_addr: Option<SocketAddr>,
    pub state: TcpState,                // Pour TCP
    pub inode: u64,
}

/// Memory map entry
pub struct MemoryMap {
    pub start_addr: u64,
    pub end_addr: u64,
    pub permissions: String,            // rwxp
    pub offset: u64,
    pub device: String,
    pub inode: u64,
    pub pathname: String,
}

/// Limites du process
pub struct ProcessLimits {
    pub max_cpu_time: (u64, u64),       // (soft, hard)
    pub max_file_size: (u64, u64),
    pub max_data_size: (u64, u64),
    pub max_stack_size: (u64, u64),
    pub max_core_size: (u64, u64),
    pub max_resident_set: (u64, u64),
    pub max_processes: (u64, u64),
    pub max_open_files: (u64, u64),
    pub max_locked_memory: (u64, u64),
    pub max_address_space: (u64, u64),
    pub max_file_locks: (u64, u64),
    pub max_pending_signals: (u64, u64),
    pub max_msgqueue_size: (u64, u64),
    pub max_nice_priority: (u64, u64),
    pub max_realtime_priority: (u64, u64),
    pub max_realtime_timeout: (u64, u64),
}
```

### 1.3 Sources Linux

| Information | Source | Parsing |
|-------------|--------|---------|
| Basic info | `/proc/[pid]/stat` | Space-separated, handle (comm) |
| Status | `/proc/[pid]/status` | Key: Value format |
| Command line | `/proc/[pid]/cmdline` | Null-separated |
| Environment | `/proc/[pid]/environ` | Null-separated key=value |
| I/O stats | `/proc/[pid]/io` | Key: Value format |
| Memory maps | `/proc/[pid]/maps` | Fixed format per line |
| Detailed memory | `/proc/[pid]/smaps` | Multi-line per mapping |
| File descriptors | `/proc/[pid]/fd/*` | Symlinks |
| Limits | `/proc/[pid]/limits` | Table format |
| Cgroups | `/proc/[pid]/cgroup` | hierarchy:controllers:path |
| Namespaces | `/proc/[pid]/ns/*` | Symlinks with inode |
| OOM | `/proc/[pid]/oom_score*` | Single integer |

---

## 2. NETWORK - Connexions Complètes

### 2.1 Ce qui existe ✅

```rust
pub struct NetStats {
    pub interface: String,
    pub rx_bytes: u64,
    pub tx_bytes: u64,
    pub rx_packets: u64,
    pub tx_packets: u64,
    pub rx_errors: u64,
    pub tx_errors: u64,
}
```

### 2.2 Ce qu'il faut ajouter ⏳

```rust
/// === TCP CONNECTIONS ===
pub struct TcpConnection {
    pub local_addr: IpAddr,
    pub local_port: u16,
    pub remote_addr: IpAddr,
    pub remote_port: u16,
    pub state: TcpState,
    pub tx_queue: u32,
    pub rx_queue: u32,
    pub timer_active: u8,
    pub timer_jiffies: u32,
    pub retransmits: u8,
    pub uid: u32,
    pub timeout: u32,
    pub inode: u64,
    // Liaison avec process
    pub pid: Option<i32>,               // Résolu via /proc/[pid]/fd
    pub process_name: Option<String>,
}

#[derive(Debug, Clone, Copy)]
pub enum TcpState {
    Established = 1,
    SynSent = 2,
    SynRecv = 3,
    FinWait1 = 4,
    FinWait2 = 5,
    TimeWait = 6,
    Close = 7,
    CloseWait = 8,
    LastAck = 9,
    Listen = 10,
    Closing = 11,
}

/// Statistiques TCP agrégées
pub struct TcpStats {
    pub connections_established: u32,
    pub connections_syn_sent: u32,
    pub connections_syn_recv: u32,
    pub connections_fin_wait1: u32,
    pub connections_fin_wait2: u32,
    pub connections_time_wait: u32,
    pub connections_close: u32,
    pub connections_close_wait: u32,
    pub connections_last_ack: u32,
    pub connections_listen: u32,
    pub connections_closing: u32,

    // SNMP stats
    pub active_opens: u64,
    pub passive_opens: u64,
    pub attempt_fails: u64,
    pub estab_resets: u64,
    pub curr_estab: u64,
    pub in_segs: u64,
    pub out_segs: u64,
    pub retrans_segs: u64,
    pub in_errs: u64,
    pub out_rsts: u64,
}

/// === UDP CONNECTIONS ===
pub struct UdpConnection {
    pub local_addr: IpAddr,
    pub local_port: u16,
    pub remote_addr: Option<IpAddr>,    // Peut être 0.0.0.0
    pub remote_port: Option<u16>,
    pub tx_queue: u32,
    pub rx_queue: u32,
    pub uid: u32,
    pub inode: u64,
    pub pid: Option<i32>,
    pub process_name: Option<String>,
}

pub struct UdpStats {
    pub in_datagrams: u64,
    pub no_ports: u64,
    pub in_errors: u64,
    pub out_datagrams: u64,
    pub rcvbuf_errors: u64,
    pub sndbuf_errors: u64,
}

/// === UNIX SOCKETS ===
pub struct UnixSocket {
    pub path: String,                   // Peut être vide (abstract)
    pub socket_type: UnixSocketType,    // Stream, Dgram, Seqpacket
    pub state: UnixState,
    pub inode: u64,
    pub pid: Option<i32>,
    pub peer_inode: Option<u64>,        // Socket connecté
}

pub enum UnixSocketType {
    Stream = 1,
    Dgram = 2,
    Seqpacket = 5,
}

/// === RAW SOCKETS ===
pub struct RawSocket {
    pub local_addr: IpAddr,
    pub remote_addr: IpAddr,
    pub protocol: u8,                   // ICMP=1, etc.
    pub uid: u32,
    pub inode: u64,
    pub pid: Option<i32>,
}

/// === NETLINK SOCKETS ===
pub struct NetlinkSocket {
    pub protocol: u32,                  // NETLINK_ROUTE, etc.
    pub port_id: u32,
    pub groups: u32,
    pub uid: u32,
    pub inode: u64,
    pub pid: Option<i32>,
}

/// === INTERFACE COMPLÈTE ===
pub struct NetInterfaceFull {
    pub name: String,
    pub index: u32,
    pub mac_address: String,
    pub mtu: u32,
    pub flags: InterfaceFlags,
    pub state: InterfaceState,          // up/down/unknown
    pub speed_mbps: Option<u32>,        // /sys/class/net/*/speed
    pub duplex: Option<Duplex>,         // /sys/class/net/*/duplex
    pub carrier: bool,                  // /sys/class/net/*/carrier
    pub carrier_changes: u64,           // /sys/class/net/*/carrier_changes

    // Adresses
    pub ipv4_addresses: Vec<Ipv4Info>,
    pub ipv6_addresses: Vec<Ipv6Info>,

    // Statistiques étendues
    pub stats: NetInterfaceStats,

    // Queues
    pub tx_queue_len: u32,
    pub rx_queue_len: u32,

    // Driver info
    pub driver: Option<String>,
    pub driver_version: Option<String>,
    pub firmware_version: Option<String>,
    pub bus_info: Option<String>,
}

pub struct Ipv4Info {
    pub address: Ipv4Addr,
    pub netmask: Ipv4Addr,
    pub broadcast: Option<Ipv4Addr>,
    pub scope: AddressScope,
}

pub struct Ipv6Info {
    pub address: Ipv6Addr,
    pub prefix_len: u8,
    pub scope: AddressScope,
    pub flags: u32,
}

pub struct NetInterfaceStats {
    // Standard
    pub rx_bytes: u64,
    pub tx_bytes: u64,
    pub rx_packets: u64,
    pub tx_packets: u64,
    pub rx_errors: u64,
    pub tx_errors: u64,
    pub rx_dropped: u64,
    pub tx_dropped: u64,

    // Extended
    pub rx_fifo_errors: u64,
    pub tx_fifo_errors: u64,
    pub rx_frame_errors: u64,
    pub rx_compressed: u64,
    pub tx_compressed: u64,
    pub multicast: u64,
    pub collisions: u64,
    pub rx_length_errors: u64,
    pub rx_over_errors: u64,
    pub rx_crc_errors: u64,
    pub rx_missed_errors: u64,
    pub tx_aborted_errors: u64,
    pub tx_carrier_errors: u64,
    pub tx_heartbeat_errors: u64,
    pub tx_window_errors: u64,
}

/// === ROUTING ===
pub struct Route {
    pub destination: IpAddr,
    pub gateway: Option<IpAddr>,
    pub genmask: IpAddr,
    pub flags: RouteFlags,
    pub metric: u32,
    pub interface: String,
    pub mtu: Option<u32>,
    pub window: Option<u32>,
    pub irtt: Option<u32>,
}

/// === ARP TABLE ===
pub struct ArpEntry {
    pub ip_address: Ipv4Addr,
    pub hw_type: u16,
    pub flags: u16,
    pub hw_address: String,             // MAC
    pub device: String,
}

/// === CONNTRACK (Netfilter) ===
pub struct ConntrackEntry {
    pub protocol: String,               // tcp, udp, icmp
    pub src_addr: IpAddr,
    pub dst_addr: IpAddr,
    pub src_port: Option<u16>,
    pub dst_port: Option<u16>,
    pub state: Option<String>,          // Pour TCP
    pub timeout: u32,
    pub packets_orig: u64,
    pub bytes_orig: u64,
    pub packets_reply: u64,
    pub bytes_reply: u64,
    pub mark: u32,
    pub zone: u16,
}

/// === SOCKET BUFFERS ===
pub struct SocketStats {
    pub sockets_used: u32,
    pub tcp_inuse: u32,
    pub tcp_orphan: u32,
    pub tcp_tw: u32,                    // TIME_WAIT
    pub tcp_alloc: u32,
    pub tcp_mem: (u64, u64, u64),       // (low, pressure, high)
    pub udp_inuse: u32,
    pub udp_mem: (u64, u64, u64),
    pub udplite_inuse: u32,
    pub raw_inuse: u32,
    pub frag_inuse: u32,
    pub frag_memory: u64,
}
```

### 2.3 Sources Linux

| Information | Source | Format |
|-------------|--------|--------|
| TCP connections | `/proc/net/tcp`, `/proc/net/tcp6` | Hex addresses |
| UDP connections | `/proc/net/udp`, `/proc/net/udp6` | Hex addresses |
| Unix sockets | `/proc/net/unix` | Table format |
| Raw sockets | `/proc/net/raw`, `/proc/net/raw6` | Hex addresses |
| Netlink | `/proc/net/netlink` | Table format |
| Routes | `/proc/net/route` | Table format |
| ARP | `/proc/net/arp` | Table format |
| Conntrack | `/proc/net/nf_conntrack` | Key=value |
| Socket stats | `/proc/net/sockstat` | Key: values |
| SNMP | `/proc/net/snmp` | Multi-line tables |
| Interface extended | `/sys/class/net/*/statistics/*` | Single values |
| Interface info | Via netlink `RTM_GETLINK` | Binary |
| IP addresses | Via netlink `RTM_GETADDR` | Binary |

### 2.4 Algorithme de Liaison PID ↔ Socket

```rust
/// Résoudre le PID propriétaire d'un socket via son inode
fn resolve_socket_pid(inode: u64) -> Option<(i32, String)> {
    // 1. Lister tous les /proc/[pid]/fd/
    // 2. Pour chaque symlink, vérifier si target = "socket:[{inode}]"
    // 3. Retourner (pid, /proc/[pid]/comm)

    for pid in list_processes()? {
        let fd_dir = format!("/proc/{}/fd", pid);
        for entry in read_dir(fd_dir)? {
            let link = read_link(entry.path())?;
            if link == format!("socket:[{}]", inode) {
                let comm = read_to_string(format!("/proc/{}/comm", pid))?;
                return Some((pid, comm.trim().to_string()));
            }
        }
    }
    None
}
```

---

## 3. GPU - Métriques Complètes

### 3.1 Structure Cible

```rust
/// GPU complet
pub struct GpuMetrics {
    pub index: u32,
    pub uuid: String,
    pub name: String,
    pub vendor: GpuVendor,
    pub pci_bus_id: String,

    // Utilisation
    pub utilization_gpu: f64,           // 0-100%
    pub utilization_memory: f64,        // 0-100%
    pub utilization_encoder: Option<f64>,
    pub utilization_decoder: Option<f64>,

    // Mémoire
    pub memory_total: u64,
    pub memory_used: u64,
    pub memory_free: u64,
    pub memory_reserved: Option<u64>,

    // Température
    pub temperature_gpu: f64,           // Celsius
    pub temperature_memory: Option<f64>,
    pub temperature_slowdown: Option<f64>, // Thermal throttle point
    pub temperature_shutdown: Option<f64>,

    // Power
    pub power_draw: Option<f64>,        // Watts
    pub power_limit: Option<f64>,
    pub power_default_limit: Option<f64>,
    pub power_min_limit: Option<f64>,
    pub power_max_limit: Option<f64>,
    pub power_state: Option<u8>,        // P0-P12

    // Clocks
    pub clock_graphics: u64,            // MHz
    pub clock_sm: Option<u64>,
    pub clock_memory: u64,
    pub clock_video: Option<u64>,
    pub clock_max_graphics: Option<u64>,
    pub clock_max_memory: Option<u64>,

    // PCIe
    pub pcie_link_gen_current: Option<u8>,
    pub pcie_link_gen_max: Option<u8>,
    pub pcie_link_width_current: Option<u8>,
    pub pcie_link_width_max: Option<u8>,
    pub pcie_tx_throughput: Option<u64>, // KB/s
    pub pcie_rx_throughput: Option<u64>,

    // ECC (si disponible)
    pub ecc_enabled: Option<bool>,
    pub ecc_errors_corrected: Option<u64>,
    pub ecc_errors_uncorrected: Option<u64>,

    // Processus sur GPU
    pub processes: Vec<GpuProcess>,

    // Fan (si disponible)
    pub fan_speed: Option<u8>,          // 0-100%

    // Compute
    pub compute_mode: Option<ComputeMode>,
    pub compute_capability: Option<(u8, u8)>, // (major, minor)
}

pub struct GpuProcess {
    pub pid: i32,
    pub process_name: String,
    pub used_memory: u64,
    pub compute_instance_id: Option<u32>,
    pub gpu_instance_id: Option<u32>,
}

pub enum GpuVendor {
    Nvidia,
    Amd,
    Intel,
    Unknown,
}
```

### 3.2 Sources par Vendor

#### NVIDIA (via NVML)

```c
// Fonctions NVML à wrapper
nvmlDeviceGetCount()
nvmlDeviceGetHandleByIndex()
nvmlDeviceGetUUID()
nvmlDeviceGetName()
nvmlDeviceGetPciInfo()
nvmlDeviceGetUtilizationRates()
nvmlDeviceGetMemoryInfo()
nvmlDeviceGetTemperature()
nvmlDeviceGetPowerUsage()
nvmlDeviceGetClockInfo()
nvmlDeviceGetComputeRunningProcesses()
nvmlDeviceGetFanSpeed()
nvmlDeviceGetPowerManagementLimit()
nvmlDeviceGetPcieLink*()
nvmlDeviceGetEccMode()
nvmlDeviceGetMemoryErrorCounter()
```

#### AMD (via sysfs + ROCm-SMI)

```bash
# Paths sysfs pour AMD
/sys/class/drm/card*/device/
├── gpu_busy_percent
├── mem_info_vram_total
├── mem_info_vram_used
├── mem_info_vis_vram_total
├── mem_info_vis_vram_used
├── current_link_speed
├── current_link_width
├── power_dpm_state
├── pp_cur_state
├── hwmon/hwmon*/
│   ├── temp*_input
│   ├── power*_average
│   ├── fan*_input
│   └── freq*_input
└── device              # PCI device ID
```

#### Intel (via sysfs i915)

```bash
# Paths sysfs pour Intel
/sys/class/drm/card*/
├── gt_cur_freq_mhz
├── gt_max_freq_mhz
├── gt_min_freq_mhz
├── gt_boost_freq_mhz
└── gt/
    └── gt0/
        └── rps_cur_freq_mhz
```

---

## 4. POWER - Consommation Énergétique

### 4.1 Structure Cible

```rust
/// Power metrics complet
pub struct PowerMetrics {
    // RAPL (Intel/AMD)
    pub rapl_available: bool,
    pub domains: Vec<RaplDomain>,

    // Battery (laptops)
    pub battery: Option<BatteryInfo>,

    // AC adapter
    pub ac_online: Option<bool>,
}

pub struct RaplDomain {
    pub name: String,                   // package-0, core, uncore, dram
    pub energy_uj: u64,                 // Microjoules counter
    pub max_energy_uj: u64,             // Wrap point
    pub power_watts: f64,               // Calculated
    pub power_limit_watts: Option<f64>,
    pub time_window_us: Option<u64>,
}

pub struct BatteryInfo {
    pub present: bool,
    pub status: BatteryStatus,          // Charging, Discharging, Full, Not charging
    pub capacity_percent: u8,
    pub capacity_level: String,         // Full, Normal, Low, Critical
    pub energy_now_uwh: u64,
    pub energy_full_uwh: u64,
    pub energy_full_design_uwh: u64,
    pub power_now_uw: u64,
    pub voltage_now_uv: u64,
    pub voltage_min_design_uv: u64,
    pub current_now_ua: i64,
    pub charge_now_uah: Option<u64>,
    pub charge_full_uah: Option<u64>,
    pub charge_full_design_uah: Option<u64>,
    pub cycle_count: Option<u32>,
    pub technology: String,             // Li-ion, Li-poly, etc.
    pub manufacturer: String,
    pub model_name: String,
    pub serial_number: String,
    pub time_to_empty_min: Option<u32>,
    pub time_to_full_min: Option<u32>,
}
```

### 4.2 Sources Linux

```bash
# RAPL
/sys/class/powercap/intel-rapl:*/
├── name
├── energy_uj
├── max_energy_range_uj
├── constraint_*_power_limit_uw
└── constraint_*_time_window_us

# Battery
/sys/class/power_supply/BAT*/
├── status
├── present
├── capacity
├── energy_now
├── energy_full
├── energy_full_design
├── power_now
├── voltage_now
├── current_now
├── charge_now
├── cycle_count
├── technology
├── manufacturer
├── model_name
└── serial_number

# AC Adapter
/sys/class/power_supply/AC*/
└── online
```

---

## 5. STORAGE AVANCÉ

### 5.1 SMART Health

```rust
pub struct SmartMetrics {
    pub device: String,
    pub model: String,
    pub serial: String,
    pub firmware: String,
    pub capacity_bytes: u64,
    pub rotation_rate: Option<u16>,     // 0 = SSD
    pub form_factor: String,
    pub protocol: StorageProtocol,      // SATA, NVMe, SAS
    pub smart_enabled: bool,
    pub smart_status: SmartStatus,      // Passed, Failed

    // Attributs critiques
    pub temperature_celsius: i16,
    pub power_on_hours: u64,
    pub power_cycle_count: u64,
    pub reallocated_sector_count: u64,
    pub pending_sector_count: u64,
    pub uncorrectable_sector_count: u64,
    pub wear_leveling_count: Option<u8>, // SSD only, 0-100

    // NVMe specific
    pub nvme_spare_percent: Option<u8>,
    pub nvme_unsafe_shutdowns: Option<u64>,
    pub nvme_media_errors: Option<u64>,

    // All attributes
    pub attributes: Vec<SmartAttribute>,
}

pub struct SmartAttribute {
    pub id: u8,
    pub name: String,
    pub value: u8,
    pub worst: u8,
    pub threshold: u8,
    pub raw_value: u64,
    pub flags: u16,
}
```

### 5.2 ZFS/Btrfs/LVM

```rust
pub struct ZfsPoolMetrics {
    pub name: String,
    pub guid: u64,
    pub health: ZfsHealth,              // ONLINE, DEGRADED, FAULTED
    pub size_bytes: u64,
    pub allocated_bytes: u64,
    pub free_bytes: u64,
    pub fragmentation_percent: f64,
    pub capacity_percent: f64,
    pub dedup_ratio: f64,
    pub compression_ratio: f64,
    pub vdevs: Vec<ZfsVdev>,
    pub datasets: Vec<ZfsDataset>,

    // I/O stats
    pub read_ops: u64,
    pub write_ops: u64,
    pub read_bandwidth: u64,
    pub write_bandwidth: u64,

    // Errors
    pub read_errors: u64,
    pub write_errors: u64,
    pub checksum_errors: u64,
}

pub struct BtrfsMetrics {
    pub uuid: String,
    pub label: Option<String>,
    pub devices: Vec<BtrfsDevice>,
    pub total_bytes: u64,
    pub used_bytes: u64,
    pub data_ratio: f64,
    pub metadata_ratio: f64,
    pub generation: u64,
}

pub struct LvmMetrics {
    pub volume_groups: Vec<VolumeGroup>,
}

pub struct VolumeGroup {
    pub name: String,
    pub uuid: String,
    pub size_bytes: u64,
    pub free_bytes: u64,
    pub extent_size: u64,
    pub extent_count: u64,
    pub free_extent_count: u64,
    pub pv_count: u32,
    pub lv_count: u32,
    pub logical_volumes: Vec<LogicalVolume>,
    pub physical_volumes: Vec<PhysicalVolume>,
}
```

---

## 6. eBPF METRICS (Avancé)

### 6.1 Métriques Kernel-Level

```rust
/// Métriques eBPF collectables
pub struct EbpfMetrics {
    // Syscalls
    pub syscall_counts: HashMap<String, u64>,      // syscall_name -> count
    pub syscall_latencies: HashMap<String, LatencyHistogram>,

    // Scheduling
    pub runqueue_latency: LatencyHistogram,        // Time waiting to run
    pub cpu_off_time: HashMap<i32, u64>,          // pid -> off-cpu time ns

    // I/O
    pub block_io_latency: LatencyHistogram,
    pub block_io_by_process: HashMap<i32, BlockIoStats>,

    // Network
    pub tcp_connect_latency: LatencyHistogram,
    pub tcp_retransmits: HashMap<(IpAddr, IpAddr, u16), u64>,
    pub packet_drops: HashMap<String, u64>,        // interface -> drops

    // Memory
    pub page_faults: HashMap<i32, PageFaultStats>,
    pub mmap_count: HashMap<i32, u64>,

    // Files
    pub file_opens: HashMap<i32, Vec<FileOpenEvent>>,
    pub file_read_latency: LatencyHistogram,
    pub file_write_latency: LatencyHistogram,
}

pub struct LatencyHistogram {
    pub buckets: Vec<(u64, u64)>,       // (upper_bound_ns, count)
    pub sum_ns: u64,
    pub count: u64,
    pub min_ns: u64,
    pub max_ns: u64,
}

pub struct BlockIoStats {
    pub reads: u64,
    pub writes: u64,
    pub read_bytes: u64,
    pub write_bytes: u64,
    pub read_latency_ns: u64,
    pub write_latency_ns: u64,
}

pub struct PageFaultStats {
    pub minor: u64,
    pub major: u64,
}

pub struct FileOpenEvent {
    pub path: String,
    pub flags: u32,
    pub timestamp_ns: u64,
}
```

---

## 7. RÉSUMÉ - Métriques Totales Ciblées

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    MÉTRIQUES TOTALES - SONDE ULTIME                         │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  PROCESS COMPLET                     ~50 champs par process                │
│  ├─ Identification (pid, ppid, pgid, sid, uid, gid, etc.)                  │
│  ├─ CPU détaillé (utime, stime, affinity, scheduler)                       │
│  ├─ Mémoire détaillée (rss, vms, pss, uss, maps)                          │
│  ├─ I/O complet (bytes, syscalls, cancelled)                               │
│  ├─ FDs et sockets (avec résolution inode->pid)                            │
│  ├─ Sécurité (caps, seccomp, no_new_privs)                                │
│  ├─ Namespaces (cgroup, ipc, mnt, net, pid, user, uts)                    │
│  ├─ Limites (rlimits complets)                                             │
│  └─ Environnement et cmdline                                               │
│                                                                             │
│  NETWORK COMPLET                     ~100+ métriques réseau                │
│  ├─ TCP connections avec état et process owner                             │
│  ├─ UDP connections avec process owner                                      │
│  ├─ Unix sockets avec peer resolution                                       │
│  ├─ Raw et Netlink sockets                                                 │
│  ├─ Interfaces (speed, duplex, driver, addresses)                          │
│  ├─ Routes et ARP                                                          │
│  ├─ Conntrack (netfilter)                                                  │
│  ├─ Socket buffers stats                                                   │
│  └─ SNMP complet (TCP, UDP, IP, ICMP)                                      │
│                                                                             │
│  GPU COMPLET                         ~30 métriques par GPU                 │
│  ├─ NVIDIA via NVML                                                        │
│  ├─ AMD via sysfs/ROCm                                                     │
│  ├─ Intel via i915                                                         │
│  └─ Processus par GPU                                                      │
│                                                                             │
│  POWER COMPLET                       ~20 métriques                         │
│  ├─ RAPL (package, core, uncore, dram)                                    │
│  └─ Battery (status, capacity, cycles, time remaining)                     │
│                                                                             │
│  STORAGE AVANCÉ                      ~50 métriques                         │
│  ├─ SMART health (HDD + SSD + NVMe)                                        │
│  ├─ ZFS pools et datasets                                                  │
│  ├─ Btrfs volumes                                                          │
│  └─ LVM (VG, LV, PV)                                                       │
│                                                                             │
│  eBPF (Avancé)                       ~20 types de tracing                  │
│  ├─ Syscall latencies                                                      │
│  ├─ Block I/O latencies                                                    │
│  ├─ Network events                                                          │
│  └─ Scheduling latencies                                                    │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  TOTAL ESTIMÉ: ~300+ métriques distinctes                                  │
│                                                                             │
│  Effort estimé: ~40-50 jours de développement                              │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## 8. PRIORITÉ D'IMPLÉMENTATION

### Phase 2A - Network (5-7 jours)
1. TCP connections + process resolution
2. UDP connections
3. Unix sockets
4. Interface extended stats
5. Socket stats (sockstat)

### Phase 2B - Process (5-7 jours)
1. Process full info (cmdline, environ, cwd)
2. Process I/O (/proc/[pid]/io)
3. Open files et sockets
4. Memory maps (smaps)
5. Namespaces et cgroups

### Phase 2C - GPU (5-7 jours)
1. NVIDIA via NVML
2. AMD via sysfs
3. Intel via i915
4. Process GPU mapping

### Phase 2D - Power & Storage (5-7 jours)
1. RAPL power
2. Battery info
3. SMART health (basic)
4. ZFS/LVM detection

### Phase 3 - eBPF (10+ jours)
1. Infrastructure eBPF
2. Syscall tracing
3. Block I/O latency
4. Network tracing

---

*Document de spécification - Sonde Système Ultime*
