package actions

import (
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

func PlanDailyNote(vault obsidian.VaultManager, uri obsidian.UriManager) (string, error) {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return "", err
	}

	return uri.Construct(OnsDailyUrl, map[string]string{
		"vault": vaultName,
	}), nil
}

func DailyNote(vault obsidian.VaultManager, uri obsidian.UriManager) error {
	obsidianUri, err := PlanDailyNote(vault, uri)
	if err != nil {
		return err
	}

	err = uri.Execute(obsidianUri)
	if err != nil {
		return err
	}
	return nil
}
