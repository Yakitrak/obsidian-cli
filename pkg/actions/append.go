package actions

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
)

// AppendToDailyNote appends content to today's daily note, using per-vault settings.
func AppendToDailyNote(vault *obsidian.Vault, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("no content provided")
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	settings, err := vault.Settings()
	if err != nil {
		return err
	}

	folder := strings.TrimSpace(settings.DailyNote.Folder)
	if folder == "" {
		return errors.New("daily note is not configured (missing daily_note.folder in preferences.json)")
	}

	pattern := strings.TrimSpace(settings.DailyNote.FilenamePattern)
	if pattern == "" {
		pattern = "{YYYY-MM-DD}"
	}

	filename := expandDailyFilename(pattern, time.Now())
	relNotePath := filepath.ToSlash(filepath.Join(folder, filename))
	abs, err := safeJoinVaultPath(vaultPath, relNotePath)
	if err != nil {
		return err
	}
	if !strings.HasSuffix(strings.ToLower(abs), ".md") {
		abs += ".md"
	}

	if err := os.MkdirAll(filepath.Dir(abs), 0750); err != nil {
		return fmt.Errorf("failed to create note directory: %w", err)
	}

	mode := os.FileMode(0600)
	if info, err := os.Stat(abs); err == nil {
		mode = info.Mode()
	}

	existing, err := os.ReadFile(abs)
	if err != nil {
		if !os.IsNotExist(err) {
			return fmt.Errorf("failed to read note: %w", err)
		}
		existing = []byte{}
	}

	next := appendWithSeparator(string(existing), content)
	return os.WriteFile(abs, []byte(next), mode)
}

// PromptForContentIfEmpty returns content if non-empty, otherwise reads stdin (piped) or prompts
// for multi-line input until EOF.
func PromptForContentIfEmpty(content string) (string, error) {
	content = strings.TrimSpace(content)
	if content != "" {
		return content, nil
	}

	stat, err := os.Stdin.Stat()
	if err == nil && (stat.Mode()&os.ModeCharDevice) == 0 {
		b, err := io.ReadAll(os.Stdin)
		if err != nil {
			return "", err
		}
		s := strings.TrimSpace(string(b))
		if s == "" {
			return "", errors.New("no content provided")
		}
		return s, nil
	}

	fmt.Println("Enter text to append (Ctrl-D to save, Ctrl-C to cancel):")
	in := bufio.NewScanner(os.Stdin)
	var lines []string
	for in.Scan() {
		lines = append(lines, in.Text())
	}
	if err := in.Err(); err != nil {
		return "", err
	}
	s := strings.TrimSpace(strings.Join(lines, "\n"))
	if s == "" {
		return "", errors.New("no content provided")
	}
	return s, nil
}

func expandDailyFilename(pattern string, now time.Time) string {
	out := pattern
	out = strings.ReplaceAll(out, "{YYYY-MM-DD}", now.Format("2006-01-02"))
	out = strings.ReplaceAll(out, "YYYY-MM-DD", now.Format("2006-01-02"))
	return out
}

func appendWithSeparator(existing string, addition string) string {
	existing = strings.TrimRight(existing, "\n")
	addition = strings.TrimSpace(addition)
	if addition == "" {
		return existing + "\n"
	}
	if existing == "" {
		return addition + "\n"
	}
	return existing + "\n\n" + addition + "\n"
}

// safeJoinVaultPath joins a vault root and a relative note path and ensures the result stays within the vault.
func safeJoinVaultPath(vaultPath string, relativePath string) (string, error) {
	if filepath.IsAbs(relativePath) {
		return "", fmt.Errorf("absolute paths are not allowed: %s", relativePath)
	}
	cleaned := filepath.Clean(strings.TrimSpace(relativePath))
	cleaned = strings.TrimPrefix(cleaned, string(filepath.Separator))
	cleaned = strings.TrimPrefix(cleaned, "./")
	if cleaned == "" || cleaned == "." {
		return "", fmt.Errorf("note path cannot be empty")
	}

	absVault, err := filepath.Abs(vaultPath)
	if err != nil {
		return "", fmt.Errorf("failed to resolve vault path: %w", err)
	}

	joined := filepath.Join(absVault, filepath.FromSlash(cleaned))
	absJoined, err := filepath.Abs(joined)
	if err != nil {
		return "", fmt.Errorf("failed to resolve note path: %w", err)
	}

	if absJoined != absVault && !strings.HasPrefix(absJoined, absVault+string(filepath.Separator)) {
		return "", fmt.Errorf("note path escapes vault: %s", relativePath)
	}

	return absJoined, nil
}
