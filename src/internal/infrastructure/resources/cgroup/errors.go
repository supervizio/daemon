//go:build linux

// Package cgroup provides adapters for reading cgroup metrics.
package cgroup

import "errors"

// Package errors.
var (
	// ErrUnknownVersion is returned when the cgroup version cannot be detected.
	ErrUnknownVersion = errors.New("unknown cgroup version")

	// ErrPathNotFound is returned when the cgroup path cannot be detected.
	ErrPathNotFound = errors.New("cgroup path not found in /proc/self/cgroup")
)
