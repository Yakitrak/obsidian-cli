package pkg

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/handler"
	"github.com/Yakitrak/obsidian-cli/utils/note"
	"github.com/Yakitrak/obsidian-cli/utils/uri"
	"path/filepath"
)

func MoveNote(vaultName string, currentNoteName string, newNoteName string) (string, error) {
	vaultHandler := handler.Vault{Name: vaultName}
	vaultName, err := vaultHandler.DefaultName()
	if err != nil {
		return "", err
	}
	vaultPath, err := vaultHandler.Path()

	if err != nil {
		return "", fmt.Errorf("cannot locate vault %s", err)
	}

	currentPath := filepath.Join(vaultPath, currentNoteName)
	newPath := filepath.Join(vaultPath, newNoteName)

	err = note.Move(currentPath, newPath)
	if err != nil {
		return "", fmt.Errorf("cannot move note '%s'", currentNoteName)
	}

	vaultHandler.UpdateNoteLinks(vaultPath, currentNoteName, newNoteName)

	uri := uri.Construct(ObsOpenUrl, map[string]string{
		"file":  newNoteName,
		"vault": vaultName,
	})

	return uri, nil
}
