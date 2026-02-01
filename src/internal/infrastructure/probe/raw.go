// Package probe provides CGO bindings to the Rust probe library for
// unified cross-platform system metrics and resource quota management.
//
// This file serves as the package documentation for Raw* types.
// The Raw* types are Go-only data structures that mirror C data
// without requiring CGO for testing.
package probe

// RawTypesVersion represents the version of the Raw* type definitions.
// This is used to track compatibility between Raw types and CGO bindings.
const RawTypesVersion string = "1.0.0"
