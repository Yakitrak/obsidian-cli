package obsidian

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// TagEditResult represents the result of a tag editing operation
type TagEditResult struct {
	Content string
	Changed bool
}

// RemoveTags removes specified tags from markdown content (both frontmatter and hashtags)
func RemoveTags(content string, tagsToDelete []string) (string, bool) {
	if len(tagsToDelete) == 0 {
		return content, false
	}

	// Normalize tags for comparison
	normalizedDelete := make([]string, len(tagsToDelete))
	for i, tag := range tagsToDelete {
		normalizedDelete[i] = strings.ToLower(strings.TrimSpace(tag))
	}

	// Process frontmatter first
	result, frontmatterChanged := removeFrontmatterTags(content, normalizedDelete)

	// Process inline hashtags
	result, hashtagsChanged := removeInlineHashtags(result, normalizedDelete)

	return result, frontmatterChanged || hashtagsChanged
}

// ReplaceTags replaces specified tags with a new tag in markdown content
func ReplaceTags(content string, fromTags []string, toTag string) (string, bool) {
	if len(fromTags) == 0 || strings.TrimSpace(toTag) == "" {
		return content, false
	}

	// Normalize tags for comparison
	normalizedFrom := make([]string, len(fromTags))
	for i, tag := range fromTags {
		normalizedFrom[i] = strings.ToLower(strings.TrimSpace(tag))
	}
	normalizedTo := strings.TrimSpace(toTag)

	// Process frontmatter first
	result, frontmatterChanged := replaceFrontmatterTags(content, normalizedFrom, normalizedTo)

	// Process inline hashtags
	result, hashtagsChanged := replaceInlineHashtags(result, normalizedFrom, normalizedTo)

	return result, frontmatterChanged || hashtagsChanged
}

// removeFrontmatterTags removes tags from YAML frontmatter
func removeFrontmatterTags(content string, tagsToDelete []string) (string, bool) {
	matches := frontmatterRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return content, false // No frontmatter found
	}

	frontmatterYAML := matches[1]
	var frontmatter map[string]interface{}
	err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter)
	if err != nil {
		return content, false
	}

	// Check if tags field exists
	tagsField, exists := frontmatter["tags"]
	if !exists {
		return content, false
	}

	// Normalize existing tags
	existingTags := normalizeTags(tagsField)
	if len(existingTags) == 0 {
		return content, false
	}

	// Filter out tags to delete
	var remainingTags []string
	changed := false
	for _, tag := range existingTags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		shouldDelete := false
		for _, deleteTag := range tagsToDelete {
			if normalizedTag == deleteTag {
				shouldDelete = true
				changed = true
				break
			}
		}
		if !shouldDelete {
			remainingTags = append(remainingTags, tag)
		}
	}

	if !changed {
		return content, false
	}

	// Update frontmatter
	if len(remainingTags) == 0 {
		delete(frontmatter, "tags")
	} else {
		frontmatter["tags"] = remainingTags
	}

	return rebuildContentWithFrontmatter(content, frontmatter)
}

// replaceFrontmatterTags replaces tags in YAML frontmatter
func replaceFrontmatterTags(content string, fromTags []string, toTag string) (string, bool) {
	matches := frontmatterRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return content, false // No frontmatter found
	}

	frontmatterYAML := matches[1]
	var frontmatter map[string]interface{}
	err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter)
	if err != nil {
		return content, false
	}

	// Check if tags field exists
	tagsField, exists := frontmatter["tags"]
	if !exists {
		return content, false
	}

	// Normalize existing tags
	existingTags := normalizeTags(tagsField)
	if len(existingTags) == 0 {
		return content, false
	}

	// Replace tags and deduplicate
	var newTags []string
	seen := make(map[string]bool)
	changed := false

	for _, tag := range existingTags {
		normalizedTag := strings.ToLower(strings.TrimSpace(tag))
		shouldReplace := false
		var replacement string
		for _, fromTag := range fromTags {
			if normalizedTag == fromTag {
				shouldReplace = true
				replacement = toTag
				changed = true
				break
			}
			// hierarchical match: prefix followed by /
			if strings.HasPrefix(normalizedTag, fromTag+"/") {
				shouldReplace = true
				suffix := tag[len(fromTag):] // retain original case suffix, including leading '/'
				replacement = toTag + suffix
				changed = true
				break
			}
		}

		var tagToAdd string
		if shouldReplace {
			tagToAdd = replacement
		} else {
			tagToAdd = tag
		}

		// Deduplicate (case insensitive)
		normalizedToAdd := strings.ToLower(strings.TrimSpace(tagToAdd))
		if !seen[normalizedToAdd] {
			newTags = append(newTags, tagToAdd)
			seen[normalizedToAdd] = true
		}
	}

	if !changed {
		return content, false
	}

	// Update frontmatter
	frontmatter["tags"] = newTags
	return rebuildContentWithFrontmatter(content, frontmatter)
}

// removeInlineHashtags removes hashtags from content (excluding code blocks)
func removeInlineHashtags(content string, tagsToDelete []string) (string, bool) {
	lines := strings.Split(content, "\n")
	var inCodeBlock bool
	changed := false

	for i, line := range lines {
		// Toggle code block state
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip lines in code blocks
		if inCodeBlock {
			continue
		}

		// Process hashtags in this line
		newLine, lineChanged := removeHashtagsFromLine(line, tagsToDelete)
		if lineChanged {
			lines[i] = newLine
			changed = true
		}
	}

	return strings.Join(lines, "\n"), changed
}

// replaceInlineHashtags replaces hashtags in content (excluding code blocks)
func replaceInlineHashtags(content string, fromTags []string, toTag string) (string, bool) {
	lines := strings.Split(content, "\n")
	var inCodeBlock bool
	changed := false

	for i, line := range lines {
		// Toggle code block state
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip lines in code blocks
		if inCodeBlock {
			continue
		}

		// Process hashtags in this line
		newLine, lineChanged := replaceHashtagsInLine(line, fromTags, toTag)
		if lineChanged {
			lines[i] = newLine
			changed = true
		}
	}

	return strings.Join(lines, "\n"), changed
}

// removeHashtagsFromLine removes specific hashtags from a single line
func removeHashtagsFromLine(line string, tagsToDelete []string) (string, bool) {
	changed := false
	result := line

	for _, tag := range tagsToDelete {
		// Create regex for this specific tag (case insensitive, word boundaries)
		pattern := `(?i)(\s|^)#` + regexp.QuoteMeta(tag) + `(?:\s|$|[^\p{L}\p{N}_/\-])`
		regex := regexp.MustCompile(pattern)

		// Replace hashtag with appropriate spacing
		newResult := regex.ReplaceAllStringFunc(result, func(match string) string {
			// Preserve leading whitespace, remove hashtag, preserve trailing non-alphanumeric
			if strings.HasPrefix(match, " ") {
				if len(match) > len(tag)+2 { // " #tag" + something after
					return " " + match[len(tag)+2:] // Keep trailing char
				}
				return " " // Just " #tag"
			} else {
				if len(match) > len(tag)+1 { // "#tag" + something after
					return match[len(tag)+1:] // Keep trailing char
				}
				return "" // Just "#tag" at start
			}
		})

		if newResult != result {
			result = newResult
			changed = true
		}
	}

	// Clean up multiple spaces and fix spacing issues
	if changed {
		// Replace multiple spaces with single space
		result = regexp.MustCompile(`\s+`).ReplaceAllString(result, " ")
		// Remove space before punctuation
		result = regexp.MustCompile(`\s+([.,!?])`).ReplaceAllString(result, "$1")
		result = strings.TrimSpace(result)
	}

	return result, changed
}

// replaceHashtagsInLine replaces specific hashtags in a single line
func replaceHashtagsInLine(line string, fromTags []string, toTag string) (string, bool) {
	changed := false
	result := line

	for _, tag := range fromTags {
		// Create a combined regex that matches both exact and hierarchical
		pattern := `(?i)(\s|^)(#)` + regexp.QuoteMeta(tag) + `((?:/[\p{L}\p{N}_/\-]+)?)(?:\s|$|[^\p{L}\p{N}_/\-])`
		regex := regexp.MustCompile(pattern)

		// Keep replacing until no more matches (handles overlapping cases)
		for {
			oldResult := result
			result = regex.ReplaceAllStringFunc(result, func(match string) string {
				submatches := regex.FindStringSubmatch(match)
				if len(submatches) < 4 {
					return match
				}
				prefix := submatches[1] + submatches[2] // space + #
				suffix := submatches[3]                 // hierarchical part like "/sub"
				// Preserve any trailing character that was matched for word boundary
				trailing := ""
				expectedLen := len(prefix) + len(tag) + len(suffix)
				if len(match) > expectedLen {
					trailing = match[expectedLen:]
				}
				return prefix + toTag + suffix + trailing
			})
			if result == oldResult {
				break // No more matches
			}
			changed = true
		}
	}

	return result, changed
}

// rebuildContentWithFrontmatter rebuilds content with updated frontmatter
func rebuildContentWithFrontmatter(content string, frontmatter map[string]interface{}) (string, bool) {
	// Find the end of the frontmatter block
	matches := frontmatterRegex.FindStringSubmatch(content)
	if len(matches) < 1 {
		return content, false
	}

	// Get content after frontmatter
	afterFrontmatter := content[len(matches[0]):]

	// If frontmatter is empty, remove the entire block
	if len(frontmatter) == 0 {
		return afterFrontmatter, true
	}

	// Marshal updated frontmatter
	yamlBytes, err := yaml.Marshal(frontmatter)
	if err != nil {
		return content, false
	}

	// Rebuild content
	newContent := "---\n" + string(yamlBytes) + "---\n" + afterFrontmatter
	return newContent, true
}
