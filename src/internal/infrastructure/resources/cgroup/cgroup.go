//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
//
// The package supports cgroups v1 (legacy), v2 (unified hierarchy), and hybrid modes.
// Use [Detect] to auto-detect the cgroup version, or [NewReader] to create a reader
// with automatic version detection.
//
// Example:
//
//	reader, err := cgroup.NewReader()
//	if err != nil {
//	    return err
//	}
//	usage, _ := reader.CPUUsage(ctx)
//	limit, _ := reader.MemoryLimit(ctx)
package cgroup

// Re-exported for documentation purposes.
// All types and functions are defined in their respective files:
//   - detector.go: Version, Detect, DetectWithPath, Reader, NewReader
//   - v1.go: V1Reader
//   - v2.go: V2Reader
//   - errors.go: ErrUnknownVersion, ErrPathNotFound
