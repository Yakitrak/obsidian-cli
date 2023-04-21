package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type SearchParams struct {
	SearchText string
}

func SearchNotes(vaultOp obsidian.VaultOperator, uriManager obsidian.UriManager, params SearchParams) error {
	vaultName, err := vaultOp.DefaultName()
	if err != nil {
		return err
	}

	uri := uriManager.Construct(ObsSearchUrl, map[string]string{
		"vault": vaultName,
		"query": params.SearchText,
	})

	err = uriManager.Execute(uri)
	if err != nil {
		return err
	}
	return nil
}
