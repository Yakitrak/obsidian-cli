package pkg

import (
	"strconv"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func CreateNote(vaultName string, noteName string, content string, shouldAppend bool, shouldOverwrite bool) string {
	uri := ObsCreateUrl + utils.UrlConstructor(map[string]string{
		"append":    strconv.FormatBool(shouldAppend),
		"content":   content,
		"file":      noteName,
		"overwrite": strconv.FormatBool(shouldOverwrite),
		"vault":     vaultName,
	})

	return uri
}
