//go:build cgo

// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
//
// This file serves as a companion to export_internal_test.go, which
// exports internal identifiers for external testing. The export mechanism
// allows external test packages (_external_test.go files) to access
// package internals without breaking encapsulation in production code.
package probe
