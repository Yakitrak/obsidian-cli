package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
)

func OpenNote(vaultName string, noteName string) (string, error) {
	vault := vault.Vault{Name: vaultName}
	vaultName, err := vault.DefaultName()
	if err != nil {
		return "", err
	}

	uri := uri.Construct(ObsOpenUrl, map[string]string{
		"file":  noteName,
		"vault": vaultName,
	})

	return uri, nil
}
