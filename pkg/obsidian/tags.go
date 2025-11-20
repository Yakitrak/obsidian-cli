package obsidian

import (
	"regexp"
	"strings"

	"gopkg.in/yaml.v3"
)

// Pre-compiled regex patterns for better performance
var (
	frontmatterRegex = regexp.MustCompile(`(?s)^\s*---\r?\n(.*?)\r?\n---\s*\r?\n`) // Matches YAML frontmatter
	hashtagRegex     = regexp.MustCompile(`(?:^|\s)#[\p{L}\p{N}_/\-]+`)            // Matches hashtags, including hierarchical tags
)

// ExtractFrontmatter extracts YAML frontmatter from a markdown file
func ExtractFrontmatter(content string) (map[string]interface{}, error) {
	// Frontmatter must start on the first line
	if !strings.HasPrefix(content, "---") {
		return nil, nil
	}

	// Find the end of the frontmatter
	// We look for "\n---" which indicates the start of the closing delimiter on a new line
	// The closing delimiter must be followed by a newline or end of file
	const closingFence = "\n---"
	endIdx := strings.Index(content[3:], closingFence)
	if endIdx == -1 {
		return nil, nil // No closing delimiter found
	}

	// content[3:] shifts the index, so we add 3 back.
	// endIdx is the start of "\n---", so the YAML content ends there.
	// content[3 : 3+endIdx] is the YAML content.
	frontmatterYAML := content[3 : 3+endIdx]

	var frontmatter map[string]interface{}
	err := yaml.Unmarshal([]byte(frontmatterYAML), &frontmatter)
	if err != nil {
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
		// Handle comma-separated tags directly in a string
		for _, tag := range strings.Split(t, ",") {
			if tag = strings.TrimSpace(tag); tag != "" {
				result = append(result, tag)
			}
		}
	case []interface{}:
		// Process array of tags, potentially nested
		for _, item := range t {
			// Recursively normalize each item in the array and append
			result = append(result, normalizeTags(item)...)
		}
	case []string:
		// Handle simple string array
		for _, tag := range t {
			if tag = strings.TrimSpace(tag); tag != "" {
				result = append(result, tag)
			}
		}
	}

	return result
}

// ExtractHashtags extracts hashtags from markdown content, excluding code blocks and inline code.
// Note: Returned hashtags include the leading '#'. Callers should strip it if they want just the tag name.
func ExtractHashtags(content string) []string {
	var hashtags []string
	seenTags := make(map[string]bool)
	inCodeBlock := false

	for _, rawLine := range strings.Split(content, "\n") {
		line := strings.TrimRight(rawLine, "\r")
		trimLine := strings.TrimSpace(line)

		// Toggle code block state
		if strings.HasPrefix(trimLine, "```") {
			inCodeBlock = !inCodeBlock
			continue
		}

		// Skip lines in code blocks
		if inCodeBlock {
			continue
		}

		// Handle inline code by removing it or splitting
		// We use a simpler approach than the previous regex or manual loop if possible,
		// but to match previous behavior of `split on backtick and take even parts`:
		lineToSearch := line
		if strings.Contains(line, "`") {
			parts := strings.Split(line, "`")
			var sb strings.Builder
			for i, part := range parts {
				if i%2 == 0 { // Non-code parts
					sb.WriteString(part)
					// Add space to preserve word boundaries between parts (matching previous behavior)
					if i < len(parts)-1 {
						sb.WriteString(" ")
					}
				}
			}
			lineToSearch = sb.String()
		}

		for _, match := range hashtagRegex.FindAllString(lineToSearch, -1) {
			hashtag := strings.TrimSpace(match)
			if hashtag != "#" && !seenTags[hashtag] {
				hashtags = append(hashtags, hashtag)
				seenTags[hashtag] = true
			}
		}
	}

	return hashtags
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
	return regexp.MustCompile(strings.Replace(pattern, "%s", strings.Join(escapedTags, "|"), 1))
}

// HasAnyTags checks if a file has any of the specified tags (hierarchical, e.g. foo matches foo/bar)
func HasAnyTags(content string, tags []string) bool {
	if len(tags) == 0 {
		return false
	}

	// Pre-normalize search tags once
	normalizedSearch := make([]string, 0, len(tags))
	for _, t := range tags {
		nt := strings.ToLower(strings.TrimSpace(t))
		if nt != "" {
			normalizedSearch = append(normalizedSearch, nt)
		}
	}
	if len(normalizedSearch) == 0 {
		return false
	}

	// Check frontmatter tags (hierarchical)
	frontmatter, err := ExtractFrontmatter(content)
	if err != nil {
		// Could log the error here if desired, but for now we ignore frontmatter errors and continue
	} else if frontmatter != nil {
		if hasFrontmatterTag(frontmatter, normalizedSearch) {
			return true
		}
	}

	// Check inline hashtags in non-code content (hierarchical)
	// Note: ExtractHashtags returns hashtags with the leading '#', so we strip it for matching.
	hashtags := ExtractHashtags(content)
	for _, normTag := range normalizedSearch {
		for _, hashtag := range hashtags {
			cleanHashtag := strings.TrimSpace(strings.TrimPrefix(hashtag, "#"))
			lowerHashtag := strings.ToLower(cleanHashtag)
			if lowerHashtag == normTag || strings.HasPrefix(lowerHashtag, normTag+"/") {
				return true
			}
		}
	}

	return false
}

// hasFrontmatterTag checks if any of the search tags match frontmatter tags (hierarchical)
func hasFrontmatterTag(frontmatter map[string]interface{}, searchTags []string) bool {
	frontmatterTags, ok := frontmatter["tags"]
	if !ok {
		return false
	}

	// Normalize tags using the shared normalizeTags function
	tagList := normalizeTags(frontmatterTags)

	// Check for any matching tag (case insensitive, hierarchical)
	for _, fileTag := range tagList {
		normFileTag := strings.ToLower(strings.TrimSpace(fileTag))
		for _, searchTag := range searchTags {
			normSearchTag := strings.ToLower(strings.TrimSpace(searchTag))
			if normFileTag == normSearchTag || strings.HasPrefix(normFileTag, normSearchTag+"/") {
				return true
			}
		}
	}

	return false
}
