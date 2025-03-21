package obsidian

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
)

func AddMdSuffix(str string) string {
	if !strings.HasSuffix(str, ".md") {
		return str + ".md"
	}
	return str
}

func RemoveMdSuffix(str string) string {
	if strings.HasSuffix(str, ".md") {
		return strings.TrimSuffix(str, ".md")
	}
	return str
}

func GenerateNoteLinkTexts(noteName string) [3]string {
	var noteLinkTexts [3]string
	noteName = filepath.Base(noteName)
	noteName = RemoveMdSuffix(noteName)
	noteLinkTexts[0] = "[[" + noteName + "]]"
	noteLinkTexts[1] = "[[" + noteName + "|"
	noteLinkTexts[2] = "[[" + noteName + "#"
	return noteLinkTexts
}

func ReplaceContent(content []byte, replacements map[string]string) []byte {
	for o, n := range replacements {
		content = bytes.ReplaceAll(content, []byte(o), []byte(n))
	}
	return content
}

func ShouldSkipDirectoryOrFile(info os.FileInfo) bool {
	isDirectory := info.IsDir()
	isHidden := info.Name()[0] == '.'
	isNonMarkdownFile := filepath.Ext(info.Name()) != ".md"
	if isDirectory || isHidden || isNonMarkdownFile {
		return true
	}
	return false
}

// DeduplicateResults removes duplicate entries from a slice of strings
func DeduplicateResults(results []string) []string {
	seen := make(map[string]bool)
	var unique []string
	for _, result := range results {
		if !seen[result] {
			seen[result] = true
			unique = append(unique, result)
		}
	}
	return unique
}

// NormalizePath normalizes a path for comparison
func NormalizePath(path string) string {
	// Convert all path separators to forward slashes
	path = strings.ReplaceAll(path, "\\", "/")
	path = strings.TrimPrefix(path, "./")
	path = strings.TrimPrefix(path, "../")
	return path
}
