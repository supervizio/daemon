//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
package probe

// ContextSwitchInfoJSON contains context switch counts for a process.
// It separates voluntary and involuntary context switches.
type ContextSwitchInfoJSON struct {
	Voluntary   uint64 `json:"voluntary"`
	Involuntary uint64 `json:"involuntary"`
}
