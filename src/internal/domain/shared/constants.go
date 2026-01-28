// Package shared provides common domain types used across multiple domain packages.
package shared

// Network constants.
const (
	// MaxValidPort is the maximum valid TCP/UDP port number.
	// Ports are 16-bit unsigned integers (0-65535).
	MaxValidPort int = 65535
)

// Numeric conversion constants.
const (
	// Base10 is the numeric base for decimal parsing.
	Base10 int = 10
	// BitSize64 is the bit size for int64 parsing.
	BitSize64 int = 64
)

// Unit conversion constants.
const (
	// PercentMultiplier converts ratio (0-1) to percentage (0-100).
	PercentMultiplier float64 = 100.0
	// BitsPerByte is the number of bits in a byte.
	BitsPerByte int = 8
)
