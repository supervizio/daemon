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
func (s SubjectState) IsReady() bool {
	return s == SubjectReady || s == SubjectRunning
}

// IsListening returns true if this state indicates listening.
// Ready state is also considered listening since a ready listener is still accepting connections.
func (s SubjectState) IsListening() bool {
	return s == SubjectListening || s == SubjectReady
}

// IsClosed returns true if this state indicates closed/stopped/failed.
func (s SubjectState) IsClosed() bool {
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
func NewSubjectSnapshot(name, kind string, state SubjectState) SubjectSnapshot {
	return SubjectSnapshot{
		Name:  name,
		Kind:  kind,
		State: state,
	}
}

// IsReady returns true if the subject is in a ready/running state.
func (s SubjectSnapshot) IsReady() bool {
	return s.State == SubjectReady || s.State == SubjectRunning
}

// IsListening returns true if the subject is listening (including ready state).
func (s SubjectSnapshot) IsListening() bool {
	return s.State == SubjectListening || s.State == SubjectReady
}

// IsClosed returns true if the subject is closed/stopped/failed.
func (s SubjectSnapshot) IsClosed() bool {
	return s.State == SubjectClosed || s.State == SubjectStopped || s.State == SubjectFailed
}
