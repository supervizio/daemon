// Package health provides domain entities for health checking.
package health

// SubjectState represents the state of a monitored subject.
// This is a domain-owned type that doesn't depend on external packages.
type SubjectState string

const (
	// SubjectUnknown indicates the state is not yet determined.
	SubjectUnknown SubjectState = "unknown"
	// SubjectReady indicates the subject is fully operational.
	SubjectReady SubjectState = "ready"
	// SubjectListening indicates the subject is accepting but not fully ready.
	SubjectListening SubjectState = "listening"
	// SubjectClosed indicates the subject is not operational.
	SubjectClosed SubjectState = "closed"
	// SubjectRunning indicates a process is running.
	SubjectRunning SubjectState = "running"
	// SubjectStopped indicates a process is stopped.
	SubjectStopped SubjectState = "stopped"
	// SubjectFailed indicates a process has failed.
	SubjectFailed SubjectState = "failed"
)

// IsReady returns true if this state indicates ready/running.
//
// Returns:
//   - bool: true if state is ready or running, false otherwise.
func (s SubjectState) IsReady() bool {
	// check for ready or running state
	return s == SubjectReady || s == SubjectRunning
}

// IsListening returns true if this state indicates listening.
// Ready state is also considered listening since a ready listener is still accepting connections.
//
// Returns:
//   - bool: true if state is listening or ready, false otherwise.
func (s SubjectState) IsListening() bool {
	// check for listening or ready state
	return s == SubjectListening || s == SubjectReady
}

// IsClosed returns true if this state indicates closed/stopped/failed.
//
// Returns:
//   - bool: true if state is closed, stopped, or failed, false otherwise.
func (s SubjectState) IsClosed() bool {
	// check for closed, stopped, or failed state
	return s == SubjectClosed || s == SubjectStopped || s == SubjectFailed
}

// SubjectSnapshot represents a point-in-time view of a subject's state.
// The application layer creates these from concrete types (listener.State, process.State).
type SubjectSnapshot struct {
	// Name is the identifier of the subject.
	Name string
	// Kind identifies the type of subject ("listener", "process").
	Kind string
	// State is the current state.
	State SubjectState
}

// NewSubjectSnapshot creates a new snapshot.
//
// Params:
//   - name: the identifier of the subject.
//   - kind: the type of subject ("listener", "process").
//   - state: the current state of the subject.
//
// Returns:
//   - SubjectSnapshot: a new subject snapshot with the specified values.
func NewSubjectSnapshot(name, kind string, state SubjectState) SubjectSnapshot {
	// create snapshot with all fields
	return SubjectSnapshot{
		Name:  name,
		Kind:  kind,
		State: state,
	}
}

// IsReady returns true if the subject is in a ready/running state.
//
// Returns:
//   - bool: true if subject is ready or running, false otherwise.
func (s SubjectSnapshot) IsReady() bool {
	// check for ready or running state
	return s.State == SubjectReady || s.State == SubjectRunning
}

// IsListening returns true if the subject is listening (including ready state).
//
// Returns:
//   - bool: true if subject is listening or ready, false otherwise.
func (s SubjectSnapshot) IsListening() bool {
	// check for listening or ready state
	return s.State == SubjectListening || s.State == SubjectReady
}

// IsClosed returns true if the subject is closed/stopped/failed.
//
// Returns:
//   - bool: true if subject is closed, stopped, or failed, false otherwise.
func (s SubjectSnapshot) IsClosed() bool {
	// check for closed, stopped, or failed state
	return s.State == SubjectClosed || s.State == SubjectStopped || s.State == SubjectFailed
}
