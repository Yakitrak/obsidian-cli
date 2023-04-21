package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/note"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"path/filepath"
)

type MoveParams struct {
	CurrentNoteName string
	NewNoteName     string
	ShouldOpen      bool
}

func MoveNote(vaultOp obsidian.VaultOperator, noteManager note.ManagerInterface, uriManager obsidian.UriManager, params MoveParams) error {
	vaultName, err := vaultOp.DefaultName()
	if err != nil {
		return err
	}
	vaultPath, err := vaultOp.Path()

	if err != nil {
		return err
	}

	currentPath := filepath.Join(vaultPath, params.CurrentNoteName)
	newPath := filepath.Join(vaultPath, params.NewNoteName)

	err = noteManager.Move(currentPath, newPath)
	if err != nil {
		return err
	}

	err = noteManager.UpdateLinks(vaultPath, params.CurrentNoteName, params.NewNoteName)
	if err != nil {
		return err
	}

	if params.ShouldOpen {
		uri := uriManager.Construct(ObsOpenUrl, map[string]string{
			"file":     params.NewNoteName,
			"obsidian": vaultName,
		})

		err := uriManager.Execute(uri)
		if err != nil {
			return err
		}
	}

	return nil
}
