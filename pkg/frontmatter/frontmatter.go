package frontmatter

import (
	"errors"
	"strings"

	"github.com/adrg/frontmatter"
	"gopkg.in/yaml.v3"
)

const (
	Delimiter              = "---"
	NoFrontmatterError     = "note does not contain frontmatter"
	InvalidFrontmatterError = "frontmatter contains invalid YAML"
)

// Parse extracts and parses frontmatter from note content.
// Returns the frontmatter as a map, the body content, and any error.
func Parse(content string) (map[string]interface{}, string, error) {
	var fm map[string]interface{}
	rest, err := frontmatter.Parse(strings.NewReader(content), &fm)
	if err != nil {
		return nil, "", errors.New(InvalidFrontmatterError)
	}
	return fm, string(rest), nil
}

// Format converts a frontmatter map to a YAML string.
func Format(fm map[string]interface{}) (string, error) {
	if fm == nil || len(fm) == 0 {
		return "", nil
	}
	data, err := yaml.Marshal(fm)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

// HasFrontmatter checks if content starts with frontmatter delimiters.
func HasFrontmatter(content string) bool {
	lines := strings.Split(content, "\n")
	if len(lines) == 0 {
		return false
	}
	return strings.TrimSpace(lines[0]) == Delimiter
}

// SetKey updates or adds a key in the frontmatter, returning the full updated content.
// If no frontmatter exists, it creates new frontmatter with the key.
func SetKey(content, key, value string) (string, error) {
	parsedValue := parseValue(value)

	if !HasFrontmatter(content) {
		// Create new frontmatter
		fm := map[string]interface{}{key: parsedValue}
		fmStr, err := yaml.Marshal(fm)
		if err != nil {
			return "", err
		}
		return Delimiter + "\n" + string(fmStr) + Delimiter + "\n" + content, nil
	}

	// Parse existing frontmatter
	fm, body, err := Parse(content)
	if err != nil {
		return "", err
	}

	if fm == nil {
		fm = make(map[string]interface{})
	}

	// Update the key
	fm[key] = parsedValue

	// Reconstruct content
	fmStr, err := yaml.Marshal(fm)
	if err != nil {
		return "", err
	}

	return Delimiter + "\n" + string(fmStr) + Delimiter + "\n" + body, nil
}

// DeleteKey removes a key from the frontmatter, returning the full updated content.
func DeleteKey(content, key string) (string, error) {
	if !HasFrontmatter(content) {
		return "", errors.New(NoFrontmatterError)
	}

	fm, body, err := Parse(content)
	if err != nil {
		return "", err
	}

	if fm == nil {
		return "", errors.New(NoFrontmatterError)
	}

	// Delete the key
	delete(fm, key)

	// If no keys left, return just the body
	if len(fm) == 0 {
		return strings.TrimPrefix(body, "\n"), nil
	}

	// Reconstruct content
	fmStr, err := yaml.Marshal(fm)
	if err != nil {
		return "", err
	}

	return Delimiter + "\n" + string(fmStr) + Delimiter + "\n" + body, nil
}

// parseValue attempts to parse the value into appropriate Go types.
// Supports: booleans, arrays (comma-separated in brackets), strings.
func parseValue(value string) interface{} {
	// Try boolean
	if value == "true" {
		return true
	}
	if value == "false" {
		return false
	}

	// Try array (comma-separated in brackets)
	if strings.HasPrefix(value, "[") && strings.HasSuffix(value, "]") {
		inner := value[1 : len(value)-1]
		if inner == "" {
			return []string{}
		}
		parts := strings.Split(inner, ",")
		result := make([]string, 0, len(parts))
		for _, p := range parts {
			result = append(result, strings.TrimSpace(p))
		}
		return result
	}

	// Default to string
	return value
}
