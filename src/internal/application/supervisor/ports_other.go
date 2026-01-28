//go:build !linux

package supervisor

// getListeningPorts returns TCP/UDP ports the process is listening on.
// Not implemented on non-Linux platforms.
func getListeningPorts(_ int) []int {
	return nil
}
