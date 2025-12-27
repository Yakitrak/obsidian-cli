package cmd

import "strings"

// wizard_ui.go defines shared constants and functions for interactive wizard UI styling.
// This ensures consistent border styling across all wizard prompts.

// DefaultBorderWidth is the standard width for wizard borders
const DefaultBorderWidth = 76

// DoubleLine returns a double-line border of the specified width.
// Used for major section headers (wizard start, completion).
func DoubleLine(width int) string {
	if width < 0 {
		width = 0
	}
	return strings.Repeat("═", width)
}

// SingleLine returns a single-line border of the specified width.
// Used for step headers within a wizard.
func SingleLine(width int) string {
	if width < 0 {
		width = 0
	}
	return strings.Repeat("─", width)
}

