package pkg

import (
	"github.com/Yakitrak/obsidian-cli/utils"
)

func OpenNote(noteName string, vaultName string) string {
	if vaultName == "" {
		vaultName = utils.GetDefaultVault()
	}

	uri := ObsOpenUrl + utils.UrlConstructor(map[string]string{
		"file":  noteName,
		"vault": vaultName,
	})

	return uri
}
