package pkg

import (
	"log"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func MoveNote(vaultName string, currentNoteName string, newNoteName string) string {
	if vaultName == "" {
		vaultName = utils.GetDefaultVault()
	}
	// Find obsidian vault directory
	vaultPath, err := utils.FindVaultPathFromName(vaultName)

	if err != nil {
		log.Fatal(err)
	}

	// Move / rename note
	currentPath := vaultPath + "/" + currentNoteName
	newPath := vaultPath + "/" + newNoteName
	utils.MoveNote(currentPath, newPath)

	// Update links inside note
	utils.UpdateLinksInVault(vaultPath, currentNoteName, newNoteName)

	// Open renamed note
	uri := ObsOpenUrl + utils.UrlConstructor(map[string]string{
		"file":  newNoteName,
		"vault": vaultName,
	})

	return uri

}
