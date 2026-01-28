// Package daemon provides daemon event logging infrastructure.
package daemon

// JSONLogEntry is the JSON structure for log events.
// Metadata fields are inlined into the root of the JSON object.
type JSONLogEntry struct {
	Timestamp string         `json:"ts"`
	Level     string         `json:"level"`
	Service   string         `json:"service,omitempty"`
	Event     string         `json:"event"`
	Message   string         `json:"message,omitempty"`
	Metadata  map[string]any `json:",inline"`
}
