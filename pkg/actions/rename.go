package actions

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/atomicobject/obsidian-cli/pkg/obsidian"
)

type RenameParams struct {
	Source          string
	Target          string
	Overwrite       bool
	UpdateBacklinks bool
	IgnoredPaths    []string
}

type RenameResult struct {
	RenamedPath         string
	LinkUpdates         int
	Skipped             []string
	GitHistoryPreserved bool
}

func RenameNote(vault obsidian.VaultManager, params RenameParams) (RenameResult, error) {
	var result RenameResult
	if strings.TrimSpace(params.Source) == "" || strings.TrimSpace(params.Target) == "" {
		return result, errors.New("source and target note names are required")
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return result, err
	}

	sourceRel := obsidian.NormalizePath(obsidian.AddMdSuffix(params.Source))
	targetRel := obsidian.NormalizePath(obsidian.AddMdSuffix(params.Target))
	sourceAbs := filepath.Join(vaultPath, sourceRel)
	targetAbs := filepath.Join(vaultPath, targetRel)

	if _, err := os.Stat(sourceAbs); err != nil {
		return result, fmt.Errorf("source note does not exist: %w", err)
	}
	targetExists := false
	if _, err := os.Stat(targetAbs); err == nil {
		targetExists = true
	}
	if !params.Overwrite && targetExists {
		return result, fmt.Errorf("target note already exists: %s", targetRel)
	}
	if err := os.MkdirAll(filepath.Dir(targetAbs), 0o755); err != nil {
		return result, fmt.Errorf("unable to prepare target directory: %w", err)
	}

	isGit := isGitRepo(vaultPath)

	attemptGit := isGit
	if params.Overwrite && targetExists {
		if attemptGit {
			if err := gitRemove(vaultPath, targetRel); err != nil {
				attemptGit = false
				if err := os.RemoveAll(targetAbs); err != nil {
					return result, fmt.Errorf("failed to remove existing target: %w", err)
				}
			}
		} else {
			if err := os.RemoveAll(targetAbs); err != nil {
				return result, fmt.Errorf("failed to remove existing target: %w", err)
			}
		}
	}

	if attemptGit {
		if err := gitMove(vaultPath, sourceRel, targetRel); err == nil {
			result.GitHistoryPreserved = true
		} else {
			attemptGit = false
		}
	}

	if !result.GitHistoryPreserved {
		if err := os.Rename(sourceAbs, targetAbs); err != nil {
			return result, fmt.Errorf("filesystem rename failed: %w", err)
		}
		result.GitHistoryPreserved = false
	}

	linkUpdates := 0
	var skipped []string
	if params.UpdateBacklinks {
		updates, skippedFiles, err := rewriteVaultLinks(vaultPath, sourceRel, targetRel, params.IgnoredPaths)
		if err != nil {
			return result, err
		}
		linkUpdates = updates
		skipped = skippedFiles
	}

	result.RenamedPath = targetRel
	result.LinkUpdates = linkUpdates
	result.Skipped = skipped
	return result, nil
}

func isGitRepo(path string) bool {
	info, err := os.Stat(filepath.Join(path, ".git"))
	return err == nil && info.IsDir()
}

func gitMove(root, sourceRel, targetRel string) error {
	cmd := exec.Command("git", "-C", root, "mv", "--", sourceRel, targetRel)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func gitRemove(root, targetRel string) error {
	cmd := exec.Command("git", "-C", root, "rm", "-rf", "--", targetRel)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run()
}

func rewriteVaultLinks(vaultPath, oldRel, newRel string, ignored []string) (int, []string, error) {
	oldRel = obsidian.NormalizePath(oldRel)
	newRel = obsidian.NormalizePath(newRel)
	totalUpdates := 0
	var skipped []string

	// First pass: check if the basename is unique in the vault
	// Use case-insensitive comparison on macOS/Windows to match filesystem behavior
	oldBasename := filepath.Base(strings.TrimSuffix(oldRel, ".md"))
	oldBasenameNorm := obsidian.NormalizeForComparison(oldBasename)
	newBasename := filepath.Base(strings.TrimSuffix(newRel, ".md"))
	newBasenameNorm := obsidian.NormalizeForComparison(newBasename)
	basenameCount := 0
	_ = filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || filepath.Ext(d.Name()) != ".md" {
			return nil
		}
		if obsidian.ShouldIgnorePath(vaultPath, path, ignored) {
			return nil
		}
		if obsidian.NormalizeForComparison(strings.TrimSuffix(d.Name(), ".md")) == oldBasenameNorm {
			basenameCount++
		}
		return nil
	})
	// If the note was renamed to a different basename, add it back into the count so we
	// consider the pre-rename state when deciding if bare wikilinks are ambiguous.
	if oldBasenameNorm != newBasenameNorm && !obsidian.ShouldIgnorePath(vaultPath, filepath.Join(vaultPath, oldRel), ignored) {
		basenameCount++
	}
	basenameUnique := basenameCount <= 1

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
		rewritten, count := obsidian.RewriteLinksInContentWithOptions(content, oldRel, newRel, basenameUnique)
		if count == 0 {
			return nil
		}
		totalUpdates += count
		info, statErr := d.Info()
		if statErr != nil {
			return statErr
		}
		return os.WriteFile(path, []byte(rewritten), info.Mode())
	})

	if err != nil {
		return 0, skipped, err
	}
	return totalUpdates, skipped, nil
}
