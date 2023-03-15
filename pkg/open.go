package pkg

import (
	"github.com/Yakitrak/obsidian-cli/utils"
)

func OpenNote(fileName string, vaultName string) string {
	if vaultName == "" {
		vaultName = utils.GetDefaultVault()
	}

	uri := ObsOpenUrl + utils.UrlConstructor(map[string]string{
		"file":  fileName,
		"vault": vaultName,
	})

	return uri
}
