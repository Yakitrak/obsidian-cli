package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"strconv"
)

func CreateNote(vaultName string, noteName string, content string, shouldAppend bool, shouldOverwrite bool) (string, error) {
	vault := vault.Vault{Name: vaultName}
	vaultName, err := vault.DefaultName()
	if err != nil {
		return "", err
	}

	uri := uri.Construct(ObsCreateUrl, map[string]string{
		"append":    strconv.FormatBool(shouldAppend),
		"content":   content,
		"file":      noteName,
		"overwrite": strconv.FormatBool(shouldOverwrite),
		"vault":     vaultName,
	})

	return uri, nil
}
