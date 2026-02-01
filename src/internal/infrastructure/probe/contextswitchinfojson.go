//go:build cgo

package probe

// ContextSwitchInfoJSON contains context switch counts for a process.
// It separates voluntary and involuntary context switches.
type ContextSwitchInfoJSON struct {
	Voluntary   uint64 `json:"voluntary"`
	Involuntary uint64 `json:"involuntary"`
}
