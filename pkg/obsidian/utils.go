package obsidian

import (
	"bytes"
	"fmt"
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
// It supports common GUI editors with appropriate wait flags
func OpenInEditor(filePath string) error {
	editor := os.Getenv("EDITOR")
	if editor == "" {
		editor = "vim" // Default fallback
	}

	// Parse editor command to handle GUI editors that need wait flags
	var cmd *exec.Cmd
	editorLower := strings.ToLower(filepath.Base(editor))
	
	switch {
	case strings.Contains(editorLower, "code") || strings.Contains(editorLower, "vscode"):
		// VSCode needs --wait flag to block
		cmd = exec.Command(editor, "--wait", filePath)
	case strings.Contains(editorLower, "subl"):
		// Sublime Text needs --wait flag
		cmd = exec.Command(editor, "--wait", filePath)
	case strings.Contains(editorLower, "atom"):
		// Atom needs --wait flag
		cmd = exec.Command(editor, "--wait", filePath)
	case strings.Contains(editorLower, "mate"):
		// TextMate needs --wait flag
		cmd = exec.Command(editor, "--wait", filePath)
	default:
		// For vim, nano, emacs, and other terminal editors, or unknown editors
		cmd = exec.Command(editor, filePath)
	}

	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to open file in editor '%s': %w", editor, err)
	}

	return nil
}
