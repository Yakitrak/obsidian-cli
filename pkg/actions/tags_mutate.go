package actions

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

// TagMutationSummary represents the result of a tag mutation operation
type TagMutationSummary struct {
	NotesTouched int            `json:"notesTouched"`
	TagChanges   map[string]int `json:"tagChanges"` // tag -> number of notes where this tag was changed
	FilesChanged []string       `json:"filesChanged,omitempty"`
}

// delta represents the result of processing a single file
type delta struct {
	notesTouched bool
	tagChanges   map[string]int // tag → 1 if changed in this file
	fileChanged  string         // empty if unchanged
	err          error          // non-nil on any failure
}

// parallelProcessor defines the interface for processing files in parallel
type parallelProcessor func(ctx context.Context, vaultPath, notePath string) delta

// runInParallel processes files using a worker pool and aggregates results
func runInParallel(ctx context.Context, cancel context.CancelFunc, vaultPath string, allNotes []string, processor parallelProcessor) (TagMutationSummary, error) {
	return runInParallelWithWorkers(ctx, cancel, vaultPath, allNotes, processor, runtime.NumCPU())
}

// runInParallelWithWorkers processes files using a worker pool with specified worker count
func runInParallelWithWorkers(ctx context.Context, cancel context.CancelFunc, vaultPath string, allNotes []string, processor parallelProcessor, workerCount int) (TagMutationSummary, error) {
	if workerCount < 1 {
		workerCount = 1
	}

	jobs := make(chan string, len(allNotes))
	results := make(chan delta, len(allNotes))

	// Start workers
	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for notePath := range jobs {
				select {
				case <-ctx.Done():
					return
				case results <- processor(ctx, vaultPath, notePath):
				}
			}
		}()
	}

	// Send jobs
	go func() {
		defer close(jobs)
		for _, notePath := range allNotes {
			select {
			case <-ctx.Done():
				return
			case jobs <- notePath:
			}
		}
	}()

	// Close results when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()

	// Aggregate results
	summary := TagMutationSummary{
		TagChanges:   make(map[string]int),
		FilesChanged: make([]string, 0),
	}

	var firstError error
	var firstErrOnce sync.Once

	for delta := range results {
		if delta.err != nil {
			firstErrOnce.Do(func() {
				firstError = delta.err
				// Cancel context to stop other workers
				cancel()
			})
			continue
		}

		if delta.notesTouched {
			summary.NotesTouched++
		}

		for tag, count := range delta.tagChanges {
			summary.TagChanges[tag] += count
		}

		if delta.fileChanged != "" {
			summary.FilesChanged = append(summary.FilesChanged, delta.fileChanged)
		}
	}

	return summary, firstError
}

// processDeleteFile handles deletion of tags from a single file
func processDeleteFile(tagsToDelete []string, dryRun bool) parallelProcessor {
	return func(ctx context.Context, vaultPath, notePath string) delta {
		select {
		case <-ctx.Done():
			return delta{err: ctx.Err()}
		default:
		}

		full := filepath.Join(vaultPath, notePath)
		data, err := os.ReadFile(full)
		if err != nil {
			// Skip files we can't read, don't treat as fatal error
			return delta{}
		}
		content := string(data)

		newContent, changed := obsidian.RemoveTags(content, tagsToDelete)
		if !changed {
			return delta{}
		}

		result := delta{
			notesTouched: true,
			tagChanges:   make(map[string]int),
			fileChanged:  notePath,
		}

		// Track which tags were actually changed in this file
		for _, tag := range tagsToDelete {
			if hasTag(content, tag) {
				result.tagChanges[tag] = 1
			}
		}

		// Write the file if not dry run
		if !dryRun {
			err = obsidian.WriteFileAtomic(full, []byte(newContent), 0644)
			if err != nil {
				result.err = fmt.Errorf("failed to write file %s: %w", notePath, err)
				return result
			}
		}

		return result
	}
}

// processRenameFile handles renaming of tags in a single file
func processRenameFile(fromTags []string, toTag string, dryRun bool) parallelProcessor {
	return func(ctx context.Context, vaultPath, notePath string) delta {
		select {
		case <-ctx.Done():
			return delta{err: ctx.Err()}
		default:
		}

		full := filepath.Join(vaultPath, notePath)
		data, err := os.ReadFile(full)
		if err != nil {
			// Skip files we can't read, don't treat as fatal error
			return delta{}
		}
		content := string(data)

		newContent, changed := obsidian.ReplaceTags(content, fromTags, toTag)
		if !changed {
			return delta{}
		}

		result := delta{
			notesTouched: true,
			tagChanges:   make(map[string]int),
			fileChanged:  notePath,
		}

		// Track which tags were actually changed in this file
		for _, tag := range fromTags {
			if hasTag(content, tag) {
				result.tagChanges[tag] = 1
			}
		}

		// Write the file if not dry run
		if !dryRun {
			err = obsidian.WriteFileAtomic(full, []byte(newContent), 0644)
			if err != nil {
				result.err = fmt.Errorf("failed to write file %s: %w", notePath, err)
				return result
			}
		}

		return result
	}
}

// DeleteTags removes specified tags from all notes in the vault
func DeleteTags(vault obsidian.VaultManager, note obsidian.NoteManager, tagsToDelete []string, dryRun bool) (TagMutationSummary, error) {
	return DeleteTagsWithWorkers(vault, note, tagsToDelete, dryRun, runtime.NumCPU())
}

// DeleteTagsWithWorkers removes specified tags from all notes in the vault using specified worker count
func DeleteTagsWithWorkers(vault obsidian.VaultManager, note obsidian.NoteManager, tagsToDelete []string, dryRun bool, workers int) (TagMutationSummary, error) {
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

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return runInParallelWithWorkers(ctx, cancel, vaultPath, allNotes, processDeleteFile(tagsToDelete, dryRun), workers)
}

// RenameTags replaces specified tags with a new tag in all notes in the vault
func RenameTags(vault obsidian.VaultManager, note obsidian.NoteManager, fromTags []string, toTag string, dryRun bool) (TagMutationSummary, error) {
	return RenameTagsWithWorkers(vault, note, fromTags, toTag, dryRun, runtime.NumCPU())
}

// RenameTagsWithWorkers replaces specified tags with a new tag in all notes in the vault using specified worker count
func RenameTagsWithWorkers(vault obsidian.VaultManager, note obsidian.NoteManager, fromTags []string, toTag string, dryRun bool, workers int) (TagMutationSummary, error) {
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

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return runInParallelWithWorkers(ctx, cancel, vaultPath, allNotes, processRenameFile(fromTags, toTag, dryRun), workers)
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

// processAddFile handles addition of tags to a single file
func processAddFile(tagsToAdd []string, dryRun bool) parallelProcessor {
	return func(ctx context.Context, vaultPath, notePath string) delta {
		select {
		case <-ctx.Done():
			return delta{err: ctx.Err()}
		default:
		}

		full := filepath.Join(vaultPath, notePath)
		data, err := os.ReadFile(full)
		if err != nil {
			// Skip files we can't read, don't treat as fatal error
			return delta{}
		}
		content := string(data)

		newContent, changed := obsidian.AddTags(content, tagsToAdd)
		if !changed {
			return delta{}
		}

		result := delta{
			notesTouched: true,
			tagChanges:   make(map[string]int),
			fileChanged:  notePath,
		}

		// Track which tags were actually added to this file
		for _, tag := range tagsToAdd {
			normalizedTag := normalizeTagForComparison(tag)
			if !hasTag(content, tag) { // Tag wasn't present before
				result.tagChanges[normalizedTag] = 1
			}
		}

		// Write the file if not dry run
		if !dryRun {
			err = obsidian.WriteFileAtomic(full, []byte(newContent), 0644)
			if err != nil {
				result.err = fmt.Errorf("failed to write file %s: %w", notePath, err)
				return result
			}
		}

		return result
	}
}

// AddTags adds specified tags to all notes in the vault
func AddTags(vault obsidian.VaultManager, note obsidian.NoteManager, tagsToAdd []string, dryRun bool) (TagMutationSummary, error) {
	return AddTagsWithWorkers(vault, note, tagsToAdd, dryRun, runtime.NumCPU())
}

// AddTagsWithWorkers adds specified tags to all notes in the vault using specified worker count
func AddTagsWithWorkers(vault obsidian.VaultManager, note obsidian.NoteManager, tagsToAdd []string, dryRun bool, workers int) (TagMutationSummary, error) {
	if len(tagsToAdd) == 0 {
		return TagMutationSummary{}, fmt.Errorf("no tags specified for addition")
	}

	// Validate tags
	for _, tag := range tagsToAdd {
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

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return runInParallelWithWorkers(ctx, cancel, vaultPath, allNotes, processAddFile(tagsToAdd, dryRun), workers)
}

// AddTagsToFiles adds specified tags to a specific list of files
func AddTagsToFiles(vault obsidian.VaultManager, note obsidian.NoteManager, tagsToAdd []string, files []string, dryRun bool) (TagMutationSummary, error) {
	return AddTagsToFilesWithWorkers(vault, note, tagsToAdd, files, dryRun, runtime.NumCPU())
}

// AddTagsToFilesWithWorkers adds specified tags to a specific list of files using specified worker count
func AddTagsToFilesWithWorkers(vault obsidian.VaultManager, note obsidian.NoteManager, tagsToAdd []string, files []string, dryRun bool, workers int) (TagMutationSummary, error) {
	if len(tagsToAdd) == 0 {
		return TagMutationSummary{}, fmt.Errorf("no tags specified for addition")
	}

	if len(files) == 0 {
		return TagMutationSummary{}, fmt.Errorf("no files specified")
	}

	// Validate tags
	for _, tag := range tagsToAdd {
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

	// Create context with cancellation
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	return runInParallelWithWorkers(ctx, cancel, vaultPath, files, processAddFile(tagsToAdd, dryRun), workers)
}
