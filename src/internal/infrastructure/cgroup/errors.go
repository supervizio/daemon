//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

import "errors"

// Package errors.
var (
	// ErrV1NotSupported is returned when cgroups v1 is detected but not supported.
	ErrV1NotSupported = errors.New("cgroups v1 is not supported")

	// ErrUnknownVersion is returned when the cgroup version cannot be detected.
	ErrUnknownVersion = errors.New("unknown cgroup version")

	// ErrPathNotFound is returned when the cgroup path cannot be detected.
	ErrPathNotFound = errors.New("cgroup v2 path not found in /proc/self/cgroup")
)
