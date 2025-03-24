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
var embedRegex = regexp.MustCompile(`!\[\[(.*?)(?:\|.*?)?\]\]`)   // Matches embedded wiki-style links

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

// WikilinkOptions defines options for extracting wikilinks
type WikilinkOptions struct {
	SkipAnchors bool // Skip links containing anchors (# symbol)
	SkipEmbeds  bool // Skip embedded links (![[...]])
}

// DefaultWikilinkOptions provides standard options for wikilink extraction
var DefaultWikilinkOptions = WikilinkOptions{
	SkipAnchors: false,
	SkipEmbeds:  false,
}

// ExtractWikilinks extracts wikilinks from markdown content with configurable options
func ExtractWikilinks(content string, options WikilinkOptions) []string {
	// If we need to skip embedded links, remove them from the content first
	contentToProcess := content
	if options.SkipEmbeds {
		contentToProcess = embedRegex.ReplaceAllString(content, "")
	}

	// Extract wikilinks from the processed content
	matches := wikilinkRegex.FindAllStringSubmatch(contentToProcess, -1)
	var links []string

	for _, match := range matches {
		if len(match) > 1 {
			link := filepath.ToSlash(match[1]) // Normalize path separators

			// Skip links with anchors if requested
			if options.SkipAnchors && strings.Contains(link, "#") {
				continue
			}

			links = append(links, link)
		}
	}

	return links
}

// FollowWikilinksOptions contains options for following wikilinks
type FollowWikilinksOptions struct {
	WikilinkOptions     // Embed the WikilinkOptions struct
	MaxDepth        int // Maximum depth to follow links
}

// DefaultFollowWikilinksOptions provides standard options for following wikilinks
var DefaultFollowWikilinksOptions = FollowWikilinksOptions{
	WikilinkOptions: DefaultWikilinkOptions,
	MaxDepth:        -1, // -1 indicates no limit
}

// CreateWikilinksOptions is a helper function to create a FollowWikilinksOptions struct
// from individual parameters for easier migration from legacy code
func CreateWikilinksOptions(maxDepth int, skipAnchors bool, skipEmbeds bool) FollowWikilinksOptions {
	return FollowWikilinksOptions{
		WikilinkOptions: WikilinkOptions{
			SkipAnchors: skipAnchors,
			SkipEmbeds:  skipEmbeds,
		},
		MaxDepth: maxDepth,
	}
}

// FollowWikilinks recursively follows wikilinks according to provided options
func FollowWikilinks(vaultPath string, note NoteManager, startFile string, visited map[string]bool, cache *NotePathCache, options FollowWikilinksOptions) ([]string, error) {
	if visited[startFile] {
		return nil, nil
	}
	visited[startFile] = true

	content, err := note.GetContents(vaultPath, startFile)
	if err != nil {
		return nil, err
	}

	result := []string{startFile}

	// Only follow links if maxDepth != 0
	if options.MaxDepth != 0 {
		links := ExtractWikilinks(content, options.WikilinkOptions)

		for _, link := range links {
			if actualPath, exists := cache.ResolveNote(link); exists {
				// Create a new options with decremented MaxDepth for recursion
				nextOptions := options
				if nextOptions.MaxDepth > 0 {
					nextOptions.MaxDepth--
				}

				if followed, err := FollowWikilinks(vaultPath, note, actualPath, visited, cache, nextOptions); err == nil {
					result = append(result, followed...)
				}
			}
		}
	}

	return result, nil
}
