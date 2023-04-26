package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type SearchParams struct {
	SearchText string
}

func SearchNotes(vault obsidian.VaultManager, uri obsidian.UriManager, params SearchParams) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	obsidianUri := uri.Construct(ObsSearchUrl, map[string]string{
		"vault": vaultName,
		"query": params.SearchText,
	})

	err = uri.Execute(obsidianUri)
	if err != nil {
		return err
	}
	return nil
}
