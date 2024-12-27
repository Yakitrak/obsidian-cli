package obsidian

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

type Note struct {
}

type NoteManager interface {
	Move(string, string) error
	Delete(string) error
	UpdateLinks(string, string, string) error
	GetContents(string, string) (string, error)
	GetNotesList(string) ([]string, error)
}

func (m *Note) Move(originalPath string, newPath string) error {
	o := AddMdSuffix(originalPath)
	n := AddMdSuffix(newPath)

	err := os.Rename(o, n)

	if err != nil {
		return errors.New(NoteDoesNotExistError)
	}

	message := fmt.Sprintf(`Moved note 
from %s
to %s`, o, n)

	fmt.Println(message)
	return nil
}
func (m *Note) Delete(path string) error {
	note := AddMdSuffix(path)
	err := os.Remove(note)
	if err != nil {
		return errors.New(NoteDoesNotExistError)
	}
	fmt.Println("Deleted note: ", note)
	return nil
}

func (m *Note) GetContents(vaultPath string, noteName string) (string, error) {
	note := AddMdSuffix(noteName)

	var notePath string
	err := filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err // Continue to the next path if there's an error
		}
		if d.IsDir() {
			return nil // Skip directories
		}
		if filepath.Base(path) == note {
			notePath = path
			return filepath.SkipDir
		}
		return nil
	})

	if err != nil || notePath == "" {
		return "", errors.New(NoteDoesNotExistError)
	}

	file, err := os.Open(notePath)
	if err != nil {
		return "", errors.New(VaultReadError)
	}
	defer file.Close()

	content, err := io.ReadAll(file)
	if err != nil {
		return "", errors.New(VaultReadError)
	}

	return string(content), nil
}

func (m *Note) UpdateLinks(vaultPath string, oldNoteName string, newNoteName string) error {
	err := filepath.Walk(vaultPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return errors.New(VaultAccessError)
		}

		if ShouldSkipDirectoryOrFile(info) {
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return errors.New(VaultReadError)
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
			return errors.New(VaultWriteError)
		}
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (m *Note) GetNotesList(vaultPath string) ([]string, error) {
	var notes []string
	err := filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			relPath, err := filepath.Rel(vaultPath, path)
			if err != nil {
				return err
			}
			notes = append(notes, relPath)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return notes, nil
}
