//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for unified
// cross-platform system metrics and resource quota management.
package probe

/*
#include "probe.h"
*/
import "C"

// ContextSwitches contains context switch statistics.
// It tracks both voluntary and involuntary context switches for a process.
type ContextSwitches struct {
	// Voluntary context switches (process yielded CPU).
	Voluntary uint64
	// Involuntary context switches (preempted by scheduler).
	Involuntary uint64
	// SystemTotal is the system-wide total context switches.
	SystemTotal uint64
}

// CollectSystemContextSwitches collects the system-wide context switch count.
//
// Returns:
//   - uint64: total number of context switches since system boot
//   - error: nil on success, error if probe not initialized or collection fails
func CollectSystemContextSwitches() (uint64, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return zero with initialization error.
		return 0, err
	}

	var count C.uint64_t
	result := C.probe_collect_system_context_switches(&count)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return zero with collection error.
		return 0, err
	}

	// Return the collected context switch count.
	return uint64(count), nil
}

// CollectProcessContextSwitches collects context switches for a specific process.
//
// Params:
//   - pid: process ID to collect context switches for
//
// Returns:
//   - *ContextSwitches: context switch statistics for the process
//   - error: nil on success, error if probe not initialized or collection fails
func CollectProcessContextSwitches(pid int32) (*ContextSwitches, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return nil with initialization error.
		return nil, err
	}

	var cs C.ContextSwitches
	result := C.probe_collect_process_context_switches(C.int32_t(pid), &cs)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return nil with collection error.
		return nil, err
	}

	// Return collected context switch statistics.
	return &ContextSwitches{
		Voluntary:   uint64(cs.voluntary),
		Involuntary: uint64(cs.involuntary),
		SystemTotal: uint64(cs.system_total),
	}, nil
}

// CollectSelfContextSwitches collects context switches for the current process.
//
// Returns:
//   - *ContextSwitches: context switch statistics for this process
//   - error: nil on success, error if probe not initialized or collection fails
func CollectSelfContextSwitches() (*ContextSwitches, error) {
	// Verify probe library is initialized before collecting.
	if err := checkInitialized(); err != nil {
		// Return nil with initialization error.
		return nil, err
	}

	var cs C.ContextSwitches
	result := C.probe_collect_self_context_switches(&cs)
	// Check if the FFI call succeeded.
	if err := resultToError(result); err != nil {
		// Return nil with collection error.
		return nil, err
	}

	// Return collected context switch statistics.
	return &ContextSwitches{
		Voluntary:   uint64(cs.voluntary),
		Involuntary: uint64(cs.involuntary),
		SystemTotal: uint64(cs.system_total),
	}, nil
}
