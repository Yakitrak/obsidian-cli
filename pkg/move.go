package pkg

import (
	"fmt"
	"os"
)

func MoveNote(uriConstructor UriConstructorFunc, findVaultPathFromName FindVaultPathFromNameFunc, moveNote MoveNoteFunc, updateLinksInVault UpdateLinksInVaultFunc, vaultName string, currentNoteName string, newNoteName string) (string, error) {
	// Find obsidian vault directory
	userConfigDir, err := os.UserConfigDir()
	if err != nil {
		return "", fmt.Errorf("user config directory not found %g", err)
	}
	configFilePath := userConfigDir + ObsidianConfigPath
	vaultPath, err := findVaultPathFromName(vaultName, configFilePath)

	if err != nil {
		return "", fmt.Errorf("cannot locate vault %g", err)
	}

	// Move / rename note
	currentPath := vaultPath + "/" + currentNoteName
	newPath := vaultPath + "/" + newNoteName
	err = moveNote(currentPath, newPath)
	if err != nil {
		return "", err
	}
	updateLinksInVault(vaultPath, currentNoteName, newNoteName)

	// Open renamed note
	uri := uriConstructor(ObsOpenUrl, map[string]string{
		"file":  newNoteName,
		"vault": vaultName,
	})

	return uri, nil
}
