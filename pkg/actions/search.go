package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
)

func SearchNotes(vaultOp vault.VaultOperator, searchText string) (string, error) {
	vaultName, err := vaultOp.DefaultName()
	if err != nil {
		return "", err
	}

	uri := uri.Construct(ObsSearchUrl, map[string]string{
		"query": searchText,
		"vault": vaultName,
	})

	return uri, nil
}
