package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"strconv"
	"strings"
)

type CreateParams struct {
	NoteName        string
	ShouldAppend    bool
	ShouldOverwrite bool
	Content         string
	ShouldOpen      bool
}

func CreateNote(vault obsidian.VaultManager, uri obsidian.UriManager, params CreateParams) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	normalizedContent := NormalizeContent(params.Content)

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
