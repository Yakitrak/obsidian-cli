package actions

import (
	"path/filepath"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type MoveParams struct {
	CurrentNoteName string
	NewNoteName     string
	ShouldOpen      bool
	UseEditor       bool
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
		if params.UseEditor {
			filePathWithExt := filepath.Join(vaultPath, obsidian.AddMdSuffix(params.NewNoteName))
			return obsidian.OpenInEditor(filePathWithExt)
		}

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
