//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// TcpConnection represents a TCP connection with process information.
// It includes local/remote endpoints, state, and owning process details.
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
