package actions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

type MoveParams struct {
	// Backwards compatibility fields; if Moves is empty these are used.
	CurrentNoteName string
	NewNoteName     string

	// Preferred: explicit list of moves.
	Moves []MoveRequest

	Overwrite       bool
	UpdateBacklinks bool
	IgnoredPaths    []string
	ShouldOpen      bool
}

// MoveRequest represents a single move/rename operation.
type MoveRequest struct {
	Source string
	Target string
}

// MoveResult captures the outcome for a single move.
type MoveResult struct {
	Source              string
	Target              string
	LinkUpdates         int
	GitHistoryPreserved bool
}

// MoveSummary aggregates results for a batch of moves.
type MoveSummary struct {
	Results          []MoveResult
	TotalLinkUpdates int
	Skipped          []string
}

// MoveNote preserves the CLI entrypoint signature; it delegates to MoveNotes with a single move.
func MoveNote(vault obsidian.VaultManager, note obsidian.NoteManager, uri obsidian.UriManager, params MoveParams) error {
	_ = note // maintained for backwards compatibility with callers/mocks

	if len(params.Moves) == 0 && params.CurrentNoteName != "" && params.NewNoteName != "" {
		params.Moves = []MoveRequest{
			{Source: params.CurrentNoteName, Target: params.NewNoteName},
		}
	}
	_, err := MoveNotes(vault, uri, params)
	return err
}

// MoveNotes performs one or more move operations, defaulting to no backlink rewrites (Obsidian links by note name).
// When UpdateBacklinks is true, all rewrites are done in a single vault scan for efficiency.
func MoveNotes(vault obsidian.VaultManager, uri obsidian.UriManager, params MoveParams) (MoveSummary, error) {
	var summary MoveSummary

	if len(params.Moves) == 0 {
		return summary, errors.New("at least one move is required")
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return summary, err
	}

	// Normalize and validate move set
	mappings := make([]linkMapping, 0, len(params.Moves))
	seenTargets := make(map[string]string)
	for _, mv := range params.Moves {
		src := strings.TrimSpace(mv.Source)
		dst := strings.TrimSpace(mv.Target)
		if src == "" || dst == "" {
			return summary, fmt.Errorf("invalid move (source and target required): %+v", mv)
		}

		srcNorm := obsidian.NormalizeWithDefaultExt(src, ".md")
		dstNorm := obsidian.NormalizeWithDefaultExt(dst, ".md")

		if srcNorm == dstNorm {
			return summary, fmt.Errorf("source and target are the same: %s", srcNorm)
		}

		if existingSrc, exists := seenTargets[dstNorm]; exists {
			return summary, fmt.Errorf("duplicate target %s for sources %s and %s", dstNorm, existingSrc, srcNorm)
		}
		seenTargets[dstNorm] = srcNorm

		mappings = append(mappings, linkMapping{Old: srcNorm, New: dstNorm})
	}

	// Perform filesystem/git moves first (no backlink rewrites yet)
	for _, mv := range mappings {
		res, err := RenameNote(vault, RenameParams{
			Source:          mv.Old,
			Target:          mv.New,
			Overwrite:       params.Overwrite,
			UpdateBacklinks: false,
			IgnoredPaths:    params.IgnoredPaths,
		})
		if err != nil {
			return summary, err
		}

		summary.Results = append(summary.Results, MoveResult{
			Source:              mv.Old,
			Target:              res.RenamedPath,
			GitHistoryPreserved: res.GitHistoryPreserved,
		})
	}

	// Optional backlink rewrite in a single pass
	if params.UpdateBacklinks {
		total, perTarget, skipped, err := rewriteVaultLinksBatch(vaultPath, mappings, params.IgnoredPaths)
		if err != nil {
			return summary, err
		}
		summary.TotalLinkUpdates = total
		summary.Skipped = skipped
		for i, res := range summary.Results {
			normalizedTarget := obsidian.NormalizePath(res.Target)
			summary.Results[i].LinkUpdates = perTarget[normalizedTarget]
		}
	}

	// Open the moved note if requested and only one target is present
	if params.ShouldOpen && len(summary.Results) == 1 {
		vaultName, err := vault.DefaultName()
		if err != nil {
			return summary, err
		}
		target := summary.Results[0].Target
		obsidianURI := uri.Construct(ObsOpenUrl, map[string]string{
			"file":  target,
			"vault": vaultName,
		})
		if err := uri.Execute(obsidianURI); err != nil {
			return summary, err
		}
	}

	return summary, nil
}

type linkMapping struct {
	Old            string
	New            string
	BasenameUnique bool // true if no other note in the vault shares this basename
}

// rewriteVaultLinksBatch rewrites backlinks for multiple moves in a single vault walk.
func rewriteVaultLinksBatch(vaultPath string, mappings []linkMapping, ignored []string) (int, map[string]int, []string, error) {
	totalUpdates := 0
	perTarget := make(map[string]int)
	var skipped []string

	// First pass: collect all basenames in the vault to detect duplicates
	// Use case-insensitive keys on macOS/Windows to match filesystem behavior
	basenameCount := make(map[string]int)
	_ = filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(d.Name()) != ".md" {
			return nil
		}
		if obsidian.ShouldIgnorePath(vaultPath, path, ignored) {
			return nil
		}
		basename := strings.TrimSuffix(d.Name(), ".md")
		basenameCount[obsidian.NormalizeForComparison(basename)]++
		return nil
	})

	// Mark which mappings have unique basenames
	for _, m := range mappings {
		oldExt := strings.ToLower(filepath.Ext(m.Old))
		if oldExt != ".md" {
			continue
		}
		oldBase := filepath.Base(strings.TrimSuffix(m.Old, oldExt))
		newBase := filepath.Base(strings.TrimSuffix(m.New, strings.ToLower(filepath.Ext(m.New))))
		if obsidian.NormalizeForComparison(oldBase) != obsidian.NormalizeForComparison(newBase) &&
			!obsidian.ShouldIgnorePath(vaultPath, filepath.Join(vaultPath, m.Old), ignored) {
			basenameCount[obsidian.NormalizeForComparison(oldBase)]++
		}
	}

	for i := range mappings {
		oldExt := strings.ToLower(filepath.Ext(mappings[i].Old))
		if oldExt != ".md" {
			mappings[i].BasenameUnique = true
			continue
		}
		oldBasename := filepath.Base(strings.TrimSuffix(mappings[i].Old, oldExt))
		mappings[i].BasenameUnique = basenameCount[obsidian.NormalizeForComparison(oldBasename)] <= 1
	}

	err := filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if d.IsDir() {
			if obsidian.ShouldIgnorePath(vaultPath, path, ignored) {
				return filepath.SkipDir
			}
			return nil
		}
		if obsidian.ShouldIgnorePath(vaultPath, path, ignored) {
			rel, _ := filepath.Rel(vaultPath, path)
			skipped = append(skipped, obsidian.NormalizePath(rel))
			return nil
		}
		if filepath.Ext(d.Name()) != ".md" {
			return nil
		}

		contentBytes, readErr := os.ReadFile(path)
		if readErr != nil {
			return readErr
		}

		content := string(contentBytes)
		fileUpdates := 0
		rewritten := content
		for _, m := range mappings {
			next, count := obsidian.RewriteLinksInContentWithOptions(rewritten, m.Old, m.New, m.BasenameUnique)
			if count > 0 {
				fileUpdates += count
				perTarget[obsidian.NormalizePath(m.New)] += count
				rewritten = next
			}
		}

		if fileUpdates == 0 {
			return nil
		}

		info, statErr := d.Info()
		if statErr != nil {
			return statErr
		}
		if writeErr := os.WriteFile(path, []byte(rewritten), info.Mode()); writeErr != nil {
			return writeErr
		}
		totalUpdates += fileUpdates
		return nil
	})

	if err != nil {
		return 0, perTarget, skipped, err
	}
	return totalUpdates, perTarget, skipped, nil
}
