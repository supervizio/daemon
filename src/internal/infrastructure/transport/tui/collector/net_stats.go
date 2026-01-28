// Package collector provides data collectors for TUI snapshot.
package collector

// netStats is a private helper struct for NetworkCollector.
type netStats struct {
	rxBytes uint64
	txBytes uint64
}
