package actions

import (
	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

// FileInfo represents the information about a file
type FileInfo struct {
	Frontmatter map[string]interface{}
	Tags        []string
}

// GetFileInfo retrieves the frontmatter and tags from a file
func GetFileInfo(vault obsidian.VaultManager, note obsidian.NoteManager, filePath string) (*FileInfo, error) {
	vaultPath, err := vault.Path()
	if err != nil {
		return nil, err
	}

	content, err := note.GetContents(vaultPath, filePath)
	if err != nil {
		return nil, err
	}

	// Extract frontmatter
	frontmatter, err := obsidian.ExtractFrontmatter(content)
	if err != nil {
		return nil, err
	}

	// Get all hashtags from the content
	hashtags := obsidian.ExtractHashtags(content)

	// Get tags from frontmatter
	var frontmatterTags []string
	if frontmatter != nil {
		if tags, ok := frontmatter["tags"]; ok {
			switch t := tags.(type) {
			case []string:
				frontmatterTags = t
			case string:
				// Handle comma-separated tags
				frontmatterTags = []string{t}
			}
		}
	}

	// Combine and deduplicate tags
	seenTags := make(map[string]bool)
	var allTags []string

	// Add frontmatter tags
	for _, tag := range frontmatterTags {
		if !seenTags[tag] {
			seenTags[tag] = true
			allTags = append(allTags, tag)
		}
	}

	// Add hashtags (without the # prefix)
	for _, tag := range hashtags {
		tag = tag[1:] // Remove the # prefix
		if !seenTags[tag] {
			seenTags[tag] = true
			allTags = append(allTags, tag)
		}
	}

	return &FileInfo{
		Frontmatter: frontmatter,
		Tags:        allTags,
	}, nil
}
