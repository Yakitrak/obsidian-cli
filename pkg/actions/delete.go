package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
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

	// Validate path stays within vault directory
	notePath, err := obsidian.ValidatePath(vaultPath, params.NotePath)
	if err != nil {
		return err
	}

	err = note.Delete(notePath)
	if err != nil {
		return err
	}
	return nil
}
