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

type FrontmatterValue struct {
	Value any
	Found bool
}

// GetFrontmatterValue locates a note and returns the value for a specific frontmatter key.
func GetFrontmatterValue(vault obsidian.VaultManager, noteName string, key string) (*FrontmatterValue, error) {
	if noteName == "" {
		return nil, errors.New("note name is required")
	}
	if key == "" {
		return nil, errors.New("key is required")
	}

	// Ensure default vault is set
	if _, err := vault.DefaultName(); err != nil {
		return nil, err
	}
	vaultPath, err := vault.Path()
	if err != nil {
		return nil, err
	}

	// Find note path
	targetFile := obsidian.AddMdSuffix(noteName)
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
		return nil, errors.New(obsidian.NoteDoesNotExistError)
	}

	data, err := os.ReadFile(notePath)
	if err != nil {
		return nil, errors.New(obsidian.VaultReadError)
	}

	fmMap, _, _, err := parseFrontmatter(string(data))
	if err != nil {
		return nil, err
	}
	if fmMap == nil {
		return &FrontmatterValue{Value: nil, Found: false}, nil
	}
	val, ok := fmMap[key]
	return &FrontmatterValue{Value: val, Found: ok}, nil
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

	// Compute new value, supporting empty --value to add empty keys
	var newVal any
	if params.Value == "" {
		// Empty value: tags -> [], others -> ""
		if strings.EqualFold(params.Key, "tags") {
			newVal = []interface{}{}
		} else {
			newVal = ""
		}
	} else {
		// Try to parse YAML for flexibility
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
	// Ensure a single newline between YAML and body when body is non-empty
	if len(body) > 0 {
		if !strings.HasPrefix(body, "\n") {
			body = "\n" + body
		}
	}
	return "---\n" + yml + "---\n" + body
}

// ClearFrontmatter clears the content of an existing frontmatter key.
// Behavior:
// - If key exists and is a sequence: set to empty list []
// - If key exists and is a mapping: set to empty map {}
// - Otherwise: set to empty string ""
// - If key does not exist: no changes are made.
func ClearFrontmatter(vault obsidian.VaultManager, noteName string, key string) error {
	if noteName == "" {
		return errors.New("note name is required")
	}
	if key == "" {
		return errors.New("key is required")
	}
	if _, err := vault.DefaultName(); err != nil {
		return err
	}
	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}
	targetFile := obsidian.AddMdSuffix(noteName)
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
	info, err := os.Stat(notePath)
	if err != nil {
		return errors.New(obsidian.VaultReadError)
	}
	data, err := os.ReadFile(notePath)
	if err != nil {
		return errors.New(obsidian.VaultReadError)
	}
	fmMap, body, hadFM, err := parseFrontmatter(string(data))
	if err != nil {
		return err
	}
	if fmMap == nil {
		// no FM to clear
		return nil
	}
	if _, exists := fmMap[key]; !exists {
		return nil
	}
	// Determine empty based on current type
	switch fmMap[key].(type) {
	case []interface{}:
		fmMap[key] = []interface{}{}
	case map[string]interface{}:
		fmMap[key] = map[string]interface{}{}
	default:
		fmMap[key] = ""
	}
	yml, err := yaml.Marshal(fmMap)
	if err != nil {
		return errors.New(obsidian.VaultWriteError)
	}
	newContent := ""
	if hadFM {
		newContent = buildWithFrontmatter(string(yml), body)
	} else {
		// shouldn't happen because fmMap!=nil implies hadFM, but keep safe
		newContent = buildWithFrontmatter(string(yml), body)
	}
	if err := os.WriteFile(notePath, []byte(newContent), info.Mode()); err != nil {
		return errors.New(obsidian.VaultWriteError)
	}
	fmt.Printf("Cleared frontmatter key in %s: %s\n", filepath.Base(notePath), key)
	return nil
}

// RemoveFrontmatterKey removes a key (and its value) from frontmatter. If this
// results in an empty frontmatter, the whole frontmatter block is removed.
func RemoveFrontmatterKey(vault obsidian.VaultManager, noteName string, key string) error {
	if noteName == "" {
		return errors.New("note name is required")
	}
	if key == "" {
		return errors.New("key is required")
	}
	if _, err := vault.DefaultName(); err != nil {
		return err
	}
	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}
	targetFile := obsidian.AddMdSuffix(noteName)
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
	info, err := os.Stat(notePath)
	if err != nil {
		return errors.New(obsidian.VaultReadError)
	}
	data, err := os.ReadFile(notePath)
	if err != nil {
		return errors.New(obsidian.VaultReadError)
	}
	fmMap, body, hadFM, err := parseFrontmatter(string(data))
	if err != nil {
		return err
	}
	if fmMap == nil {
		// nothing to remove
		return nil
	}
	if _, exists := fmMap[key]; !exists {
		return nil
	}
	delete(fmMap, key)
	var newContent string
	if len(fmMap) == 0 {
		// remove the whole frontmatter block; ensure body has no leading stray newline
		newContent = body
		// If body begins with a newline due to previous build rules, trim a single leading newline
		if strings.HasPrefix(newContent, "\n") {
			newContent = newContent[1:]
		}
	} else {
		yml, err := yaml.Marshal(fmMap)
		if err != nil {
			return errors.New(obsidian.VaultWriteError)
		}
		if hadFM {
			newContent = buildWithFrontmatter(string(yml), body)
		} else {
			newContent = buildWithFrontmatter(string(yml), body)
		}
	}
	if err := os.WriteFile(notePath, []byte(newContent), info.Mode()); err != nil {
		return errors.New(obsidian.VaultWriteError)
	}
	fmt.Printf("Removed frontmatter key in %s: %s\n", filepath.Base(notePath), key)
	return nil
}

// AddEmptyFrontmatterKey adds the given key to the frontmatter with an empty value.
// If the key already exists, it will be set to an empty value.
// Rules:
// - If key is "tags" (case-insensitive), set to an empty list []
// - Otherwise, set to empty string ""
func AddEmptyFrontmatterKey(vault obsidian.VaultManager, noteName string, key string) error {
	if noteName == "" {
		return errors.New("note name is required")
	}
	if key == "" {
		return errors.New("key is required")
	}
	if _, err := vault.DefaultName(); err != nil {
		return err
	}
	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}
	targetFile := obsidian.AddMdSuffix(noteName)
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
	info, err := os.Stat(notePath)
	if err != nil {
		return errors.New(obsidian.VaultReadError)
	}
	data, err := os.ReadFile(notePath)
	if err != nil {
		return errors.New(obsidian.VaultReadError)
	}
	fmMap, body, _, err := parseFrontmatter(string(data))
	if err != nil {
		return err
	}
	if fmMap == nil {
		fmMap = map[string]any{}
	}
	// Set empty value depending on key semantics
	if strings.EqualFold(key, "tags") {
		fmMap[key] = []interface{}{}
	} else {
		fmMap[key] = ""
	}
	yml, err := yaml.Marshal(fmMap)
	if err != nil {
		return errors.New(obsidian.VaultWriteError)
	}
	newContent := buildWithFrontmatter(string(yml), body)
	if err := os.WriteFile(notePath, []byte(newContent), info.Mode()); err != nil {
		return errors.New(obsidian.VaultWriteError)
	}
	fmt.Printf("Added empty frontmatter key in %s: %s\n", filepath.Base(notePath), key)
	return nil
}
