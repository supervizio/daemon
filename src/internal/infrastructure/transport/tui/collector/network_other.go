//go:build !linux

package collector

// getInterfaceStats returns placeholder stats on non-Linux platforms.
// TODO: Implement using syscalls on macOS/BSD.
func getInterfaceStats(name string) (rxBytes, txBytes, speed uint64) {
	// Platform-specific implementation needed.
	// For now, return zeros (best effort).
	_ = name
	return 0, 0, 0
}

// UpdateAdaptiveSpeed is a no-op on non-Linux platforms.
func UpdateAdaptiveSpeed(_ string, _ uint64) {
	// No-op: adaptive speed only implemented on Linux.
}
