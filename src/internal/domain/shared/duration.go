// Package shared provides common domain types used across multiple domain packages.
package shared

import "time"

// Duration is a domain wrapper around time.Duration.
// It provides a clean domain type without serialization concerns.
type Duration time.Duration

// Duration returns the underlying time.Duration value.
//
// Returns:
//   - time.Duration: the wrapped duration value
func (d Duration) Duration() time.Duration {
	return time.Duration(d)
}

// Seconds returns the duration in seconds.
//
// Returns:
//   - float64: the duration expressed in seconds
func (d Duration) Seconds() float64 {
	return time.Duration(d).Seconds()
}

// Milliseconds returns the duration in milliseconds.
//
// Returns:
//   - int64: the duration expressed in milliseconds
func (d Duration) Milliseconds() int64 {
	return time.Duration(d).Milliseconds()
}

// String returns the string representation.
//
// Returns:
//   - string: human-readable duration format
func (d Duration) String() string {
	return time.Duration(d).String()
}

// Common duration constructors for convenience.

// Seconds creates a Duration from seconds.
//
// Params:
//   - s: the number of seconds
//
// Returns:
//   - Duration: the duration value
func Seconds(s int) Duration {
	return Duration(time.Duration(s) * time.Second)
}

// Minutes creates a Duration from minutes.
//
// Params:
//   - m: the number of minutes
//
// Returns:
//   - Duration: the duration value
func Minutes(m int) Duration {
	return Duration(time.Duration(m) * time.Minute)
}

// FromTimeDuration converts time.Duration to shared.Duration.
//
// Params:
//   - d: the time.Duration to convert
//
// Returns:
//   - Duration: the wrapped duration value
func FromTimeDuration(d time.Duration) Duration {
	return Duration(d)
}
