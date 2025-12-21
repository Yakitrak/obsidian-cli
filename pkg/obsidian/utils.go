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

// normalizePathSeparators converts backslashes to forward slashes for cross-platform consistency.
// Obsidian uses forward slashes in links regardless of OS.
func normalizePathSeparators(path string) string {
	return strings.ReplaceAll(path, "\\", "/")
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

// GenerateLinkReplacements creates all replacement patterns for updating links when moving a note.
// This handles:
// - Simple wikilinks: [[note]], [[note|alias]], [[note#heading]]
// - Path-based wikilinks: [[folder/note]], [[folder/note|alias]], [[folder/note#heading]]
// - Markdown links: [text](folder/note.md), [text](./folder/note.md)
func GenerateLinkReplacements(oldNotePath, newNotePath string) map[string]string {
	replacements := make(map[string]string)

	// Normalize paths to forward slashes for consistent matching
	oldNormalized := normalizePathSeparators(oldNotePath)
	newNormalized := normalizePathSeparators(newNotePath)

	// Get basename without .md extension
	oldBase := RemoveMdSuffix(filepath.Base(oldNotePath))
	newBase := RemoveMdSuffix(filepath.Base(newNotePath))

	// Get full path without .md extension
	oldPathNoExt := RemoveMdSuffix(oldNormalized)
	newPathNoExt := RemoveMdSuffix(newNormalized)

	// 1. Simple wikilinks (basename only) - for backward compatibility
	replacements["[["+oldBase+"]]"] = "[[" + newBase + "]]"
	replacements["[["+oldBase+"|"] = "[[" + newBase + "|"
	replacements["[["+oldBase+"#"] = "[[" + newBase + "#"

	// 2. Path-based wikilinks (only if path differs from basename)
	if oldPathNoExt != oldBase {
		replacements["[["+oldPathNoExt+"]]"] = "[[" + newPathNoExt + "]]"
		replacements["[["+oldPathNoExt+"|"] = "[[" + newPathNoExt + "|"
		replacements["[["+oldPathNoExt+"#"] = "[[" + newPathNoExt + "#"
	}

	// 3. Markdown links (various formats)
	oldMd := AddMdSuffix(oldNormalized)
	newMd := AddMdSuffix(newNormalized)

	// Standard markdown link: [text](folder/note.md)
	replacements["]("+oldMd+")"] = "](" + newMd + ")"
	replacements["]("+oldPathNoExt+")"] = "](" + newPathNoExt + ")"

	// Relative markdown link: [text](./folder/note.md)
	replacements["](./"+oldMd+")"] = "](./" + newMd + ")"
	replacements["](./"+oldPathNoExt+")"] = "](./" + newPathNoExt + ")"

	return replacements
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
