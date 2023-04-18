package note

import (
	"os"
	"path/filepath"
)

// Skip hidden directories and non-markdown files
func ShouldSkipDirectoryOrFile(info os.FileInfo) bool {
	isDirectory := info.IsDir()
	isHidden := info.Name()[0] == '.'
	isNonMarkdownFile := filepath.Ext(info.Name()) != ".md"

	if isDirectory || isHidden || isNonMarkdownFile {
		return true
	}

	return false
}
