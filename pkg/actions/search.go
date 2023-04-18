package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
)

func SearchNotes(vaultName string, searchText string) (string, error) {
	vault := vault.Vault{Name: vaultName}
	vaultName, err := vault.DefaultName()
	if err != nil {
		return "", err
	}

	uri := uri.Construct(ObsSearchUrl, map[string]string{
		"query": searchText,
		"vault": vaultName,
	})

	return uri, nil
}
