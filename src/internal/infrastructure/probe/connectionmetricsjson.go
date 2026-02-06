//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ConnectionMetricsJSON contains network connection information.
// It includes TCP stats, connections, UDP sockets, and Unix sockets.
type ConnectionMetricsJSON struct {
	TCPStats       *TcpStatsJSON    `json:"tcp_stats,omitempty"`
	TCPConnections []TcpConnJSON    `json:"tcp_connections,omitempty"`
	UDPSockets     []UdpConnJSON    `json:"udp_sockets,omitempty"`
	UnixSockets    []UnixSockJSON   `json:"unix_sockets,omitempty"`
	ListeningPorts []ListenInfoJSON `json:"listening_ports,omitempty"`
}
