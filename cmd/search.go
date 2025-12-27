package cmd

import (
	"fmt"
	"strings"

	"github.com/Yakitrak/obsidian-cli/pkg/actions"
	"github.com/Yakitrak/obsidian-cli/pkg/obsidian"

	"github.com/spf13/cobra"
)

var searchCmd = &cobra.Command{
	Use:     "search",
	Aliases: []string{"s"},
	Short:   "Fuzzy searches and opens note in vault",
	Long: `Opens an interactive fuzzy finder to search and open notes.

Type to filter notes by filename. Press Enter to open the selected
note in Obsidian, or use --editor to open in your $EDITOR.`,
	Example: `  # Interactive search
  obsidian-cli search

  # Search and open in editor
  obsidian-cli search --editor

  # Search in specific vault
  obsidian-cli search --vault "Work"`,
	Args: cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		vault := obsidian.Vault{Name: vaultName}
		useEditor, err := cmd.Flags().GetBool("editor")
		if err != nil {
			return fmt.Errorf("failed to retrieve 'editor' flag: %w", err)
		}

		if _, err := vault.DefaultName(); err != nil {
			return err
		}
		vaultPath, err := vault.Path()
		if err != nil {
			return err
		}

		notePath, err := pickNotePathOrNew(vaultPath)
		if err != nil {
			return err
		}
		if strings.TrimSpace(notePath) == "" {
			return fmt.Errorf("no note selected")
		}

		if useEditor {
			fmt.Printf("Opening note: %s\n", notePath)
			rel := notePath
			if !strings.HasSuffix(strings.ToLower(rel), ".md") {
				rel += ".md"
			}
			abs, err := obsidian.SafeJoinVaultPath(vaultPath, rel)
			if err != nil {
				return err
			}
			return obsidian.OpenInEditor(abs)
		}

		uri := obsidian.Uri{}
		params := actions.OpenParams{NoteName: notePath}
		return actions.OpenNote(&vault, &uri, params)
	},
}

func init() {
	searchCmd.Flags().StringVarP(&vaultName, "vault", "v", "", "vault name")
	searchCmd.Flags().BoolP("editor", "e", false, "open in editor instead of Obsidian")
	rootCmd.AddCommand(searchCmd)
}
