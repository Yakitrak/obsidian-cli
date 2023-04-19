package note

import (
	"errors"
	"os"
	"path/filepath"
)

func UpdateNoteLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	err := filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New("Failed to access vault directory")
		}

		if ShouldSkipDirectoryOrFile(info) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return errors.New("Failed to read files in vault")
		}

		oldNoteLinkTexts := GenerateNoteLinkTexts(oldNoteName)
		newNoteLinkTexts := GenerateNoteLinkTexts(newNoteName)

		content = ReplaceContent(content, map[string]string{
			oldNoteLinkTexts[0]: newNoteLinkTexts[0],
			oldNoteLinkTexts[1]: newNoteLinkTexts[1],
			oldNoteLinkTexts[2]: newNoteLinkTexts[2],
		})

		err = os.WriteFile(path, content, info.Mode())
		if err != nil {
			return errors.New("Failed to write to files in vault")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
