package actions

import (
	"fmt"
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/Yakitrak/obsidian-cli/pkg/vault"
	"path/filepath"
)

func DeleteNote(vaultOp vault.VaultOperator, notePath string) error {
	vaultOp.DefaultName()
	vaultPath, err := vaultOp.Path()

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
