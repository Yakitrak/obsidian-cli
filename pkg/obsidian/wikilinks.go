package obsidian

import (
	"path/filepath"
	"regexp"
	"strings"
)

// NotePathCache maps note names to their full paths for efficient wikilink resolution
type NotePathCache struct {
	// Map from note name (without extension) to full path
	// e.g. "my note" -> "Notes/my note.md"
	// and  "Notes/my note" -> "Notes/my note.md"
	Paths map[string]string
}

// Pre-compiled regex pattern for better performance
var wikilinkRegex = regexp.MustCompile(`\[\[(.*?)(?:\|.*?)?\]\]`) // Matches wiki-style links

// BuildNotePathCache creates a cache of note paths for efficient wikilink resolution
func BuildNotePathCache(allNotes []string) *NotePathCache {
	cache := &NotePathCache{
		Paths: make(map[string]string),
	}

	for _, notePath := range allNotes {
		// Store the full path version
		baseName := strings.TrimSuffix(notePath, filepath.Ext(notePath))
		cache.Paths[baseName] = notePath

		// Store the filename-only version
		fileName := filepath.Base(baseName)
		// Only store filename if it doesn't exist or if current path is shorter
		if existing, ok := cache.Paths[fileName]; !ok || len(notePath) < len(existing) {
			cache.Paths[fileName] = notePath
		}
	}

	return cache
}

// ResolveNote finds the actual note path for a wikilink
func (c *NotePathCache) ResolveNote(link string) (string, bool) {
	// Remove anchors (anything after #) if present
	if idx := strings.Index(link, "#"); idx >= 0 {
		link = link[:idx]
	}

	// Remove extension if present
	baseName := strings.TrimSuffix(link, filepath.Ext(link))

	// Try exact match first
	if path, ok := c.Paths[baseName]; ok {
		return path, true
	}

	// If link contains path separators, try without them
	if strings.Contains(baseName, "/") {
		fileName := filepath.Base(baseName)
		if path, ok := c.Paths[fileName]; ok {
			return path, true
		}
	}

	return "", false
}

// ExtractWikilinks extracts wikilinks from markdown content
func ExtractWikilinks(content string) []string {
	matches := wikilinkRegex.FindAllStringSubmatch(content, -1)
	var links []string

	for _, match := range matches {
		if len(match) > 1 {
			link := filepath.ToSlash(match[1]) // Normalize path separators
			links = append(links, link)
		}
	}

	return links
}

// ExtractWikilinksSkipAnchors extracts wikilinks from markdown content
// but skips any wikilinks that contain anchors (# symbol)
func ExtractWikilinksSkipAnchors(content string) []string {
	matches := wikilinkRegex.FindAllStringSubmatch(content, -1)
	var links []string

	for _, match := range matches {
		if len(match) > 1 {
			link := filepath.ToSlash(match[1]) // Normalize path separators
			
			// Skip links with anchors (containing # symbol)
			if strings.Contains(link, "#") {
				continue
			}
			
			links = append(links, link)
		}
	}

	return links
}

// FollowWikilinks recursively follows wikilinks up to maxDepth
func FollowWikilinks(vaultPath string, note NoteManager, startFile string, maxDepth int, visited map[string]bool, cache *NotePathCache) ([]string, error) {
	return FollowWikilinksWithOptions(vaultPath, note, startFile, maxDepth, visited, cache, false)
}

// FollowWikilinksWithOptions recursively follows wikilinks up to maxDepth with additional options
func FollowWikilinksWithOptions(vaultPath string, note NoteManager, startFile string, maxDepth int, visited map[string]bool, cache *NotePathCache, skipAnchors bool) ([]string, error) {
	if visited[startFile] {
		return nil, nil
	}
	visited[startFile] = true

	content, err := note.GetContents(vaultPath, startFile)
	if err != nil {
		return nil, err
	}

	result := []string{startFile}

	// Only follow links if maxDepth > 0
	if maxDepth > 0 {
		var links []string
		if skipAnchors {
			links = ExtractWikilinksSkipAnchors(content)
		} else {
			links = ExtractWikilinks(content)
		}
		
		for _, link := range links {
			if actualPath, exists := cache.ResolveNote(link); exists {
				if followed, err := FollowWikilinksWithOptions(vaultPath, note, actualPath, maxDepth-1, visited, cache, skipAnchors); err == nil {
					result = append(result, followed...)
				}
			}
		}
	}

	return result, nil
}