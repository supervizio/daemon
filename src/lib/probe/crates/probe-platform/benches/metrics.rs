//! Benchmark suite for probe metrics collection.
//!
//! Run with: `cargo bench -p probe-platform`

use criterion::{black_box, criterion_group, criterion_main, Criterion, Throughput};
use probe_platform::{SystemCollector, new_collector};

/// Benchmark system CPU collection.
fn bench_cpu_collect_system(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("cpu_collect_system", |b| {
        b.iter(|| {
            black_box(collector.cpu().collect_system()).ok()
        })
    });
}

/// Benchmark CPU pressure collection (Linux only).
fn bench_cpu_collect_pressure(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("cpu_collect_pressure", |b| {
        b.iter(|| {
            black_box(collector.cpu().collect_pressure()).ok()
        })
    });
}

/// Benchmark system memory collection.
fn bench_memory_collect_system(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("memory_collect_system", |b| {
        b.iter(|| {
            black_box(collector.memory().collect_system()).ok()
        })
    });
}

/// Benchmark memory pressure collection (Linux only).
fn bench_memory_collect_pressure(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("memory_collect_pressure", |b| {
        b.iter(|| {
            black_box(collector.memory().collect_pressure()).ok()
        })
    });
}

/// Benchmark load average collection.
fn bench_load_collect(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("load_collect", |b| {
        b.iter(|| {
            black_box(collector.load().collect()).ok()
        })
    });
}

/// Benchmark single process metrics collection.
fn bench_process_collect_single(c: &mut Criterion) {
    let collector = new_collector();
    let pid = std::process::id() as i32;

    c.bench_function("process_collect_single", |b| {
        b.iter(|| {
            black_box(collector.process().collect(pid)).ok()
        })
    });
}

/// Benchmark all processes enumeration and collection.
fn bench_process_collect_all(c: &mut Criterion) {
    let collector = new_collector();

    let mut group = c.benchmark_group("process_collect_all");
    group.throughput(Throughput::Elements(1));
    group.bench_function("full", |b| {
        b.iter(|| {
            black_box(collector.process().collect_all()).ok()
        })
    });
    group.finish();
}

/// Benchmark disk partition listing.
fn bench_disk_list_partitions(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("disk_list_partitions", |b| {
        b.iter(|| {
            black_box(collector.disk().list_partitions()).ok()
        })
    });
}

/// Benchmark disk usage collection for root.
fn bench_disk_collect_usage(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("disk_collect_usage", |b| {
        b.iter(|| {
            black_box(collector.disk().collect_usage("/")).ok()
        })
    });
}

/// Benchmark disk I/O statistics collection.
fn bench_disk_collect_io(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("disk_collect_io", |b| {
        b.iter(|| {
            black_box(collector.disk().collect_io()).ok()
        })
    });
}

/// Benchmark network interface listing.
fn bench_network_list_interfaces(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("network_list_interfaces", |b| {
        b.iter(|| {
            black_box(collector.network().list_interfaces()).ok()
        })
    });
}

/// Benchmark network statistics collection for all interfaces.
fn bench_network_collect_all_stats(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("network_collect_all_stats", |b| {
        b.iter(|| {
            black_box(collector.network().collect_all_stats()).ok()
        })
    });
}

/// Benchmark system I/O statistics collection.
fn bench_io_collect_stats(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("io_collect_stats", |b| {
        b.iter(|| {
            black_box(collector.io().collect_stats()).ok()
        })
    });
}

/// Benchmark I/O pressure collection (Linux only).
fn bench_io_collect_pressure(c: &mut Criterion) {
    let collector = new_collector();

    c.bench_function("io_collect_pressure", |b| {
        b.iter(|| {
            black_box(collector.io().collect_pressure()).ok()
        })
    });
}

/// Benchmark aggregated collect_all (all metrics at once).
fn bench_collect_all(c: &mut Criterion) {
    let collector = new_collector();

    let mut group = c.benchmark_group("collect_all");
    group.throughput(Throughput::Elements(1));
    group.bench_function("full", |b| {
        b.iter(|| {
            black_box(collector.collect_all()).ok()
        })
    });
    group.finish();
}

/// Benchmark thermal zone collection (Linux only).
#[cfg(target_os = "linux")]
fn bench_thermal_collect(c: &mut Criterion) {
    use probe_platform::linux::read_thermal_zones;

    c.bench_function("thermal_collect", |b| {
        b.iter(|| {
            black_box(read_thermal_zones()).ok()
        })
    });
}

/// Benchmark context switch reading.
#[cfg(target_os = "linux")]
fn bench_context_switches(c: &mut Criterion) {
    use probe_platform::linux::{read_self_context_switches, read_system_context_switches};

    c.bench_function("context_switches_system", |b| {
        b.iter(|| {
            black_box(read_system_context_switches()).ok()
        })
    });

    c.bench_function("context_switches_self", |b| {
        b.iter(|| {
            black_box(read_self_context_switches()).ok()
        })
    });
}

// Group all basic benchmarks
criterion_group!(
    basic_benches,
    bench_cpu_collect_system,
    bench_cpu_collect_pressure,
    bench_memory_collect_system,
    bench_memory_collect_pressure,
    bench_load_collect,
);

// Group disk and network benchmarks
criterion_group!(
    io_benches,
    bench_disk_list_partitions,
    bench_disk_collect_usage,
    bench_disk_collect_io,
    bench_network_list_interfaces,
    bench_network_collect_all_stats,
    bench_io_collect_stats,
    bench_io_collect_pressure,
);

// Group process benchmarks (potentially slower)
criterion_group!(
    process_benches,
    bench_process_collect_single,
    bench_process_collect_all,
);

// Group aggregated benchmarks
criterion_group!(
    aggregate_benches,
    bench_collect_all,
);

// Linux-specific benchmarks
#[cfg(target_os = "linux")]
criterion_group!(
    linux_benches,
    bench_thermal_collect,
    bench_context_switches,
);

#[cfg(target_os = "linux")]
criterion_main!(basic_benches, io_benches, process_benches, aggregate_benches, linux_benches);

#[cfg(not(target_os = "linux"))]
criterion_main!(basic_benches, io_benches, process_benches, aggregate_benches);
