package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

func DailyNote(vault obsidian.VaultManager, uri obsidian.UriManager) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	obsidianUri := uri.Construct(OnsDailyUrl, map[string]string{
		"vault": vaultName,
	})

	err = uri.Execute(obsidianUri)
	if err != nil {
		return err
	}
	return nil
}
