package obsidian

import (
	"strings"
)

// FuzzyMatch determines if a file path matches a search pattern
// Used for non-interactive file filtering in the list and prompt commands
func FuzzyMatch(pattern, path string) bool {
	if pattern == "" || path == "" {
		return false
	}

	// Handle the directory and content differently based on pattern format
	hasDirectorySpecifier := strings.Contains(pattern, "/")
	
	// Prevent patterns with multiple slashes (not supported)
	if hasDirectorySpecifier && strings.Count(pattern, "/") > 1 {
		return false
	}

	// Normalize for case sensitivity
	patternLower := strings.ToLower(pattern)
	pathLower := strings.ToLower(path)

	// Handle directory-specific search
	if hasDirectorySpecifier {
		dirPattern, contentPattern := splitDirectoryAndContent(patternLower)
		
		// First check if directory part matches
		if !matchesDirectory(dirPattern, pathLower) {
			return false
		}
		
		// If we have a content part to match as well
		if contentPattern != "" {
			// Split path to get content section (everything after first /)
			parts := strings.SplitN(pathLower, "/", 2)
			if len(parts) < 2 {
				return false // No content section in path
			}
			
			return matchesContent(contentPattern, parts[1])
		}
		
		return true // Only directory matched, but that's all we asked for
	}
	
	// Content-only search (no directory specifier)
	return matchesContentOnly(patternLower, pathLower)
}

// splitDirectoryAndContent splits a pattern into directory and content components
func splitDirectoryAndContent(pattern string) (string, string) {
	parts := strings.SplitN(pattern, "/", 2)
	dirPattern := parts[0]
	contentPattern := ""
	if len(parts) > 1 {
		contentPattern = parts[1]
	}
	return dirPattern, contentPattern
}

// matchesDirectory checks if a directory pattern matches the beginning of a path
// Handles single character prefix matching
func matchesDirectory(dirPattern string, path string) bool {
	pathParts := strings.Split(path, "/")
	if len(pathParts) == 0 {
		return false
	}
	
	firstPathSegment := pathParts[0]
	
	// Single character directory match allows prefix matching
	if len(dirPattern) == 1 {
		return strings.HasPrefix(firstPathSegment, dirPattern)
	}
	
	// Otherwise, exact directory match is required
	return firstPathSegment == dirPattern
}

// matchesContent checks if content words match sequentially with word boundaries
func matchesContent(contentPattern string, content string) bool {
	patternWords := splitWords(contentPattern)
	
	// No words to match
	if len(patternWords) == 0 {
		return true
	}
	
	// Try to match all pattern words in order
	searchText := content
	for _, word := range patternWords {
		found := false
		for {
			index := strings.Index(searchText, word)
			if index == -1 {
				return false // Word not found
			}
			if isWordBoundary(searchText, index) {
				searchText = searchText[index+len(word):]
				found = true
				break
			}
			searchText = searchText[index+1:]
		}
		if !found {
			return false
		}
	}
	
	return true
}

// matchesContentOnly checks if a pattern matches anywhere in the path
func matchesContentOnly(pattern string, path string) bool {
	patternWords := splitWords(pattern)
	
	// Try to match all pattern words in order
	searchText := path
	for _, word := range patternWords {
		found := false
		for {
			index := strings.Index(searchText, word)
			if index == -1 {
				return false // Word not found
			}
			if isWordBoundary(searchText, index) {
				searchText = searchText[index+len(word):]
				found = true
				break
			}
			searchText = searchText[index+1:]
		}
		if !found {
			return false
		}
	}
	
	return true
}

// isDelimiter checks if a byte is a delimiter character
func isDelimiter(b byte) bool {
	switch b {
	case '/', '-', '_', ' ', '.', ',', '(', ')':
		return true
	default:
		return false
	}
}

// isWordBoundary checks if a position represents a word boundary
func isWordBoundary(text string, pos int) bool {
	if pos == 0 {
		return true
	}
	if pos >= len(text) {
		return false
	}
	return isDelimiter(text[pos-1])
}

// splitWords splits text into words, treating hyphens and underscores as separators
func splitWords(text string) []string {
	parts := strings.Fields(text)
	var result []string

	for _, part := range parts {
		// Split by hyphen and underscore
		for _, hp := range strings.Split(part, "-") {
			for _, up := range strings.Split(hp, "_") {
				if up != "" {
					result = append(result, up)
				}
			}
		}
	}

	return result
}