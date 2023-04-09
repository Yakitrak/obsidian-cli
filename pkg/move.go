package pkg

import (
	"log"
)

func MoveNote(uriConstructor UriConstructorFunc, findVaultPathFromName FindVaultPathFromNameFunc, moveNote MoveNoteFunc, updateLinksInVault UpdateLinksInVaultFunc, vaultName string, currentNoteName string, newNoteName string) string {
	// Find obsidian vault directory
	vaultPath, err := findVaultPathFromName(vaultName)

	if err != nil {
		log.Fatal(err)
	}

	// Move / rename note
	currentPath := vaultPath + "/" + currentNoteName
	newPath := vaultPath + "/" + newNoteName
	moveNote(currentPath, newPath)

	updateLinksInVault(vaultPath, currentNoteName, newNoteName)

	// Open renamed note
	uri := uriConstructor(ObsOpenUrl, map[string]string{
		"file":  newNoteName,
		"vault": vaultName,
	})

	return uri

}
