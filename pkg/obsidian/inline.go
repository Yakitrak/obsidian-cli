package obsidian

import (
	"strings"
	"unicode"
)

// ExtractInlineProperties collects Dataview-style inline properties (Key:: Value) from markdown content outside frontmatter.
// Returns a map of property name to list of raw string values.
func ExtractInlineProperties(content string) map[string][]string {
	result := make(map[string][]string)
	lines := strings.Split(content, "\n")

	inFrontmatter := false
	inCodeFence := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		if i == 0 && strings.HasPrefix(trimmed, "---") {
			inFrontmatter = true
			continue
		}
		if inFrontmatter {
			if strings.TrimSpace(trimmed) == "---" {
				inFrontmatter = false
			}
			continue
		}
		if strings.HasPrefix(trimmed, "```") || strings.HasPrefix(trimmed, "~~~") {
			inCodeFence = !inCodeFence
			continue
		}
		if inCodeFence {
			continue
		}
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if strings.HasPrefix(trimmed, "- ") || strings.HasPrefix(trimmed, "* ") || strings.HasPrefix(trimmed, "+ ") {
			continue
		}
		if len(trimmed) > 1 && trimmed[0] >= '0' && trimmed[0] <= '9' && strings.Contains(trimmed, ". ") {
			continue
		}
		if strings.HasPrefix(trimmed, "|") {
			continue
		}

		parts := strings.SplitN(trimmed, "::", 2)
		if len(parts) != 2 || strings.Contains(parts[0], " ") {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" || val == "" {
			continue
		}
		if !isValidInlineKey(key) {
			continue
		}

		result[key] = append(result[key], val)
	}

	return result
}

func isValidInlineKey(key string) bool {
	runes := []rune(key)
	if len(runes) == 0 {
		return false
	}
	first := runes[0]
	if !unicode.IsLetter(first) && !unicode.IsDigit(first) {
		return false
	}
	for _, r := range runes {
		if unicode.IsLetter(r) || unicode.IsDigit(r) || r == ' ' || r == '-' || r == '_' {
			continue
		}
		return false
	}
	return true
}
