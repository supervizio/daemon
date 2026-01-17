# Linux Metrics - Parsing /proc

Collecte métriques via le pseudo-filesystem `/proc`.

## Fichiers Parsés

| Fichier | Données |
|---------|---------|
| `/proc/stat` | CPU temps global (jiffies) |
| `/proc/meminfo` | Mémoire système (kB → bytes) |
| `/proc/[pid]/stat` | CPU par processus |
| `/proc/[pid]/status` | Mémoire par processus |

## Structure

| Fichier | Rôle |
|---------|------|
| `collector.go` | `Collector` combiné CPU+Memory |
| `cpu.go` | `CPUCollector` - parsing `/proc/stat` |
| `memory.go` | `MemoryCollector` - parsing `/proc/meminfo` |

## CPU (/proc/stat)

```
cpu  user nice system idle iowait irq softirq steal guest guest_nice
```

Colonnes 1-10 mappées vers `CPUMetrics`.

## Memory (/proc/meminfo)

```
MemTotal:       16384000 kB
MemFree:         1234567 kB
MemAvailable:    8765432 kB
...
```

Valeurs en kB, converties en bytes (*1024).

## Constructeurs

```go
NewCollector() *Collector           // Combiné
NewCPUCollector() *CPUCollector     // CPU seul
NewMemoryCollector() *MemoryCollector  // Memory seul

// Pour tests avec mock /proc
NewCPUCollectorWithPath(path string) *CPUCollector
```

## Build Tag

```go
//go:build linux
```
