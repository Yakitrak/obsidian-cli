package actions

import (
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type CreateParams struct {
	NoteName        string
	ShouldAppend    bool
	ShouldOverwrite bool
	Content         string
	ShouldOpen      bool
	UseEditor       bool
}

func CreateNote(vault obsidian.VaultManager, uri obsidian.UriManager, params CreateParams) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	normalizedContent := NormalizeContent(params.Content)

	// If using editor and should open, use editor mode instead of Obsidian
	if params.UseEditor && params.ShouldOpen {
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}
		// Note: When using editor mode, the note must be created via Obsidian URI first
		// before opening in editor, to respect append/overwrite flags
		obsidianUri := uri.Construct(ObsCreateUrl, map[string]string{
			"vault":     vaultName,
			"append":    strconv.FormatBool(params.ShouldAppend),
			"overwrite": strconv.FormatBool(params.ShouldOverwrite),
			"content":   normalizedContent,
			"file":      params.NoteName,
			"silent":    "true", // Don't open in Obsidian
		})

		if err := uri.Execute(obsidianUri); err != nil {
			return err
		}

		// Now open in editor
		filePath := filepath.Join(vaultPath, params.NoteName)
		return obsidian.OpenInEditor(filePath)
	}

	obsidianUri := uri.Construct(ObsCreateUrl, map[string]string{
		"vault":     vaultName,
		"append":    strconv.FormatBool(params.ShouldAppend),
		"overwrite": strconv.FormatBool(params.ShouldOverwrite),
		"content":   normalizedContent,
		"file":      params.NoteName,
		"silent":    strconv.FormatBool(!params.ShouldOpen),
	})

	if err := uri.Execute(obsidianUri); err != nil {
		return err
	}

	return nil
}

func NormalizeContent(content string) string {
	replacer := strings.NewReplacer(
		"\\n", "\n",
		"\\r", "\r",
		"\\t", "\t",
		"\\\\", "\\",
		"\\\"", "\"",
		"\\'", "'",
	)
	return replacer.Replace(content)
}
