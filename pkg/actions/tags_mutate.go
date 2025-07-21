package actions

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

// TagMutationSummary represents the result of a tag mutation operation
type TagMutationSummary struct {
	NotesTouched int            `json:"notesTouched"`
	TagChanges   map[string]int `json:"tagChanges"` // tag -> number of notes where this tag was changed
	FilesChanged []string       `json:"filesChanged,omitempty"`
}

// DeleteTags removes specified tags from all notes in the vault
func DeleteTags(vault obsidian.VaultManager, note obsidian.NoteManager, tagsToDelete []string, dryRun bool) (TagMutationSummary, error) {
	if len(tagsToDelete) == 0 {
		return TagMutationSummary{}, fmt.Errorf("no tags specified for deletion")
	}

	// Validate tags
	for _, tag := range tagsToDelete {
		if !isValidTagForOperation(tag) {
			return TagMutationSummary{}, fmt.Errorf("invalid tag: %s", tag)
		}
	}

	vaultPath, err := vault.Path()
	if err != nil {
		// If vault.Path() failed (e.g., during unit tests), attempt to use vault.Name directly if it is an existing directory.
		if v, ok := vault.(*obsidian.Vault); ok {
			if stat, statErr := os.Stat(v.Name); statErr == nil && stat.IsDir() {
				vaultPath = v.Name
			} else {
				return TagMutationSummary{}, fmt.Errorf("failed to get vault path: %w", err)
			}
		} else {
			return TagMutationSummary{}, fmt.Errorf("failed to get vault path: %w", err)
		}
	}

	allNotes, err := note.GetNotesList(vaultPath)
	if err != nil {
		return TagMutationSummary{}, fmt.Errorf("failed to get notes list: %w", err)
	}

	summary := TagMutationSummary{
		TagChanges:   make(map[string]int),
		FilesChanged: make([]string, 0),
	}

	for _, notePath := range allNotes {
		full := filepath.Join(vaultPath, notePath)
		data, err := os.ReadFile(full)
		if err != nil {
			continue // Skip notes we can't read
		}
		content := string(data)

		newContent, changed := obsidian.RemoveTags(content, tagsToDelete)
		if !changed {
			continue
		}

		summary.NotesTouched++
		summary.FilesChanged = append(summary.FilesChanged, notePath)

		// Track which tags were actually changed in this file
		for _, tag := range tagsToDelete {
			if hasTag(content, tag) {
				summary.TagChanges[tag]++
			}
		}

		// Write the file if not dry run
		if !dryRun {
			err = obsidian.WriteFileAtomic(full, []byte(newContent), 0644)
			if err != nil {
				return summary, fmt.Errorf("failed to write file %s: %w", notePath, err)
			}
		}
	}

	return summary, nil
}

// RenameTags replaces specified tags with a new tag in all notes in the vault
func RenameTags(vault obsidian.VaultManager, note obsidian.NoteManager, fromTags []string, toTag string, dryRun bool) (TagMutationSummary, error) {
	if len(fromTags) == 0 {
		return TagMutationSummary{}, fmt.Errorf("no source tags specified for rename")
	}

	if toTag == "" {
		return TagMutationSummary{}, fmt.Errorf("destination tag cannot be empty")
	}

	// Validate all tags
	for _, tag := range fromTags {
		if !isValidTagForOperation(tag) {
			return TagMutationSummary{}, fmt.Errorf("invalid source tag: %s", tag)
		}
	}

	if !isValidTagForOperation(toTag) {
		return TagMutationSummary{}, fmt.Errorf("invalid destination tag: %s", toTag)
	}

	// Check for circular rename (trying to rename to one of the source tags)
	normalizedTo := normalizeTagForComparison(toTag)
	for _, fromTag := range fromTags {
		if normalizeTagForComparison(fromTag) == normalizedTo {
			return TagMutationSummary{}, fmt.Errorf("cannot rename tag %s to itself", fromTag)
		}
	}

	vaultPath, err := vault.Path()
	if err != nil {
		if v, ok := vault.(*obsidian.Vault); ok {
			if stat, statErr := os.Stat(v.Name); statErr == nil && stat.IsDir() {
				vaultPath = v.Name
			} else {
				return TagMutationSummary{}, fmt.Errorf("failed to get vault path: %w", err)
			}
		} else {
			return TagMutationSummary{}, fmt.Errorf("failed to get vault path: %w", err)
		}
	}

	allNotes, err := note.GetNotesList(vaultPath)
	if err != nil {
		return TagMutationSummary{}, fmt.Errorf("failed to get notes list: %w", err)
	}

	summary := TagMutationSummary{
		TagChanges:   make(map[string]int),
		FilesChanged: make([]string, 0),
	}

	for _, notePath := range allNotes {
		full := filepath.Join(vaultPath, notePath)
		data, err := os.ReadFile(full)
		if err != nil {
			continue // Skip notes we can't read
		}
		content := string(data)

		newContent, changed := obsidian.ReplaceTags(content, fromTags, toTag)
		if !changed {
			continue
		}

		summary.NotesTouched++
		summary.FilesChanged = append(summary.FilesChanged, notePath)

		// Track which tags were actually changed in this file
		for _, tag := range fromTags {
			if hasTag(content, tag) {
				summary.TagChanges[tag]++
			}
		}

		// Write the file if not dry run
		if !dryRun {
			err = obsidian.WriteFileAtomic(full, []byte(newContent), 0644)
			if err != nil {
				return summary, fmt.Errorf("failed to write file %s: %w", notePath, err)
			}
		}
	}

	return summary, nil
}

// isValidTagForOperation checks if a tag is valid for mutation operations
func isValidTagForOperation(tag string) bool {
	if tag == "" {
		return false
	}

	cleanTag := normalizeTagForComparison(tag)
	return obsidian.IsValidTag(cleanTag)
}

// normalizeTagForComparison normalizes a tag for comparison (removes # prefix, trims, lowercases)
func normalizeTagForComparison(tag string) string {
	if strings.HasPrefix(tag, "#") {
		tag = tag[1:]
	}
	return obsidian.NormalizeTag(tag)
}

// hasTag checks if content contains a specific tag (case insensitive)
func hasTag(content, tag string) bool {
	normalizedTag := normalizeTagForComparison(tag)

	// Check frontmatter
	frontmatter, err := obsidian.ExtractFrontmatter(content)
	if err == nil && frontmatter != nil {
		if tags, ok := frontmatter["tags"]; ok {
			tagList := normalizeFrontmatterTags(tags)
			for _, fmTag := range tagList {
				if obsidian.NormalizeTag(fmTag) == normalizedTag {
					return true
				}
			}
		}
	}

	// Check hashtags
	hashtags := obsidian.ExtractHashtags(content)
	for _, hashtag := range hashtags {
		cleanHashtag := hashtag
		if cleanHashtag != "" && cleanHashtag[0] == '#' {
			cleanHashtag = cleanHashtag[1:]
		}
		if obsidian.NormalizeTag(cleanHashtag) == normalizedTag {
			return true
		}
	}

	return false
}

// normalizeFrontmatterTags normalizes tag values from various formats into a clean string slice
// This is a local copy of the normalizeTags function from obsidian package
func normalizeFrontmatterTags(tags interface{}) []string {
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
			result = append(result, normalizeFrontmatterTags(item)...)
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
