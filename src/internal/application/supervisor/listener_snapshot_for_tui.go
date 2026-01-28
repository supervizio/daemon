// Package supervisor provides the application service for orchestrating multiple services.
package supervisor

// Listener status codes for TUI display.
const (
	// ListenerStatusOK indicates the listener is healthy (green).
	ListenerStatusOK int = 0
	// ListenerStatusWarning indicates a warning state (yellow).
	ListenerStatusWarning int = 1
	// ListenerStatusError indicates an error state (red).
	ListenerStatusError int = 2
)

// ListenerSnapshotForTUI contains listener info for TUI display.
// This struct uses basic types to avoid import cycles with TUI packages.
// Each listener tracks its configuration and runtime status for visualization.
type ListenerSnapshotForTUI struct {
	Name      string
	Port      int
	Protocol  string
	Exposed   bool // Whether the port should be publicly accessible
	Listening bool // Whether the port is actually listening
	StatusInt int  // 0=OK (green), 1=Warning (yellow), 2=Error (red)
}
