package pkg

import (
	"fmt"
	"log"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func RenameNote(vaultName string, currentNoteName string, newNoteName string) string {
	fmt.Println("This feature is new, please report any bugs!")

	if vaultName == "" {
		vaultName = utils.GetDefaultVault()
	}
	// Find obsidian vault directory
	vaultPath, err := utils.FindVaultPathFromName(vaultName)

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(vaultPath)

	// Find name of note to rename -> replace name
	fmt.Println("Renaming note " + currentNoteName + " to " + newNoteName)
	currentPath := vaultPath + "/" + currentNoteName
	newPath := vaultPath + "/" + newNoteName
	utils.RenameNote(currentPath, newPath)

	fmt.Println("Updated links inside: abc")
	// utils.UpdateLinks(vaultName, currentNoteName, newNoteName)

	// Open renamed note
	uri := ObsOpenUrl + utils.UrlConstructor(map[string]string{
		"file":  newNoteName,
		"vault": vaultName,
	})

	return uri

}
