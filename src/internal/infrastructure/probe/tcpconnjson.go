//go:build cgo

package probe

// TcpConnJSON contains information about a TCP connection.
// It includes local/remote endpoints, state, and owning process.
type TcpConnJSON struct {
	Family      string `json:"family"`
	LocalAddr   string `json:"local_addr"`
	LocalPort   uint16 `json:"local_port"`
	RemoteAddr  string `json:"remote_addr"`
	RemotePort  uint16 `json:"remote_port"`
	State       string `json:"state"`
	PID         int32  `json:"pid"`
	ProcessName string `json:"process_name,omitempty"`
}
