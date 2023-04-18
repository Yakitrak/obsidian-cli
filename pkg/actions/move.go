package actions

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/Yakitrak/obsidian-cli/pkg/uri"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"path/filepath"
)

func MoveNote(vaultName string, currentNoteName string, newNoteName string) (string, error) {
	vault := vault.Vault{Name: vaultName}
	vaultName, err := vault.DefaultName()
	if err != nil {
		return "", err
	}
	vaultPath, err := vault.Path()

	if err != nil {
		return "", fmt.Errorf("cannot locate vault %s", err)
	}

	currentPath := filepath.Join(vaultPath, currentNoteName)
	newPath := filepath.Join(vaultPath, newNoteName)

	err = note.Move(currentPath, newPath)
	if err != nil {
		return "", fmt.Errorf("cannot move note '%s'", currentNoteName)
	}

	vault.UpdateNoteLinks(vaultPath, currentNoteName, newNoteName)

	uri := uri.Construct(ObsOpenUrl, map[string]string{
		"file":  newNoteName,
		"vault": vaultName,
	})

	return uri, nil
}
