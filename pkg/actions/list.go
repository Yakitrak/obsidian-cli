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
	InputTypeFile     InputType = iota // File path input
	InputTypeTag                       // Tag search input
	InputTypeFind                      // Fuzzy find input
	InputTypeProperty                  // Property key:value input
)

// ListInput represents a single input for listing files
type ListInput struct {
	Type     InputType // Type of the input
	Value    string    // Value of the input
	Property string    // Property name for InputTypeProperty
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
	Expression               *InputExpression                // Optional boolean expression to evaluate
	OnMatch                  func(string)                    // Callback function to report matches as they're found
	IncludeBacklinks         bool                            // Whether to collect first-degree backlinks for matched files
	Backlinks                *map[string][]obsidian.Backlink // Optional output map for backlinks keyed by normalized target path
	PrimaryMatches           *[]string                       // Optional output slice capturing the matches before link following
	obsidian.WikilinkOptions                                 // options influencing backlink parsing
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
	expr := params.Expression
	if expr == nil {
		expr = buildOrExpression(params.Inputs)
	}

	var matches []string
	if expr != nil {
		matches = evaluateExpressionMatches(allNotes, vaultPath, note, expr)
	}
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

func buildOrExpression(inputs []ListInput) *InputExpression {
	if len(inputs) == 0 {
		return nil
	}
	var expr *InputExpression
	for i := len(inputs) - 1; i >= 0; i-- {
		current := &InputExpression{
			Type:  exprLeaf,
			Input: &inputs[i],
		}
		if expr == nil {
			expr = current
			continue
		}
		expr = &InputExpression{
			Type:  exprOr,
			Left:  current,
			Right: expr,
		}
	}
	return expr
}

func determineWorkerCount(noteCount int) int {
	numWorkers := runtime.NumCPU()
	if noteCount < numWorkers {
		numWorkers = noteCount
	}
	return numWorkers
}

func evaluateExpressionMatches(allNotes []string, vaultPath string, note obsidian.NoteManager, expr *InputExpression) []string {
	if expr == nil {
		return nil
	}

	if len(allNotes) == 0 {
		return nil
	}

	results := make(chan string, len(allNotes))
	numWorkers := determineWorkerCount(len(allNotes))
	if numWorkers == 0 {
		return nil
	}
	batchSize := (len(allNotes) + numWorkers - 1) / numWorkers

	var wg sync.WaitGroup
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
			for _, path := range files {
				ctx := &noteContext{
					path:      path,
					vaultPath: vaultPath,
					note:      note,
				}
				if evaluateExpression(expr, ctx) {
					results <- path
				}
			}
		}(allNotes[start:end])
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	var matches []string
	for path := range results {
		matches = append(matches, path)
	}
	return matches
}

func evaluateExpression(expr *InputExpression, ctx *noteContext) bool {
	if expr == nil {
		return false
	}
	switch expr.Type {
	case exprLeaf:
		return ctx.matchesInput(expr.Input)
	case exprAnd:
		return evaluateExpression(expr.Left, ctx) && evaluateExpression(expr.Right, ctx)
	case exprOr:
		return evaluateExpression(expr.Left, ctx) || evaluateExpression(expr.Right, ctx)
	case exprNot:
		return !evaluateExpression(expr.Left, ctx)
	default:
		return false
	}
}

type noteContext struct {
	path      string
	vaultPath string
	note      obsidian.NoteManager

	contentOnce sync.Once
	content     string
	contentErr  error

	frontmatterOnce sync.Once
	frontmatter     map[string]interface{}

	inlineOnce     sync.Once
	inlineProps    map[string][]string
	inlinePropsErr error
}

func (c *noteContext) matchesInput(input *ListInput) bool {
	if input == nil {
		return false
	}
	switch input.Type {
	case InputTypeFile:
		return matchFilePath(c.path, input.Value)
	case InputTypeFind:
		return obsidian.FuzzyMatch(input.Value, c.path)
	case InputTypeTag:
		content, err := c.getContent()
		if err != nil {
			return false
		}
		return obsidian.HasAnyTags(content, []string{input.Value})
	case InputTypeProperty:
		return c.propertyHasValue(input.Property, input.Value)
	default:
		return false
	}
}

func matchFilePath(notePath, input string) bool {
	normalizedInputPath := obsidian.NormalizePath(input)

	if normalizedInputPath == "*" {
		return true
	}

	normalizedNotePath := obsidian.NormalizePath(notePath)
	dirPrefix := normalizedInputPath + "/"
	return normalizedNotePath == normalizedInputPath || strings.HasPrefix(normalizedNotePath, dirPrefix)
}

func (c *noteContext) getContent() (string, error) {
	c.contentOnce.Do(func() {
		content, err := c.note.GetContents(c.vaultPath, c.path)
		if err != nil {
			c.contentErr = err
			return
		}
		c.content = content
	})
	return c.content, c.contentErr
}

func (c *noteContext) getFrontmatter() map[string]interface{} {
	c.frontmatterOnce.Do(func() {
		content, err := c.getContent()
		if err != nil {
			return
		}
		fm, _ := obsidian.ExtractFrontmatter(content)
		c.frontmatter = fm
	})
	return c.frontmatter
}

func (c *noteContext) getInlineProperties() map[string][]string {
	c.inlineOnce.Do(func() {
		content, err := c.getContent()
		if err != nil {
			c.inlinePropsErr = err
			return
		}
		c.inlineProps = obsidian.ExtractInlineProperties(content)
	})
	return c.inlineProps
}

func (c *noteContext) propertyHasValue(property string, target string) bool {
	if strings.TrimSpace(property) == "" || strings.TrimSpace(target) == "" {
		return false
	}

	content, err := c.getContent()
	if err != nil {
		return false
	}
	return propertyHasValue(content, property, target)
}

// propertyHasValue checks if a note content contains the given property with the target value.
func propertyHasValue(content string, property string, target string) bool {
	if strings.TrimSpace(property) == "" || strings.TrimSpace(target) == "" {
		return false
	}

	targetNorm := normalizePropertyValue(target)
	propKey := strings.ToLower(strings.TrimSpace(property))

	checkValues := func(vals []string) bool {
		for _, v := range vals {
			if normalizePropertyValue(v) == targetNorm {
				return true
			}
		}
		return false
	}

	frontmatter, _ := obsidian.ExtractFrontmatter(content)
	if frontmatter != nil {
		for k, v := range frontmatter {
			if strings.ToLower(strings.TrimSpace(k)) != propKey {
				continue
			}
			info := obsidian.AnalyzePropertyValue(v)
			if checkValues(info.Values) {
				return true
			}
		}
	}

	inline := obsidian.ExtractInlineProperties(content)
	for k, vals := range inline {
		if strings.ToLower(strings.TrimSpace(k)) != propKey {
			continue
		}
		if checkValues(vals) {
			return true
		}
	}

	return false
}

// normalizePropertyValue lowercases, trims, and removes wikilink brackets for comparison.
func normalizePropertyValue(v string) string {
	val := strings.ToLower(strings.TrimSpace(v))
	if strings.HasPrefix(val, "[[") && strings.HasSuffix(val, "]]") {
		val = strings.TrimSuffix(strings.TrimPrefix(val, "[["), "]]")
	}
	if strings.Contains(val, "|") {
		if parts := strings.SplitN(val, "|", 2); len(parts) > 0 {
			val = strings.TrimSpace(parts[0])
		}
	}
	return val
}
