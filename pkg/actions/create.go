package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"strconv"
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

	obsidianUri := uri.Construct(ObsCreateUrl, map[string]string{
		"vault":     vaultName,
		"append":    strconv.FormatBool(params.ShouldAppend),
		"overwrite": strconv.FormatBool(params.ShouldOverwrite),
		"content":   params.Content,
		"file":      params.NoteName,
		"silent":    strconv.FormatBool(!params.ShouldOpen),
	})

	err = uri.Execute(obsidianUri)
	if err != nil {
		return err
	}

	return nil
}
