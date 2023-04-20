package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
)

func OpenNote(vaultOp vault.VaultOperator, noteName string) (string, error) {
	vaultName, err := vaultOp.DefaultName()
	if err != nil {
		return "", err
	}

	uri := uri.Construct(ObsOpenUrl, map[string]string{
		"file":  noteName,
		"vault": vaultName,
	})

	return uri, nil
}
