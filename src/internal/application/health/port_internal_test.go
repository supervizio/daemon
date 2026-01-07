// Package health provides internal tests for port.go.
// It tests internal implementation details using white-box testing.
package health

// Note: Internal tests for port.go are minimal as it only defines
// interfaces and a configuration struct. The Checker and Creator
// interfaces are tested through external tests with mock implementations.
// The CheckerConfig struct has no unexported fields or methods to test.
