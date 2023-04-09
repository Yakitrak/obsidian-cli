package pkg

import (
	"strconv"
)

func CreateNote(uriConstructor UriConstructorFunc, vaultName string, noteName string, content string, shouldAppend bool, shouldOverwrite bool) string {
	uri := uriConstructor(ObsCreateUrl, map[string]string{
		"append":    strconv.FormatBool(shouldAppend),
		"content":   content,
		"file":      noteName,
		"overwrite": strconv.FormatBool(shouldOverwrite),
		"vault":     vaultName,
	})

	return uri
}
