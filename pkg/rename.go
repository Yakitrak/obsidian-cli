package pkg

import (
	"fmt"
	"log"

	"github.com/Yakitrak/obsidian-cli/utils"
)

func RenameNote(vaultName string, currentFileName string, newNoteName string) string {
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

	// Find name of file to rename -> replace name
	fmt.Println("Renaming file " + currentFileName + " to " + newNoteName)
	// utils.RenameNote(currentFileName, newNoteName)

	fmt.Println("Updated links inside: abc")
	// utils.UpdateLinks(vaultName, currentFileName, newNoteName)

	// Open renamed file
	uri := ObsOpenUrl + utils.UrlConstructor(map[string]string{
		"file":  newNoteName,
		"vault": vaultName,
	})

	return uri

}
