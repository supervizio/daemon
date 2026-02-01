//! Thermal zone monitoring for BSD systems.
//!
//! FreeBSD: Reads ACPI thermal zones via sysctl (hw.acpi.thermal.tz*)
//! OpenBSD/NetBSD: Not supported (returns empty results)
//!
//! # FreeBSD ACPI Thermal Zones
//!
//! FreeBSD exposes ACPI thermal zones through sysctl:
//! - `hw.acpi.thermal.tz0.temperature` - Current temperature in deciKelvin (tenths of Kelvin)
//! - `hw.acpi.thermal.tz0._CRT` - Critical temperature in deciKelvin
//! - `hw.acpi.thermal.tz0._HOT` - Hot temperature in deciKelvin
//! - `hw.acpi.thermal.tz1.*` - Additional thermal zones
//!
//! Temperature conversion: Celsius = (deciKelvin / 10.0) - 273.15

use crate::{Error, Result, ThermalZone};
use std::ffi::CString;
use std::mem;
use std::process::Command;

/// Maximum number of thermal zones to probe.
const MAX_THERMAL_ZONES: usize = 16;

/// Read all thermal zones from the system.
///
/// # Platform Support
///
/// - **FreeBSD**: Reads ACPI thermal zones via sysctl
/// - **OpenBSD**: Not supported, returns empty vector
/// - **NetBSD**: Not supported, returns empty vector
///
/// # Errors
///
/// Returns [`Error::NotSupported`] if thermal monitoring is not available.
///
/// # Examples
///
/// ```no_run
/// use probe_platform::bsd::thermal::read_thermal_zones;
///
/// match read_thermal_zones() {
///     Ok(zones) => {
///         for zone in zones {
///             println!("{}: {:.1}°C", zone.name, zone.temp_celsius);
///         }
///     }
///     Err(e) => eprintln!("Thermal monitoring not available: {}", e),
/// }
/// ```
pub fn read_thermal_zones() -> Result<Vec<ThermalZone>> {
    #[cfg(target_os = "freebsd")]
    {
        read_thermal_zones_freebsd()
    }

    #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
    {
        // OpenBSD and NetBSD don't have standardized thermal zone interfaces
        Ok(Vec::new())
    }
}

/// Check if thermal monitoring is supported on this system.
///
/// # Examples
///
/// ```no_run
/// use probe_platform::bsd::thermal::is_thermal_supported;
///
/// if is_thermal_supported() {
///     println!("Thermal monitoring available");
/// } else {
///     println!("Thermal monitoring not supported");
/// }
/// ```
#[must_use]
pub fn is_thermal_supported() -> bool {
    #[cfg(target_os = "freebsd")]
    {
        is_thermal_supported_freebsd()
    }

    #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
    {
        false
    }
}

// ============================================================================
// FreeBSD Implementation
// ============================================================================

#[cfg(target_os = "freebsd")]
fn read_thermal_zones_freebsd() -> Result<Vec<ThermalZone>> {
    let mut zones = Vec::new();

    for zone_idx in 0..MAX_THERMAL_ZONES {
        match read_thermal_zone_freebsd(zone_idx) {
            Ok(zone) => zones.push(zone),
            Err(Error::NotFound(_)) => {
                // No more zones, stop probing
                break;
            }
            Err(_) => {
                // Zone exists but read failed, continue to next zone
                continue;
            }
        }
    }

    if zones.is_empty() {
        return Err(Error::NotSupported);
    }

    Ok(zones)
}

#[cfg(target_os = "freebsd")]
fn read_thermal_zone_freebsd(zone_idx: usize) -> Result<ThermalZone> {
    let name = format!("tz{zone_idx}");

    // Read current temperature
    let temp_sysctl = format!("hw.acpi.thermal.{name}.temperature");
    let temp_deci_kelvin = read_sysctl_i32(&temp_sysctl)
        .map_err(|_| Error::NotFound(format!("thermal zone {name} not found")))?;

    let temp_celsius = deci_kelvin_to_celsius(temp_deci_kelvin);

    // Read critical temperature (optional)
    let crit_sysctl = format!("hw.acpi.thermal.{name}._CRT");
    let temp_crit = read_sysctl_i32(&crit_sysctl)
        .ok()
        .map(deci_kelvin_to_celsius);

    // Read hot temperature (optional, used as max if available)
    let hot_sysctl = format!("hw.acpi.thermal.{name}._HOT");
    let temp_max = read_sysctl_i32(&hot_sysctl)
        .ok()
        .map(deci_kelvin_to_celsius);

    Ok(ThermalZone {
        name: "acpi".to_string(),
        label: name,
        temp_celsius,
        temp_max,
        temp_crit,
    })
}

#[cfg(target_os = "freebsd")]
fn is_thermal_supported_freebsd() -> bool {
    // Check if the first thermal zone exists
    Command::new("sysctl")
        .args(["-n", "hw.acpi.thermal.tz0.temperature"])
        .output()
        .map_or(false, |output| output.status.success())
}

#[cfg(target_os = "freebsd")]
fn read_sysctl_i32(name: &str) -> Result<i32> {
    unsafe {
        let c_name =
            CString::new(name).map_err(|_| Error::Platform("invalid sysctl name".to_string()))?;
        let mut value: i32 = 0;
        let mut len = mem::size_of::<i32>();

        let result = libc::sysctlbyname(
            c_name.as_ptr(),
            &mut value as *mut _ as *mut libc::c_void,
            &mut len,
            std::ptr::null_mut(),
            0,
        );

        if result != 0 {
            return Err(Error::NotFound(format!("sysctl {name} not found")));
        }

        Ok(value)
    }
}

// ============================================================================
// Helper Functions
// ============================================================================

/// Convert deciKelvin to Celsius.
///
/// FreeBSD ACPI thermal zones report temperatures in deciKelvin (tenths of Kelvin).
///
/// # Examples
///
/// ```
/// # use probe_platform::bsd::thermal::deci_kelvin_to_celsius;
/// assert_eq!(deci_kelvin_to_celsius(2981), 25.0); // 298.1 K = 25°C
/// assert_eq!(deci_kelvin_to_celsius(3231), 50.0); // 323.1 K = 50°C
/// ```
#[must_use]
pub fn deci_kelvin_to_celsius(deci_kelvin: i32) -> f64 {
    const ABSOLUTE_ZERO_CELSIUS: f64 = 273.15;
    const DECI_KELVIN_SCALE: f64 = 10.0;

    (f64::from(deci_kelvin) / DECI_KELVIN_SCALE) - ABSOLUTE_ZERO_CELSIUS
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_deci_kelvin_to_celsius_conversion() {
        // 0°C = 273.15 K = 2731.5 deciKelvin
        assert!((deci_kelvin_to_celsius(2731) - 0.0).abs() < 0.2);

        // 25°C = 298.15 K = 2981.5 deciKelvin
        assert!((deci_kelvin_to_celsius(2981) - 25.0).abs() < 0.2);

        // 50°C = 323.15 K = 3231.5 deciKelvin
        assert!((deci_kelvin_to_celsius(3231) - 50.0).abs() < 0.2);

        // 100°C = 373.15 K = 3731.5 deciKelvin
        assert!((deci_kelvin_to_celsius(3731) - 100.0).abs() < 0.2);

        // -40°C = 233.15 K = 2331.5 deciKelvin
        assert!((deci_kelvin_to_celsius(2331) - (-40.0)).abs() < 0.2);
    }

    #[test]
    fn test_is_thermal_supported() {
        // This test will pass on all platforms
        // On FreeBSD, it may return true or false depending on ACPI support
        // On OpenBSD/NetBSD, it should always return false
        let supported = is_thermal_supported();

        #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
        assert!(
            !supported,
            "OpenBSD/NetBSD should not support thermal monitoring"
        );

        #[cfg(target_os = "freebsd")]
        {
            // On FreeBSD, just verify it doesn't panic
            println!("FreeBSD thermal support: {supported}");
        }
    }

    #[test]
    fn test_read_thermal_zones() {
        // This test will behave differently based on the platform
        let result = read_thermal_zones();

        #[cfg(any(target_os = "openbsd", target_os = "netbsd"))]
        {
            // OpenBSD/NetBSD should return empty vector
            assert!(result.is_ok());
            assert!(result.unwrap().is_empty());
        }

        #[cfg(target_os = "freebsd")]
        {
            // On FreeBSD, may succeed or fail depending on ACPI support
            match result {
                Ok(zones) => {
                    println!("Found {} thermal zones", zones.len());
                    for zone in &zones {
                        println!(
                            "  {} ({}): {:.1}°C (max: {:?}, crit: {:?})",
                            zone.name, zone.label, zone.temp_celsius, zone.temp_max, zone.temp_crit
                        );
                        // Sanity check: temperature should be within reasonable range
                        assert!(
                            zone.temp_celsius > -50.0 && zone.temp_celsius < 150.0,
                            "Temperature out of reasonable range: {}°C",
                            zone.temp_celsius
                        );
                    }
                }
                Err(Error::NotSupported) => {
                    println!("ACPI thermal zones not available on this FreeBSD system");
                }
                Err(e) => {
                    println!("Error reading thermal zones: {e}");
                }
            }
        }
    }

    #[cfg(target_os = "freebsd")]
    #[test]
    fn test_read_sysctl_i32_invalid() {
        // Try to read a non-existent sysctl
        let result = read_sysctl_i32("hw.acpi.thermal.nonexistent.temperature");
        assert!(result.is_err());
    }

    #[test]
    fn test_max_thermal_zones_constant() {
        // Ensure the constant is reasonable
        assert!(
            MAX_THERMAL_ZONES > 0 && MAX_THERMAL_ZONES <= 256,
            "MAX_THERMAL_ZONES should be between 1 and 256"
        );
    }
}
