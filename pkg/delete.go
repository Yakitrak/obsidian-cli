package pkg

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/handler"
	"github.com/Yakitrak/obsidian-cli/utils/note"
	"path/filepath"
)

func DeleteNote(vaultName string, notePath string) error {
	vaultHandler := handler.Vault{Name: vaultName}
	vaultHandler.DefaultName()
	vaultPath, err := vaultHandler.Path()

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
