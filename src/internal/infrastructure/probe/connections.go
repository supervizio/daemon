//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

import (
	"context"
	"unsafe"
)

// notFoundPID represents the PID value returned when no process is found.
const notFoundPID int32 = -1

// establishedConnMultiplier is used to calculate 70% of connections (7/10).
const establishedConnMultiplier int = 7

// String constants for address families (interned to avoid allocations).
const (
	strIPv4    string = "ipv4"
	strIPv6    string = "ipv6"
	strUnknown string = "unknown"
)

// String constants for socket states (interned to avoid allocations).
const (
	strStateUnknown     string = "UNKNOWN"
	strStateEstablished string = "ESTABLISHED"
	strStateSynSent     string = "SYN_SENT"
	strStateSynRecv     string = "SYN_RECV"
	strStateFinWait1    string = "FIN_WAIT1"
	strStateFinWait2    string = "FIN_WAIT2"
	strStateTimeWait    string = "TIME_WAIT"
	strStateClose       string = "CLOSE"
	strStateCloseWait   string = "CLOSE_WAIT"
	strStateLastAck     string = "LAST_ACK"
	strStateListen      string = "LISTEN"
	strStateClosing     string = "CLOSING"
)

// SocketState represents the state of a network socket.
type SocketState uint8

// Socket state constants representing TCP connection states.
const (
	// SocketStateUnknown indicates an unknown socket state.
	SocketStateUnknown SocketState = 0
	// SocketStateEstablished indicates an established connection.
	SocketStateEstablished SocketState = 1
	// SocketStateSynSent indicates SYN packet sent, awaiting response.
	SocketStateSynSent SocketState = 2
	// SocketStateSynRecv indicates SYN received, awaiting ACK.
	SocketStateSynRecv SocketState = 3
	// SocketStateFinWait1 indicates FIN sent, awaiting ACK.
	SocketStateFinWait1 SocketState = 4
	// SocketStateFinWait2 indicates FIN acknowledged, awaiting remote FIN.
	SocketStateFinWait2 SocketState = 5
	// SocketStateTimeWait indicates waiting for enough time to ensure remote received ACK.
	SocketStateTimeWait SocketState = 6
	// SocketStateClose indicates socket is closed.
	SocketStateClose SocketState = 7
	// SocketStateCloseWait indicates remote has closed, waiting for local close.
	SocketStateCloseWait SocketState = 8
	// SocketStateLastAck indicates waiting for final ACK after sending FIN.
	SocketStateLastAck SocketState = 9
	// SocketStateListen indicates socket is listening for connections.
	SocketStateListen SocketState = 10
	// SocketStateClosing indicates both sides sent FIN simultaneously.
	SocketStateClosing SocketState = 11
)

// AddressFamily represents the address family of a network connection.
type AddressFamily uint8

// Address family constants.
const (
	// AddressFamilyIPv4 indicates IPv4 address family.
	AddressFamilyIPv4 AddressFamily = 4
	// AddressFamilyIPv6 indicates IPv6 address family.
	AddressFamilyIPv6 AddressFamily = 6
)

// socketStateNames maps socket states to their string representations.
var socketStateNames map[SocketState]string = map[SocketState]string{
	SocketStateUnknown:     strStateUnknown,
	SocketStateEstablished: strStateEstablished,
	SocketStateSynSent:     strStateSynSent,
	SocketStateSynRecv:     strStateSynRecv,
	SocketStateFinWait1:    strStateFinWait1,
	SocketStateFinWait2:    strStateFinWait2,
	SocketStateTimeWait:    strStateTimeWait,
	SocketStateClose:       strStateClose,
	SocketStateCloseWait:   strStateCloseWait,
	SocketStateLastAck:     strStateLastAck,
	SocketStateListen:      strStateListen,
	SocketStateClosing:     strStateClosing,
}

// String returns the string representation of the socket state.
//
// Returns:
//   - string: human-readable name of the socket state
func (s SocketState) String() string {
	// look up state name in the map
	if name, ok := socketStateNames[s]; ok {
		// return the mapped name
		return name
	}
	// return UNKNOWN for any unrecognized value
	return strStateUnknown
}

// String returns the string representation of the address family.
//
// Returns:
//   - string: human-readable name of the address family
func (f AddressFamily) String() string {
	// check for IPv4 family
	if f == AddressFamilyIPv4 {
		// return IPv4 string
		return strIPv4
	}
	// check for IPv6 family
	if f == AddressFamilyIPv6 {
		// return IPv6 string
		return strIPv6
	}
	// return Unknown for any unrecognized value
	return strUnknown
}

// ConnectionCollector provides network connection metrics via the Rust probe library.
// It collects TCP, UDP, and Unix socket information.
type ConnectionCollector struct{}

// NewConnectionCollector creates a new connection collector.
//
// Returns:
//   - *ConnectionCollector: new collector instance
func NewConnectionCollector() *ConnectionCollector {
	// return a new empty collector
	return &ConnectionCollector{}
}

// CollectTCP returns all TCP connections with process information.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - []TcpConnection: list of TCP connections
//   - error: nil on success, error if probe not initialized or collection fails
func (c *ConnectionCollector) CollectTCP(ctx context.Context) ([]TcpConnection, error) {
	// validate context and initialization state
	if err := validateCollectionContext(ctx); err != nil {
		// return nil on validation failure
		return nil, err
	}
	// collect TCP connections from C library
	var list C.TcpConnectionList
	result := C.probe_collect_tcp_connections(&list)
	// check if collection failed
	if err := resultToError(result); err != nil {
		// return nil on collection failure
		return nil, err
	}
	defer C.probe_free_tcp_connection_list(&list)
	// convert C list to Go slice and return
	return convertTCPConnections(&list), nil
}

// convertTCPConnections converts a C TCP connection list to Go slice.
//
// Params:
//   - list: pointer to the C TCP connection list
//
// Returns:
//   - []TcpConnection: the converted Go slice
func convertTCPConnections(list *C.TcpConnectionList) []TcpConnection {
	// allocate slice with capacity matching list count
	connections := make([]TcpConnection, 0, list.count)
	items := unsafe.Slice(list.items, list.count)
	// iterate over each connection item
	for _, item := range items {
		connections = append(connections, TcpConnection{
			Family:      AddressFamily(item.family),
			LocalAddr:   cCharArrayToString(item.local_addr[:]),
			LocalPort:   uint16(item.local_port),
			RemoteAddr:  cCharArrayToString(item.remote_addr[:]),
			RemotePort:  uint16(item.remote_port),
			State:       SocketState(item.state),
			PID:         int32(item.pid),
			ProcessName: cCharArrayToString(item.process_name[:]),
			Inode:       uint64(item.inode),
			RxQueue:     uint32(item.rx_queue),
			TxQueue:     uint32(item.tx_queue),
		})
	}
	// return the converted connections
	return connections
}

// CollectUDP returns all UDP sockets with process information.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - []UdpConnection: list of UDP connections
//   - error: nil on success, error if probe not initialized or collection fails
func (c *ConnectionCollector) CollectUDP(ctx context.Context) ([]UdpConnection, error) {
	// validate context and initialization state
	if err := validateCollectionContext(ctx); err != nil {
		// return nil on validation failure
		return nil, err
	}
	// collect UDP connections from C library
	var list C.UdpConnectionList
	result := C.probe_collect_udp_connections(&list)
	// check if collection failed
	if err := resultToError(result); err != nil {
		// return nil on collection failure
		return nil, err
	}
	defer C.probe_free_udp_connection_list(&list)
	// convert C list to Go slice and return
	return convertUDPConnections(&list), nil
}

// convertUDPConnections converts a C UDP connection list to Go slice.
//
// Params:
//   - list: pointer to the C UDP connection list
//
// Returns:
//   - []UdpConnection: the converted Go slice
func convertUDPConnections(list *C.UdpConnectionList) []UdpConnection {
	// allocate slice with capacity matching list count
	connections := make([]UdpConnection, 0, list.count)
	items := unsafe.Slice(list.items, list.count)
	// iterate over each connection item
	for _, item := range items {
		connections = append(connections, UdpConnection{
			Family:      AddressFamily(item.family),
			LocalAddr:   cCharArrayToString(item.local_addr[:]),
			LocalPort:   uint16(item.local_port),
			RemoteAddr:  cCharArrayToString(item.remote_addr[:]),
			RemotePort:  uint16(item.remote_port),
			State:       SocketState(item.state),
			PID:         int32(item.pid),
			ProcessName: cCharArrayToString(item.process_name[:]),
			Inode:       uint64(item.inode),
			RxQueue:     uint32(item.rx_queue),
			TxQueue:     uint32(item.tx_queue),
		})
	}
	// return the converted connections
	return connections
}

// CollectUnix returns all Unix domain sockets with process information.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - []UnixSocket: list of Unix sockets
//   - error: nil on success, error if probe not initialized or collection fails
func (c *ConnectionCollector) CollectUnix(ctx context.Context) ([]UnixSocket, error) {
	// check if context has been cancelled before expensive FFI call
	if err := checkContext(ctx); err != nil {
		// return nil slice with context error
		return nil, err
	}
	// check if probe is initialized
	if err := checkInitialized(); err != nil {
		// return early if not initialized
		return nil, err
	}

	var list C.UnixSocketList
	result := C.probe_collect_unix_sockets(&list)
	// check collection result
	if err := resultToError(result); err != nil {
		// return early on collection failure
		return nil, err
	}
	defer C.probe_free_unix_socket_list(&list)

	// convert C list to Go slice with capacity
	sockets := make([]UnixSocket, 0, list.count)
	items := unsafe.Slice(list.items, list.count)
	// iterate over each socket item
	for _, item := range items {
		sockets = append(sockets, UnixSocket{
			Path:        cCharArrayToString(item.path[:]),
			SocketType:  cCharArrayToString(item.socket_type[:]),
			State:       SocketState(item.state),
			PID:         int32(item.pid),
			ProcessName: cCharArrayToString(item.process_name[:]),
			Inode:       uint64(item.inode),
		})
	}

	// return the collected sockets
	return sockets, nil
}

// CollectTCPStats returns aggregated TCP connection statistics.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//
// Returns:
//   - *TcpStats: aggregated TCP statistics
//   - error: nil on success, error if probe not initialized or collection fails
func (c *ConnectionCollector) CollectTCPStats(ctx context.Context) (*TcpStats, error) {
	// check if context has been cancelled before expensive FFI call
	if err := checkContext(ctx); err != nil {
		// return nil with context error
		return nil, err
	}
	// check if probe is initialized
	if err := checkInitialized(); err != nil {
		// return early if not initialized
		return nil, err
	}

	var stats C.TcpStats
	result := C.probe_collect_tcp_stats(&stats)
	// check collection result
	if err := resultToError(result); err != nil {
		// return early on collection failure
		return nil, err
	}

	// return the collected stats
	return &TcpStats{
		Established: uint32(stats.established),
		SynSent:     uint32(stats.syn_sent),
		SynRecv:     uint32(stats.syn_recv),
		FinWait1:    uint32(stats.fin_wait1),
		FinWait2:    uint32(stats.fin_wait2),
		TimeWait:    uint32(stats.time_wait),
		Close:       uint32(stats.close),
		CloseWait:   uint32(stats.close_wait),
		LastAck:     uint32(stats.last_ack),
		Listen:      uint32(stats.listen),
		Closing:     uint32(stats.closing),
	}, nil
}

// FindProcessByPort finds which process owns a specific port.
// Returns -1 if no process is found.
//
// Params:
//   - ctx: context for cancellation (unused, reserved for future use)
//   - port: port number to look up
//   - tcp: true for TCP, false for UDP
//
// Returns:
//   - int32: process ID owning the port, or -1 if not found
//   - error: nil on success, error if probe not initialized or lookup fails
func (c *ConnectionCollector) FindProcessByPort(ctx context.Context, port uint16, tcp bool) (int32, error) {
	// check if context has been cancelled before expensive FFI call
	if err := checkContext(ctx); err != nil {
		// return not found with context error
		return notFoundPID, err
	}
	// check if probe is initialized
	if err := checkInitialized(); err != nil {
		// return not found with error if not initialized
		return notFoundPID, err
	}

	var pid C.int32_t
	result := C.probe_find_process_by_port(C.uint16_t(port), C.bool(tcp), &pid)
	// check lookup result
	if err := resultToError(result); err != nil {
		// return not found with error on lookup failure
		return notFoundPID, err
	}

	// return the found process ID
	return int32(pid), nil
}

// CollectListeningPorts returns all ports that are in LISTEN state.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []TcpConnection: list of listening TCP connections
//   - error: nil on success, error if collection fails
func (c *ConnectionCollector) CollectListeningPorts(ctx context.Context) ([]TcpConnection, error) {
	// collect all TCP connections first
	connections, err := c.CollectTCP(ctx)
	// check if collection failed
	if err != nil {
		// return early on collection failure
		return nil, err
	}

	// filter connections in LISTEN state
	// capacity heuristic: listening ports are typically 1-5% of total connections
	estimatedCap := len(connections) / listeningPortPercentage
	// ensure minimum capacity for small connection counts
	if estimatedCap < minListeningCapacity {
		estimatedCap = minListeningCapacity
	}
	listening := make([]TcpConnection, 0, estimatedCap)
	// iterate over each connection
	for _, conn := range connections {
		// check if connection is in LISTEN state
		if conn.State == SocketStateListen {
			listening = append(listening, conn)
		}
	}

	// return the filtered connections
	return listening, nil
}

// CollectEstablishedConnections returns all TCP connections in ESTABLISHED state.
//
// Params:
//   - ctx: context for cancellation
//
// Returns:
//   - []TcpConnection: list of established TCP connections
//   - error: nil on success, error if collection fails
func (c *ConnectionCollector) CollectEstablishedConnections(ctx context.Context) ([]TcpConnection, error) {
	// collect all TCP connections first
	connections, err := c.CollectTCP(ctx)
	// check if collection failed
	if err != nil {
		// return early on collection failure
		return nil, err
	}

	// filter connections in ESTABLISHED state
	// capacity heuristic: established connections are typically 70% of total
	estimatedCap := (len(connections) * establishedConnMultiplier) / establishedConnPercentage
	// ensure minimum capacity for small connection counts
	if estimatedCap < minEstablishedCapacity {
		estimatedCap = minEstablishedCapacity
	}
	established := make([]TcpConnection, 0, estimatedCap)
	// iterate over each connection
	for _, conn := range connections {
		// check if connection is in ESTABLISHED state
		if conn.State == SocketStateEstablished {
			established = append(established, conn)
		}
	}

	// return the filtered connections
	return established, nil
}

// CollectProcessConnections returns all TCP and UDP connections for a specific process.
//
// Params:
//   - ctx: context for cancellation
//   - pid: process ID to filter connections for
//
// Returns:
//   - []TcpConnection: list of TCP connections for the process
//   - []UdpConnection: list of UDP connections for the process
//   - error: nil on success, error if collection fails
func (c *ConnectionCollector) CollectProcessConnections(ctx context.Context, pid int32) ([]TcpConnection, []UdpConnection, error) {
	// collect all TCP connections first
	tcpConns, err := c.CollectTCP(ctx)
	// check if TCP collection failed
	if err != nil {
		// return early on TCP collection failure
		return nil, nil, err
	}

	// collect all UDP connections
	udpConns, err := c.CollectUDP(ctx)
	// check if UDP collection failed
	if err != nil {
		// return early on UDP collection failure
		return nil, nil, err
	}

	// filter TCP connections by PID
	// capacity heuristic: a process typically owns 5-10% of total connections
	tcpEstimatedCap := len(tcpConns) / processConnPercentage
	// ensure minimum capacity for small connection counts
	if tcpEstimatedCap < minProcessConnCapacity {
		tcpEstimatedCap = minProcessConnCapacity
	}
	tcpFiltered := make([]TcpConnection, 0, tcpEstimatedCap)
	// iterate over each TCP connection
	for _, conn := range tcpConns {
		// check if connection belongs to the process
		if conn.PID == pid {
			tcpFiltered = append(tcpFiltered, conn)
		}
	}

	// filter UDP connections by PID
	// capacity heuristic: a process typically owns 5-10% of total connections
	udpEstimatedCap := len(udpConns) / processConnPercentage
	// ensure minimum capacity for small connection counts
	if udpEstimatedCap < minProcessConnCapacity {
		udpEstimatedCap = minProcessConnCapacity
	}
	udpFiltered := make([]UdpConnection, 0, udpEstimatedCap)
	// iterate over each UDP connection
	for _, conn := range udpConns {
		// check if connection belongs to the process
		if conn.PID == pid {
			udpFiltered = append(udpFiltered, conn)
		}
	}

	// return the filtered connections
	return tcpFiltered, udpFiltered, nil
}
