//! macOS/Darwin-specific utilities.

use std::path::Path;

/// Check if Docker Desktop is installed.
pub fn has_docker_desktop() -> bool {
    Path::new("/Applications/Docker.app").exists()
}

/// Check if Colima is installed.
pub fn has_colima() -> bool {
    if let Ok(home) = std::env::var("HOME") {
        return Path::new(&format!("{}/.colima", home)).exists();
    }
    false
}

/// Check if Podman Desktop is installed.
pub fn has_podman_desktop() -> bool {
    Path::new("/Applications/Podman Desktop.app").exists()
}

/// Get the Docker socket path for macOS.
pub fn get_docker_socket() -> Option<String> {
    let home = std::env::var("HOME").ok()?;

    // Check various locations
    let paths = [
        format!("{}/.docker/run/docker.sock", home),
        format!(
            "{}/Library/Containers/com.docker.docker/Data/docker.sock",
            home
        ),
        format!("{}/.colima/default/docker.sock", home),
    ];

    for path in paths {
        if Path::new(&path).exists() {
            return Some(path);
        }
    }

    // Check standard Unix path as fallback
    if Path::new("/var/run/docker.sock").exists() {
        return Some("/var/run/docker.sock".to_string());
    }

    None
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_has_docker_desktop() {
        // Just verify it doesn't panic
        let _ = has_docker_desktop();
    }

    #[test]
    fn test_get_docker_socket() {
        let _ = get_docker_socket();
    }
}
