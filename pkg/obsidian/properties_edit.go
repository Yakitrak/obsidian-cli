package obsidian

import (
	"fmt"
	"reflect"
	"strings"

	"gopkg.in/yaml.v3"
)

// SetFrontmatterProperty sets a property in YAML frontmatter. If the property already exists and
// overwrite is false, no change is made. When no frontmatter exists, a new block is created.
func SetFrontmatterProperty(content, property string, value interface{}, overwrite bool) (string, bool, error) {
	property = strings.TrimSpace(property)
	if property == "" {
		return content, false, fmt.Errorf("property name is required")
	}

	return editFrontmatter(content, true, func(fm map[string]interface{}) (bool, error) {
		existingKey := findFrontmatterKey(fm, property)
		if existingKey != "" && !overwrite {
			return false, nil
		}

		targetKey := property
		if existingKey != "" {
			targetKey = existingKey // preserve existing casing if present
		}

		fm[targetKey] = value
		return true, nil
	})
}

// DeleteFrontmatterProperties removes the provided properties (case-insensitive) from frontmatter.
// If no frontmatter exists, no change is made.
func DeleteFrontmatterProperties(content string, properties []string) (string, bool, error) {
	if len(properties) == 0 {
		return content, false, fmt.Errorf("no properties specified for deletion")
	}

	lowered := make([]string, len(properties))
	for i, p := range properties {
		lowered[i] = strings.ToLower(strings.TrimSpace(p))
	}

	return editFrontmatter(content, false, func(fm map[string]interface{}) (bool, error) {
		changed := false
		for key := range fm {
			if containsFolded(lowered, key) {
				delete(fm, key)
				changed = true
			}
		}
		return changed, nil
	})
}

// RenameFrontmatterProperties renames one or more properties to a single destination. When merge is true
// and the destination already exists, values are merged (flattening lists and deduping). If merge is false
// and the destination exists, the destination value is left intact and sources are removed.
func RenameFrontmatterProperties(content string, from []string, to string, merge bool) (string, bool, error) {
	if len(from) == 0 {
		return content, false, fmt.Errorf("no properties specified to rename")
	}
	to = strings.TrimSpace(to)
	if to == "" {
		return content, false, fmt.Errorf("destination property is required")
	}

	loweredFrom := make([]string, len(from))
	for i, p := range from {
		loweredFrom[i] = strings.ToLower(strings.TrimSpace(p))
	}

	return editFrontmatter(content, false, func(fm map[string]interface{}) (bool, error) {
		var collected []interface{}
		var toKey string
		changed := false

		// capture existing destination key (preserve casing)
		toKey = findFrontmatterKey(fm, to)
		if toKey == "" {
			toKey = to
		}

		for key, val := range fm {
			if containsFolded(loweredFrom, key) {
				collected = append(collected, val)
				delete(fm, key)
				changed = true
			}
		}

		if !changed {
			return false, nil
		}

		if existing, ok := fm[toKey]; ok {
			if merge {
				fm[toKey] = mergeValues(existing, collected)
			}
			// if merge is false, leave existing value unchanged
		} else if len(collected) == 1 {
			fm[toKey] = collected[0]
		} else if len(collected) > 1 {
			fm[toKey] = flattenList(collected)
		}

		return true, nil
	})
}

// editFrontmatter loads, mutates, and writes frontmatter. If ensureFrontmatter is true, a missing frontmatter
// block will be created when the mutator reports a change.
func editFrontmatter(content string, ensureFrontmatter bool, mutator func(map[string]interface{}) (bool, error)) (string, bool, error) {
	fm, err := ExtractFrontmatter(content)
	if err != nil {
		return content, false, err
	}
	hasFrontmatter := fm != nil
	if fm == nil {
		if !ensureFrontmatter {
			return content, false, nil
		}
		fm = make(map[string]interface{})
	}

	changed, err := mutator(fm)
	if err != nil {
		return content, false, err
	}
	if !changed {
		return content, false, nil
	}

	if hasFrontmatter {
		rebuilt, ok := rebuildContentWithFrontmatter(content, fm)
		return rebuilt, ok, nil
	}

	// No prior frontmatter: create a new block
	yamlBytes, err := yaml.Marshal(fm)
	if err != nil {
		return content, false, err
	}
	newContent := "---\n" + string(yamlBytes) + "---\n" + content
	return newContent, true, nil
}

func findFrontmatterKey(fm map[string]interface{}, target string) string {
	for key := range fm {
		if strings.EqualFold(key, target) {
			return key
		}
	}
	return ""
}

func containsFolded(values []string, candidate string) bool {
	for _, v := range values {
		if strings.EqualFold(v, candidate) {
			return true
		}
	}
	return false
}

func mergeValues(existing interface{}, collected []interface{}) interface{} {
	merged := flattenList(collected)

	switch ev := existing.(type) {
	case []interface{}:
		merged = append(ev, merged...)
	default:
		merged = append([]interface{}{ev}, merged...)
	}

	return dedupeList(merged)
}

func flattenList(values []interface{}) []interface{} {
	var flattened []interface{}
	for _, v := range values {
		switch t := v.(type) {
		case []interface{}:
			flattened = append(flattened, flattenList(t)...)
		case []string:
			for _, s := range t {
				flattened = append(flattened, s)
			}
		default:
			flattened = append(flattened, t)
		}
	}
	return flattened
}

func dedupeList(values []interface{}) interface{} {
	var unique []interface{}
	for _, v := range values {
		found := false
		for _, u := range unique {
			if reflect.DeepEqual(u, v) {
				found = true
				break
			}
		}
		if !found {
			unique = append(unique, v)
		}
	}

	if len(unique) == 1 {
		return unique[0]
	}
	return unique
}
