package actions

import (
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	if params.UseEditor && params.ShouldOpen {
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}
		// Create note via Obsidian URI to respect append/overwrite flags
		obsidianUri := uri.Construct(ObsCreateUrl, map[string]string{
			"vault":     vaultName,
			"append":    strconv.FormatBool(params.ShouldAppend),
			"overwrite": strconv.FormatBool(params.ShouldOverwrite),
			"content":   normalizedContent,
			"file":      params.NoteName,
			"silent":    "true",
		})

		if err := uri.Execute(obsidianUri); err != nil {
			return err
		}

		// Wait for Obsidian to finish creating the file before opening in editor.
		// The URI command is async, so we need a brief delay to ensure the file exists.
		time.Sleep(200 * time.Millisecond)

		filePath := filepath.Join(vaultPath, obsidian.AddMdSuffix(params.NoteName))
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
