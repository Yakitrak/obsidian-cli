package obsidian

import (
	"strings"
)

// ExtractInlineProperties collects Dataview-style inline properties (Key:: Value) from markdown content outside frontmatter.
// Returns a map of property name to list of raw string values.
func ExtractInlineProperties(content string) map[string][]string {
	result := make(map[string][]string)
	lines := strings.Split(content, "\n")

	inFrontmatter := false
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
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}

		parts := strings.SplitN(trimmed, "::", 2)
		if len(parts) != 2 {
			continue
		}

		key := strings.TrimSpace(parts[0])
		val := strings.TrimSpace(parts[1])
		if key == "" || val == "" {
			continue
		}

		result[key] = append(result[key], val)
	}

	return result
}
