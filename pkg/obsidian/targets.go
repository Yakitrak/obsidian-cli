package obsidian

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"gopkg.in/yaml.v3"
)

// Target represents a named capture target.
// Targets are configured in targets.yaml stored next to preferences.json.
type Target struct {
	Type     string `yaml:"type,omitempty"`           // "folder" or "file"
	Folder   string `yaml:"folder,omitempty"`         // for folder targets
	Pattern  string `yaml:"pattern,omitempty"`        // for folder targets (Obsidian-style or legacy tokens)
	Template string `yaml:"template,omitempty"`       // optional template path
	File     string `yaml:"file,omitempty"`           // for file targets
	Note     string `yaml:"note,omitempty"`           // legacy/simplified key (treated as file)
	Vault    string `yaml:"vault,omitempty"`          // optional: override default vault
	Format   string `yaml:"pattern_format,omitempty"` // "auto" (default), "obsidian", or "tokens"
}

func (t *Target) UnmarshalYAML(value *yaml.Node) error {
	switch value.Kind {
	case yaml.ScalarNode:
		// Simplified format: id: "Some/Path" → treat as file target.
		t.Type = "file"
		t.File = strings.TrimSpace(value.Value)
		return nil
	case yaml.MappingNode:
		type raw Target
		var r raw
		if err := value.Decode(&r); err != nil {
			return err
		}
		*t = Target(r)
		if strings.TrimSpace(t.File) == "" && strings.TrimSpace(t.Note) != "" {
			t.File = t.Note
			if strings.TrimSpace(t.Type) == "" {
				t.Type = "file"
			}
		}
		if strings.TrimSpace(t.Type) == "" {
			// Infer if possible.
			if strings.TrimSpace(t.Folder) != "" || strings.TrimSpace(t.Pattern) != "" {
				t.Type = "folder"
			} else if strings.TrimSpace(t.File) != "" {
				t.Type = "file"
			}
		}
		return nil
	default:
		return fmt.Errorf("invalid target definition")
	}
}

// TargetsConfig is the map of target name to Target.
type TargetsConfig map[string]Target

var reservedTargetNames = map[string]bool{
	"add":      true,
	"remove":   true,
	"rm":       true,
	"list":     true,
	"ls":       true,
	"edit":     true,
	"validate": true,
	"test":     true,
	"help":     true,
}

type TargetPlan struct {
	TargetName        string
	TargetType        string
	VaultName         string
	VaultPath         string
	RelativeNotePath  string
	AbsoluteNotePath  string
	TemplatePath      string
	AbsoluteTemplate  string
	WillCreateDirs    bool
	WillCreateFile    bool
	WillApplyTemplate bool
}

func TargetsFile() (string, error) {
	_, path, err := config.TargetsPath()
	return path, err
}

func EnsureTargetsFileExists() (string, error) {
	dir, path, err := config.TargetsPath()
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(dir, 0750); err != nil {
		return "", err
	}
	if _, err := os.Stat(path); err == nil {
		return path, nil
	} else if err != nil && !os.IsNotExist(err) {
		return "", err
	}
	if err := os.WriteFile(path, []byte(targetsFileHeader), 0600); err != nil {
		return "", err
	}
	return path, nil
}

// LoadTargets loads targets from targets.yaml.
func LoadTargets() (TargetsConfig, error) {
	path, err := TargetsFile()
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	cfg := TargetsConfig{}
	if err := yaml.Unmarshal(raw, &cfg); err != nil {
		return nil, err
	}
	return cfg, nil
}

// SaveTargets saves targets to targets.yaml.
func SaveTargets(cfg TargetsConfig) error {
	path, err := EnsureTargetsFileExists()
	if err != nil {
		return err
	}
	out, err := yaml.Marshal(cfg)
	if err != nil {
		return err
	}
	content := append([]byte(targetsFileHeader), out...)
	return os.WriteFile(path, content, 0600)
}

func ListTargetNames(cfg TargetsConfig) []string {
	var names []string
	for name := range cfg {
		names = append(names, name)
	}
	sort.Slice(names, func(i, j int) bool { return strings.ToLower(names[i]) < strings.ToLower(names[j]) })
	return names
}

func ValidateTargetName(name string) error {
	n := strings.TrimSpace(name)
	if n == "" {
		return errors.New("target name is required")
	}
	if strings.ContainsAny(n, " \t\r\n") {
		return errors.New("target name cannot contain whitespace")
	}
	if reservedTargetNames[strings.ToLower(n)] {
		return fmt.Errorf("target name '%s' is reserved", n)
	}
	return nil
}

func ValidateTarget(t Target) error {
	tt := strings.ToLower(strings.TrimSpace(t.Type))
	switch tt {
	case "file":
		if strings.TrimSpace(t.File) == "" {
			return errors.New("file target requires 'file' field")
		}
		return nil
	case "folder":
		if strings.TrimSpace(t.Folder) == "" {
			return errors.New("folder target requires 'folder' field")
		}
		if strings.TrimSpace(t.Pattern) == "" {
			return errors.New("folder target requires 'pattern' field")
		}
		return nil
	default:
		if tt == "" {
			return errors.New("target type is required: must be 'folder' or 'file'")
		}
		return fmt.Errorf("invalid target type '%s': must be 'folder' or 'file'", t.Type)
	}
}

func ResolveTargetNotePath(vaultPath string, target Target, now time.Time) (relative string, absolute string, err error) {
	if err := ValidateTarget(target); err != nil {
		return "", "", err
	}

	switch strings.ToLower(strings.TrimSpace(target.Type)) {
	case "file":
		relative = strings.TrimSpace(target.File)
	case "folder":
		filename := ExpandDatePattern(strings.TrimSpace(target.Pattern), now)
		if strings.TrimSpace(filename) == "" {
			return "", "", errors.New("expanded filename is empty")
		}
		relative = filepath.ToSlash(filepath.Join(strings.TrimSpace(target.Folder), filename))
	}

	absolute, err = SafeJoinVaultPath(vaultPath, filepath.ToSlash(relative))
	if err != nil {
		return "", "", err
	}
	if !strings.HasSuffix(strings.ToLower(absolute), ".md") {
		absolute += ".md"
	}
	return relative, absolute, nil
}

func PlanTargetAppend(vaultPath string, targetName string, target Target, now time.Time) (TargetPlan, error) {
	rel, abs, err := ResolveTargetNotePath(vaultPath, target, now)
	if err != nil {
		return TargetPlan{}, err
	}

	plan := TargetPlan{
		TargetName:       targetName,
		TargetType:       strings.ToLower(strings.TrimSpace(target.Type)),
		RelativeNotePath: rel,
		AbsoluteNotePath: abs,
	}

	if _, err := os.Stat(filepath.Dir(abs)); err != nil {
		if os.IsNotExist(err) {
			plan.WillCreateDirs = true
		} else {
			return TargetPlan{}, err
		}
	}

	if _, err := os.Stat(abs); err != nil {
		if os.IsNotExist(err) {
			plan.WillCreateFile = true
		} else {
			return TargetPlan{}, err
		}
	}

	templateRel := strings.TrimSpace(target.Template)
	if plan.WillCreateFile && templateRel != "" {
		templateAbs, err := SafeJoinVaultPath(vaultPath, filepath.ToSlash(templateRel))
		if err != nil {
			return TargetPlan{}, fmt.Errorf("invalid template path: %w", err)
		}
		if !strings.HasSuffix(strings.ToLower(templateAbs), ".md") {
			templateAbs += ".md"
		}
		plan.TemplatePath = templateRel
		plan.AbsoluteTemplate = templateAbs
		plan.WillApplyTemplate = true
	}

	return plan, nil
}

const targetsFileHeader = `# Obsidian CLI Targets
#
# This file defines capture targets for:
#   obsidian-cli target <id> [text]
#
# Targets can be:
#   - type: file   (append to a fixed note)
#   - type: folder (append to a dated note determined by folder + pattern)
#
# Fields:
#   type:     "file" or "folder"
#   file:     (file target) path relative to vault root
#   folder:   (folder target) folder path relative to vault root
#   pattern:  (folder target) filename pattern, without ".md"
#   template: (optional) note path to use as initial content when creating a new file
#   vault:    (optional) override the default vault
#
# Pattern format:
#   - Supports Obsidian-style tokens like YYYY, MM, DD, HH, mm, ss, ddd, dddd, MMM, MMMM
#   - Supports [literal] blocks, e.g. YYYY-[log]-MM
#   - Supports the "zettel" timestamp: YYYYMMDDHHmmss
#
# Examples:
#   inbox:
#     type: file
#     file: Inbox.md
#
#   hourly-log:
#     type: folder
#     folder: Logs
#     pattern: YYYY-MM-DD_HH
#
`
