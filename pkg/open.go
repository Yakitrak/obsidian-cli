package pkg

import (
	"github.com/Yakitrak/obsidian-cli/handler"
	"github.com/Yakitrak/obsidian-cli/utils/uri"
)

func OpenNote(vaultName string, noteName string) (string, error) {
	vaultHandler := handler.Vault{Name: vaultName}
	vaultName, err := vaultHandler.DefaultName()
	if err != nil {
		return "", err
	}

	uri := uri.Construct(ObsOpenUrl, map[string]string{
		"file":  noteName,
		"vault": vaultName,
	})

	return uri, nil
}
