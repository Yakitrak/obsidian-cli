package actions

import (
	"path/filepath"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type OpenParams struct {
	NoteName  string
	UseEditor bool
}

func OpenNote(vault obsidian.VaultManager, uri obsidian.UriManager, params OpenParams) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	if params.UseEditor {
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}
		filePath := filepath.Join(vaultPath, params.NoteName)
		return obsidian.OpenInEditor(filePath)
	}

	obsidianUri := uri.Construct(ObsOpenUrl, map[string]string{
		"vault": vaultName,
		"file":  params.NoteName,
	})

	err = uri.Execute(obsidianUri)
	if err != nil {
		return err
	}
	return nil
}
