# Vers une Sonde Parfaite - Analyse Détaillée

## Vision: La Sonde Système Ultime

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│   ╔═══════════════════════════════════════════════════════════════════╗    │
│   ║                    SONDE PARFAITE                                  ║    │
│   ║                                                                    ║    │
│   ║   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐ ║    │
│   ║   │   CPU   │  │ Memory  │  │  Disk   │  │ Network │  │   GPU   │ ║    │
│   ║   │ ✅ 100% │  │ ✅ 100% │  │ 🔶 80%  │  │ 🔶 70%  │  │ ⏳ 0%   │ ║    │
│   ║   └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘ ║    │
│   ║                                                                    ║    │
│   ║   ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐ ║    │
│   ║   │ Process │  │ Thermal │  │  Power  │  │  Cloud  │  │  eBPF   │ ║    │
│   ║   │ ✅ 90%  │  │ ✅ 100% │  │ ⏳ 0%   │  │ ⏳ 0%   │  │ ⏳ 0%   │ ║    │
│   ║   └─────────┘  └─────────┘  └─────────┘  └─────────┘  └─────────┘ ║    │
│   ║                                                                    ║    │
│   ╚═══════════════════════════════════════════════════════════════════╝    │
│                                                                             │
│   Caractéristiques Cibles:                                                  │
│   • Latence < 100µs pour collect_all()                                     │
│   • Support 5+ plateformes (Linux, macOS, FreeBSD, OpenBSD, NetBSD)        │
│   • Détection automatique de 20+ runtimes                                   │
│   • Export natif OpenTelemetry                                              │
│   • Zero allocations en mode hot-path                                       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Analyse par Catégorie

### A. Métriques CPU - État Complet

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           CPU METRICS                                       │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  IMPLÉMENTÉ ✅                                                       │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • User/System/Idle %         /proc/stat                            │   │
│  │  • I/O Wait %                 /proc/stat                            │   │
│  │  • Steal % (VM)               /proc/stat                            │   │
│  │  • Core count                 /proc/cpuinfo                         │   │
│  │  • Frequency MHz              /proc/cpuinfo                         │   │
│  │  • PSI Pressure               /proc/pressure/cpu                    │   │
│  │  • Context Switches           /proc/stat, /proc/[pid]/status        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  MANQUANT ⏳                                                         │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • Per-core utilization       /proc/stat (cpuN lines)               │   │
│  │  • CPU times (tick→ms)        Conversion hz_to_ms                   │   │
│  │  • Interrupts count           /proc/interrupts                      │   │
│  │  • Softirqs count             /proc/softirqs                        │   │
│  │  • CPU topology (NUMA)        /sys/devices/system/cpu/cpu*/topology │   │
│  │  • C-states residency         /sys/devices/system/cpu/cpu*/cpuidle  │   │
│  │  • Turbo boost status         /sys/devices/system/cpu/intel_pstate  │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  COMPLÉTUDE: ████████░░ 80%                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### B. Métriques Mémoire - État Complet

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          MEMORY METRICS                                     │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  IMPLÉMENTÉ ✅                                                       │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • Total/Used/Available       /proc/meminfo                         │   │
│  │  • Cached/Buffers             /proc/meminfo                         │   │
│  │  • Swap Total/Used/Free       /proc/meminfo                         │   │
│  │  • PSI Pressure (some/full)   /proc/pressure/memory                 │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  MANQUANT ⏳                                                         │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • Slab memory                /proc/meminfo (Slab)                  │   │
│  │  • Dirty pages                /proc/meminfo (Dirty)                 │   │
│  │  • Writeback pages            /proc/meminfo (Writeback)             │   │
│  │  • Mapped memory              /proc/meminfo (Mapped)                │   │
│  │  • Shared memory              /proc/meminfo (Shmem)                 │   │
│  │  • Page tables                /proc/meminfo (PageTables)            │   │
│  │  • Huge pages                 /proc/meminfo (Hugepages*)            │   │
│  │  • NUMA stats                 /sys/devices/system/node/node*/       │   │
│  │  • OOM score                  /proc/[pid]/oom_score                 │   │
│  │  • Memory cgroups             /sys/fs/cgroup/memory/                │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  COMPLÉTUDE: ██████░░░░ 60%                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### C. Métriques Disque - État Complet

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           DISK METRICS                                      │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  IMPLÉMENTÉ ✅                                                       │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • Partitions list            /proc/mounts                          │   │
│  │  • Space usage (statvfs)      statvfs() syscall                     │   │
│  │  • Inode usage                statvfs() syscall                     │   │
│  │  • I/O stats (reads/writes)   /proc/diskstats                       │   │
│  │  • I/O time                   /proc/diskstats                       │   │
│  │  • I/O pressure (PSI)         /proc/pressure/io                     │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  MANQUANT ⏳                                                         │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • SMART health               libatasmart / smartctl                │   │
│  │  • NVMe health                /sys/block/nvme*/device/              │   │
│  │  • I/O latency histograms     blktrace / eBPF                       │   │
│  │  • Queue depth                /sys/block/*/queue/                   │   │
│  │  • Scheduler stats            /sys/block/*/queue/scheduler          │   │
│  │  • ZFS pool stats             /proc/spl/kstat/zfs/                  │   │
│  │  • Btrfs stats                /sys/fs/btrfs/                        │   │
│  │  • LVM stats                  /sys/block/dm-*/                      │   │
│  │  • RAID stats                 /proc/mdstat                          │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  COMPLÉTUDE: ██████░░░░ 55%                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### D. Métriques Réseau - État Complet

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          NETWORK METRICS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  IMPLÉMENTÉ ✅                                                       │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • Interface list             /sys/class/net/                       │   │
│  │  • MAC address                /sys/class/net/*/address              │   │
│  │  • MTU                        /sys/class/net/*/mtu                  │   │
│  │  • RX/TX bytes/packets        /proc/net/dev                         │   │
│  │  • RX/TX errors/drops         /proc/net/dev                         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  MANQUANT ⏳                                                         │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • TCP connections state      /proc/net/tcp, /proc/net/tcp6         │   │
│  │  • UDP stats                  /proc/net/udp, /proc/net/udp6         │   │
│  │  • Socket buffers             /proc/net/sockstat                    │   │
│  │  • TCP retransmits            /proc/net/snmp (Tcp)                  │   │
│  │  • IP addresses               netlink / ioctl                       │   │
│  │  • Routing table              /proc/net/route                       │   │
│  │  • ARP table                  /proc/net/arp                         │   │
│  │  • Netfilter stats            /proc/net/nf_conntrack                │   │
│  │  • Interface speed            /sys/class/net/*/speed                │   │
│  │  • Duplex mode                /sys/class/net/*/duplex               │   │
│  │  • ICMP stats                 /proc/net/snmp (Icmp)                 │   │
│  │  • Unix sockets               /proc/net/unix                        │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  COMPLÉTUDE: ████░░░░░░ 40%                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### E. Métriques Process - État Complet

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          PROCESS METRICS                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  IMPLÉMENTÉ ✅                                                       │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • PID list                   /proc/                                │   │
│  │  • State (R/S/D/Z/T)          /proc/[pid]/stat                      │   │
│  │  • Thread count               /proc/[pid]/stat                      │   │
│  │  • Memory RSS/VMS             /proc/[pid]/status                    │   │
│  │  • FD count                   /proc/[pid]/fd/                       │   │
│  │  • Context switches           /proc/[pid]/status                    │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  ┌─────────────────────────────────────────────────────────────────────┐   │
│  │  MANQUANT ⏳                                                         │   │
│  ├─────────────────────────────────────────────────────────────────────┤   │
│  │  • CPU time (utime/stime)     /proc/[pid]/stat (avec calcul %)      │   │
│  │  • Memory maps                /proc/[pid]/maps, /proc/[pid]/smaps   │   │
│  │  • I/O stats                  /proc/[pid]/io                        │   │
│  │  • Limits                     /proc/[pid]/limits                    │   │
│  │  • Environment                /proc/[pid]/environ                   │   │
│  │  • Command line               /proc/[pid]/cmdline                   │   │
│  │  • Working directory          /proc/[pid]/cwd                       │   │
│  │  • Start time                 /proc/[pid]/stat (starttime)          │   │
│  │  • Parent PID                 /proc/[pid]/stat (ppid)               │   │
│  │  • User/Group                 /proc/[pid]/status (Uid/Gid)          │   │
│  │  • Cgroup membership          /proc/[pid]/cgroup                    │   │
│  │  • Namespaces                 /proc/[pid]/ns/                       │   │
│  │  • Capabilities               /proc/[pid]/status (Cap*)             │   │
│  │  • Seccomp status             /proc/[pid]/status (Seccomp)          │   │
│  │  • Network connections        /proc/[pid]/fd/ + /proc/net/*         │   │
│  └─────────────────────────────────────────────────────────────────────┘   │
│                                                                             │
│  COMPLÉTUDE: ████░░░░░░ 35%                                                │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Catégories Non Implémentées

### F. GPU Metrics (Phase 2)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                           GPU METRICS ⏳                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   NVIDIA (via NVML)                    AMD (via sysfs/ROCm)                │
│   ┌─────────────────────┐             ┌─────────────────────┐              │
│   │ • Utilization %     │             │ • gpu_busy_percent  │              │
│   │ • Memory used/total │             │ • mem_info_vram_*   │              │
│   │ • Temperature       │             │ • temp*_input       │              │
│   │ • Power draw        │             │ • power*_average    │              │
│   │ • Clock speeds      │             │ • freq*_input       │              │
│   │ • PCIe bandwidth    │             │ • PCIe stats        │              │
│   │ • ECC errors        │             │ • ras_*_count       │              │
│   │ • Processes list    │             │ • fdinfo/*          │              │
│   └─────────────────────┘             └─────────────────────┘              │
│                                                                             │
│   Intel (via sysfs/i915)                                                   │
│   ┌─────────────────────┐                                                  │
│   │ • gt_cur_freq_mhz   │                                                  │
│   │ • rc6_residency_ms  │                                                  │
│   │ • energy counters   │                                                  │
│   └─────────────────────┘                                                  │
│                                                                             │
│  COMPLÉTUDE: ░░░░░░░░░░ 0%                                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### G. Power Metrics (Phase 2)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          POWER METRICS ⏳                                   │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   RAPL (Intel/AMD)                                                         │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │  /sys/class/powercap/intel-rapl:*/                                  │  │
│   │                                                                     │  │
│   │  ├─ intel-rapl:0/              Package (CPU + uncore)               │  │
│   │  │   ├─ energy_uj              Énergie en microjoules               │  │
│   │  │   ├─ max_energy_range_uj    Wrap counter value                   │  │
│   │  │   └─ name                   "package-0"                          │  │
│   │  │                                                                  │  │
│   │  ├─ intel-rapl:0:0/            CPU cores only                       │  │
│   │  │   └─ energy_uj                                                   │  │
│   │  │                                                                  │  │
│   │  └─ intel-rapl:0:1/            DRAM (si disponible)                 │  │
│   │      └─ energy_uj                                                   │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   Calcul de puissance:                                                     │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                                                                     │  │
│   │   power_watts = (energy_uj_2 - energy_uj_1) / (time_us * 1_000_000)│  │
│   │                                                                     │  │
│   │   Attention au wrap: si energy_uj_2 < energy_uj_1:                  │  │
│   │     energy_uj_2 += max_energy_range_uj                              │  │
│   │                                                                     │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  COMPLÉTUDE: ░░░░░░░░░░ 0%                                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

### H. Advanced eBPF Metrics (Phase 5)

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                          eBPF METRICS ⏳                                    │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │                    eBPF PROGRAMS                                    │  │
│   │                                                                     │  │
│   │   ┌─────────────┐    ┌─────────────┐    ┌─────────────┐            │  │
│   │   │  Tracepoint │    │    kprobe   │    │     XDP     │            │  │
│   │   │  (syscalls) │    │  (functions)│    │  (network)  │            │  │
│   │   └──────┬──────┘    └──────┬──────┘    └──────┬──────┘            │  │
│   │          │                  │                  │                    │  │
│   │          └──────────────────┼──────────────────┘                    │  │
│   │                             │                                       │  │
│   │                             ▼                                       │  │
│   │   ┌─────────────────────────────────────────────────────────────┐  │  │
│   │   │                   BPF MAPS                                  │  │  │
│   │   │  • Histogram (latences I/O)                                 │  │  │
│   │   │  • Hash map (syscalls par process)                          │  │  │
│   │   │  • Ring buffer (events temps réel)                          │  │  │
│   │   └─────────────────────────────────────────────────────────────┘  │  │
│   │                             │                                       │  │
│   │                             ▼                                       │  │
│   │   ┌─────────────────────────────────────────────────────────────┐  │  │
│   │   │              MÉTRIQUES COLLECTABLES                         │  │  │
│   │   │                                                             │  │  │
│   │   │  • Syscall latency (read, write, open, close, etc.)         │  │  │
│   │   │  • Block I/O latency histograms                             │  │  │
│   │   │  • TCP connection events (connect, accept, close)           │  │  │
│   │   │  • TCP retransmission tracking                              │  │  │
│   │   │  • Process execution events (exec, fork, exit)              │  │  │
│   │   │  • Memory allocation tracking                               │  │  │
│   │   │  • File system events (open, read, write, unlink)           │  │  │
│   │   │  • Scheduling latency (runqueue wait time)                  │  │  │
│   │   │  • Page faults (minor, major)                               │  │  │
│   │   │  • Lock contention analysis                                 │  │  │
│   │   │                                                             │  │  │
│   │   └─────────────────────────────────────────────────────────────┘  │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│  COMPLÉTUDE: ░░░░░░░░░░ 0%                                                 │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Comparaison avec Outils Existants

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    COMPARAISON FONCTIONNALITÉS                              │
├──────────────────┬─────────┬─────────┬─────────┬─────────┬─────────────────┤
│ Fonctionnalité   │  probe  │ gopsutil│  node   │  cadvisor│  prometheus    │
│                  │  (nous) │  (Go)   │ exporter│   (Go)  │ node_exporter  │
├──────────────────┼─────────┼─────────┼─────────┼─────────┼─────────────────┤
│ CPU basic        │   ✅    │   ✅    │   ✅    │   ✅    │      ✅         │
│ CPU per-core     │   ⏳    │   ✅    │   ✅    │   ✅    │      ✅         │
│ Memory basic     │   ✅    │   ✅    │   ✅    │   ✅    │      ✅         │
│ Memory detailed  │   🔶    │   ✅    │   ✅    │   ✅    │      ✅         │
│ Disk usage       │   ✅    │   ✅    │   ✅    │   ✅    │      ✅         │
│ Disk I/O         │   ✅    │   ✅    │   ✅    │   ✅    │      ✅         │
│ Network basic    │   ✅    │   ✅    │   ✅    │   ✅    │      ✅         │
│ TCP connections  │   ⏳    │   ✅    │   ✅    │   ❌    │      ✅         │
│ Process list     │   ✅    │   ✅    │   ⏳    │   ✅    │      ✅         │
│ Process detailed │   🔶    │   ✅    │   ⏳    │   ✅    │      🔶         │
│ PSI (pressure)   │   ✅    │   ❌    │   ✅    │   ❌    │      ✅         │
│ Thermal          │   ✅    │   ✅    │   ✅    │   ❌    │      ✅         │
│ GPU              │   ⏳    │   ⏳    │   ✅    │   ✅    │      🔶         │
│ Power (RAPL)     │   ⏳    │   ❌    │   ✅    │   ❌    │      ✅         │
│ Container detect │   ✅    │   🔶    │   ❌    │   ✅    │      ❌         │
│ Quotas (cgroups) │   ✅    │   🔶    │   ❌    │   ✅    │      ✅         │
│ eBPF             │   ⏳    │   ❌    │   ❌    │   ❌    │      🔶         │
├──────────────────┼─────────┼─────────┼─────────┼─────────┼─────────────────┤
│ Cross-platform   │   ✅    │   ✅    │  Linux  │  Linux  │    Multi       │
│ Performance      │  ~25µs  │ ~100µs  │  ~50µs  │ ~200µs  │    ~100µs      │
│ Memory footprint │  ~2MB   │  ~10MB  │  ~15MB  │  ~50MB  │    ~20MB       │
│ CGO dependency   │   ✅    │   ❌    │   ❌    │   ❌    │      ❌        │
├──────────────────┴─────────┴─────────┴─────────┴─────────┴─────────────────┤
│                                                                             │
│  AVANTAGES probe:                         INCONVÉNIENTS probe:             │
│  • Performance (Rust + cache)             • Nécessite CGO                  │
│  • Détection runtime complète             • Plus complexe à builder        │
│  • PSI natif                              • Moins de métriques actuellement│
│  • Architecture modulaire                 • Pas encore d'export OTLP       │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Recommandations Prioritaires

### Pour atteindre la "Sonde Parfaite"

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                                                                             │
│   PRIORITÉ 1: FONCTIONNALITÉS CRITIQUES (Impact Business)                  │
│   ─────────────────────────────────────────────────────────                │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 1. TCP Connection Stats      │ Indispensable pour monitoring réseau │  │
│   │    /proc/net/tcp parsing     │ ~2 jours d'effort                    │  │
│   ├──────────────────────────────┼──────────────────────────────────────┤  │
│   │ 2. Per-process CPU %         │ Calcul réel via utime/stime diff     │  │
│   │    Tracking temporel         │ ~1 jour d'effort                     │  │
│   ├──────────────────────────────┼──────────────────────────────────────┤  │
│   │ 3. Process I/O stats         │ /proc/[pid]/io parsing               │  │
│   │    read_bytes, write_bytes   │ ~0.5 jour d'effort                   │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   PRIORITÉ 2: PERFORMANCE (Différentiation)                                │
│   ──────────────────────────────────────────                               │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 4. Rayon parallelization     │ process_collect_all 60% plus rapide  │  │
│   │    Parallel iterator         │ ~1 jour d'effort                     │  │
│   ├──────────────────────────────┼──────────────────────────────────────┤  │
│   │ 5. io_uring batch            │ Batch reads pour /proc (Linux 5.1+)  │  │
│   │    Async I/O                 │ ~3 jours d'effort                    │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   PRIORITÉ 3: CLOUD NATIVE (Adoption)                                      │
│   ────────────────────────────────────                                     │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 6. Cloud metadata            │ AWS/GCP/Azure instance info          │  │
│   │    HTTP metadata endpoints   │ ~2 jours d'effort                    │  │
│   ├──────────────────────────────┼──────────────────────────────────────┤  │
│   │ 7. OpenTelemetry export      │ OTLP push/pull natif                 │  │
│   │    Prometheus compatible     │ ~3 jours d'effort                    │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
│   PRIORITÉ 4: AVANCÉ (Différentiation Technique)                           │
│   ─────────────────────────────────────────────                            │
│                                                                             │
│   ┌─────────────────────────────────────────────────────────────────────┐  │
│   │ 8. GPU metrics               │ NVIDIA/AMD monitoring                │  │
│   │    NVML + sysfs              │ ~4 jours d'effort                    │  │
│   ├──────────────────────────────┼──────────────────────────────────────┤  │
│   │ 9. eBPF integration          │ Kernel-level tracing                 │  │
│   │    libbpf-rs                 │ ~5 jours d'effort                    │  │
│   └─────────────────────────────────────────────────────────────────────┘  │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

## Estimation Effort Total

```
┌─────────────────────────────────────────────────────────────────────────────┐
│                    EFFORT POUR SONDE PARFAITE                               │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   Phase 1 (Complète)           5 jours   ████████████████████ FAIT         │
│                                                                             │
│   Phase 2 (Métriques)         10 jours   ░░░░░░░░░░░░░░░░░░░░ À faire      │
│   ├─ GPU Metrics               4j                                          │
│   ├─ TCP Stats                 2j                                          │
│   ├─ RAPL Power                2j                                          │
│   └─ SMART/ZFS                 2j                                          │
│                                                                             │
│   Phase 3 (Cloud)              5 jours   ░░░░░░░░░░░░░░░░░░░░ À faire      │
│   ├─ AWS/GCP/Azure             3j                                          │
│   ├─ Serverless                1j                                          │
│   └─ Kata/gVisor               1j                                          │
│                                                                             │
│   Phase 4 (Performance)        7 jours   ░░░░░░░░░░░░░░░░░░░░ À faire      │
│   ├─ io_uring                  3j                                          │
│   ├─ Rayon                     1j                                          │
│   ├─ mmap                      1j                                          │
│   └─ Buffer pools              2j                                          │
│                                                                             │
│   Phase 5 (Avancé)            10 jours   ░░░░░░░░░░░░░░░░░░░░ À faire      │
│   ├─ eBPF                      5j                                          │
│   ├─ perf_event                2j                                          │
│   ├─ OpenTelemetry             3j                                          │
│   └─ Anomaly                   (bonus)                                     │
│                                                                             │
├─────────────────────────────────────────────────────────────────────────────┤
│                                                                             │
│   TOTAL ESTIMÉ: ~37 jours de travail                                       │
│                                                                             │
│   Progression actuelle: 5/37 jours = 13.5%                                 │
│                                                                             │
│   Pour une sonde "parfaite" à 90%+ complétude:                             │
│   → Phases 2-3 suffisent (~15 jours additionnels)                          │
│   → Avec Phase 4: Performance optimale (~22 jours)                         │
│   → Avec Phase 5: État de l'art (~32 jours additionnels)                   │
│                                                                             │
└─────────────────────────────────────────────────────────────────────────────┘
```

---

*Document généré automatiquement - 2026-01-30*
