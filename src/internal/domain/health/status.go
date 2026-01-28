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
	switch s {
	case StatusHealthy:
		return "healthy"
	case StatusUnhealthy:
		return "unhealthy"
	case StatusDegraded:
		return "degraded"
	case StatusUnknown:
		return "unknown"
	default:
		return "unknown"
	}
}

// IsHealthy returns true if status indicates healthy state.
//
// Returns:
//   - bool: true if status equals StatusHealthy, false otherwise
func (s Status) IsHealthy() bool {
	return s == StatusHealthy
}

// IsUnhealthy returns true if status indicates unhealthy state.
//
// Returns:
//   - bool: true if status equals StatusUnhealthy, false otherwise
func (s Status) IsUnhealthy() bool {
	return s == StatusUnhealthy
}
