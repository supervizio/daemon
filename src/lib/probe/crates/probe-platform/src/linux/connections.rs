//! Network connection parsing for Linux.
//!
//! Parses /proc/net/tcp, /proc/net/tcp6, /proc/net/udp, /proc/net/udp6
//! and resolves process ownership via /proc/[pid]/fd.

use crate::{
    AddressFamily, Error, Result, SocketState, TcpConnection, TcpStats, UdpConnection, UnixSocket,
};
use std::collections::HashMap;
use std::fs;
use std::path::Path;

/// Parse an IPv4 address from hex format (little-endian).
fn parse_ipv4_addr(hex: &str) -> String {
    if hex.len() != 8 {
        return "0.0.0.0".to_string();
    }
    let bytes: Vec<u8> =
        (0..4).filter_map(|i| u8::from_str_radix(&hex[i * 2..i * 2 + 2], 16).ok()).collect();
    if bytes.len() != 4 {
        return "0.0.0.0".to_string();
    }
    // Linux stores in little-endian, reverse for display
    format!("{}.{}.{}.{}", bytes[3], bytes[2], bytes[1], bytes[0])
}

/// Parse an IPv6 address from hex format.
fn parse_ipv6_addr(hex: &str) -> String {
    if hex.len() != 32 {
        return "::".to_string();
    }
    // IPv6 is stored as 4 32-bit words in little-endian
    let mut parts = Vec::new();
    for i in 0..4 {
        let word_start = i * 8;
        let word = &hex[word_start..word_start + 8];
        // Each 32-bit word is little-endian, convert to big-endian for display
        let b0 = &word[6..8];
        let b1 = &word[4..6];
        let b2 = &word[2..4];
        let b3 = &word[0..2];
        parts.push(format!("{}{}", b0, b1));
        parts.push(format!("{}{}", b2, b3));
    }
    // Simplify notation
    let full = parts.join(":");
    // Basic compression (could be improved)
    full.to_lowercase()
}

/// Parse address:port from hex format.
fn parse_addr_port(addr_port: &str, ipv6: bool) -> (String, u16) {
    let parts: Vec<&str> = addr_port.split(':').collect();
    if parts.len() != 2 {
        return (String::new(), 0);
    }
    let addr = if ipv6 { parse_ipv6_addr(parts[0]) } else { parse_ipv4_addr(parts[0]) };
    let port = u16::from_str_radix(parts[1], 16).unwrap_or(0);
    (addr, port)
}

/// Build a map of socket inode -> (pid, process_name) for all processes.
pub fn build_socket_pid_map() -> HashMap<u64, (i32, String)> {
    let mut map = HashMap::new();

    let proc_path = Path::new("/proc");
    let entries = match fs::read_dir(proc_path) {
        Ok(e) => e,
        Err(_) => return map,
    };

    for entry in entries.flatten() {
        let name = entry.file_name();
        let name_str = name.to_string_lossy();

        // Only process numeric directories (PIDs)
        let pid: i32 = match name_str.parse() {
            Ok(p) => p,
            Err(_) => continue,
        };

        // Read process name
        let comm_path = proc_path.join(&name).join("comm");
        let process_name = fs::read_to_string(&comm_path).unwrap_or_default().trim().to_string();

        // Scan fd directory for socket links
        let fd_path = proc_path.join(&name).join("fd");
        let fd_entries = match fs::read_dir(&fd_path) {
            Ok(e) => e,
            Err(_) => continue,
        };

        for fd_entry in fd_entries.flatten() {
            let link = match fs::read_link(fd_entry.path()) {
                Ok(l) => l,
                Err(_) => continue,
            };

            let link_str = link.to_string_lossy();
            // Socket links look like: socket:[12345]
            if let Some(inode_str) =
                link_str.strip_prefix("socket:[").and_then(|s| s.strip_suffix(']'))
                && let Ok(inode) = inode_str.parse::<u64>()
            {
                map.insert(inode, (pid, process_name.clone()));
            }
        }
    }

    map
}

/// Parse /proc/net/tcp or /proc/net/tcp6 file.
fn parse_tcp_file(
    path: &str,
    ipv6: bool,
    socket_map: &HashMap<u64, (i32, String)>,
) -> Result<Vec<TcpConnection>> {
    let content = fs::read_to_string(path)?;
    let mut connections = Vec::new();

    for line in content.lines().skip(1) {
        // Skip header
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 10 {
            continue;
        }

        // Format: sl local_address rem_address st tx_queue:rx_queue tr:tm->when retrnsmt uid timeout inode
        let (local_addr, local_port) = parse_addr_port(parts[1], ipv6);
        let (remote_addr, remote_port) = parse_addr_port(parts[2], ipv6);

        let state_hex = u8::from_str_radix(parts[3], 16).unwrap_or(0);
        let state = SocketState::from_linux_state(state_hex);

        // Parse tx_queue:rx_queue
        let queue_parts: Vec<&str> = parts[4].split(':').collect();
        let tx_queue =
            queue_parts.first().and_then(|s| u32::from_str_radix(s, 16).ok()).unwrap_or(0);
        let rx_queue =
            queue_parts.get(1).and_then(|s| u32::from_str_radix(s, 16).ok()).unwrap_or(0);

        let inode = parts.get(9).and_then(|s| s.parse::<u64>().ok()).unwrap_or(0);

        let (pid, process_name) = socket_map.get(&inode).cloned().unwrap_or((-1, String::new()));

        connections.push(TcpConnection {
            family: if ipv6 { AddressFamily::IPv6 } else { AddressFamily::IPv4 },
            local_addr,
            local_port,
            remote_addr,
            remote_port,
            state,
            pid,
            process_name,
            inode,
            rx_queue,
            tx_queue,
        });
    }

    Ok(connections)
}

/// Parse /proc/net/udp or /proc/net/udp6 file.
fn parse_udp_file(
    path: &str,
    ipv6: bool,
    socket_map: &HashMap<u64, (i32, String)>,
) -> Result<Vec<UdpConnection>> {
    let content = fs::read_to_string(path)?;
    let mut connections = Vec::new();

    for line in content.lines().skip(1) {
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 10 {
            continue;
        }

        let (local_addr, local_port) = parse_addr_port(parts[1], ipv6);
        let (remote_addr, remote_port) = parse_addr_port(parts[2], ipv6);

        let state_hex = u8::from_str_radix(parts[3], 16).unwrap_or(0);
        let state = SocketState::from_linux_state(state_hex);

        let queue_parts: Vec<&str> = parts[4].split(':').collect();
        let tx_queue =
            queue_parts.first().and_then(|s| u32::from_str_radix(s, 16).ok()).unwrap_or(0);
        let rx_queue =
            queue_parts.get(1).and_then(|s| u32::from_str_radix(s, 16).ok()).unwrap_or(0);

        let inode = parts.get(9).and_then(|s| s.parse::<u64>().ok()).unwrap_or(0);

        let (pid, process_name) = socket_map.get(&inode).cloned().unwrap_or((-1, String::new()));

        connections.push(UdpConnection {
            family: if ipv6 { AddressFamily::IPv6 } else { AddressFamily::IPv4 },
            local_addr,
            local_port,
            remote_addr,
            remote_port,
            state,
            pid,
            process_name,
            inode,
            rx_queue,
            tx_queue,
        });
    }

    Ok(connections)
}

/// Parse /proc/net/unix file.
fn parse_unix_file(socket_map: &HashMap<u64, (i32, String)>) -> Result<Vec<UnixSocket>> {
    let content = fs::read_to_string("/proc/net/unix")?;
    let mut sockets = Vec::new();

    for line in content.lines().skip(1) {
        let parts: Vec<&str> = line.split_whitespace().collect();
        if parts.len() < 7 {
            continue;
        }

        // Format: Num RefCount Protocol Flags Type St Inode Path
        let socket_type = match parts[4] {
            "0001" => "stream",
            "0002" => "dgram",
            "0005" => "seqpacket",
            _ => "unknown",
        }
        .to_string();

        let state_num: u8 = parts[5].parse().unwrap_or(0);
        let state = match state_num {
            1 => SocketState::Listen,
            3 => SocketState::Established,
            _ => SocketState::Unknown,
        };

        let inode = parts[6].parse::<u64>().unwrap_or(0);

        let path = parts.get(7).map(|s| s.to_string()).unwrap_or_default();

        let (pid, process_name) = socket_map.get(&inode).cloned().unwrap_or((-1, String::new()));

        sockets.push(UnixSocket { path, socket_type, state, pid, process_name, inode });
    }

    Ok(sockets)
}

/// Collect all TCP connections (IPv4 and IPv6).
pub fn collect_tcp_connections() -> Result<Vec<TcpConnection>> {
    let socket_map = build_socket_pid_map();
    let mut connections = Vec::new();

    // IPv4
    if let Ok(mut tcp4) = parse_tcp_file("/proc/net/tcp", false, &socket_map) {
        connections.append(&mut tcp4);
    }

    // IPv6
    if let Ok(mut tcp6) = parse_tcp_file("/proc/net/tcp6", true, &socket_map) {
        connections.append(&mut tcp6);
    }

    Ok(connections)
}

/// Collect all UDP sockets (IPv4 and IPv6).
pub fn collect_udp_connections() -> Result<Vec<UdpConnection>> {
    let socket_map = build_socket_pid_map();
    let mut connections = Vec::new();

    // IPv4
    if let Ok(mut udp4) = parse_udp_file("/proc/net/udp", false, &socket_map) {
        connections.append(&mut udp4);
    }

    // IPv6
    if let Ok(mut udp6) = parse_udp_file("/proc/net/udp6", true, &socket_map) {
        connections.append(&mut udp6);
    }

    Ok(connections)
}

/// Collect all Unix domain sockets.
pub fn collect_unix_sockets() -> Result<Vec<UnixSocket>> {
    let socket_map = build_socket_pid_map();
    parse_unix_file(&socket_map)
}

/// Calculate TCP connection statistics.
pub fn collect_tcp_stats() -> Result<TcpStats> {
    let connections = collect_tcp_connections()?;
    let mut stats = TcpStats::default();

    for conn in connections {
        match conn.state {
            SocketState::Established => stats.established += 1,
            SocketState::SynSent => stats.syn_sent += 1,
            SocketState::SynRecv => stats.syn_recv += 1,
            SocketState::FinWait1 => stats.fin_wait1 += 1,
            SocketState::FinWait2 => stats.fin_wait2 += 1,
            SocketState::TimeWait => stats.time_wait += 1,
            SocketState::Close => stats.close += 1,
            SocketState::CloseWait => stats.close_wait += 1,
            SocketState::LastAck => stats.last_ack += 1,
            SocketState::Listen => stats.listen += 1,
            SocketState::Closing => stats.closing += 1,
            SocketState::Unknown => {}
        }
    }

    Ok(stats)
}

/// Collect connections for a specific process.
pub fn collect_process_connections(pid: i32) -> Result<(Vec<TcpConnection>, Vec<UdpConnection>)> {
    // Build socket map for just this process
    let mut socket_map = HashMap::new();

    let proc_path = Path::new("/proc").join(pid.to_string());
    if !proc_path.exists() {
        return Err(Error::NotFound(format!("process {} not found", pid)));
    }

    let comm_path = proc_path.join("comm");
    let process_name = fs::read_to_string(&comm_path).unwrap_or_default().trim().to_string();

    let fd_path = proc_path.join("fd");
    if let Ok(entries) = fs::read_dir(&fd_path) {
        for entry in entries.flatten() {
            if let Ok(link) = fs::read_link(entry.path()) {
                let link_str = link.to_string_lossy();
                if let Some(inode_str) =
                    link_str.strip_prefix("socket:[").and_then(|s| s.strip_suffix(']'))
                    && let Ok(inode) = inode_str.parse::<u64>()
                {
                    socket_map.insert(inode, (pid, process_name.clone()));
                }
            }
        }
    }

    // Parse TCP connections and filter by this process's sockets
    let mut tcp_conns = Vec::new();
    if let Ok(tcp4) = parse_tcp_file("/proc/net/tcp", false, &socket_map) {
        tcp_conns.extend(tcp4.into_iter().filter(|c| c.pid == pid));
    }
    if let Ok(tcp6) = parse_tcp_file("/proc/net/tcp6", true, &socket_map) {
        tcp_conns.extend(tcp6.into_iter().filter(|c| c.pid == pid));
    }

    // Parse UDP connections and filter
    let mut udp_conns = Vec::new();
    if let Ok(udp4) = parse_udp_file("/proc/net/udp", false, &socket_map) {
        udp_conns.extend(udp4.into_iter().filter(|c| c.pid == pid));
    }
    if let Ok(udp6) = parse_udp_file("/proc/net/udp6", true, &socket_map) {
        udp_conns.extend(udp6.into_iter().filter(|c| c.pid == pid));
    }

    Ok((tcp_conns, udp_conns))
}

/// Find which process owns a specific port.
pub fn find_process_by_port(port: u16, tcp: bool) -> Result<Option<i32>> {
    if tcp {
        let connections = collect_tcp_connections()?;
        for conn in connections {
            if conn.local_port == port && conn.pid > 0 {
                return Ok(Some(conn.pid));
            }
        }
    } else {
        let connections = collect_udp_connections()?;
        for conn in connections {
            if conn.local_port == port && conn.pid > 0 {
                return Ok(Some(conn.pid));
            }
        }
    }
    Ok(None)
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_ipv4_addr() {
        // 127.0.0.1 in little-endian hex
        assert_eq!(parse_ipv4_addr("0100007F"), "127.0.0.1");
        // 0.0.0.0
        assert_eq!(parse_ipv4_addr("00000000"), "0.0.0.0");
    }

    #[test]
    fn test_socket_state_from_linux() {
        assert_eq!(SocketState::from_linux_state(1), SocketState::Established);
        assert_eq!(SocketState::from_linux_state(10), SocketState::Listen);
        assert_eq!(SocketState::from_linux_state(99), SocketState::Unknown);
    }

    #[test]
    fn test_collect_tcp_connections() {
        // This test requires /proc/net/tcp to exist
        let result = collect_tcp_connections();
        // Should at least not error on Linux
        assert!(result.is_ok());
    }

    #[test]
    fn test_collect_tcp_stats() {
        let result = collect_tcp_stats();
        assert!(result.is_ok());
    }
}
