//go:build cgo

package probe

/*
#include "probe.h"
*/
import "C"

// ContextSwitches contains context switch statistics.
type ContextSwitches struct {
	// Voluntary context switches (process yielded CPU).
	Voluntary uint64
	// Involuntary context switches (preempted by scheduler).
	Involuntary uint64
	// SystemTotal is the system-wide total context switches.
	SystemTotal uint64
}

// CollectSystemContextSwitches collects the system-wide context switch count.
func CollectSystemContextSwitches() (uint64, error) {
	if err := checkInitialized(); err != nil {
		return 0, err
	}

	var count C.uint64_t
	result := C.probe_collect_system_context_switches(&count)
	if err := resultToError(result); err != nil {
		return 0, err
	}

	return uint64(count), nil
}

// CollectProcessContextSwitches collects context switches for a specific process.
func CollectProcessContextSwitches(pid int32) (*ContextSwitches, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var cs C.ContextSwitches
	result := C.probe_collect_process_context_switches(C.int32_t(pid), &cs)
	if err := resultToError(result); err != nil {
		return nil, err
	}

	return &ContextSwitches{
		Voluntary:   uint64(cs.voluntary),
		Involuntary: uint64(cs.involuntary),
		SystemTotal: uint64(cs.system_total),
	}, nil
}

// CollectSelfContextSwitches collects context switches for the current process.
func CollectSelfContextSwitches() (*ContextSwitches, error) {
	if err := checkInitialized(); err != nil {
		return nil, err
	}

	var cs C.ContextSwitches
	result := C.probe_collect_self_context_switches(&cs)
	if err := resultToError(result); err != nil {
		return nil, err
	}

	return &ContextSwitches{
		Voluntary:   uint64(cs.voluntary),
		Involuntary: uint64(cs.involuntary),
		SystemTotal: uint64(cs.system_total),
	}, nil
}
