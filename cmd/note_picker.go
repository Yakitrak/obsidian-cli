package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/ktr0731/go-fuzzyfinder"
)

const (
	notePickerNewNote   = "(Create new note...)"
	notePickerVaultRoot = "(Vault root)"
)

func pickExistingNotePath(vaultPath string) (string, error) {
	notes, err := listVaultNotes(vaultPath)
	if err != nil {
		return "", err
	}
	if len(notes) == 0 {
		return "", errors.New("no notes found in vault")
	}
	sort.Strings(notes)

	idx, err := fuzzyfinder.Find(notes, func(i int) string { return notes[i] })
	if err != nil {
		return "", err
	}
	return notes[idx], nil
}

func pickNotePathOrNew(vaultPath string) (string, error) {
	notes, err := listVaultNotes(vaultPath)
	if err != nil {
		return "", err
	}
	notes = append(notes, notePickerNewNote)
	sort.Strings(notes)

	idx, err := fuzzyfinder.Find(notes, func(i int) string { return notes[i] })
	if err != nil {
		return "", err
	}
	choice := notes[idx]
	if choice != notePickerNewNote {
		return choice, nil
	}
	return promptNewNotePath(vaultPath)
}

func promptNewNotePath(vaultPath string) (string, error) {
	folders, err := listVaultFolders(vaultPath)
	if err != nil {
		return "", err
	}
	folders = append(folders, notePickerVaultRoot, "(Create new folder...)")
	sort.Strings(folders)

	idx, err := fuzzyfinder.Find(folders, func(i int) string { return folders[i] })
	if err != nil {
		return "", err
	}
	choice := folders[idx]
	switch choice {
	case "(Create new folder...)":
		created, err := promptCreateFolder(vaultPath)
		if err != nil {
			return "", err
		}
		choice = created
	case notePickerVaultRoot:
		choice = ""
	}

	in := bufio.NewReader(os.Stdin)
	fmt.Println("Enter new note name (relative to the selected folder; .md optional):")
	fmt.Print("> ")
	name, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}
	name = strings.TrimSpace(name)
	if name == "" {
		return "", errors.New("note name is required")
	}
	name = strings.TrimSuffix(name, ".md")

	rel := strings.TrimSpace(filepath.ToSlash(filepath.Join(choice, name)))
	if rel == "" {
		return "", errors.New("note path is required")
	}
	if _, err := obsidian.SafeJoinVaultPath(vaultPath, rel); err != nil {
		return "", err
	}
	return rel, nil
}

