package actions

import (
	"fmt"
	"os"
	"runtime"
	"strings"
	"sync"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

// InputType represents the type of input for listing files
type InputType int

const (
	InputTypeFile InputType = iota // File path input
	InputTypeTag                   // Tag search input
	InputTypeFind                  // Fuzzy find input
)

// ListInput represents a single input for listing files
type ListInput struct {
	Type  InputType // Type of the input
	Value string    // Value of the input
}

// ListParams represents parameters for listing files
type ListParams struct {
	Inputs                   []ListInput                     // List of inputs to process
	FollowLinks              bool                            // Whether to follow wiki links
	MaxDepth                 int                             // Maximum depth for following links
	SkipAnchors              bool                            // Whether to skip wikilinks with anchors (e.g. [[Note#Section]])
	SkipEmbeds               bool                            // Whether to skip embedded wikilinks (e.g. ![[Embedded Note]])
	AbsolutePaths            bool                            // Whether to return absolute paths
	SuppressedTags           []string                        // Tags to exclude from results
	OnMatch                  func(string)                    // Callback function to report matches as they're found
	IncludeBacklinks         bool                            // Whether to collect first-degree backlinks for matched files
	Backlinks                *map[string][]obsidian.Backlink // Optional output map for backlinks keyed by normalized target path
	PrimaryMatches           *[]string                       // Optional output slice capturing the matches before link following
	obsidian.WikilinkOptions                                 // options influencing backlink parsing
}

// ParseInputs parses command-line arguments into ListInput objects.
// Returns an error if tag: or find: inputs have empty or wildcard values.
func ParseInputs(args []string) ([]ListInput, error) {
	var inputs []ListInput
	for _, arg := range args {
		if strings.HasPrefix(arg, "tag:") {
			// Handle quoted tags
			tag := strings.TrimPrefix(arg, "tag:")
			if strings.HasPrefix(tag, "\"") && strings.HasSuffix(tag, "\"") {
				tag = strings.Trim(tag, "\"")
			}

			// Validate tag value
			if tag == "" || tag == "*" {
				return nil, fmt.Errorf("invalid tag value in %q: tag cannot be empty or a wildcard (*)", arg)
			}

			inputs = append(inputs, ListInput{
				Type:  InputTypeTag,
				Value: tag,
			})
		} else if strings.HasPrefix(arg, "find:") {
			// Handle find input
			searchTerm := strings.TrimPrefix(arg, "find:")
			if strings.HasPrefix(searchTerm, "\"") && strings.HasSuffix(searchTerm, "\"") {
				searchTerm = strings.Trim(searchTerm, "\"")
			}

			// Validate find value
			if searchTerm == "" || searchTerm == "*" {
				return nil, fmt.Errorf("invalid find value in %q: find cannot be empty or a wildcard (*)", arg)
			}

			inputs = append(inputs, ListInput{
				Type:  InputTypeFind,
				Value: searchTerm,
			})
		} else {
			// Handle file paths
			inputs = append(inputs, ListInput{
				Type:  InputTypeFile,
				Value: arg,
			})
		}
	}

	return inputs, nil
}

// Debug controls whether debug output is printed
var Debug bool

// debugf prints debug output if Debug is true
func debugf(format string, args ...interface{}) {
	if Debug {
		fmt.Fprintf(os.Stderr, format, args...)
	}
}

// hasAnySuppressedTags checks if a file contains any of the suppressed tags
func hasAnySuppressedTags(content string, suppressedTags []string) bool {
	if len(suppressedTags) == 0 {
		return false
	}

	// Extract all tags from the file
	allTags := extractAllTags(content)

	// Normalize suppressed tags for comparison
	normalizedSuppressed := make(map[string]bool)
	for _, tag := range suppressedTags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalized != "" {
			normalizedSuppressed[normalized] = true
		}
	}

	// Check if any tag in the file matches a suppressed tag
	for _, tag := range allTags {
		normalized := strings.ToLower(strings.TrimSpace(tag))
		if normalizedSuppressed[normalized] {
			return true
		}
	}

	return false
}

// filterSuppressedFiles removes files that contain suppressed tags
func filterSuppressedFiles(files []string, vaultPath string, note obsidian.NoteManager, suppressedTags []string) []string {
	if len(suppressedTags) == 0 {
		return files
	}

	var filtered []string
	for _, file := range files {
		content, err := note.GetContents(vaultPath, file)
		if err != nil {
			debugf("Error reading file %s for suppression check: %v\n", file, err)
			// Include the file if we can't read it (safer default)
			filtered = append(filtered, file)
			continue
		}

		if !hasAnySuppressedTags(content, suppressedTags) {
			filtered = append(filtered, file)
		} else {
			debugf("Suppressing file %s due to suppressed tags\n", file)
		}
	}

	return filtered
}

// filterSuppressedBacklinks removes backlink referrers whose contents include suppressed tags.
func filterSuppressedBacklinks(backlinks map[string][]obsidian.Backlink, vaultPath string, note obsidian.NoteManager, suppressedTags []string) map[string][]obsidian.Backlink {
	if len(suppressedTags) == 0 {
		return backlinks
	}

	for target, refs := range backlinks {
		var kept []obsidian.Backlink
		for _, bl := range refs {
			content, err := note.GetContents(vaultPath, bl.Referrer)
			if err != nil {
				// Skip unreadable referrers to avoid leaking suppressed content
				debugf("Skipping backlink referrer %s due to read error: %v\n", bl.Referrer, err)
				continue
			}
			if hasAnySuppressedTags(content, suppressedTags) {
				debugf("Suppressing backlink referrer %s for target %s\n", bl.Referrer, target)
				continue
			}
			kept = append(kept, bl)
		}
		backlinks[target] = kept
	}

	return backlinks
}

// ListFiles is the main function that lists files based on the provided parameters
func ListFiles(vault obsidian.VaultManager, note obsidian.NoteManager, params ListParams) ([]string, error) {
	vaultPath, err := vault.Path()
	if err != nil {
		return nil, err
	}

	allNotes, err := note.GetNotesList(vaultPath)
	if err != nil {
		return nil, err
	}

	// Process all inputs to get matching files
	matches := processInputs(allNotes, vaultPath, note, params)
	debugf("Found %d initial matching files\n", len(matches))

	// Filter out suppressed files
	matches = filterSuppressedFiles(matches, vaultPath, note, params.SuppressedTags)
	debugf("After suppression filtering: %d files\n", len(matches))

	if params.PrimaryMatches != nil {
		copied := make([]string, len(matches))
		copy(copied, matches)
		*params.PrimaryMatches = copied
	}

	// If following links, get all connected files
	if params.FollowLinks {
		linkedFiles := followMatchedFiles(matches, vaultPath, note, params)
		debugf("Found %d total files after following links\n", len(linkedFiles))

		// Apply suppression filter to linked files as well
		linkedFiles = filterSuppressedFiles(linkedFiles, vaultPath, note, params.SuppressedTags)
		debugf("After suppression filtering linked files: %d files\n", len(linkedFiles))

		if params.IncludeBacklinks && params.Backlinks != nil {
			backlinks, err := obsidian.CollectBacklinks(vaultPath, note, matches, params.WikilinkOptions, params.SuppressedTags)
			if err != nil {
				return nil, err
			}
			*params.Backlinks = backlinks
		}

		// Call OnMatch for each linked file
		debugf("Notifying OnMatch with %d files\n", len(linkedFiles))
		notifyMatches(linkedFiles, params.OnMatch)
		return linkedFiles, nil
	}

	if params.IncludeBacklinks && params.Backlinks != nil {
		backlinks, err := obsidian.CollectBacklinks(vaultPath, note, matches, params.WikilinkOptions, params.SuppressedTags)
		if err != nil {
			return nil, err
		}
		*params.Backlinks = backlinks
	}

	// Call OnMatch for each matched file - this only happens when not following links
	debugf("Notifying OnMatch with %d files (no link following)\n", len(matches))
	notifyMatches(matches, params.OnMatch)
	return matches, nil
}

// notifyMatches calls the OnMatch callback for each file if a callback is provided
func notifyMatches(files []string, onMatch func(string)) {
	if onMatch == nil {
		return
	}
	for _, file := range files {
		onMatch(file)
	}
}

// processInputs processes all inputs and returns matching files
func processInputs(allNotes []string, vaultPath string, note obsidian.NoteManager, params ListParams) []string {
	seen := make(map[string]bool)
	var matches []string

	// Group inputs by type for efficient processing
	tagInputs, otherInputs := groupInputsByType(params.Inputs)

	// Process tag inputs together if there are any
	if len(tagInputs) > 0 {
		tagMatches := processTagInput(tagInputs, allNotes, vaultPath, note, params)
		addUniqueMatches(&matches, tagMatches, seen)
	}

	// Process other inputs individually
	for _, input := range otherInputs {
		var inputMatches []string
		switch input.Type {
		case InputTypeFile:
			inputMatches = processFilePathInput(input, allNotes, params)
		case InputTypeFind:
			inputMatches = processFuzzyFindInput(input, allNotes, params)
		}
		addUniqueMatches(&matches, inputMatches, seen)
	}

	return matches
}

// groupInputsByType separates tag inputs from other inputs
func groupInputsByType(inputs []ListInput) (tagInputs, otherInputs []ListInput) {
	for _, input := range inputs {
		if input.Type == InputTypeTag {
			tagInputs = append(tagInputs, input)
		} else {
			otherInputs = append(otherInputs, input)
		}
	}
	return
}

// addUniqueMatches adds matches to the result slice while avoiding duplicates
func addUniqueMatches(matches *[]string, newMatches []string, seen map[string]bool) {
	for _, match := range newMatches {
		if !seen[match] {
			seen[match] = true
			*matches = append(*matches, match)
		}
	}
}

// processFilePathInput processes a single file path input
func processFilePathInput(input ListInput, allNotes []string, params ListParams) []string {
	normalizedInputPath := obsidian.NormalizePath(input.Value)

	// Handle wildcard pattern
	if normalizedInputPath == "*" {
		return allNotes
	}

	// Handle regular path matching
	var matches []string
	dirPrefix := normalizedInputPath + "/"
	for _, notePath := range allNotes {
		normalizedNotePath := obsidian.NormalizePath(notePath)
		if normalizedNotePath == normalizedInputPath || strings.HasPrefix(normalizedNotePath, dirPrefix) {
			matches = append(matches, notePath)
		}
	}
	return matches
}

// processTagInput processes multiple tag inputs concurrently
func processTagInput(inputs []ListInput, allNotes []string, vaultPath string, note obsidian.NoteManager, params ListParams) []string {
	var matches []string
	var mu sync.Mutex
	results := make(chan string, len(allNotes))

	tags := extractTagValues(inputs)
	numWorkers := determineWorkerCount(len(allNotes))
	processBatches(allNotes, numWorkers, vaultPath, note, tags, results)

	// Collect results without triggering OnMatch (it will be called later)
	for notePath := range results {
		mu.Lock()
		matches = append(matches, notePath)
		mu.Unlock()
		// Don't call OnMatch here - all matching will be handled by ListFiles
	}

	return matches
}

// extractTagValues extracts tag values from inputs
func extractTagValues(inputs []ListInput) []string {
	tags := make([]string, len(inputs))
	for i, input := range inputs {
		tags[i] = input.Value
	}
	return tags
}

// determineWorkerCount determines the optimal number of worker goroutines
func determineWorkerCount(noteCount int) int {
	numWorkers := runtime.NumCPU()
	if noteCount < numWorkers {
		numWorkers = noteCount
	}
	return numWorkers
}

// processBatches processes notes in batches using multiple goroutines
func processBatches(allNotes []string, numWorkers int, vaultPath string, note obsidian.NoteManager, tags []string, results chan<- string) {
	var wg sync.WaitGroup
	batchSize := (len(allNotes) + numWorkers - 1) / numWorkers

	for i := 0; i < numWorkers; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(allNotes) {
			end = len(allNotes)
		}
		if start >= len(allNotes) {
			continue
		}

		wg.Add(1)
		go func(files []string) {
			defer wg.Done()
			processTagBatch(files, vaultPath, note, tags, results)
		}(allNotes[start:end])
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()
}

// processTagBatch processes a batch of files for tag matching
func processTagBatch(files []string, vaultPath string, note obsidian.NoteManager, tags []string, results chan<- string) {
	for _, notePath := range files {
		content, err := note.GetContents(vaultPath, notePath)
		if err != nil {
			continue
		}
		if obsidian.HasAnyTags(content, tags) {
			results <- notePath
		}
	}
}

// processFuzzyFindInput processes a single fuzzy find input concurrently
func processFuzzyFindInput(input ListInput, allNotes []string, params ListParams) []string {
	var matches []string
	var mu sync.Mutex
	results := make(chan string, len(allNotes))

	numWorkers := determineWorkerCount(len(allNotes))
	processFuzzyBatches(allNotes, numWorkers, input.Value, results)

	// Collect results without triggering OnMatch (it will be called later)
	for notePath := range results {
		mu.Lock()
		matches = append(matches, notePath)
		mu.Unlock()
		// Don't call OnMatch here - all matching will be handled by ListFiles
	}

	return matches
}

// processFuzzyBatches processes notes in batches for fuzzy matching
func processFuzzyBatches(allNotes []string, numWorkers int, pattern string, results chan<- string) {
	var wg sync.WaitGroup
	batchSize := (len(allNotes) + numWorkers - 1) / numWorkers

	for i := 0; i < numWorkers; i++ {
		start := i * batchSize
		end := start + batchSize
		if end > len(allNotes) {
			end = len(allNotes)
		}
		if start >= len(allNotes) {
			continue
		}

		wg.Add(1)
		go func(files []string) {
			defer wg.Done()
			processFuzzyBatch(files, pattern, results)
		}(allNotes[start:end])
	}

	// Close results channel when all workers are done
	go func() {
		wg.Wait()
		close(results)
	}()
}

// processFuzzyBatch processes a batch of files for fuzzy matching
func processFuzzyBatch(files []string, pattern string, results chan<- string) {
	for _, notePath := range files {
		if obsidian.FuzzyMatch(pattern, notePath) {
			results <- notePath
		}
	}
}

// followMatchedFiles follows wikilinks for matched files
func followMatchedFiles(matches []string, vaultPath string, note obsidian.NoteManager, params ListParams) []string {
	// Get all notes first
	allNotes, err := note.GetNotesList(vaultPath)
	if err != nil {
		debugf("Error getting notes list: %v\n", err)
		return matches
	}

	debugf("Found %d total notes in vault\n", len(allNotes))

	// Build the note path cache
	cache := obsidian.BuildNotePathCache(allNotes)
	debugf("Built cache with %d entries\n", len(cache.Paths))

	visited := make(map[string]bool)
	var result []string

	// Create wikilinks options from parameters
	options := obsidian.CreateWikilinksOptions(params.MaxDepth, params.SkipAnchors, params.SkipEmbeds)

	for _, notePath := range matches {
		debugf("Following links for note: %s\n", notePath)
		files, err := obsidian.FollowWikilinks(vaultPath, note, notePath, visited, cache, options)
		if err != nil {
			debugf("Error following links for %s: %v\n", notePath, err)
			continue
		}
		debugf("Found %d linked files for %s\n", len(files), notePath)
		result = append(result, files...)
	}

	return obsidian.DeduplicateResults(result)
}
