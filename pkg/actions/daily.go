package actions

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

func DailyNote(vault obsidian.VaultManager, uri obsidian.UriManager, useEditor bool) error {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return err
	}

	if useEditor {
		// Note: For editor mode, we use a simple date format for the daily note
		// This may not match Obsidian's daily note configuration, but provides basic functionality
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}
		// Use a standard daily note format: YYYY-MM-DD.md
		dailyNoteName := time.Now().Format("2006-01-02") + ".md"
		filePath := filepath.Join(vaultPath, dailyNoteName)
		fmt.Printf("Opening daily note: %s\n", dailyNoteName)
		return obsidian.OpenInEditor(filePath)
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
