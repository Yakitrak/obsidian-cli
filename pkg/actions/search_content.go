package actions

import (
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

func SearchNotesContent(vault obsidian.VaultManager, note obsidian.NoteManager, uri obsidian.UriManager, fuzzyFinder obsidian.FuzzyFinderManager, searchTerm string) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	matches, err := note.SearchNotesWithSnippets(vaultPath, searchTerm)
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		fmt.Printf("No notes found containing '%s'\n", searchTerm)
		return nil
	}

	if len(matches) == 1 {
		obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
			"file":  matches[0].FilePath,
			"vault": vaultName,
		})
		fmt.Printf("Opening note: %s\n", matches[0].FilePath)
		return uri.Execute(obsidianUri)
	}

	displayItems := formatMatchesForDisplay(matches)

	index, err := fuzzyFinder.Find(displayItems, func(i int) string {
		return displayItems[i]
	})
	if err != nil {
		return err
	}

	selectedMatch := matches[index]
	obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
		"file":  selectedMatch.FilePath,
		"vault": vaultName,
	})
	return uri.Execute(obsidianUri)
}

func formatMatchesForDisplay(matches []obsidian.NoteMatch) []string {
	maxPathLength := calculateMaxPathLength(matches)

	var displayItems []string
	for _, match := range matches {
		displayStr := formatSingleMatch(match, maxPathLength)
		displayItems = append(displayItems, displayStr)
	}

	return displayItems
}

func calculateMaxPathLength(matches []obsidian.NoteMatch) int {
	maxLength := 0
	for _, match := range matches {
		pathWithLine := formatPathWithLine(match)
		if len(pathWithLine) > maxLength {
			maxLength = len(pathWithLine)
		}
	}
	return maxLength
}

func formatPathWithLine(match obsidian.NoteMatch) string {
	if match.LineNumber > 0 {
		return fmt.Sprintf("%s:%d", match.FilePath, match.LineNumber)
	}
	return match.FilePath
}

func formatSingleMatch(match obsidian.NoteMatch, maxPathLength int) string {
	if match.LineNumber == 0 {
		// Filename match - just show the file path
		return fmt.Sprintf("%-*s", maxPathLength, match.FilePath)
	}
	// Content match - show path:line | snippet
	pathWithLine := formatPathWithLine(match)
	return fmt.Sprintf("%-*s | %s", maxPathLength, pathWithLine, match.MatchLine)
}
