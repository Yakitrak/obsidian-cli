package obsidian

import (
	"bytes"
	"os"
	"os/exec"
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

// OpenInEditor opens the specified file path in the user's preferred editor
func OpenInEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Default fallback
	}
	
	cmd := exec.Command(editor, filePath)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	
	return cmd.Run()
}
