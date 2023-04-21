package note

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

type Manager struct {
}

type ManagerInterface interface {
	Move(string, string) error
	Delete(string) error
	UpdateLinks(string, string, string) error
}

func (m *Manager) Move(originalPath string, newPath string) error {
	o := AddMdSuffix(originalPath)
	n := AddMdSuffix(newPath)

	err := os.Rename(o, n)

	if err != nil {
		return err
	} else {
		message := fmt.Sprintf(`Moved note 
from %s
to %s`, o, n)
		fmt.Println(message)
	}
	return nil
}
func (m *Manager) Delete(path string) error {
	note := AddMdSuffix(path)
	err := os.Remove(note)
	if err != nil {
		return errors.New("note does not exist")
	} else {
		fmt.Println("Deleted note: ", note)
	}
	return nil
}

func (m *Manager) UpdateLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	err := filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New("Failed to access obsidian directory")
		}

		if ShouldSkipDirectoryOrFile(info) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return errors.New("Failed to read files in obsidian")
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
			return errors.New("Failed to write to files in obsidian")
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}
