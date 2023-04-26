package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"path/filepath"
)

type DeleteParams struct {
	NotePath string
}

func DeleteNote(vault obsidian.VaultManager, note obsidian.NoteManager, params DeleteParams) error {
	_, err := vault.DefaultName()
	if err != nil {
		return err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}
	notePath := filepath.Join(vaultPath, params.NotePath)

	err = note.Delete(notePath)
	if err != nil {
		return err
	}
	return nil
}
