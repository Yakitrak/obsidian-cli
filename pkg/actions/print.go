package actions

import (
	"fmt"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type PrintParams struct {
	NoteName        string
	IncludeMentions bool
}

func PrintNote(vault obsidian.VaultManager, note obsidian.NoteManager, params PrintParams) (string, error) {
	_, err := vault.DefaultName()
	if err != nil {
		return "", err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return "", err
	}

	contents, err := note.GetContents(vaultPath, params.NoteName)
	if err != nil {
		return "", err
	}

	if params.IncludeMentions {
		backlinks, err := note.FindBacklinks(vaultPath, params.NoteName)
		if err != nil {
			return "", err
		}

		if len(backlinks) > 0 {
			contents += formatMentions(backlinks)
		}
	}

	return contents, nil
}

func formatMentions(backlinks []obsidian.NoteMatch) string {
	var sb strings.Builder
	sb.WriteString("\n\n## Linked Mentions\n")

	// Group matches by file path, preserving order
	grouped := make(map[string][]obsidian.NoteMatch)
	var order []string

	for _, match := range backlinks {
		if _, exists := grouped[match.FilePath]; !exists {
			order = append(order, match.FilePath)
		}
		grouped[match.FilePath] = append(grouped[match.FilePath], match)
	}

	for _, filePath := range order {
		noteName := strings.TrimSuffix(filePath, ".md")
		fmt.Fprintf(&sb, "\n**[[%s]]**\n", noteName)
		for _, match := range grouped[filePath] {
			sb.WriteString("- ")
			sb.WriteString(match.MatchLine)
			sb.WriteByte('\n')
		}
	}

	return sb.String()
}
