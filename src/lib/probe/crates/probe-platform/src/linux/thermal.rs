//! Thermal zone monitoring for Linux via /sys/class/hwmon
//!
//! Reads temperature sensors from hwmon interface.

use crate::{Error, Result, ThermalZone};
use std::fs;
use std::path::Path;

/// Read temperature sensors from /sys/class/hwmon.
///
/// Each hwmon device may have multiple temperature inputs (temp1, temp2, etc.)
/// Path structure:
/// - /sys/class/hwmon/hwmon*/name - Device name
/// - /sys/class/hwmon/hwmon*/temp*_input - Temperature in millidegrees
/// - /sys/class/hwmon/hwmon*/temp*_label - Zone label (optional)
/// - /sys/class/hwmon/hwmon*/temp*_max - Max safe temp (optional)
/// - /sys/class/hwmon/hwmon*/temp*_crit - Critical temp (optional)
pub fn read_thermal_zones() -> Result<Vec<ThermalZone>> {
    let hwmon_path = Path::new("/sys/class/hwmon");
    if !hwmon_path.exists() {
        return Err(Error::NotSupported);
    }

    let mut zones = Vec::new();

    let entries = fs::read_dir(hwmon_path)?;
    for entry in entries.flatten() {
        let hwmon_dir = entry.path();
        if !hwmon_dir.is_dir() {
            continue;
        }

        // Read device name
        let name = fs::read_to_string(hwmon_dir.join("name"))
            .map(|s| s.trim().to_string())
            .unwrap_or_default();

        // Find all temp*_input files
        if let Ok(files) = fs::read_dir(&hwmon_dir) {
            for file in files.flatten() {
                let file_name = file.file_name().to_string_lossy().to_string();
                if file_name.starts_with("temp") && file_name.ends_with("_input") {
                    // Extract sensor number (e.g., "temp1_input" -> "1")
                    let prefix = file_name.trim_end_matches("_input");

                    // Read temperature (in millidegrees Celsius)
                    let temp_millidegrees: i64 = fs::read_to_string(file.path())
                        .ok()
                        .and_then(|s| s.trim().parse().ok())
                        .unwrap_or(0);

                    let temp_celsius = temp_millidegrees as f64 / 1000.0;

                    // Read label (optional)
                    let label = fs::read_to_string(hwmon_dir.join(format!("{}_label", prefix)))
                        .map(|s| s.trim().to_string())
                        .unwrap_or_default();

                    // Read max temperature (optional)
                    let temp_max = fs::read_to_string(hwmon_dir.join(format!("{}_max", prefix)))
                        .ok()
                        .and_then(|s| s.trim().parse::<i64>().ok())
                        .map(|t| t as f64 / 1000.0);

                    // Read critical temperature (optional)
                    let temp_crit = fs::read_to_string(hwmon_dir.join(format!("{}_crit", prefix)))
                        .ok()
                        .and_then(|s| s.trim().parse::<i64>().ok())
                        .map(|t| t as f64 / 1000.0);

                    zones.push(ThermalZone {
                        name: name.clone(),
                        label,
                        temp_celsius,
                        temp_max,
                        temp_crit,
                    });
                }
            }
        }
    }

    Ok(zones)
}

/// Check if thermal monitoring is supported on this system.
pub fn is_thermal_supported() -> bool {
    Path::new("/sys/class/hwmon").exists()
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_is_thermal_supported() {
        // In a container, hwmon may or may not be available
        let supported = is_thermal_supported();
        // Just verify it doesn't panic
        println!("Thermal monitoring supported: {}", supported);
    }

    #[test]
    fn test_read_thermal_zones() {
        let result = read_thermal_zones();
        // May succeed or fail depending on environment
        match result {
            Ok(zones) => {
                println!("Found {} thermal zones", zones.len());
                for zone in &zones {
                    println!(
                        "  {} ({}): {:.1}°C",
                        zone.name, zone.label, zone.temp_celsius
                    );
                }
            }
            Err(e) => println!("Thermal zones not available: {}", e),
        }
    }
}
