package actions

import (
	"errors"
	"fmt"

	"github.com/Yakitrak/obsidian-cli/pkg/frontmatter"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

type FrontmatterParams struct {
	NoteName string
	Print    bool
	Edit     bool
	Delete   bool
	Key      string
	Value    string
}

// Frontmatter handles viewing and modifying note frontmatter.
// Based on flags, it will print, edit, or delete frontmatter keys.
func Frontmatter(vault obsidian.VaultManager, note obsidian.NoteManager, params FrontmatterParams) (string, error) {
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

	// Handle print operation
	if params.Print {
		return handlePrint(contents)
	}

	// Handle edit operation
	if params.Edit {
		return handleEdit(note, vaultPath, params.NoteName, contents, params.Key, params.Value)
	}

	// Handle delete operation
	if params.Delete {
		return handleDelete(note, vaultPath, params.NoteName, contents, params.Key)
	}

	return "", errors.New("no operation specified: use --print, --edit, or --delete")
}

func handlePrint(contents string) (string, error) {
	if !frontmatter.HasFrontmatter(contents) {
		return "", nil // Return empty for notes without frontmatter
	}

	fm, _, err := frontmatter.Parse(contents)
	if err != nil {
		return "", err
	}

	formatted, err := frontmatter.Format(fm)
	if err != nil {
		return "", err
	}

	return formatted, nil
}

func handleEdit(note obsidian.NoteManager, vaultPath, noteName, contents, key, value string) (string, error) {
	if key == "" {
		return "", errors.New("--key is required for edit operation")
	}
	if value == "" {
		return "", errors.New("--value is required for edit operation")
	}

	updatedContent, err := frontmatter.SetKey(contents, key, value)
	if err != nil {
		return "", err
	}

	err = note.SetContents(vaultPath, noteName, updatedContent)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Updated frontmatter key '%s' in %s", key, noteName), nil
}

func handleDelete(note obsidian.NoteManager, vaultPath, noteName, contents, key string) (string, error) {
	if key == "" {
		return "", errors.New("--key is required for delete operation")
	}

	updatedContent, err := frontmatter.DeleteKey(contents, key)
	if err != nil {
		return "", err
	}

	err = note.SetContents(vaultPath, noteName, updatedContent)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("Deleted frontmatter key '%s' from %s", key, noteName), nil
}
