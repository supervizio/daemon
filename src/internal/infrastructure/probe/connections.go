//go:build cgo

package probe

/*
#include "probe.h"
*/
import "C"

import (
	"context"
	"unsafe"
)

// SocketState represents the state of a network socket.
type SocketState uint8

const (
	SocketStateUnknown     SocketState = 0
	SocketStateEstablished SocketState = 1
	SocketStateSynSent     SocketState = 2
	SocketStateSynRecv     SocketState = 3
	SocketStateFinWait1    SocketState = 4
	SocketStateFinWait2    SocketState = 5
	SocketStateTimeWait    SocketState = 6
	SocketStateClose       SocketState = 7
	SocketStateCloseWait   SocketState = 8
	SocketStateLastAck     SocketState = 9
	SocketStateListen      SocketState = 10
	SocketStateClosing     SocketState = 11
)

// String returns the string representation of the socket state.
func (s SocketState) String() string {
	switch s {
	case SocketStateEstablished:
		return "ESTABLISHED"
	case SocketStateSynSent:
		return "SYN_SENT"
	case SocketStateSynRecv:
		return "SYN_RECV"
	case SocketStateFinWait1:
		return "FIN_WAIT1"
	case SocketStateFinWait2:
		return "FIN_WAIT2"
	case SocketStateTimeWait:
		return "TIME_WAIT"
	case SocketStateClose:
		return "CLOSE"
	case SocketStateCloseWait:
		return "CLOSE_WAIT"
	case SocketStateLastAck:
		return "LAST_ACK"
	case SocketStateListen:
		return "LISTEN"
	case SocketStateClosing:
		return "CLOSING"
	default:
		return "UNKNOWN"
	}
}

// AddressFamily represents the address family of a network connection.
type AddressFamily uint8

const (
	AddressFamilyIPv4 AddressFamily = 4
	AddressFamilyIPv6 AddressFamily = 6
)

// String returns the string representation of the address family.
func (f AddressFamily) String() string {
	switch f {
	case AddressFamilyIPv4:
		return "IPv4"
	case AddressFamilyIPv6:
		return "IPv6"
	default:
		return "Unknown"
	}
}

// TcpConnection represents a TCP connection with process information.
type TcpConnection struct {
	Family      AddressFamily
	LocalAddr   string
	LocalPort   uint16
	RemoteAddr  string
	RemotePort  uint16
	State       SocketState
	PID         int32
	ProcessName string
	Inode       uint64
	RxQueue     uint32
	TxQueue     uint32
}

// UdpConnection represents a UDP socket with process information.
type UdpConnection struct {
	Family      AddressFamily
	LocalAddr   string
	LocalPort   uint16
	RemoteAddr  string
	RemotePort  uint16
	State       SocketState
	PID         int32
	ProcessName string
	Inode       uint64
	RxQueue     uint32
	TxQueue     uint32
}

// UnixSocket represents a Unix domain socket with process information.
type UnixSocket struct {
	Path        string
	SocketType  string
	State       SocketState
	PID         int32
	ProcessName string
	Inode       uint64
}

// TcpStats contains aggregated TCP connection statistics.
type TcpStats struct {
	Established uint32
	SynSent     uint32
	SynRecv     uint32
	FinWait1    uint32
	FinWait2    uint32
	TimeWait    uint32
	Close       uint32
	CloseWait   uint32
	LastAck     uint32
	Listen      uint32
	Closing     uint32
}

// Total returns the total number of TCP connections.
func (s *TcpStats) Total() uint32 {
	return s.Established + s.SynSent + s.SynRecv + s.FinWait1 + s.FinWait2 +
		s.TimeWait + s.Close + s.CloseWait + s.LastAck + s.Listen + s.Closing
}

// ConnectionCollector provides network connection metrics via the Rust probe library.
type ConnectionCollector struct{}

// NewConnectionCollector creates a new connection collector.
func NewConnectionCollector() *ConnectionCollector {
	return &ConnectionCollector{}
}

// CollectTCP returns all TCP connections with process information.
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func (c *ConnectionCollector) CollectTCP(_ context.Context) ([]TcpConnection, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var list C.TcpConnectionList
	result := C.probe_collect_tcp_connections(&list)
	if err := resultToError(result); err != nil {
		return nil, err
	}
	defer C.probe_free_tcp_connection_list(&list)

	connections := make([]TcpConnection, list.count)
	items := unsafe.Slice(list.items, list.count)
	for i, item := range items {
		connections[i] = TcpConnection{
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
		}
	}

	return connections, nil
}

// CollectUDP returns all UDP sockets with process information.
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func (c *ConnectionCollector) CollectUDP(_ context.Context) ([]UdpConnection, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var list C.UdpConnectionList
	result := C.probe_collect_udp_connections(&list)
	if err := resultToError(result); err != nil {
		return nil, err
	}
	defer C.probe_free_udp_connection_list(&list)

	connections := make([]UdpConnection, list.count)
	items := unsafe.Slice(list.items, list.count)
	for i, item := range items {
		connections[i] = UdpConnection{
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
		}
	}

	return connections, nil
}

// CollectUnix returns all Unix domain sockets with process information.
//
//nolint:gocritic // dupSubExpr false positive from CGO list operations
func (c *ConnectionCollector) CollectUnix(_ context.Context) ([]UnixSocket, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var list C.UnixSocketList
	result := C.probe_collect_unix_sockets(&list)
	if err := resultToError(result); err != nil {
		return nil, err
	}
	defer C.probe_free_unix_socket_list(&list)

	sockets := make([]UnixSocket, list.count)
	items := unsafe.Slice(list.items, list.count)
	for i, item := range items {
		sockets[i] = UnixSocket{
			Path:        cCharArrayToString(item.path[:]),
			SocketType:  cCharArrayToString(item.socket_type[:]),
			State:       SocketState(item.state),
			PID:         int32(item.pid),
			ProcessName: cCharArrayToString(item.process_name[:]),
			Inode:       uint64(item.inode),
		}
	}

	return sockets, nil
}

// CollectTCPStats returns aggregated TCP connection statistics.
func (c *ConnectionCollector) CollectTCPStats(_ context.Context) (*TcpStats, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var stats C.TcpStats
	result := C.probe_collect_tcp_stats(&stats)
	if err := resultToError(result); err != nil {
		return nil, err
	}

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
func (c *ConnectionCollector) FindProcessByPort(_ context.Context, port uint16, tcp bool) (int32, error) {
	if err := checkInitialized(); err != nil {
		return -1, err
	}

	var pid C.int32_t
	result := C.probe_find_process_by_port(C.uint16_t(port), C.bool(tcp), &pid)
	if err := resultToError(result); err != nil {
		return -1, err
	}

	return int32(pid), nil
}

// CollectListeningPorts returns all ports that are in LISTEN state.
func (c *ConnectionCollector) CollectListeningPorts(ctx context.Context) ([]TcpConnection, error) {
	connections, err := c.CollectTCP(ctx)
	if err != nil {
		return nil, err
	}

	var listening []TcpConnection
	for _, conn := range connections {
		if conn.State == SocketStateListen {
			listening = append(listening, conn)
		}
	}

	return listening, nil
}

// CollectEstablishedConnections returns all TCP connections in ESTABLISHED state.
func (c *ConnectionCollector) CollectEstablishedConnections(ctx context.Context) ([]TcpConnection, error) {
	connections, err := c.CollectTCP(ctx)
	if err != nil {
		return nil, err
	}

	var established []TcpConnection
	for _, conn := range connections {
		if conn.State == SocketStateEstablished {
			established = append(established, conn)
		}
	}

	return established, nil
}

// CollectProcessConnections returns all TCP and UDP connections for a specific process.
func (c *ConnectionCollector) CollectProcessConnections(ctx context.Context, pid int32) ([]TcpConnection, []UdpConnection, error) {
	tcpConns, err := c.CollectTCP(ctx)
	if err != nil {
		return nil, nil, err
	}

	udpConns, err := c.CollectUDP(ctx)
	if err != nil {
		return nil, nil, err
	}

	var tcpFiltered []TcpConnection
	for _, conn := range tcpConns {
		if conn.PID == pid {
			tcpFiltered = append(tcpFiltered, conn)
		}
	}

	var udpFiltered []UdpConnection
	for _, conn := range udpConns {
		if conn.PID == pid {
			udpFiltered = append(udpFiltered, conn)
		}
	}

	return tcpFiltered, udpFiltered, nil
}
