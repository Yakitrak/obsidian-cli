package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"path/filepath"
)

type MoveParams struct {
	CurrentNoteName string
	NewNoteName     string
	ShouldOpen      bool
}

func MoveNote(vault obsidian.VaultManager, note obsidian.NoteManager, uri obsidian.UriManager, params MoveParams) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}
	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	currentPath := filepath.Join(vaultPath, params.CurrentNoteName)
	newPath := filepath.Join(vaultPath, params.NewNoteName)

	err = note.Move(currentPath, newPath)
	if err != nil {
		return err
	}

	err = note.UpdateLinks(vaultPath, params.CurrentNoteName, params.NewNoteName)
	if err != nil {
		return err
	}

	if params.ShouldOpen {
		obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
			"file":  params.NewNoteName,
			"vault": vaultName,
		})

		err := uri.Execute(obsidianUri)
		if err != nil {
			return err
		}
	}

	return nil
}
