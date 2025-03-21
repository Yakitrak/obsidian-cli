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
	Inputs        []ListInput  // List of inputs to process
	FollowLinks   bool         // Whether to follow wiki links
	MaxDepth      int          // Maximum depth for following links
	AbsolutePaths bool         // Whether to return absolute paths
	OnMatch       func(string) // Callback function to report matches as they're found
}

// Debug controls whether debug output is printed
var Debug bool

// debugf prints debug output if Debug is true
func debugf(format string, args ...interface{}) {
	if Debug {
		fmt.Fprintf(os.Stderr, format, args...)
	}
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

	// If no inputs specified and not following links, return all files
	if len(params.Inputs) == 0 && !params.FollowLinks {
		return allNotes, nil
	}

	// Process all inputs to get matching files
	matches := processInputs(allNotes, vaultPath, note, params)

	// If following links, get all connected files
	if params.FollowLinks {
		linkedFiles := followMatchedFiles(matches, vaultPath, note, params.MaxDepth)
		debugf("Found %d total files after following links\n", len(linkedFiles))
		// Call OnMatch for each linked file
		for _, file := range linkedFiles {
			if params.OnMatch != nil {
				params.OnMatch(file)
			}
		}
		return linkedFiles, nil
	}

	return matches, nil
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
	var matches []string
	normalizedInputPath := obsidian.NormalizePath(input.Value)
	dirPrefix := normalizedInputPath + "/"

	for _, notePath := range allNotes {
		normalizedNotePath := obsidian.NormalizePath(notePath)
		if normalizedNotePath == normalizedInputPath || strings.HasPrefix(normalizedNotePath, dirPrefix) {
			matches = append(matches, notePath)
			if params.OnMatch != nil {
				params.OnMatch(notePath)
			}
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

	// Collect results
	for notePath := range results {
		mu.Lock()
		matches = append(matches, notePath)
		mu.Unlock()
		if params.OnMatch != nil {
			params.OnMatch(notePath)
		}
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

	// Collect results
	for notePath := range results {
		mu.Lock()
		matches = append(matches, notePath)
		mu.Unlock()
		if params.OnMatch != nil {
			params.OnMatch(notePath)
		}
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
func followMatchedFiles(matches []string, vaultPath string, note obsidian.NoteManager, maxDepth int) []string {
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

	for _, notePath := range matches {
		debugf("Following links for note: %s\n", notePath)
		files, err := obsidian.FollowWikilinks(vaultPath, note, notePath, maxDepth, visited, cache)
		if err != nil {
			debugf("Error following links for %s: %v\n", notePath, err)
			continue
		}
		debugf("Found %d linked files for %s\n", len(files), notePath)
		result = append(result, files...)
	}

	return obsidian.DeduplicateResults(result)
}