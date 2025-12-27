package obsidian

import (
	"bytes"
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

type NoteMatch struct {
	FilePath   string
	LineNumber int
	MatchLine  string
}

type NoteManager interface {
	Move(string, string) error
	Delete(string) error
	UpdateLinks(string, string, string) error
	GetContents(string, string) (string, error)
	GetNotesList(string) ([]string, error)
	SearchNotesWithSnippets(string, string) ([]NoteMatch, error)
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

		// Check for full path match first
		relPath, err := filepath.Rel(vaultPath, path)
		if err == nil && relPath == note {
			notePath = path
			return filepath.SkipDir
		}

		// Fall back to basename match for backward compatibility
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

		originalContent, err := os.ReadFile(path)
		if err != nil {
			return errors.New(VaultReadError)
		}

		replacements := GenerateLinkReplacements(oldNoteName, newNoteName)
		updatedContent := ReplaceContent(originalContent, replacements)

		if bytes.Equal(originalContent, updatedContent) {
			return nil
		}

		err = os.WriteFile(path, updatedContent, info.Mode())
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

func (m *Note) SearchNotesWithSnippets(vaultPath string, query string) ([]NoteMatch, error) {
	var matches []NoteMatch
	queryLower := strings.ToLower(query)

	err := filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && strings.HasSuffix(d.Name(), ".md") {
			relPath, err := filepath.Rel(vaultPath, path)
			if err != nil {
				return err
			}

			fileNameMatches := strings.Contains(strings.ToLower(relPath), queryLower)
			var hasContentMatch bool

			// Check file size to avoid reading very large files (>10MB)
			if info, err := d.Info(); err == nil && info.Size() < 10*1024*1024 {
				content, err := os.ReadFile(path)
				if err == nil {
					lines := strings.Split(string(content), "\n")
					for lineNum, line := range lines {
						if strings.Contains(strings.ToLower(line), queryLower) {
							hasContentMatch = true
							matchLine := strings.TrimSpace(line)
							if len(matchLine) > 80 {
								// Find the query position and center around it
								queryPos := strings.Index(strings.ToLower(matchLine), queryLower)
								if queryPos != -1 {
									start := queryPos - 20
									end := queryPos + len(query) + 20
									if start < 0 {
										start = 0
									}
									if end > len(matchLine) {
										end = len(matchLine)
									}
									if start > 0 {
										matchLine = "..." + matchLine[start:]
									}
									if end < len(strings.TrimSpace(line)) {
										matchLine = matchLine[:end-start] + "..."
									}
								} else {
									matchLine = matchLine[:80] + "..."
								}
							}

							matches = append(matches, NoteMatch{
								FilePath:   relPath,
								LineNumber: lineNum + 1,
								MatchLine:  matchLine,
							})
						}
					}
				}
			}

			// Only add filename match if there are no content matches
			if fileNameMatches && !hasContentMatch {
				matches = append(matches, NoteMatch{
					FilePath:   relPath,
					LineNumber: 0,
					MatchLine:  fmt.Sprintf("(filename match: %s)", filepath.Base(relPath)),
				})
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return matches, nil
}
