package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type PrintParams struct {
	NoteName string
}

func PrintNote(vault obsidian.VaultManager, note obsidian.NoteManager, params PrintParams) (string, error) {
	_, err := vault.DefaultName()
	if err != nil {
		return "", err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return "", err
	}

	contents, err := note.GetContents(vaultPath, params.NoteName)
	if err != nil {
		return "", err
	}

	return contents, nil
}
