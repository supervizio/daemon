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
	// map status to string
	switch s {
	// healthy status
	case StatusHealthy:
		// return healthy label
		return "healthy"
	// unhealthy status
	case StatusUnhealthy:
		// return unhealthy label
		return "unhealthy"
	// degraded status
	case StatusDegraded:
		// return degraded label
		return "degraded"
	// unknown status
	case StatusUnknown:
		// return unknown label
		return "unknown"
	// fallback for invalid status
	default:
		// return default unknown label
		return "unknown"
	}
}

// IsHealthy returns true if status indicates healthy state.
//
// Returns:
//   - bool: true if status equals StatusHealthy, false otherwise
func (s Status) IsHealthy() bool {
	// check for healthy status
	return s == StatusHealthy
}

// IsUnhealthy returns true if status indicates unhealthy state.
//
// Returns:
//   - bool: true if status equals StatusUnhealthy, false otherwise
func (s Status) IsUnhealthy() bool {
	// check for unhealthy status
	return s == StatusUnhealthy
}
