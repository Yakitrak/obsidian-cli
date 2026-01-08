package actions

import (
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

	// Validate paths stay within vault directory
	currentPath, err := obsidian.ValidatePath(vaultPath, params.CurrentNoteName)
	if err != nil {
		return err
	}
	newPath, err := obsidian.ValidatePath(vaultPath, params.NewNoteName)
	if err != nil {
		return err
	}

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
			filePathWithExt, err := obsidian.ValidatePath(vaultPath, obsidian.AddMdSuffix(params.NewNoteName))
			if err != nil {
				return err
			}
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
