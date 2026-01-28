// Package component provides reusable Bubble Tea components.
package component

// serviceColumns is a private helper struct for ServicesPanel.
type serviceColumns struct {
	stateIcon  string
	stateText  string
	healthText string
	uptime     string
	pid        string
	restarts   string
	cpu        string
	mem        string
	ports      string
}
