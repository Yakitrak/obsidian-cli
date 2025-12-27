package cmd

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/spf13/cobra"
)

type aliasShell string

const (
	aliasShellBash       aliasShell = "bash"
	aliasShellZsh        aliasShell = "zsh"
	aliasShellFish       aliasShell = "fish"
	aliasShellPowerShell aliasShell = "powershell"
	aliasShellCmd        aliasShell = "cmd"
)

var aliasCmdName string
var aliasCmdShell string
var aliasCmdPrint bool
var aliasCmdSymlink bool
var aliasCmdSymlinkDir string
var aliasCmdForce bool
var aliasCmdDryRun bool

var aliasCmd = &cobra.Command{
	Use:   "alias [name]",
	Short: "Generate a shell alias snippet or install a symlink shortcut",
	Long: `Aliases are typically a shell feature. This command helps by generating an alias snippet you can add to your shell profile,
or by creating a symlink (e.g. ~/.local/bin/obsi -> obsidian-cli) so you can run the CLI with a shorter name.`,
	Args: cobra.MaximumNArgs(1),
	Example: `  # Print an alias snippet (paste into your shell profile, or eval it)
  obsidian-cli alias obsi --shell zsh
  eval "$(obsidian-cli alias obsi --shell zsh)"

  # Install a symlink shortcut (recommended for a persistent shortcut)
  obsidian-cli alias obsi --symlink --dir "$HOME/.local/bin"`,
	RunE: func(cmd *cobra.Command, args []string) error {
		if len(args) == 1 && aliasCmdName == "" {
			aliasCmdName = args[0]
		}
		if aliasCmdName == "" {
			return errors.New("alias name is required (provide [name] or --name)")
		}

		if !isValidAliasName(aliasCmdName) {
			return fmt.Errorf("invalid alias name %q (use letters/numbers/underscore/dash; must start with a letter)", aliasCmdName)
		}

		shell := normalizeShell(aliasCmdShell, os.Getenv("SHELL"))

		if aliasCmdSymlink {
			if aliasCmdSymlinkDir == "" {
				return errors.New("--dir is required when using --symlink")
			}
			if err := installSymlinkAlias(aliasCmdName, aliasCmdSymlinkDir, aliasCmdForce, aliasCmdDryRun); err != nil {
				return err
			}
			if !aliasCmdPrint {
				return nil
			}
		}

		if aliasCmdPrint {
			fmt.Print(renderAliasSnippet(aliasCmdName, shell))
		}
		return nil
	},
}

func init() {
	aliasCmd.Flags().StringVar(&aliasCmdName, "name", "", "alias name (e.g. obsi)")
	aliasCmd.Flags().StringVar(&aliasCmdShell, "shell", "", "shell for snippet output: bash, zsh, fish, powershell, cmd (default: detect from $SHELL)")
	aliasCmd.Flags().BoolVar(&aliasCmdPrint, "print", true, "print the alias snippet to stdout")

	aliasCmd.Flags().BoolVar(&aliasCmdSymlink, "symlink", false, "install a symlink shortcut in --dir pointing to this executable")
	aliasCmd.Flags().StringVar(&aliasCmdSymlinkDir, "dir", filepath.Join(os.Getenv("HOME"), ".local", "bin"), "directory to place the symlink (used with --symlink)")
	aliasCmd.Flags().BoolVar(&aliasCmdForce, "force", false, "overwrite an existing file at the symlink path")
	aliasCmd.Flags().BoolVar(&aliasCmdDryRun, "dry-run", false, "show what would be done without writing anything")

	rootCmd.AddCommand(aliasCmd)
}

func isValidAliasName(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	if !((first >= 'A' && first <= 'Z') || (first >= 'a' && first <= 'z')) {
		return false
	}
	for i := 0; i < len(name); i++ {
		c := name[i]
		isLetter := (c >= 'A' && c <= 'Z') || (c >= 'a' && c <= 'z')
		isDigit := c >= '0' && c <= '9'
		isOK := isLetter || isDigit || c == '_' || c == '-'
		if !isOK {
			return false
		}
	}
	return true
}

func normalizeShell(flag string, envShell string) aliasShell {
	if flag != "" {
		return aliasShell(strings.ToLower(strings.TrimSpace(flag)))
	}
	base := strings.ToLower(filepath.Base(envShell))
	switch base {
	case "bash":
		return aliasShellBash
	case "zsh":
		return aliasShellZsh
	case "fish":
		return aliasShellFish
	case "pwsh", "powershell":
		return aliasShellPowerShell
	case "cmd", "cmd.exe":
		return aliasShellCmd
	default:
		return aliasShellZsh
	}
}

func renderAliasSnippet(name string, shell aliasShell) string {
	switch shell {
	case aliasShellFish:
		return fmt.Sprintf("alias %s 'obsidian-cli'\n", name)
	case aliasShellPowerShell:
		return fmt.Sprintf("Set-Alias -Name %s -Value obsidian-cli\n", name)
	case aliasShellCmd:
		return fmt.Sprintf("doskey %s=obsidian-cli $*\n", name)
	case aliasShellBash, aliasShellZsh:
		fallthrough
	default:
		return fmt.Sprintf("alias %s=\"obsidian-cli\"\n", name)
	}
}

func installSymlinkAlias(name string, dir string, force bool, dryRun bool) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	exe, err = filepath.EvalSymlinks(exe)
	if err != nil {
		return err
	}

	if err := os.MkdirAll(dir, 0o755); err != nil {
		return err
	}

	linkName := name
	if runtime.GOOS == "windows" && !strings.HasSuffix(strings.ToLower(linkName), ".exe") {
		linkName += ".exe"
	}
	dst := filepath.Join(dir, linkName)

	if _, err := os.Lstat(dst); err == nil {
		if !force {
			return fmt.Errorf("refusing to overwrite existing path: %s (use --force)", dst)
		}
		if dryRun {
			fmt.Printf("Would remove existing path: %s\n", dst)
		} else if err := os.Remove(dst); err != nil {
			return err
		}
	} else if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}

	if dryRun {
		fmt.Printf("Would create symlink: %s -> %s\n", dst, exe)
		return nil
	}
	if err := os.Symlink(exe, dst); err != nil {
		if runtime.GOOS == "windows" {
			return fmt.Errorf("failed to create symlink %s -> %s: %w (Windows may require admin/developer mode)", dst, exe, err)
		}
		return err
	}
	return nil
}

