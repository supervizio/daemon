//go:build cgo

package probe

// UdpConnJSON contains information about a UDP socket.
// It includes local/remote endpoints and owning process.
type UdpConnJSON struct {
	Family      string `json:"family"`
	LocalAddr   string `json:"local_addr"`
	LocalPort   uint16 `json:"local_port"`
	RemoteAddr  string `json:"remote_addr"`
	RemotePort  uint16 `json:"remote_port"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}
