package actions

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"path/filepath"
)

func DeleteNote(vaultName string, notePath string) error {
	vault := vault.Vault{Name: vaultName}
	vault.DefaultName()
	vaultPath, err := vault.Path()

	if err != nil {
		return fmt.Errorf("cannot locate vault %g", err)
	}
	notePath = filepath.Join(vaultPath, notePath)

	err = note.Delete(notePath)
	if err != nil {
		return err
	}
	return nil
}
