package cmd

import (
	"bufio"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/config"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:     "init",
	Short:   "Interactive setup wizard",
	Long:    "Interactive setup wizard to configure your default vault, daily note settings, and targets.",
	Args:    cobra.NoArgs,
	Example: "  obsidian-cli init",
	RunE: func(cmd *cobra.Command, args []string) error {
		in := bufio.NewReader(os.Stdin)

		fmt.Println()
		fmt.Println(DoubleLine(DefaultBorderWidth))
		fmt.Println("                            OBSIDIAN-CLI SETUP WIZARD                        ")
		fmt.Println(DoubleLine(DefaultBorderWidth))
		fmt.Println()

		return runInitWizard(in)
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

type discoveredVault struct {
	Name string
	Path string
	ID   string
}

func runInitWizard(in *bufio.Reader) error {
	var vaultName string
	var vaultPath string

	var settings obsidian.VaultSettings
	var daily obsidian.DailyNoteSettings

vaultSelection:
	// Step 1: vault selection
	for {
		fmt.Println(SingleLine(DefaultBorderWidth))
		fmt.Println("  Step 1: Default vault")
		fmt.Println(SingleLine(DefaultBorderWidth))
		fmt.Println("Type '?' for help, or 'back' to cancel.")
		fmt.Println()

		vaults, err := discoverVaults()
		if err != nil || len(vaults) == 0 {
			if err != nil {
				fmt.Printf("Could not discover vaults: %v\n", err)
			} else {
				fmt.Println("No vaults discovered.")
			}
			fmt.Println("Enter the vault name (usually the folder name), or type '?' for help.")
			name, action, err := promptLine(in, "> ")
			if err != nil {
				return err
			}
			switch action {
			case actionBack:
				return errors.New("cancelled")
			case actionHelp:
				printVaultHelp()
				continue
			default:
			}
			name = strings.TrimSpace(name)
			if name == "" {
				fmt.Println("Vault name is required.")
				continue
			}
			vault := obsidian.Vault{Name: name}
			path, err := vault.Path()
			if err != nil {
				fmt.Printf("Vault not found: %v\n", err)
				fmt.Println("Tip: ensure Obsidian has created an obsidian.json with your vaults.")
				continue
			}
			vaultName = name
			vaultPath = path
			break
		}

		fmt.Println("Discovered vaults:")
		for i, v := range vaults {
			fmt.Printf("  %d) %s\n", i+1, v.Name)
			fmt.Printf("     %s\n", v.Path)
		}
		fmt.Println()
		fmt.Println("Choose a vault:")
		fmt.Println("  - Enter a number")
		fmt.Println("  - Type the exact name")
		fmt.Println("  - Type 'select' to choose with fuzzy finder")

		line, action, err := promptLine(in, "> ")
		if err != nil {
			return err
		}
		switch action {
		case actionBack:
			return errors.New("cancelled")
		case actionSkip:
			fmt.Println("Skip is not available here. Choose a vault or type 'back' to cancel.")
			continue
		case actionHelp:
			printVaultHelp()
			continue
		default:
		}

		line = strings.TrimSpace(line)
		if strings.EqualFold(line, "select") {
			name, path, err := pickVaultFuzzy(vaults)
			if err != nil {
				fmt.Printf("Selection cancelled: %v\n", err)
				continue
			}
			vaultName = name
			vaultPath = path
			break
		}

		if n, ok := parseNumber(line); ok {
			if n < 1 || n > len(vaults) {
				fmt.Printf("Invalid selection: enter 1-%d\n", len(vaults))
				continue
			}
			vaultName = vaults[n-1].Name
			vaultPath = vaults[n-1].Path
			break
		}

		if name, path, ok := matchVaultName(vaults, line); ok {
			vaultName = name
			vaultPath = path
			break
		}

		fmt.Printf("Vault '%s' not found. Try again.\n", line)
	}

	// Confirm and write default vault name
	for {
		fmt.Println()
		fmt.Printf("Set default vault to: %s\n", vaultName)
		fmt.Println("This writes to preferences.json in your user config directory.")
		fmt.Println("Continue? (y/N)  Type 'back' to re-select vault.")
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return err
		}
		if action == actionBack {
			goto vaultSelection
		}
		if action == actionHelp {
			fmt.Println("This sets which vault is used when --vault is not provided.")
			continue
		}
		if action == actionSkip {
			fmt.Println("Skip is not available here. Answer 'y' to write preferences.json, or type 'back' to re-select.")
			continue
		}
		if !isYes(line) {
			return errors.New("cancelled")
		}

		v := obsidian.Vault{}
		if err := v.SetDefaultName(vaultName); err != nil {
			fmt.Printf("Failed to write preferences.json: %v\n", err)
			fmt.Println("Tip: check permissions on your user config directory.")
			continue
		}
		break
	}

	vault := obsidian.Vault{Name: vaultName}

dailySettings:
	// Step 2: daily folder
	settings, _ = vault.Settings()
	daily = settings.DailyNote

	folder, err := promptDailyFolder(in, vaultPath, daily.Folder)
	if err != nil {
		if errors.Is(err, errBack) {
			goto vaultSelection
		}
		return err
	}
	daily.Folder = folder

	// Step 3: pattern
	pattern, err := promptDailyPattern(in, daily.FilenamePattern)
	if err != nil {
		if errors.Is(err, errBack) {
			goto dailySettings
		}
		return err
	}
	daily.FilenamePattern = pattern

	// Step 4: template
	templatePath, err := promptDailyTemplate(in, vaultPath, daily.TemplatePath)
	if err != nil {
		if errors.Is(err, errBack) {
			goto dailySettings
		}
		return err
	}
	daily.TemplatePath = templatePath

	// Step 5: save daily settings
	daily.CreateIfMissing = true
	settings.DailyNote = daily

	fmt.Println()
	fmt.Println(SingleLine(DefaultBorderWidth))
	fmt.Println("  Step 2: Daily note settings summary")
	fmt.Println(SingleLine(DefaultBorderWidth))
	fmt.Printf("  folder:   %s\n", daily.Folder)
	fmt.Printf("  pattern:  %s\n", daily.FilenamePattern)
	fmt.Printf("  example:  %s.md\n", obsidian.ExpandDatePattern(daily.FilenamePattern, time.Now()))
	if strings.TrimSpace(daily.TemplatePath) == "" {
		fmt.Println("  template: (none)")
	} else {
		fmt.Printf("  template: %s\n", daily.TemplatePath)
	}
	fmt.Println()
	fmt.Println("Save these daily note settings? (y/N)  Type 'back' to edit.")

	for {
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return err
		}
		if action == actionBack {
			goto dailySettings
		}
		if action == actionHelp {
			printDailyHelp()
			continue
		}
		if !isYes(line) {
			return errors.New("cancelled")
		}
		if err := vault.SetSettings(settings); err != nil {
			return err
		}
		break
	}

targetsStep:
	// Step 6: targets migration / setup
	if err := runTargetsMigration(in); err != nil {
		if errors.Is(err, errBack) {
			goto dailySettings
		}
		return err
	}

	// Step 7: optionally add a first target
	if err := runInitAddTargets(in, vaultPath); err != nil {
		if errors.Is(err, errBack) {
			goto targetsStep
		}
		return err
	}

	// Summary
	fmt.Println()
	fmt.Println(DoubleLine(DefaultBorderWidth))
	fmt.Println("                                  SETUP COMPLETE                              ")
	fmt.Println(DoubleLine(DefaultBorderWidth))
	fmt.Println()
	fmt.Printf("Default vault: %s\n", vaultName)
	fmt.Printf("Vault path:    %s\n", vaultPath)
	fmt.Println()
	fmt.Println("Next steps:")
	fmt.Println("  - Append to today's daily note:")
	fmt.Println("      obsidian-cli append \"some text\"")
	fmt.Println("      obsidian-cli a \"some text\"")
	fmt.Println("  - Append multi-line content (Ctrl-D to save, Ctrl-C to cancel):")
	fmt.Println("      obsidian-cli append")
	fmt.Println("  - Capture to a target:")
	fmt.Println("      obsidian-cli target --select")
	fmt.Println("      obsidian-cli target inbox \"some text\"")
	return nil
}

func parseNumber(s string) (int, bool) {
	var n int
	_, err := fmt.Sscanf(strings.TrimSpace(s), "%d", &n)
	return n, err == nil
}

func discoverVaults() ([]discoveredVault, error) {
	obsidianConfigFile, err := config.ObsidianFile()
	if err != nil {
		return nil, err
	}
	content, err := os.ReadFile(obsidianConfigFile)
	if err != nil {
		return nil, err
	}
	var cfg struct {
		Vaults map[string]struct {
			Path string `json:"path"`
		} `json:"vaults"`
	}
	if err := json.Unmarshal(content, &cfg); err != nil {
		return nil, err
	}
	if len(cfg.Vaults) == 0 {
		return nil, errors.New("no vaults found in obsidian.json")
	}
	var out []discoveredVault
	for id, v := range cfg.Vaults {
		p := v.Path
		if strings.TrimSpace(p) == "" {
			continue
		}
		base := filepath.Base(filepath.Clean(p))
		if base == "." || base == string(filepath.Separator) || base == "" {
			continue
		}
		out = append(out, discoveredVault{Name: base, Path: p, ID: id})
	}
	sort.Slice(out, func(i, j int) bool { return strings.ToLower(out[i].Name) < strings.ToLower(out[j].Name) })
	return out, nil
}

func matchVaultName(vaults []discoveredVault, input string) (string, string, bool) {
	in := strings.TrimSpace(input)
	if in == "" {
		return "", "", false
	}
	for _, v := range vaults {
		if strings.EqualFold(v.Name, in) || (v.ID != "" && strings.EqualFold(v.ID, in)) {
			return v.Name, v.Path, true
		}
	}
	return "", "", false
}

func pickVaultFuzzy(vaults []discoveredVault) (string, string, error) {
	items := make([]string, 0, len(vaults))
	for _, v := range vaults {
		items = append(items, fmt.Sprintf("%s  (%s)", v.Name, v.Path))
	}
	idx, err := fuzzyfinder.Find(items, func(i int) string { return items[i] })
	if err != nil {
		return "", "", err
	}
	return vaults[idx].Name, vaults[idx].Path, nil
}

func promptDailyFolder(in *bufio.Reader, vaultPath string, existing string) (string, error) {
	const defaultDailyFolder = "Daily"
	for {
		fmt.Println()
		fmt.Println(SingleLine(DefaultBorderWidth))
		fmt.Println("  Step 2: Daily note folder")
		fmt.Println(SingleLine(DefaultBorderWidth))
		if strings.TrimSpace(existing) != "" {
			fmt.Printf("Current: %s (press Enter to keep)\n", existing)
		}
		fmt.Println("Enter a folder path relative to the vault, or type 'ls' to browse.")
		fmt.Println("Type 'skip' to accept a default, '?' for help, or 'back' to go back.")

		line, action, err := promptLine(in, "> ")
		if err != nil {
			return "", err
		}
		if action == actionBack {
			return "", errBack
		}
		if action == actionSkip {
			if strings.TrimSpace(existing) != "" {
				return existing, nil
			}
			fmt.Printf("Using default: %s\n", defaultDailyFolder)
			return defaultDailyFolder, nil
		}
		if action == actionHelp {
			printDailyHelp()
			continue
		}
		if line == "" && strings.TrimSpace(existing) != "" {
			return existing, nil
		}
		if strings.EqualFold(line, "ls") {
			p, err := pickOrCreateFolderPath(vaultPath)
			if err != nil {
				fmt.Printf("Selection error: %v\n", err)
				continue
			}
			return p, nil
		}
		if strings.TrimSpace(line) == "" {
			fmt.Println("Folder is required.")
			continue
		}
		if _, err := obsidian.SafeJoinVaultPath(vaultPath, filepath.ToSlash(line)); err != nil {
			fmt.Printf("Invalid folder path: %v\n", err)
			continue
		}
		return line, nil
	}
}

func promptDailyPattern(in *bufio.Reader, existing string) (string, error) {
	const defaultDailyPattern = "YYYY-MM-DD"
	for {
		fmt.Println()
		fmt.Println(SingleLine(DefaultBorderWidth))
		fmt.Println("  Step 3: Daily note filename pattern")
		fmt.Println(SingleLine(DefaultBorderWidth))
		if strings.TrimSpace(existing) != "" {
			fmt.Printf("Current: %s (press Enter to keep)\n", existing)
		}
		if strings.TrimSpace(existing) == "" {
			fmt.Printf("Tip: type 'skip' to accept the default (%s).\n", defaultDailyPattern)
		}

		pattern, err := promptForTargetPattern(in, existing)
		if err != nil {
			if errors.Is(err, errBack) {
				return "", errBack
			}
			return "", err
		}
		if strings.TrimSpace(pattern) == "" {
			fmt.Println("Pattern is required.")
			continue
		}
		return pattern, nil
	}
}

func promptDailyTemplate(in *bufio.Reader, vaultPath string, existing string) (string, error) {
	for {
		fmt.Println()
		fmt.Println(SingleLine(DefaultBorderWidth))
		fmt.Println("  Step 4: Daily note template (optional)")
		fmt.Println(SingleLine(DefaultBorderWidth))
		if strings.TrimSpace(existing) != "" {
			fmt.Printf("Current: %s (press Enter to keep)\n", existing)
		}
		fmt.Println("Press Enter for none, type a path, or type 'ls' to browse.")
		fmt.Println("Type 'skip' to skip this step, '?' for help, or 'back' to go back.")

		line, action, err := promptLine(in, "> ")
		if err != nil {
			return "", err
		}
		switch action {
		case actionBack:
			return "", errBack
		case actionSkip:
			return existing, nil
		case actionHelp:
			printTemplateHelp()
			continue
		default:
		}
		if line == "" && strings.TrimSpace(existing) != "" {
			return existing, nil
		}
		if line == "" {
			return "", nil
		}
		if strings.EqualFold(line, "ls") {
			p, err := pickOrCreateNotePath(vaultPath)
			if err != nil {
				fmt.Printf("Selection error: %v\n", err)
				continue
			}
			return p, nil
		}
		if _, err := obsidian.SafeJoinVaultPath(vaultPath, filepath.ToSlash(line)); err != nil {
			fmt.Printf("Invalid path: %v\n", err)
			continue
		}
		return line, nil
	}
}

func runTargetsMigration(in *bufio.Reader) error {
	fmt.Println()
	fmt.Println(SingleLine(DefaultBorderWidth))
	fmt.Println("  Step 5: Targets (optional)")
	fmt.Println(SingleLine(DefaultBorderWidth))
	fmt.Println("Type 'skip' to skip, 'back' to go back, '?' for help.")

	_, targetsFile, err := config.TargetsPath()
	if err != nil {
		return err
	}
	raw, err := os.ReadFile(targetsFile)
	if err != nil {
		if os.IsNotExist(err) {
			fmt.Println("No targets.yaml found.")
			fmt.Println("Create it now? (y/N)  Type 'skip' to skip, 'back' to go back.")
			for {
				line, action, err := promptLine(in, "> ")
				if err != nil {
					return err
				}
				if action == actionBack {
					return errBack
				}
				if action == actionSkip {
					return nil
				}
				if action == actionHelp {
					printTargetsHelp()
					continue
				}
				if isYes(line) {
					path, err := obsidian.EnsureTargetsFileExists()
					if err != nil {
						return err
					}
					fmt.Printf("Created: %s\n", path)
				}
				return nil
			}
		}
		return err
	}

	scalarNames, err := detectScalarTargets(raw)
	if err != nil {
		fmt.Printf("Could not parse targets.yaml: %v\n", err)
		fmt.Println("Options:")
		fmt.Println("  - Run: obsidian-cli target edit")
		fmt.Println("  - Or open it now in your editor")
		fmt.Println("Open in editor now? (y/N)  Type 'back' to go back, 'skip' to skip.")
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return err
		}
		if action == actionBack {
			return errBack
		}
		if action == actionSkip {
			return nil
		}
		if isYes(line) {
			return obsidian.OpenInEditor(targetsFile)
		}
		return nil
	}
	if len(scalarNames) == 0 {
		fmt.Println("targets.yaml looks OK (no simplified scalar targets detected).")
		return nil
	}

	fmt.Printf("Found %d simplified targets (e.g. inbox: Inbox.md).\n", len(scalarNames))
	fmt.Println("Migrate them to the canonical format (type/file fields)?")
	fmt.Println("This will rewrite targets.yaml and may remove custom comments (a standard header will be added).")
	fmt.Println("Continue? (y/N)  Type 'skip' to skip, 'back' to go back.")
	for {
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return err
		}
		if action == actionBack {
			return errBack
		}
		if action == actionSkip {
			return nil
		}
		if action == actionHelp {
			printTargetsHelp()
			continue
		}
		if !isYes(line) {
			return nil
		}
		break
	}

	cfg, err := obsidian.LoadTargets()
	if err != nil {
		return err
	}

	before := bytes.TrimSpace(raw)
	if err := obsidian.SaveTargets(cfg); err != nil {
		return err
	}
	afterRaw, _ := os.ReadFile(targetsFile)
	after := bytes.TrimSpace(afterRaw)
	if bytes.Equal(before, after) {
		fmt.Println("No changes made.")
	} else {
		fmt.Println("Migrated targets.yaml.")
	}
	return nil
}

func detectScalarTargets(raw []byte) ([]string, error) {
	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, err
	}
	if len(node.Content) == 0 {
		return nil, nil
	}
	root := node.Content[0]
	if root.Kind != yaml.MappingNode {
		return nil, nil
	}

	var names []string
	for i := 0; i+1 < len(root.Content); i += 2 {
		k := root.Content[i]
		v := root.Content[i+1]
		if k.Kind == yaml.ScalarNode && v.Kind == yaml.ScalarNode {
			n := strings.TrimSpace(k.Value)
			if n != "" {
				names = append(names, n)
			}
		}
	}
	sort.Strings(names)
	return names, nil
}

func runInitAddTargets(in *bufio.Reader, vaultPath string) error {
	cfg, err := loadTargetsOrEmpty()
	if err != nil {
		return err
	}

	fmt.Println()
	fmt.Println("Add a target now? (y/N)  Type 'skip' to skip, 'back' to go back, '?' for help.")
	for {
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return err
		}
		if action == actionBack {
			return errBack
		}
		if action == actionSkip {
			return nil
		}
		if action == actionHelp {
			printTargetsHelp()
			continue
		}
		if !isYes(line) {
			return nil
		}
		break
	}

	for {
		name, target, err := runTargetAddWizard(in, vaultPath, "", cfg)
		if err != nil {
			fmt.Printf("Error: %v\n", err)
		} else {
			cfg[name] = target
			if err := obsidian.SaveTargets(cfg); err != nil {
				return err
			}
			fmt.Printf("Saved target: %s\n", name)
		}

		fmt.Println("Add another? (y/N)")
		line, _, err := promptLine(in, "> ")
		if err != nil {
			return err
		}
		if !isYes(line) {
			break
		}
	}
	return nil
}

func printVaultHelp() {
	fmt.Println()
	fmt.Println("Vault help:")
	fmt.Println("- Vault name is usually the folder name of your vault path.")
	fmt.Println("- obsidian-cli reads vaults from Obsidian's obsidian.json in your user config directory.")
	fmt.Println("- If no vaults are discovered, open Obsidian once and ensure at least one vault is configured.")
}

func printDailyHelp() {
	fmt.Println()
	fmt.Println("Daily note help:")
	fmt.Println("- Folder is a path relative to the vault root (example: Daily).")
	fmt.Println("- Pattern controls the daily note filename and supports Obsidian-style tokens like YYYY, MM, DD, HH, mm, ss.")
	fmt.Println("- [literal] blocks are preserved, e.g. YYYY-[log]-MM.")
}

func printTemplateHelp() {
	fmt.Println()
	fmt.Println("Template help:")
	fmt.Println("- Template is a note path relative to the vault root (example: Templates/Daily).")
	fmt.Println("- When a daily note is created for the first time, template content is copied and variables like {{date}}, {{time}}, {{title}} are expanded.")
}

func printTargetsHelp() {
	fmt.Println()
	fmt.Println("Targets help:")
	fmt.Println("- Targets let you capture quickly to a configured note or dated note pattern.")
	fmt.Println("- Use: obsidian-cli target add    (guided workflow)")
	fmt.Println("- Use: obsidian-cli target --select")
	fmt.Println("- targets.yaml is stored next to preferences.json.")
}
