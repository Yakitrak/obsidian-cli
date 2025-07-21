package obsidian

import "strings"

// NormalizeTag trims spaces and lower-cases the tag for canonical comparison.
// It does not strip a leading # – callers should do that if needed.
func NormalizeTag(tag string) string {
	return strings.ToLower(strings.TrimSpace(tag))
}

// IsValidTag returns true when the tag obeys basic Obsidian rules:
//   - not empty
//   - contains no spaces
//   - is not purely numeric
//   - contains at least one letter
func IsValidTag(tag string) bool {
	if tag == "" {
		return false
	}
	if strings.Contains(tag, " ") {
		return false
	}
	if isNumeric(tag) {
		return false
	}
	if !containsLetter(tag) {
		return false
	}
	return true
}

func isNumeric(s string) bool {
	if s == "" {
		return false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return true
}

func containsLetter(s string) bool {
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') {
			return true
		}
	}
	return false
}
