package pkg

import (
	"github.com/Yakitrak/obsidian-cli/handler"
	"github.com/Yakitrak/obsidian-cli/utils/uri"
)

func SearchNotes(vaultName string, searchText string) (string, error) {
	vaultHandler := handler.Vault{Name: vaultName}
	vaultName, err := vaultHandler.DefaultName()
	if err != nil {
		return "", err
	}

	uri := uri.Construct(ObsSearchUrl, map[string]string{
		"query": searchText,
		"vault": vaultName,
	})

	return uri, nil
}
