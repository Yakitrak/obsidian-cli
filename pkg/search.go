package pkg

import (
	"github.com/Yakitrak/obsidian-cli/utils"
)

func SearchNotes(searchText string, vaultName string) string {
	uri := ObsSearchUrl + utils.UrlConstructor(map[string]string{
		"query": searchText,
		"vault": vaultName,
	})

	return uri
}
