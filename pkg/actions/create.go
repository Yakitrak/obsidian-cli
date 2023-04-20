package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"strconv"
)

func CreateNote(vaultOp vault.VaultOperator, noteName string, content string, shouldAppend bool, shouldOverwrite bool) (string, error) {
	vaultName, err := vaultOp.DefaultName()
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
