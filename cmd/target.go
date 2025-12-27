package cmd

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"
	"github.com/ktr0731/go-fuzzyfinder"
	"github.com/spf13/cobra"
)

var (
	targetSelect bool
	targetDryRun bool
)

var targetCmd = &cobra.Command{
	Use:     "target [id] [text]",
	Aliases: []string{"t"},
	Short: "Append text to a configured target note",
	Long: `Appends text to a note configured in targets.yaml.

Targets can point at:
  - a fixed file path (always append to the same note)
  - a folder + filename pattern (append to a dated note based on the current time)

If no text is provided, content is read from stdin (piped) or entered interactively until EOF.`,
	Example: `  # Append a one-liner to a target
  obsidian-cli target inbox "Buy milk"

  # Multi-line content (Ctrl-D to save, Ctrl-C to cancel)
  obsidian-cli target inbox

  # Pick a target interactively, then enter content
  obsidian-cli target --select

  # Preview which file would be used
  obsidian-cli target inbox --dry-run`,
	Args: cobra.ArbitraryArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}

		if len(args) == 0 || targetSelect {
			name, err := pickTargetName()
			if err != nil {
				return err
			}
			if strings.TrimSpace(name) == "" {
				return errors.New("no target selected")
			}

			if targetDryRun {
				return printTargetPlan(&vault, name)
			}

			content, err := actions.PromptForContentIfEmpty("")
			if err != nil {
				return err
			}
			plan, err := actions.AppendToTarget(&vault, name, content, time.Now(), false)
			if err != nil {
				return err
			}
			fmt.Printf("Wrote to: %s\n", plan.AbsoluteNotePath)
			return nil
		}

		name := strings.TrimSpace(args[0])
		content := strings.TrimSpace(strings.Join(args[1:], " "))

		if targetDryRun {
			return printTargetPlan(&vault, name)
		}

		var err error
		content, err = actions.PromptForContentIfEmpty(content)
		if err != nil {
			return err
		}

		plan, err := actions.AppendToTarget(&vault, name, content, time.Now(), false)
		if err != nil {
			return err
		}
		fmt.Printf("Wrote to: %s\n", plan.AbsoluteNotePath)
		return nil
	},
}

var targetListCmd = &cobra.Command{
	Use:   "list",
	Short: "List configured targets",
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadTargetsOrEmpty()
		if err != nil {
			return err
		}
		names := obsidian.ListTargetNames(cfg)
		if len(names) == 0 {
			fmt.Println("No targets configured.")
			fmt.Println("Run: obsidian-cli target add")
			return nil
		}
		for _, n := range names {
			t := cfg[n]
			if strings.ToLower(strings.TrimSpace(t.Type)) == "folder" {
				fmt.Printf("- %s (folder: %s, pattern: %s)\n", n, t.Folder, t.Pattern)
			} else {
				fmt.Printf("- %s (file: %s)\n", n, firstNonEmpty(t.File, t.Note))
			}
		}
		return nil
	},
}

var targetAddCmd = &cobra.Command{
	Use:   "add [name]",
	Short: "Add a new target",
	Long: `Add a new capture target.

Run without a name to start a guided workflow.`,
	Example: `  # Guided workflow
  obsidian-cli target add

  # Add a fixed-file target
  obsidian-cli target add inbox

  # Add a folder+pattern target
  obsidian-cli target add log`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		in := bufio.NewReader(os.Stdin)

		cfg, err := loadTargetsOrEmpty()
		if err != nil {
			return err
		}

		var name string
		if len(args) == 1 {
			name = strings.TrimSpace(args[0])
		}

		vault := obsidian.Vault{Name: vaultName}
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		name, target, err := runTargetAddWizard(in, vaultPath, name, cfg)
		if err != nil {
			return err
		}

		cfg[name] = target
		if err := obsidian.SaveTargets(cfg); err != nil {
			return err
		}

		fmt.Printf("Saved target: %s\n", name)
		return nil
	},
}

var targetRemoveCmd = &cobra.Command{
	Use:     "remove [name]",
	Aliases: []string{"rm"},
	Short:   "Remove a target",
	Args:    cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg, err := loadTargetsOrEmpty()
		if err != nil {
			return err
		}

		var name string
		if len(args) == 1 {
			name = strings.TrimSpace(args[0])
		} else {
			name, err = pickTargetNameFromConfig(cfg)
			if err != nil {
				return err
			}
		}
		if strings.TrimSpace(name) == "" {
			return errors.New("target name is required")
		}

		if _, ok := cfg[name]; !ok {
			return fmt.Errorf("target not found: %s", name)
		}
		delete(cfg, name)
		if err := obsidian.SaveTargets(cfg); err != nil {
			return err
		}
		fmt.Printf("Removed target: %s\n", name)
		return nil
	},
}

var targetTestCmd = &cobra.Command{
	Use:   "test [name]",
	Short: "Preview the resolved path for a target",
	Long: `Shows which file would be created or appended to for the given target.

If no name is provided, previews all targets.`,
	Args: cobra.MaximumNArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		cfg, err := loadTargetsOrEmpty()
		if err != nil {
			return err
		}
		names := obsidian.ListTargetNames(cfg)
		if len(names) == 0 {
			return errors.New("no targets configured")
		}

		if len(args) == 1 {
			name := strings.TrimSpace(args[0])
			return printTargetPlan(&vault, name)
		}

		// Preview all targets.
		for _, name := range names {
			fmt.Printf("%s:\n", name)
			if err := printTargetPlan(&vault, name); err != nil {
				fmt.Printf("  error: %v\n", err)
			}
		}
		return nil
	},
}

var targetEditCmd = &cobra.Command{
	Use:   "edit",
	Short: "Edit targets in CLI or open targets.yaml",
	RunE: func(cmd *cobra.Command, args []string) error {
		in := bufio.NewReader(os.Stdin)

		path, err := obsidian.EnsureTargetsFileExists()
		if err != nil {
			return err
		}

		for {
			fmt.Println("Edit targets:")
			fmt.Println("  1) Stay in CLI (recommended)")
			fmt.Println("  2) Open in editor")
			fmt.Println("Press Enter for CLI, type '?' for help, or 'back' to cancel.")
			line, action, err := promptLine(in, "> ")
			if err != nil {
				return err
			}
			switch action {
			case actionBack:
				return nil
			case actionHelp:
				fmt.Println()
				fmt.Println("Edit help:")
				fmt.Printf("- CLI mode provides guided add/edit/remove/test/list workflows.\n")
				fmt.Printf("- Editor mode opens: %s\n", path)
				fmt.Println()
				continue
			case actionSkip:
				return runTargetEditor(in)
			default:
			}
			switch strings.ToLower(strings.TrimSpace(line)) {
			case "", "1", "cli":
				return runTargetEditor(in)
			case "2", "editor":
				return obsidian.OpenInEditor(path)
			default:
				fmt.Println("Invalid selection. Enter 1 or 2, or type '?' for help.")
			}
		}
	},
}

func init() {
	targetCmd.Flags().BoolVar(&targetSelect, "ls", false, "select a target interactively")
	targetCmd.Flags().BoolVar(&targetSelect, "select", false, "select a target interactively")
	targetCmd.Flags().BoolVar(&targetDryRun, "dry-run", false, "preview the resolved target path without writing")
	targetCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name (not required if default is set)")

	targetCmd.AddCommand(targetListCmd)
	targetCmd.AddCommand(targetAddCmd)
	targetCmd.AddCommand(targetRemoveCmd)
	targetCmd.AddCommand(targetTestCmd)
	targetCmd.AddCommand(targetEditCmd)

	rootCmd.AddCommand(targetCmd)
}

func loadTargetsOrEmpty() (obsidian.TargetsConfig, error) {
	_, err := obsidian.EnsureTargetsFileExists()
	if err != nil {
		return nil, err
	}
	cfg, err := obsidian.LoadTargets()
	if err != nil {
		return nil, err
	}
	return cfg, nil
}

func pickTargetName() (string, error) {
	cfg, err := loadTargetsOrEmpty()
	if err != nil {
		return "", err
	}
	return pickTargetNameFromConfig(cfg)
}

func pickTargetNameFromConfig(cfg obsidian.TargetsConfig) (string, error) {
	names := obsidian.ListTargetNames(cfg)
	if len(names) == 0 {
		return "", errors.New("no targets configured (run: obsidian-cli target add)")
	}

	idx, err := fuzzyfinder.Find(names, func(i int) string {
		return names[i]
	})
	if err != nil {
		return "", err
	}
	return names[idx], nil
}

func printTargetPlan(vault *obsidian.Vault, name string) error {
	cfg, err := loadTargetsOrEmpty()
	if err != nil {
		return err
	}
	target, ok := cfg[name]
	if !ok {
		return fmt.Errorf("target not found: %s", name)
	}

	effectiveVault := vault
	if strings.TrimSpace(target.Vault) != "" {
		effectiveVault = &obsidian.Vault{Name: strings.TrimSpace(target.Vault)}
	}

	vaultName, err := effectiveVault.DefaultName()
	if err != nil {
		return err
	}
	vaultPath, err := effectiveVault.Path()
	if err != nil {
		return err
	}

	plan, err := obsidian.PlanTargetAppend(vaultPath, name, target, time.Now())
	if err != nil {
		return err
	}

	fmt.Printf("  vault: %s\n", vaultName)
	fmt.Printf("  note: %s\n", plan.AbsoluteNotePath)
	fmt.Printf("  create_dirs: %t\n", plan.WillCreateDirs)
	fmt.Printf("  create_file: %t\n", plan.WillCreateFile)
	if plan.WillApplyTemplate {
		fmt.Printf("  template: %s\n", plan.AbsoluteTemplate)
	}
	return nil
}

func runTargetAddWizard(in *bufio.Reader, vaultPath string, existingName string, existing obsidian.TargetsConfig) (string, obsidian.Target, error) {
	step := 0
	var name string
	var t obsidian.Target

	for {
		switch step {
		case 0:
			if strings.TrimSpace(existingName) != "" {
				name = existingName
			} else {
				fmt.Println("Target name")
				fmt.Println("Type '?' for help, or 'back' to cancel.")
				line, action, err := promptLine(in, "> ")
				if err != nil {
					return "", obsidian.Target{}, err
				}
				switch action {
				case actionBack:
					return "", obsidian.Target{}, errors.New("cancelled")
				case actionHelp:
					fmt.Println()
					fmt.Println("Target name help:")
					fmt.Println("- Names cannot contain whitespace.")
					fmt.Println("- Reserved names: add, rm, ls, edit, validate, test, help.")
					fmt.Println("- Examples: inbox, todo, log")
					fmt.Println()
					continue
				case actionSkip:
					fmt.Println("Target name is required (skip is not available here).")
					continue
				default:
				}
				if err := obsidian.ValidateTargetName(line); err != nil {
					fmt.Printf("Invalid name: %v\n", err)
					continue
				}
				if _, ok := existing[line]; ok {
					fmt.Println("A target with that name already exists.")
					continue
				}
				name = line
			}
			step = 1
		case 1:
			fmt.Println("Target type:")
			fmt.Println("  1) file   (append to one fixed note)")
			fmt.Println("  2) folder (append to a dated note based on a pattern)")
			fmt.Println("Type a number, 'skip' for file (recommended), '?' for help, or 'back' to change the name.")
			line, action, err := promptLine(in, "> ")
			if err != nil {
				return "", obsidian.Target{}, err
			}
			switch action {
			case actionBack:
				step = 0
				existingName = ""
				continue
			case actionHelp:
				fmt.Println()
				fmt.Println("Type help:")
				fmt.Println("- file: always appends to the same note")
				fmt.Println("- folder: appends to a note path derived from folder + date pattern")
				fmt.Println()
				continue
			case actionSkip:
				t.Type = "file"
				step = 2
				continue
			default:
			}
			switch line {
			case "1":
				t.Type = "file"
				step = 2
			case "2":
				t.Type = "folder"
				step = 2
			default:
				fmt.Println("Invalid selection.")
			}
		case 2:
			if strings.ToLower(strings.TrimSpace(t.Type)) == "file" {
				p, err := promptForTargetFile(in, vaultPath, "", name+".md")
				if err != nil {
					if errors.Is(err, errBack) {
						step = 1
						continue
					}
					return "", obsidian.Target{}, err
				}
				t.File = p
				step = 4
			} else {
				folder, err := promptForTargetFolder(in, vaultPath, "", name)
				if err != nil {
					if errors.Is(err, errBack) {
						step = 1
						continue
					}
					return "", obsidian.Target{}, err
				}
				t.Folder = folder
				step = 3
			}
		case 3:
			pattern, err := promptForTargetPattern(in, "")
			if err != nil {
				if errors.Is(err, errBack) {
					step = 2
					continue
				}
				return "", obsidian.Target{}, err
			}
			t.Pattern = pattern
			step = 4
		case 4:
			template, err := promptForTemplatePath(in, vaultPath, t.Template)
			if err != nil {
				if errors.Is(err, errBack) {
					if strings.ToLower(strings.TrimSpace(t.Type)) == "folder" {
						step = 3
					} else {
						step = 2
					}
					continue
				}
				return "", obsidian.Target{}, err
			}
			t.Template = template
			step = 5
		case 5:
			vaultOverride, err := promptForVaultOverride(in, t.Vault)
			if err != nil {
				if errors.Is(err, errBack) {
					step = 4
					continue
				}
				return "", obsidian.Target{}, err
			}
			t.Vault = vaultOverride
			step = 6
		case 6:
			fmt.Println()
			fmt.Println("Save this target? (y/N)  Type 'back' to edit, '?' for help, or 'skip' to cancel.")
			fmt.Printf("  name: %s\n", name)
			fmt.Printf("  type: %s\n", t.Type)
			if strings.ToLower(strings.TrimSpace(t.Type)) == "folder" {
				fmt.Printf("  folder: %s\n", t.Folder)
				fmt.Printf("  pattern: %s\n", t.Pattern)
				fmt.Printf("  example: %s\n", obsidian.ExpandDatePattern(t.Pattern, time.Now()))
			} else {
				fmt.Printf("  file: %s\n", t.File)
			}
			if strings.TrimSpace(t.Template) != "" {
				fmt.Printf("  template: %s\n", t.Template)
			}
			if strings.TrimSpace(t.Vault) != "" {
				fmt.Printf("  vault: %s\n", t.Vault)
			}
			line, action, err := promptLine(in, "> ")
			if err != nil {
				return "", obsidian.Target{}, err
			}
			switch action {
			case actionBack:
				step = 5
				continue
			case actionHelp:
				fmt.Println()
				fmt.Println("Save help:")
				fmt.Println("- Answer 'y' to save to targets.yaml.")
				fmt.Println("- Answer 'n' or press Enter to cancel without saving.")
				fmt.Println()
				continue
			case actionSkip:
				return "", obsidian.Target{}, errors.New("cancelled")
			default:
			}
			if isYes(line) {
				return name, t, nil
			}
			return "", obsidian.Target{}, errors.New("cancelled")
		default:
			return "", obsidian.Target{}, errors.New("invalid wizard state")
		}
	}
}

var errBack = errors.New("back")

func promptForTargetFile(in *bufio.Reader, vaultPath string, existing string, defaultPath string) (string, error) {
	fmt.Println("Target file path (relative to vault).")
	if strings.TrimSpace(existing) != "" {
		fmt.Printf("Current: %s (press Enter to keep)\n", existing)
	}
	if strings.TrimSpace(existing) == "" && strings.TrimSpace(defaultPath) != "" {
		fmt.Printf("Default: %s\n", defaultPath)
	}
	fmt.Println("Type a path, 'ls' to browse, 'skip' to accept default/keep current, '?' for help, or 'back' to go back.")
	for {
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return "", err
		}
		if action == actionBack {
			return "", errBack
		}
		if action == actionHelp {
			fmt.Println()
			fmt.Println("File help:")
			fmt.Println("- Path is relative to the vault root (example: Inbox/Capture.md).")
			fmt.Println("- A .md extension is optional.")
			fmt.Println()
			continue
		}
		if action == actionSkip {
			if strings.TrimSpace(existing) != "" {
				return existing, nil
			}
			if strings.TrimSpace(defaultPath) == "" {
				fmt.Println("No default available; enter a file path.")
				continue
			}
			fmt.Printf("Using default: %s\n", defaultPath)
			return defaultPath, nil
		}
		if line == "" {
			if strings.TrimSpace(existing) != "" {
				return existing, nil
			}
			if strings.TrimSpace(defaultPath) != "" {
				fmt.Printf("Using default: %s\n", defaultPath)
				return defaultPath, nil
			}
			fmt.Println("File path is required.")
			continue
		}
		if strings.EqualFold(line, "ls") {
			return pickOrCreateNotePath(vaultPath)
		}
		if _, err := obsidian.SafeJoinVaultPath(vaultPath, filepath.ToSlash(line)); err != nil {
			fmt.Printf("Invalid path: %v\n", err)
			continue
		}
		return line, nil
	}
}

func promptForTargetFolder(in *bufio.Reader, vaultPath string, existing string, defaultFolder string) (string, error) {
	fmt.Println("Target folder path (relative to vault).")
	if strings.TrimSpace(existing) != "" {
		fmt.Printf("Current: %s (press Enter to keep)\n", existing)
	}
	if strings.TrimSpace(existing) == "" && strings.TrimSpace(defaultFolder) != "" {
		fmt.Printf("Default: %s\n", defaultFolder)
	}
	fmt.Println("Type a folder path, 'ls' to browse, 'skip' to accept default/keep current, '?' for help, or 'back' to go back.")
	for {
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return "", err
		}
		if action == actionBack {
			return "", errBack
		}
		if action == actionHelp {
			fmt.Println()
			fmt.Println("Folder help:")
			fmt.Println("- Path is relative to the vault root (example: Inbox).")
			fmt.Println("- A date-based filename will be appended using your selected pattern.")
			fmt.Println()
			continue
		}
		if action == actionSkip {
			if strings.TrimSpace(existing) != "" {
				return existing, nil
			}
			if strings.TrimSpace(defaultFolder) == "" {
				fmt.Println("No default available; enter a folder path.")
				continue
			}
			fmt.Printf("Using default: %s\n", defaultFolder)
			return defaultFolder, nil
		}
		if line == "" {
			if strings.TrimSpace(existing) != "" {
				return existing, nil
			}
			if strings.TrimSpace(defaultFolder) != "" {
				fmt.Printf("Using default: %s\n", defaultFolder)
				return defaultFolder, nil
			}
			fmt.Println("Folder path is required.")
			continue
		}
		if strings.EqualFold(line, "ls") {
			return pickOrCreateFolderPath(vaultPath)
		}
		if _, err := obsidian.SafeJoinVaultPath(vaultPath, filepath.ToSlash(line)); err != nil {
			fmt.Printf("Invalid path: %v\n", err)
			continue
		}
		return line, nil
	}
}

func promptForVaultOverride(in *bufio.Reader, existing string) (string, error) {
	fmt.Println("Vault override (optional).")
	if strings.TrimSpace(existing) != "" {
		fmt.Printf("Current: %s (press Enter to keep)\n", existing)
	}
	fmt.Println("Press Enter for default vault, type a vault name, or type 'skip' to keep current.")
	fmt.Println("Type '?' for help, or 'back' to go back.")
	for {
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return "", err
		}
		switch action {
		case actionBack:
			return "", errBack
		case actionHelp:
			fmt.Println()
			fmt.Println("Vault override help:")
			fmt.Println("- Leave empty to use the default vault.")
			fmt.Println("- Set to a vault name (as it appears in Obsidian) to route this target elsewhere.")
			fmt.Println()
			continue
		case actionSkip:
			return existing, nil
		default:
		}
		if line == "" && strings.TrimSpace(existing) != "" {
			return existing, nil
		}
		return line, nil
	}
}

func promptForTemplatePath(in *bufio.Reader, vaultPath string, existing string) (string, error) {
	fmt.Println("Template note path (optional, relative to vault).")
	if strings.TrimSpace(existing) != "" {
		fmt.Printf("Current: %s (press Enter to keep)\n", existing)
	}
	fmt.Println("Press Enter for none, type a path, type 'ls' to browse, or type 'skip' to keep current.")
	fmt.Println("Type '?' for help, or 'back' to go back.")
	for {
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return "", err
		}
		switch action {
		case actionBack:
			return "", errBack
		case actionHelp:
			fmt.Println()
			fmt.Println("Template help:")
			fmt.Println("- Applied only when creating a new note for the first time.")
			fmt.Println("- Supports variables like {{date}}, {{time}}, and {{title}}.")
			fmt.Println()
			continue
		case actionSkip:
			return existing, nil
		default:
		}
		if line == "" && strings.TrimSpace(existing) != "" {
			return existing, nil
		}
		if line == "" {
			return "", nil
		}
		if strings.EqualFold(line, "ls") {
			path, err := pickOrCreateNotePath(vaultPath)
			if err != nil {
				return "", err
			}
			return path, nil
		}
		if _, err := obsidian.SafeJoinVaultPath(vaultPath, filepath.ToSlash(line)); err != nil {
			fmt.Printf("Invalid path: %v\n", err)
			continue
		}
		return line, nil
	}
}

func promptForTargetPattern(in *bufio.Reader, existing string) (string, error) {
	now := time.Now()
	const defaultPattern = "YYYY-MM-DD"
	options := []struct {
		label   string
		pattern string
	}{
		{"Daily (YYYY-MM-DD)", "YYYY-MM-DD"},
		{"Hourly (YYYY-MM-DD_HH)", "YYYY-MM-DD_HH"},
		{"Minute (YYYY-MM-DD_HHmm)", "YYYY-MM-DD_HHmm"},
		{"Second (YYYY-MM-DD_HHmmss)", "YYYY-MM-DD_HHmmss"},
		{"Zettel (YYYYMMDDHHmmss)", "YYYYMMDDHHmmss"},
		{"Daily + weekday (YYYY-MM-DD_ddd)", "YYYY-MM-DD_ddd"},
		{"Daily + weekday full (YYYY-MM-DD_dddd)", "YYYY-MM-DD_dddd"},
		{"Month name (YYYY-MMMM-DD)", "YYYY-MMMM-DD"},
		{"Month abbrev (YYYY-MMM-DD)", "YYYY-MMM-DD"},
		{"AM/PM (YYYY-MM-DD_A)", "YYYY-MM-DD_A"},
		{"Literal blocks (YYYY-[log]-MM)", "YYYY-[log]-MM"},
		{"Custom pattern...", ""},
	}

	fmt.Println("Filename pattern (controls when a new file is created).")
	if strings.TrimSpace(existing) != "" {
		fmt.Printf("Current: %s (press Enter to keep)\n", existing)
	}
	fmt.Printf("Type a number, type a pattern directly, 'skip' for default (%s), '?' for help, or 'back' to go back.\n", defaultPattern)
	for i, o := range options {
		ex := o.pattern
		if ex != "" {
			ex = obsidian.ExpandDatePattern(o.pattern, now)
			ex = ex + ".md"
		}
		if o.pattern == "" {
			fmt.Printf("  %d) %s\n", i+1, o.label)
		} else {
			fmt.Printf("  %d) %s -> %s\n", i+1, o.label, ex)
		}
	}

	for {
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return "", err
		}
		if action == actionBack {
			return "", errBack
		}
		if action == actionHelp {
			printPatternHelp()
			continue
		}
		if action == actionSkip {
			if strings.TrimSpace(existing) != "" {
				return existing, nil
			}
			return defaultPattern, nil
		}
		if line == "" {
			if strings.TrimSpace(existing) != "" {
				return existing, nil
			}
			return defaultPattern, nil
		}
		// allow direct entry of a pattern
		if line != "" && (strings.ContainsAny(line, "YMDHmsd") || strings.Contains(line, "{") || strings.Contains(line, "[") || strings.Contains(line, "]")) && !isDigits(line) {
			return line, nil
		}
		n, convErr := parseChoice(line, len(options))
		if convErr != nil {
			fmt.Println(convErr.Error())
			continue
		}
		chosen := options[n-1]
		if chosen.pattern == "" {
			fmt.Println("Enter custom pattern (Obsidian-style tokens; supports [literal] blocks).")
			fmt.Println("Examples: YYYY-MM-DD_HHmm, YYYYMMDDHHmmss, YYYY-[ToDo]-MM")
			fmt.Printf("Type 'skip' for default (%s), '?' for help, or 'back' to go back.\n", defaultPattern)
			for {
				custom, customAction, err := promptLine(in, "> ")
				if err != nil {
					return "", err
				}
				switch customAction {
				case actionBack:
					goto continuePatternMenu
				case actionHelp:
					printPatternHelp()
					continue
				case actionSkip:
					if strings.TrimSpace(existing) != "" {
						return existing, nil
					}
					return defaultPattern, nil
				default:
				}
				custom = strings.TrimSpace(custom)
				if strings.TrimSpace(custom) == "" {
					fmt.Println("Pattern is required.")
					continue
				}
				return custom, nil
			}
		}
		return chosen.pattern, nil

	continuePatternMenu:
		continue
	}
}

func printPatternHelp() {
	fmt.Println()
	fmt.Println("Pattern help:")
	fmt.Println("- Tokens (curated subset): YYYY, YY, MM, M, DD, D, HH, H, mm, m, ss, s, ddd, dddd, MMM, MMMM, A, a")
	fmt.Println("- Zettel timestamp: YYYYMMDDHHmmss")
	fmt.Println("- Literal blocks: wrap text in [brackets], e.g. YYYY-[log]-MM")
	fmt.Println("- Examples:")
	fmt.Println("    YYYY-MM-DD")
	fmt.Println("    YYYY-MM-DD_HH")
	fmt.Println("    YYYY-MM-DD_HHmmss")
	fmt.Println("    YYYYMMDDHHmmss")
	fmt.Println()
}

func parseChoice(line string, max int) (int, error) {
	if line == "" {
		return 0, errors.New("enter a selection")
	}
	var n int
	_, err := fmt.Sscanf(line, "%d", &n)
	if err != nil || n < 1 || n > max {
		return 0, fmt.Errorf("invalid selection: enter 1-%d", max)
	}
	return n, nil
}

func isDigits(s string) bool {
	for _, r := range s {
		if r < '0' || r > '9' {
			return false
		}
	}
	return s != ""
}

func pickOrCreateFolderPath(vaultPath string) (string, error) {
	dirs, err := listVaultFolders(vaultPath)
	if err != nil {
		return "", err
	}
	dirs = append(dirs, "(Create new folder...)")
	sort.Strings(dirs)

	idx, err := fuzzyfinder.Find(dirs, func(i int) string { return dirs[i] })
	if err != nil {
		return "", err
	}
	choice := dirs[idx]
	if choice != "(Create new folder...)" {
		return choice, nil
	}
	return promptCreateFolder(vaultPath)
}

func pickOrCreateNotePath(vaultPath string) (string, error) {
	files, err := listVaultNotes(vaultPath)
	if err != nil {
		return "", err
	}
	files = append(files, "(Create new note...)")
	sort.Strings(files)

	idx, err := fuzzyfinder.Find(files, func(i int) string { return files[i] })
	if err != nil {
		return "", err
	}
	choice := files[idx]
	if choice != "(Create new note...)" {
		return choice, nil
	}
	return promptCreateNote(vaultPath)
}

func listVaultFolders(vaultPath string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() {
			return nil
		}
		if path == vaultPath {
			return nil
		}
		base := filepath.Base(path)
		if strings.HasPrefix(base, ".") {
			return filepath.SkipDir
		}
		rel, err := filepath.Rel(vaultPath, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(rel))
		return nil
	})
	return out, err
}

func listVaultNotes(vaultPath string) ([]string, error) {
	var out []string
	err := filepath.WalkDir(vaultPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		base := filepath.Base(path)
		if d.IsDir() {
			if strings.HasPrefix(base, ".") {
				return filepath.SkipDir
			}
			return nil
		}
		if strings.HasPrefix(base, ".") {
			return nil
		}
		if strings.ToLower(filepath.Ext(base)) != ".md" {
			return nil
		}
		rel, err := filepath.Rel(vaultPath, path)
		if err != nil {
			return err
		}
		out = append(out, filepath.ToSlash(strings.TrimSuffix(rel, ".md")))
		return nil
	})
	return out, err
}

func promptCreateFolder(vaultPath string) (string, error) {
	in := bufio.NewReader(os.Stdin)
	fmt.Println("Enter folder path to create (relative to vault):")
	fmt.Print("> ")
	line, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", errors.New("folder path is required")
	}
	abs, err := obsidian.SafeJoinVaultPath(vaultPath, filepath.ToSlash(line))
	if err != nil {
		return "", err
	}
	if err := os.MkdirAll(abs, 0750); err != nil {
		return "", err
	}
	return line, nil
}

func promptCreateNote(vaultPath string) (string, error) {
	in := bufio.NewReader(os.Stdin)
	fmt.Println("Enter note path to create (relative to vault):")
	fmt.Print("> ")
	line, err := in.ReadString('\n')
	if err != nil {
		return "", err
	}
	line = strings.TrimSpace(line)
	if line == "" {
		return "", errors.New("note path is required")
	}
	abs, err := obsidian.SafeJoinVaultPath(vaultPath, filepath.ToSlash(line))
	if err != nil {
		return "", err
	}
	if !strings.HasSuffix(strings.ToLower(abs), ".md") {
		abs += ".md"
	}
	if err := os.MkdirAll(filepath.Dir(abs), 0750); err != nil {
		return "", err
	}
	if err := os.WriteFile(abs, []byte{}, 0600); err != nil {
		return "", err
	}
	return line, nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return v
		}
	}
	return ""
}

func runTargetEditor(in *bufio.Reader) error {
	cfg, err := loadTargetsOrEmpty()
	if err != nil {
		return err
	}

	vault := obsidian.Vault{Name: vaultName}
	vaultPath, err := vault.Path()
	if err != nil {
		return err
	}

	for {
		fmt.Println()
		fmt.Println("Target editor:")
		fmt.Println("  1) Add target")
		fmt.Println("  2) Edit target")
		fmt.Println("  3) Remove target")
		fmt.Println("  4) Test target")
		fmt.Println("  5) List targets")
		fmt.Println("  6) Back")
		fmt.Println("Type '?' for help, or 'back'/'skip' to exit.")
		line, action, err := promptLine(in, "> ")
		if err != nil {
			return err
		}
		switch action {
		case actionBack, actionSkip:
			return nil
		case actionHelp:
			fmt.Println()
			fmt.Println("Editor help:")
			fmt.Println("- Add/Edit prompts support 'back', 'skip', and '?' to help you continue without restarting.")
			fmt.Println()
			continue
		default:
		}
		switch line {
		case "1":
			name, target, err := runTargetAddWizard(in, vaultPath, "", cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			cfg[name] = target
			if err := obsidian.SaveTargets(cfg); err != nil {
				return err
			}
			fmt.Printf("Saved target: %s\n", name)
		case "2":
			name, err := pickTargetNameFromConfig(cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			current := cfg[name]
			updated, err := runTargetEditWizard(in, vaultPath, name, current)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			cfg[name] = updated
			if err := obsidian.SaveTargets(cfg); err != nil {
				return err
			}
			fmt.Printf("Updated target: %s\n", name)
		case "3":
			name, err := pickTargetNameFromConfig(cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			delete(cfg, name)
			if err := obsidian.SaveTargets(cfg); err != nil {
				return err
			}
			fmt.Printf("Removed target: %s\n", name)
		case "4":
			name, err := pickTargetNameFromConfig(cfg)
			if err != nil {
				fmt.Printf("Error: %v\n", err)
				continue
			}
			if err := printTargetPlan(&vault, name); err != nil {
				fmt.Printf("Error: %v\n", err)
			}
		case "5":
			names := obsidian.ListTargetNames(cfg)
			if len(names) == 0 {
				fmt.Println("No targets configured.")
				continue
			}
			for _, n := range names {
				t := cfg[n]
				if strings.ToLower(strings.TrimSpace(t.Type)) == "folder" {
					fmt.Printf("- %s (folder: %s, pattern: %s)\n", n, t.Folder, t.Pattern)
				} else {
					fmt.Printf("- %s (file: %s)\n", n, firstNonEmpty(t.File, t.Note))
				}
			}
		case "6", "back":
			return nil
		default:
			fmt.Println("Invalid selection.")
		}
	}
}

func runTargetEditWizard(in *bufio.Reader, vaultPath string, name string, current obsidian.Target) (obsidian.Target, error) {
	step := 0
	t := current

	for {
		switch step {
		case 0:
			fmt.Printf("Editing target: %s\n", name)
			step = 1
		case 1:
			fmt.Println("Target type:")
			fmt.Printf("Current: %s (press Enter to keep)\n", strings.TrimSpace(t.Type))
			fmt.Println("  1) file   (append to one fixed note)")
			fmt.Println("  2) folder (append to a dated note based on a pattern)")
			fmt.Println("Type a number, press Enter to keep current, 'skip' to keep current, '?' for help, or 'back' to cancel.")
			line, action, err := promptLine(in, "> ")
			if err != nil {
				return obsidian.Target{}, err
			}
			switch action {
			case actionBack:
				return obsidian.Target{}, errors.New("cancelled")
			case actionHelp:
				fmt.Println()
				fmt.Println("Type help:")
				fmt.Println("- file: always appends to the same note")
				fmt.Println("- folder: appends to a note path derived from folder + date pattern")
				fmt.Println()
				continue
			case actionSkip:
				step = 2
				continue
			default:
			}
			if line == "" && strings.TrimSpace(t.Type) != "" {
				step = 2
				continue
			}
			switch line {
			case "1":
				t.Type = "file"
				step = 2
			case "2":
				t.Type = "folder"
				step = 2
			default:
				fmt.Println("Invalid selection.")
			}
		case 2:
			if strings.ToLower(strings.TrimSpace(t.Type)) == "file" {
				p, err := promptForTargetFile(in, vaultPath, firstNonEmpty(t.File, t.Note), name+".md")
				if err != nil {
					if errors.Is(err, errBack) {
						step = 1
						continue
					}
					return obsidian.Target{}, err
				}
				t.File = p
				t.Note = ""
				t.Folder = ""
				t.Pattern = ""
				step = 4
			} else {
				folder, err := promptForTargetFolder(in, vaultPath, t.Folder, name)
				if err != nil {
					if errors.Is(err, errBack) {
						step = 1
						continue
					}
					return obsidian.Target{}, err
				}
				t.Folder = folder
				step = 3
			}
		case 3:
			pattern, err := promptForTargetPattern(in, t.Pattern)
			if err != nil {
				if errors.Is(err, errBack) {
					step = 2
					continue
				}
				return obsidian.Target{}, err
			}
			t.Pattern = pattern
			step = 4
		case 4:
			template, err := promptForTemplatePath(in, vaultPath, t.Template)
			if err != nil {
				if errors.Is(err, errBack) {
					if strings.ToLower(strings.TrimSpace(t.Type)) == "folder" {
						step = 3
					} else {
						step = 2
					}
					continue
				}
				return obsidian.Target{}, err
			}
			t.Template = template
			step = 5
		case 5:
			vaultOverride, err := promptForVaultOverride(in, t.Vault)
			if err != nil {
				if errors.Is(err, errBack) {
					step = 4
					continue
				}
				return obsidian.Target{}, err
			}
			t.Vault = vaultOverride
			step = 6
		case 6:
			fmt.Println()
			fmt.Println("Save changes? (y/N)  Type 'back' to edit, '?' for help, or 'skip' to cancel.")
			fmt.Printf("  name: %s\n", name)
			fmt.Printf("  type: %s\n", t.Type)
			if strings.ToLower(strings.TrimSpace(t.Type)) == "folder" {
				fmt.Printf("  folder: %s\n", t.Folder)
				fmt.Printf("  pattern: %s\n", t.Pattern)
				fmt.Printf("  example: %s\n", obsidian.ExpandDatePattern(t.Pattern, time.Now()))
			} else {
				fmt.Printf("  file: %s\n", t.File)
			}
			if strings.TrimSpace(t.Template) != "" {
				fmt.Printf("  template: %s\n", t.Template)
			}
			if strings.TrimSpace(t.Vault) != "" {
				fmt.Printf("  vault: %s\n", t.Vault)
			}
			line, action, err := promptLine(in, "> ")
			if err != nil {
				return obsidian.Target{}, err
			}
			switch action {
			case actionBack:
				step = 5
				continue
			case actionHelp:
				fmt.Println()
				fmt.Println("Save help:")
				fmt.Println("- Answer 'y' to save to targets.yaml.")
				fmt.Println("- Answer 'n' or press Enter to cancel without saving.")
				fmt.Println()
				continue
			case actionSkip:
				return obsidian.Target{}, errors.New("cancelled")
			default:
			}
			if isYes(line) {
				return t, nil
			}
			return obsidian.Target{}, errors.New("cancelled")
		default:
			return obsidian.Target{}, errors.New("invalid wizard state")
		}
	}
}
