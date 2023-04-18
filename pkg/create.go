package pkg

import (
	"github.com/Yakitrak/obsidian-cli/handler"
	"github.com/Yakitrak/obsidian-cli/utils/uri"
	"strconv"
)

func CreateNote(vaultName string, noteName string, content string, shouldAppend bool, shouldOverwrite bool) (string, error) {
	vaultHandler := handler.Vault{Name: vaultName}
	vaultName, err := vaultHandler.DefaultName()
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
