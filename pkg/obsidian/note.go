package obsidian

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
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
	SetContents(string, string, string) error
	GetNotesList(string) ([]string, error)
	SearchNotesWithSnippets(string, string) ([]NoteMatch, error)
	FindBacklinks(string, string) ([]NoteMatch, error)
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

func (m *Note) SetContents(vaultPath string, noteName string, content string) error {
	note := AddMdSuffix(noteName)

	var notePath string
	err := filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
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
		return errors.New(NoteDoesNotExistError)
	}

	err = os.WriteFile(notePath, []byte(content), 0644)
	if err != nil {
		return errors.New(VaultWriteError)
	}

	return nil
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

const maxFileSizeBytes = 10 * 1024 * 1024 // 10MB

// containsAnyPattern checks if content contains any of the patterns (case-insensitive).
func containsAnyPattern(contentLower []byte, patterns [][]byte) bool {
	for _, pattern := range patterns {
		if bytes.Contains(contentLower, pattern) {
			return true
		}
	}
	return false
}

// findMatchingLines finds all lines containing any of the patterns.
func findMatchingLines(content []byte, patternsLower [][]byte) []NoteMatch {
	var matches []NoteMatch
	lineNum := 0

	for len(content) > 0 {
		lineNum++

		// Find end of line
		idx := bytes.IndexByte(content, '\n')
		var line []byte
		if idx == -1 {
			line = content
			content = nil
		} else {
			line = content[:idx]
			content = content[idx+1:]
		}

		// Check if line matches any pattern
		lineLower := bytes.ToLower(line)
		for _, pattern := range patternsLower {
			if bytes.Contains(lineLower, pattern) {
				matches = append(matches, NoteMatch{
					LineNumber: lineNum,
					MatchLine:  string(bytes.TrimSpace(line)),
				})
				break
			}
		}
	}

	return matches
}

func (m *Note) FindBacklinks(vaultPath, noteName string) ([]NoteMatch, error) {
	noteName = RemoveMdSuffix(noteName)

	// Generate patterns and convert to lowercase bytes once
	patterns := GenerateBacklinkSearchPatterns(noteName)
	patternsLower := make([][]byte, len(patterns))
	for i, p := range patterns {
		patternsLower[i] = []byte(strings.ToLower(p))
	}

	var matches []NoteMatch
	fileModTimes := make(map[string]int64)

	err := filepath.WalkDir(vaultPath, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || !strings.HasSuffix(d.Name(), ".md") {
			return nil
		}

		relPath, err := filepath.Rel(vaultPath, path)
		if err != nil {
			return err
		}

		// Skip the note itself (normalize for comparison)
		if RemoveMdSuffix(normalizePathSeparators(relPath)) == noteName {
			return nil
		}

		info, err := d.Info()
		if err != nil {
			return nil
		}
		if info.Size() > maxFileSizeBytes {
			fmt.Fprintf(os.Stderr, "Skipping file %s: size %d bytes exceeds limit %d bytes\n", relPath, info.Size(), maxFileSizeBytes)
			return nil
		}

		content, err := os.ReadFile(path)
		if err != nil {
			return nil
		}

		// Quick check: skip file if it doesn't contain any pattern
		contentLower := bytes.ToLower(content)
		if !containsAnyPattern(contentLower, patternsLower) {
			return nil
		}

		// Find matching lines
		modTime := info.ModTime().UnixNano()
		fileMatches := findMatchingLines(content, patternsLower)
		for i := range fileMatches {
			fileMatches[i].FilePath = relPath
			fileModTimes[relPath] = modTime
		}
		matches = append(matches, fileMatches...)

		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(matches, func(i, j int) bool {
		return fileModTimes[matches[i].FilePath] > fileModTimes[matches[j].FilePath]
	})

	return matches, nil
}
