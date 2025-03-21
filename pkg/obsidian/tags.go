package obsidian

import (
	"fmt"
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pre-compiled regex patterns for better performance
var (
	frontmatterRegex = regexp.MustCompile(`(?s)^---\n(.*?)\n---\n`)   // Matches YAML frontmatter
	hashtagRegex     = regexp.MustCompile(`(?:^|\s)#[\p{L}\p{N}_-]+`) // Matches hashtags
)

// ExtractFrontmatter extracts YAML frontmatter from a markdown file
func ExtractFrontmatter(content string) (map[string]interface{}, error) {
	matches := frontmatterRegex.FindStringSubmatch(content)
	if len(matches) < 2 {
		return nil, nil
	}

	var frontmatter map[string]interface{}
	if err := yaml.Unmarshal([]byte(matches[1]), &frontmatter); err != nil {
		return nil, err
	}

	// Process tags field if it exists
	if tags, ok := frontmatter["tags"]; ok {
		cleanTags := normalizeTags(tags)
		if len(cleanTags) > 0 {
			frontmatter["tags"] = cleanTags
		} else {
			delete(frontmatter, "tags")
		}
	}

	return frontmatter, nil
}

// normalizeTags normalizes tag values from various formats into a clean string slice
func normalizeTags(tags interface{}) []string {
	var result []string

	switch t := tags.(type) {
	case string:
		// Handle comma-separated tags
		for _, tag := range strings.Split(t, ",") {
			if tag = strings.TrimSpace(tag); tag != "" {
				result = append(result, tag)
			}
		}
	case []interface{}:
		// Process array of tags
		for _, item := range t {
			switch v := item.(type) {
			case string:
				if tag := strings.TrimSpace(v); tag != "" {
					result = append(result, tag)
				}
			case []interface{}:
				// Flatten nested arrays recursively
				result = append(result, normalizeTags(v)...)
			}
		}
	}

	return result
}

// ExtractHashtags extracts hashtags from markdown content, excluding code blocks and inline code
func ExtractHashtags(content string) []string {
	nonCodeContent := extractNonCodeContent(content)
	
	var hashtags []string
	seenTags := make(map[string]bool)
	
	for _, match := range hashtagRegex.FindAllString(nonCodeContent, -1) {
		hashtag := strings.TrimSpace(match)
		if hashtag != "#" && !seenTags[hashtag] {
			hashtags = append(hashtags, hashtag)
			seenTags[hashtag] = true
		}
	}

	return hashtags
}

// extractNonCodeContent removes code blocks and inline code sections from content
func extractNonCodeContent(content string) string {
	lines := strings.Split(content, "\n")
	var inCodeBlock bool
	var result strings.Builder

	for _, line := range lines {
		// Toggle code block state
		if strings.HasPrefix(strings.TrimSpace(line), "```") {
			inCodeBlock = !inCodeBlock
			continue
		}
		
		// Skip lines in code blocks
		if inCodeBlock {
			continue
		}

		// Handle inline code and add non-code parts
		parts := strings.Split(line, "`")
		for i, part := range parts {
			if i%2 == 0 { // Non-code parts
				result.WriteString(part)
			}
			// Add space to preserve word boundaries between parts
			if i < len(parts)-1 {
				result.WriteString(" ")
			}
		}
		result.WriteString("\n")
	}

	return result.String()
}

// CompileTagsRegex creates a regex that matches any of the given tags
func CompileTagsRegex(tags []string) *regexp.Regexp {
	if len(tags) == 0 {
		return regexp.MustCompile(`^\b$`) // Will never match
	}
	
	// Escape special characters in tags
	escapedTags := make([]string, len(tags))
	for i, tag := range tags {
		escapedTags[i] = regexp.QuoteMeta(tag)
	}

	// Create pattern for optional # prefix and word boundaries
	pattern := `(?i)(?:^|\s)(#)?(%s)(?:$|\s|[.,!?])`
	return regexp.MustCompile(fmt.Sprintf(pattern, strings.Join(escapedTags, "|")))
}

// HasAnyTags checks if a file has any of the specified tags
func HasAnyTags(content string, tags []string) bool {
	if len(tags) == 0 {
		return false
	}

	// Check frontmatter tags
	if frontmatter, err := ExtractFrontmatter(content); err == nil && frontmatter != nil {
		if hasFrontmatterTag(frontmatter, tags) {
			return true
		}
	}

	// Check inline hashtags in non-code content
	nonCodeContent := extractNonCodeContent(content)
	tagRegex := CompileTagsRegex(tags)
	
	matches := tagRegex.FindAllStringSubmatch(nonCodeContent, -1)
	for _, match := range matches {
		if len(match) > 2 && match[1] != "" { // Has # prefix
			return true
		}
	}

	return false
}

// hasFrontmatterTag checks if any of the search tags match frontmatter tags
func hasFrontmatterTag(frontmatter map[string]interface{}, searchTags []string) bool {
	frontmatterTags, ok := frontmatter["tags"]
	if !ok {
		return false
	}

	// Get tags as a string slice
	var tagList []string
	switch t := frontmatterTags.(type) {
	case []string:
		tagList = t
	case string:
		for _, tag := range strings.Split(t, ",") {
			if tag = strings.TrimSpace(tag); tag != "" {
				tagList = append(tagList, tag)
			}
		}
	default:
		return false
	}

	// Check for any matching tag (case insensitive)
	for _, fileTag := range tagList {
		for _, searchTag := range searchTags {
			if strings.EqualFold(fileTag, searchTag) {
				return true
			}
		}
	}
	
	return false
}