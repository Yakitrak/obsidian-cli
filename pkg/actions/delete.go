package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"path/filepath"
)

type DeleteParams struct {
	NotePath string
}

func DeleteNote(vaultOp obsidian.VaultOperator, noteManager note.ManagerInterface, params DeleteParams) error {
	vaultPath, err := vaultOp.Path()
	if err != nil {
		return err
	}
	notePath := filepath.Join(vaultPath, params.NotePath)

	err = noteManager.Delete(notePath)
	if err != nil {
		return err
	}
	return nil
}
