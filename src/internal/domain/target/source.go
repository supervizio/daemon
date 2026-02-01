// Package target provides domain entities for external monitoring targets.
// External targets are services, containers, or hosts that supervizio
// monitors but does not manage (no lifecycle control).
package target

// Source indicates how a target was added to the monitoring system.
type Source string

// Source constants define how targets are discovered.
const (
	// SourceStatic indicates the target was defined in configuration.
	SourceStatic Source = "static"

	// SourceDiscovered indicates the target was auto-discovered.
	SourceDiscovered Source = "discovered"
)

// String returns the string representation of the source.
//
// Returns:
//   - string: the source as a string.
func (s Source) String() string {
	// Convert Source enum to string for display and serialization.
	return string(s)
}

// IsStatic checks if the target is statically configured.
//
// Returns:
//   - bool: true if the source is static.
func (s Source) IsStatic() bool {
	// Compare with SourceStatic constant to determine if static.
	return s == SourceStatic
}

// IsDiscovered checks if the target was auto-discovered.
//
// Returns:
//   - bool: true if the source is discovered.
func (s Source) IsDiscovered() bool {
	// Compare with SourceDiscovered constant to determine if auto-discovered.
	return s == SourceDiscovered
}
