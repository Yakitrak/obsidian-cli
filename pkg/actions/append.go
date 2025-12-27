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

type DailyAppendPlan struct {
	VaultName         string
	VaultPath         string
	Folder            string
	Pattern           string
	Filename          string
	RelativeNotePath  string
	AbsoluteNotePath  string
	WillCreateDirs    bool
	WillCreateFile    bool
	TemplateRel       string
	TemplateAbs       string
	WillApplyTemplate bool
}

func PlanDailyAppend(vault *obsidian.Vault, now time.Time) (DailyAppendPlan, error) {
	vaultName, err := vault.DefaultName()
	if err != nil {
		return DailyAppendPlan{}, err
	}

	vaultPath, err := vault.Path()
	if err != nil {
		return DailyAppendPlan{}, err
	}

	settings, err := vault.Settings()
	if err != nil {
		return DailyAppendPlan{}, err
	}

	folder := strings.TrimSpace(settings.DailyNote.Folder)
	if folder == "" {
		return DailyAppendPlan{}, errors.New("daily note is not configured (missing daily_note.folder in preferences.json)")
	}

	pattern := strings.TrimSpace(settings.DailyNote.FilenamePattern)
	if pattern == "" {
		pattern = "{YYYY-MM-DD}"
	}
	filename, err := obsidian.FormatDatePattern(pattern, now)
	if err != nil {
		return DailyAppendPlan{}, err
	}

	relNotePath := filepath.ToSlash(filepath.Join(folder, filename))
	abs, err := safeJoinVaultPath(vaultPath, relNotePath)
	if err != nil {
		return DailyAppendPlan{}, err
	}
	if !strings.HasSuffix(strings.ToLower(abs), ".md") {
		abs += ".md"
	}

	plan := DailyAppendPlan{
		VaultName:        vaultName,
		VaultPath:        vaultPath,
		Folder:           folder,
		Pattern:          pattern,
		Filename:         filename,
		RelativeNotePath: relNotePath,
		AbsoluteNotePath: abs,
	}

	if _, err := os.Stat(filepath.Dir(abs)); err != nil {
		if os.IsNotExist(err) {
			plan.WillCreateDirs = true
		} else {
			return DailyAppendPlan{}, err
		}
	}

	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			plan.WillCreateFile = true
		} else {
			return DailyAppendPlan{}, err
		}
	}

	templateRel := strings.TrimSpace(settings.DailyNote.TemplatePath)
	if templateRel != "" && plan.WillCreateFile {
		templateAbs, err := obsidian.SafeJoinVaultPath(vaultPath, filepath.ToSlash(templateRel))
		if err != nil {
			return DailyAppendPlan{}, fmt.Errorf("invalid template path: %w", err)
		}
		if !strings.HasSuffix(strings.ToLower(templateAbs), ".md") {
			templateAbs += ".md"
		}
		if _, err := os.Stat(templateAbs); err != nil {
			return DailyAppendPlan{}, fmt.Errorf("failed to read template: %w", err)
		}

		plan.TemplateRel = templateRel
		plan.TemplateAbs = templateAbs
		plan.WillApplyTemplate = true
	}

	return plan, nil
}

// AppendToDailyNote appends content to today's daily note, using per-vault settings.
func AppendToDailyNote(vault *obsidian.Vault, content string) error {
	content = strings.TrimSpace(content)
	if content == "" {
		return errors.New("no content provided")
	}

	now := time.Now()

	plan, err := PlanDailyAppend(vault, now)
	if err != nil {
		return err
	}
	abs := plan.AbsoluteNotePath

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

		if plan.WillApplyTemplate {
			b, err := os.ReadFile(plan.TemplateAbs)
			if err != nil {
				return fmt.Errorf("failed to read template: %w", err)
			}
			title := filepath.Base(abs)
			b = obsidian.ExpandTemplateVariablesAt(b, title, now)
			existing = b
		}
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
	return obsidian.SafeJoinVaultPath(vaultPath, relativePath)
}
