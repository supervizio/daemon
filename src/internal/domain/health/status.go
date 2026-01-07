// Package health provides domain entities and value objects for health checking.
package health

// Status represents the health status of a service.
// It is used to track whether a service is healthy, unhealthy, or degraded.
type Status int

// Health status constants.
const (
	// StatusUnknown indicates health is not yet determined.
	StatusUnknown Status = iota
	// StatusHealthy indicates all checks pass.
	StatusHealthy
	// StatusUnhealthy indicates checks are failing.
	StatusUnhealthy
	// StatusDegraded indicates some checks are failing.
	StatusDegraded
)

// String returns the string representation of the status.
//
// Returns:
//   - string: the human-readable status name
func (s Status) String() string {
	// Switch on status value to return the corresponding string representation.
	switch s {
	// Case for healthy status.
	case StatusHealthy:
		// Return the healthy status string.
		return "healthy"
	// Case for unhealthy status.
	case StatusUnhealthy:
		// Return the unhealthy status string.
		return "unhealthy"
	// Case for degraded status.
	case StatusDegraded:
		// Return the degraded status string.
		return "degraded"
	// Default case for unknown or unrecognized status.
	default:
		// Return the unknown status string for unrecognized values.
		return "unknown"
	}
}

// IsHealthy returns true if status indicates healthy state.
//
// Returns:
//   - bool: true if status equals StatusHealthy, false otherwise
func (s Status) IsHealthy() bool {
	// Return true if status equals StatusHealthy.
	return s == StatusHealthy
}

// IsUnhealthy returns true if status indicates unhealthy state.
//
// Returns:
//   - bool: true if status equals StatusUnhealthy, false otherwise
func (s Status) IsUnhealthy() bool {
	// Return true if status equals StatusUnhealthy.
	return s == StatusUnhealthy
}
