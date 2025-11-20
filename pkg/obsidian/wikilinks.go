package obsidian

import (
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"sync"
)

// NotePathCache maps note names to their full paths for efficient wikilink resolution
type NotePathCache struct {
	// Map from note name (without extension) to full path
	// e.g. "my note" -> "Notes/my note.md"
	// and  "Notes/my note" -> "Notes/my note.md"
	Paths map[string]string
}

// BacklinkType represents the type of wikilink variant used.
type BacklinkType string

const (
	BacklinkTypeBasic   BacklinkType = "basic"
	BacklinkTypeAlias   BacklinkType = "alias"
	BacklinkTypeHeading BacklinkType = "heading"
	BacklinkTypeBlock   BacklinkType = "block"
	BacklinkTypeEmbed   BacklinkType = "embed"
)

// Backlink captures a referrer and the link variant used.
type Backlink struct {
	Referrer string       `json:"referrer"`
	LinkType BacklinkType `json:"linkType"`
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

type wikilinkDetail struct {
	Target   string
	LinkType BacklinkType
}

func classifyLink(rawMatch string, link string, isEmbed bool) BacklinkType {
	switch {
	case isEmbed:
		return BacklinkTypeEmbed
	case strings.Contains(rawMatch, "|"):
		return BacklinkTypeAlias
	case strings.Contains(link, "#^"):
		return BacklinkTypeBlock
	case strings.Contains(link, "#"):
		return BacklinkTypeHeading
	default:
		return BacklinkTypeBasic
	}
}

// extractWikilinkDetails returns wikilinks with their detected variant types.
func extractWikilinkDetails(content string, options WikilinkOptions) []wikilinkDetail {
	var details []wikilinkDetail

	if !options.SkipEmbeds {
		embedMatches := embedRegex.FindAllStringSubmatch(content, -1)
		for _, match := range embedMatches {
			if len(match) > 1 {
				link := filepath.ToSlash(match[1])
				details = append(details, wikilinkDetail{
					Target:   link,
					LinkType: BacklinkTypeEmbed,
				})
			}
		}
	}

	// Remove embeds so they are not double-counted as regular wikilinks.
	contentWithoutEmbeds := embedRegex.ReplaceAllString(content, "")
	matches := wikilinkRegex.FindAllStringSubmatch(contentWithoutEmbeds, -1)

	for _, match := range matches {
		if len(match) > 1 {
			link := filepath.ToSlash(match[1])
			if options.SkipAnchors && strings.Contains(link, "#") {
				continue
			}
			details = append(details, wikilinkDetail{
				Target:   link,
				LinkType: classifyLink(match[0], link, false),
			})
		}
	}

	return details
}

func containsSuppressedTag(content string, suppressedTags []string) bool {
	if len(suppressedTags) == 0 {
		return false
	}
	lower := strings.ToLower(content)
	for _, tag := range suppressedTags {
		tag = strings.ToLower(strings.TrimSpace(tag))
		if tag == "" {
			continue
		}
		if strings.Contains(lower, "#"+tag) || strings.Contains(lower, tag) {
			return true
		}
	}
	return false
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

// CollectBacklinks finds first-degree backlinks for the provided targets.
// Each target key is present in the returned map, even if it has no backlinks.
// Options allow skipping anchors or embeds; suppressedTags removes referrers that contain those tags.
func CollectBacklinks(vaultPath string, note NoteManager, targets []string, options WikilinkOptions, suppressedTags []string) (map[string][]Backlink, error) {
	allNotes, err := note.GetNotesList(vaultPath)
	if err != nil {
		return nil, err
	}

	cache := BuildNotePathCache(allNotes)

	targetSet := make(map[string]struct{})
	result := make(map[string][]Backlink)
	for _, target := range targets {
		normalized := NormalizePath(AddMdSuffix(target))
		targetSet[normalized] = struct{}{}
		result[normalized] = []Backlink{}
	}

	backlinkMap := make(map[string]map[string]BacklinkType)

	// Parallelize referrer scanning to use available cores.
	workerCount := runtime.NumCPU()
	if workerCount < 1 {
		workerCount = 1
	}
	jobs := make(chan string, workerCount)

	var mu sync.Mutex
	var wg sync.WaitGroup

	worker := func() {
		defer wg.Done()
		for referrer := range jobs {
			referrerNorm := NormalizePath(referrer)

			content, err := note.GetContents(vaultPath, referrer)
			if err != nil {
				// Skip unreadable files without blocking other results.
				continue
			}

			if len(suppressedTags) > 0 && containsSuppressedTag(content, suppressedTags) {
				continue
			}

			links := extractWikilinkDetails(content, options)
			for _, link := range links {
				if resolved, ok := cache.ResolveNote(link.Target); ok {
					targetPath := NormalizePath(AddMdSuffix(resolved))
					if _, isTarget := targetSet[targetPath]; !isTarget {
						continue
					}

					mu.Lock()
					if _, exists := backlinkMap[targetPath]; !exists {
						backlinkMap[targetPath] = make(map[string]BacklinkType)
					}
					if _, seen := backlinkMap[targetPath][referrerNorm]; !seen {
						backlinkMap[targetPath][referrerNorm] = link.LinkType
					}
					mu.Unlock()
				}
			}
		}
	}

	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go worker()
	}

	for _, referrer := range allNotes {
		jobs <- referrer
	}
	close(jobs)
	wg.Wait()

	for targetPath, referrers := range backlinkMap {
		for referrer, linkType := range referrers {
			result[targetPath] = append(result[targetPath], Backlink{
				Referrer: referrer,
				LinkType: linkType,
			})
		}
		sort.Slice(result[targetPath], func(i, j int) bool {
			if result[targetPath][i].Referrer == result[targetPath][j].Referrer {
				return result[targetPath][i].LinkType < result[targetPath][j].LinkType
			}
			return result[targetPath][i].Referrer < result[targetPath][j].Referrer
		})
	}

	return result, nil
}
