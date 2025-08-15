package actions

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"gopkg.in/yaml.v3"
)

type FrontmatterEditParams struct {
	NoteName string
	Key      string
	Value    string
}

// EditFrontmatter finds the note in the vault and edits/creates the specified frontmatter key.
func EditFrontmatter(vault obsidian.VaultManager, params FrontmatterEditParams) error {
	if params.NoteName == "" {
		return errors.New("note name is required")
	}
	if params.Key == "" {
		return errors.New("key is required")
	}

	// Ensure default vault is set
	_, err := vault.DefaultName()
	if err != nil {
		return err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	// Locate the note's absolute path within the vault
	targetFile := obsidian.AddMdSuffix(params.NoteName)
	var notePath string
	err = filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Base(path) == targetFile {
			notePath = path
			return filepath.SkipDir
		}
		return nil
	})
	if err != nil || notePath == "" {
		return errors.New(obsidian.NoteDoesNotExistError)
	}

	// Read the file
	info, err := os.Stat(notePath)
	if err != nil {
		return errors.New(obsidian.VaultReadError)
	}
	data, err := os.ReadFile(notePath)
	if err != nil {
		return errors.New(obsidian.VaultReadError)
	}
	content := string(data)

	// Parse frontmatter
	fmMap, body, _, err := parseFrontmatter(content)
	if err != nil {
		return err
	}

	// Compute new value type by trying to parse as YAML for flexibility
	var newVal any
	if err := yaml.Unmarshal([]byte(params.Value), &newVal); err != nil || newVal == nil {
		// Fallback to raw string
		newVal = params.Value
	}
	// Convenience: if key is tags and value contains commas, split into array of trimmed strings
	if strings.EqualFold(params.Key, "tags") {
		if strings.Contains(params.Value, ",") {
			parts := strings.Split(params.Value, ",")
			arr := make([]string, 0, len(parts))
			for _, p := range parts {
				p = strings.TrimSpace(p)
				if p != "" {
					arr = append(arr, p)
				}
			}
			newVal = arr
		}
	}

	// Update map
	if fmMap == nil {
		fmMap = map[string]any{}
	}
	fmMap[params.Key] = newVal

	// Marshal YAML
	yml, err := yaml.Marshal(fmMap)
	if err != nil {
		return errors.New(obsidian.VaultWriteError)
	}

	// Reconstruct content
	newContent := buildWithFrontmatter(string(yml), body)

	// Write back
	if err := os.WriteFile(notePath, []byte(newContent), info.Mode()); err != nil {
		return errors.New(obsidian.VaultWriteError)
	}

	fmt.Printf("Updated frontmatter in %s: %s\n", filepath.Base(notePath), params.Key)
	return nil
}

// parseFrontmatter extracts YAML frontmatter into a map and returns the body.
// It supports both \n and \r\n line endings.
func parseFrontmatter(content string) (map[string]any, string, bool, error) {
	trimmed := content
	if len(trimmed) == 0 {
		return nil, "", false, nil
	}
	// Normalize line endings to \n for parsing
	norm := strings.ReplaceAll(trimmed, "\r\n", "\n")
	if !strings.HasPrefix(norm, "---\n") {
		return nil, content, false, nil
	}
	// Find the next delimiter line starting at beginning of line
	idx := strings.Index(norm[4:], "\n---\n")
	if idx == -1 {
		// malformed frontmatter; treat whole as body
		return nil, content, false, nil
	}
	endIdx := 4 + idx + len("\n---\n")
	fm := norm[4 : 4+idx]
	body := norm[endIdx:]
	// Marshal back to original newline style when building
	var m map[string]any
	if strings.TrimSpace(fm) != "" {
		if err := yaml.Unmarshal([]byte(fm), &m); err != nil {
			return nil, "", false, errors.New(obsidian.VaultReadError)
		}
	}
	// Return body with original line endings if original had \r\n
	return m, body, true, nil
}

func buildWithFrontmatter(yml string, body string) string {
	// Ensure trailing newline after closing delimiter
	if !strings.HasPrefix(body, "\n") {
		// Ensure a single newline separation between FM and body when body is non-empty
		if len(body) > 0 && body[0] != '\n' {
			body = "\n" + body
		}
	}
	return "---\n" + yml + "---\n" + body
}
